package model

import (
	"database/sql"
	"os"
	"path/filepath"
	"runtime"

	_ "modernc.org/sqlite"
)

// ResolveDBPath determina dinamicamente o caminho do banco de dados SQLite.
// No macOS, se estiver dentro de um .app bundle (/Contents/MacOS/), utiliza ~/Library/Application Support/IPMonitor.
// No Windows e outros sistemas, utiliza a pasta onde o executável foi iniciado (portátil).
func ResolveDBPath(defaultName string) string {
	if defaultName == "" {
		defaultName = "ipMonitorDB.db"
	}

	if runtime.GOOS == "darwin" {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			appSupportDir := filepath.Join(homeDir, "Library", "Application Support", "IPMonitor")
			_ = os.MkdirAll(appSupportDir, 0755)
			return filepath.Join(appSupportDir, defaultName)
		}
	}

	exePath, err := os.Executable()
	if err == nil {
		exeDir := filepath.Dir(exePath)
		return filepath.Join(exeDir, defaultName)
	}
	// Fallback para diretório de trabalho atual
	return defaultName
}

type IPDevice struct {
	ID          int    `json:"id"`
	IP          string `json:"ip"`
	Status      string `json:"status"`
	Name        string `json:"name"`
	Method      string `json:"method"`
	ThresholdMs int    `json:"thresholdMs"`
	UUID        string `json:"uuid"`
}

type NetworkOverview struct {
	Total   int        `json:"total"`
	Online  int        `json:"online"`
	Offline int        `json:"offline"`
	Devices []IPDevice `json:"devices"`
}

type SiteItem struct {
	Name        string `json:"name"`
	Enabled     bool   `json:"enabled"`
	ID          string `json:"id"`
	URL         string `json:"url"`
	Method      string `json:"method"`
	ThresholdMs int    `json:"thresholdMs"`
}

type SitesConfig struct {
	Sites []SiteItem `json:"sites"`
}

type IPRepository struct {
	DB *sql.DB
}

