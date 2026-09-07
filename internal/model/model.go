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
	ID     int    `json:"id"`
	IP     string `json:"ip"`
	Status string `json:"status"`
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
		status TEXT
	)`)
	if err != nil {
		return nil, err
	}
	return &IPRepository{DB: db}, nil
}

func (repo *IPRepository) Add(ip string) error {
	_, err := repo.DB.Exec("INSERT OR IGNORE INTO ips(ip, status) VALUES (?, ?)", ip, "Desconhecido")
	return err
}

func (repo *IPRepository) AddMultiple(ips []string) (int, error) {
	tx, err := repo.DB.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("INSERT OR IGNORE INTO ips(ip, status) VALUES (?, ?)")
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

func (repo *IPRepository) Remove(ip string) error {
	_, err := repo.DB.Exec("DELETE FROM ips WHERE ip = ?", ip)
	return err
}

func (repo *IPRepository) List() ([]IPDevice, error) {
	rows, err := repo.DB.Query("SELECT id, ip, status FROM ips")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var devices []IPDevice
	for rows.Next() {
		var d IPDevice
		_ = rows.Scan(&d.ID, &d.IP, &d.Status)
		devices = append(devices, d)
	}
	return devices, nil
}

func (repo *IPRepository) UpdateStatus(id int, status string) error {
	_, err := repo.DB.Exec("UPDATE ips SET status=? WHERE id=?", status, id)
	return err
}

func (repo *IPRepository) Close() error {
	return repo.DB.Close()
}
