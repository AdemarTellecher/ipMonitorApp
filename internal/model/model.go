package model

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

type IPDevice struct {
	ID     int
	IP     string
	Status string
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