func NewRepository(dbFile string) (*IPRepository, error) {
	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		return nil, err
	}

	// Ativa WAL mode e timeout de contenção para suportar acessos concorrentes sem "database is locked"
	_, _ = db.Exec("PRAGMA journal_mode=WAL;")
	_, _ = db.Exec("PRAGMA busy_timeout=5000;")
	_, _ = db.Exec("PRAGMA synchronous=NORMAL;")

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS ips (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		ip TEXT NOT NULL UNIQUE,
		status TEXT,
		name TEXT DEFAULT '',
		method TEXT DEFAULT 'PING',
		threshold_ms INTEGER DEFAULT 2000,
		uuid TEXT DEFAULT ''
	)`)
	if err != nil {
		return nil, err
	}

	// Migração transparente para bancos SQLite existentes
	_ = migrateColumn(db, "name", "TEXT DEFAULT ''")
	_ = migrateColumn(db, "method", "TEXT DEFAULT 'PING'")
	_ = migrateColumn(db, "threshold_ms", "INTEGER DEFAULT 2000")
	_ = migrateColumn(db, "uuid", "TEXT DEFAULT ''")

	return &IPRepository{DB: db}, nil
}

func migrateColumn(db *sql.DB, colName, colDef string) error {
	var count int
	row := db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('ips') WHERE name=?", colName)
	if err := row.Scan(&count); err == nil && count == 0 {
		_, err = db.Exec("ALTER TABLE ips ADD COLUMN " + colName + " " + colDef)
		return err
	}
	return nil
}

func (repo *IPRepository) Add(ip string) error {
	_, err := repo.DB.Exec("INSERT OR IGNORE INTO ips(ip, status, name, method, threshold_ms, uuid) VALUES (?, ?, '', 'PING', 2000, '')", ip, "Desconhecido")
	return err
}

func (repo *IPRepository) AddDevice(device IPDevice) error {
	if device.Method == "" {
		device.Method = "PING"
	}
	if device.ThresholdMs <= 0 {
		device.ThresholdMs = 2000
	}
	_, err := repo.DB.Exec(`INSERT INTO ips(ip, status, name, method, threshold_ms, uuid) 
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(ip) DO UPDATE SET 
			name=excluded.name, 
			method=excluded.method, 
			threshold_ms=excluded.threshold_ms, 
			uuid=excluded.uuid`,
		device.IP, "Desconhecido", device.Name, device.Method, device.ThresholdMs, device.UUID)
	return err
}

func (repo *IPRepository) AddMultiple(ips []string) (int, error) {
	tx, err := repo.DB.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("INSERT OR IGNORE INTO ips(ip, status, name, method, threshold_ms, uuid) VALUES (?, ?, '', 'PING', 2000, '')")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	insertedCount := 0
	for _, ip := range ips {
		res, err := stmt.Exec(ip, "Desconhecido")
		if err != nil {
			return insertedCount, err
		}
		rowsAffected, _ := res.RowsAffected()
		if rowsAffected > 0 {
			insertedCount++
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return insertedCount, nil
}

func (repo *IPRepository) AddMultipleDevices(devices []IPDevice) (int, error) {
	tx, err := repo.DB.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`INSERT INTO ips(ip, status, name, method, threshold_ms, uuid) 
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(ip) DO UPDATE SET 
			name=excluded.name, 
			method=excluded.method, 
			threshold_ms=excluded.threshold_ms, 
			uuid=excluded.uuid`)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	count := 0
	for _, d := range devices {
		if d.Method == "" {
			d.Method = "PING"
		}
		if d.ThresholdMs <= 0 {
			d.ThresholdMs = 2000
		}
		status := d.Status
		if status == "" {
			status = "Desconhecido"
		}
		_, err := stmt.Exec(d.IP, status, d.Name, d.Method, d.ThresholdMs, d.UUID)
		if err != nil {
			return count, err
		}
		count++
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return count, nil
}

func (repo *IPRepository) Remove(ip string) error {
	_, err := repo.DB.Exec("DELETE FROM ips WHERE ip = ?", ip)
	return err
}

func (repo *IPRepository) List() ([]IPDevice, error) {
	query := `SELECT id, ip, status, COALESCE(name, ''), COALESCE(method, 'PING'), COALESCE(threshold_ms, 2000), COALESCE(uuid, '') 
		FROM ips 
		ORDER BY 
			CASE status 
				WHEN 'Offline' THEN 0 
				WHEN 'Desconhecido' THEN 1 
				WHEN 'Online' THEN 2 
				ELSE 3 
			END, 
			ip ASC`
	rows, err := repo.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var devices []IPDevice
	for rows.Next() {
		var d IPDevice
		if err := rows.Scan(&d.ID, &d.IP, &d.Status, &d.Name, &d.Method, &d.ThresholdMs, &d.UUID); err != nil {
			return nil, err
		}
		devices = append(devices, d)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return devices, nil
}

// GetOverview retorna os dispositivos ordenados e o sumário consolidado com total, online e offline
func (repo *IPRepository) GetOverview() (*NetworkOverview, error) {
	devices, err := repo.List()
	if err != nil {
		return nil, err
	}

	onlineCount := 0
	offlineCount := 0
	for _, d := range devices {
		switch d.Status {
		case "Online":
			onlineCount++
		case "Offline":
			offlineCount++
		}
	}

	if devices == nil {
		devices = []IPDevice{}
	}

	return &NetworkOverview{
		Total:   len(devices),
		Online:  onlineCount,
		Offline: offlineCount,
		Devices: devices,
	}, nil
}

func (repo *IPRepository) UpdateStatus(id int, status string) error {
	_, err := repo.DB.Exec("UPDATE ips SET status=? WHERE id=?", status, id)
	return err
}

func (repo *IPRepository) UpdateIP(id int, newIP string) error {
	_, err := repo.DB.Exec("UPDATE ips SET ip=?, status='Desconhecido' WHERE id=?", newIP, id)
	return err
}

func (repo *IPRepository) UpdateDevice(id int, newIP, name, method string, thresholdMs int, uuid string) error {
	if method == "" {
		method = "PING"
	}
	if thresholdMs <= 0 {
		thresholdMs = 2000
	}
	_, err := repo.DB.Exec("UPDATE ips SET ip=?, name=?, method=?, threshold_ms=?, uuid=?, status='Desconhecido' WHERE id=?",
		newIP, name, method, thresholdMs, uuid, id)
	return err
}

func (repo *IPRepository) Close() error {
	return repo.DB.Close()
}
