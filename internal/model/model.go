package model

import (
	"database/sql"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

// ResolveDBPath determina dinamicamente o caminho do banco de dados SQLite
// Ele garante que o banco seja criado/usado na pasta onde o executável foi iniciado
func ResolveDBPath(defaultName string) string {
	if defaultName == "" {
		defaultName = "ipmonitor.db"
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
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		return nil, err
	}
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
	rows, err := repo.DB.Query("SELECT id, ip, status, COALESCE(name, ''), COALESCE(method, 'PING'), COALESCE(threshold_ms, 2000), COALESCE(uuid, '') FROM ips")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var devices []IPDevice
	for rows.Next() {
		var d IPDevice
		_ = rows.Scan(&d.ID, &d.IP, &d.Status, &d.Name, &d.Method, &d.ThresholdMs, &d.UUID)
		devices = append(devices, d)
	}
	return devices, nil
}

func (repo *IPRepository) UpdateStatus(id int, status string) error {
	_, err := repo.DB.Exec("UPDATE ips SET status=? WHERE id=?", status, id)
	return err
}

func (repo *IPRepository) UpdateIP(id int, newIP string) error {
	_, err := repo.DB.Exec("UPDATE ips SET ip=?, status='Desconhecido' WHERE id=?", newIP, id)
	return err
}

func (repo *IPRepository) Close() error {
	return repo.DB.Close()
}
