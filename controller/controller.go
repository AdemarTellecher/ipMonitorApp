package controller

import (
	"fmt"
	"ipmonitorapp/model"
	"net"
	"time"
)

type IPController struct {
	Repo *model.IPRepository
}

func NewIPController(repo *model.IPRepository) *IPController {
	return &IPController{Repo: repo}
}

func (c *IPController) AddIP(ip string) error {
	if net.ParseIP(ip) == nil {
		return fmt.Errorf("IP inválido: %s", ip)
	}
	return c.Repo.Add(ip)
}

func (c *IPController) RemoveIP(ip string) error {
	return c.Repo.Remove(ip)
}

func (c *IPController) ListIPs() ([]model.IPDevice, error) {
	return c.Repo.List()
}

func (c *IPController) UpdateAllStatuses() {
	devices, err := c.Repo.List()
	if err != nil {
		return
	}
	for _, d := range devices {
		status := checkIPOnline(d.IP)
		_ = c.Repo.UpdateStatus(d.ID, status)
	}
}

func checkIPOnline(ip string) string {
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(ip, "80"), 2*time.Second)
	if err == nil {
		conn.Close()
		return "Online"
	}
	return "Offline"
}
