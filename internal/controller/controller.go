package controller

import (
	"fmt"
	"net"
	"time"

	"github.com/AdemarTellecher/ipmonitorapp/model"

	"github.com/go-ping/ping"
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
	pinger, err := ping.NewPinger(ip)
	if err != nil {
		return "Offline"
	}
	pinger.Count = 1
	pinger.Timeout = 2 * time.Second
	pinger.SetPrivileged(true)
	err = pinger.Run()
	if err != nil {
		return "Offline"
	}
	stats := pinger.Statistics()
	if stats.PacketsRecv > 0 {
		return "Online"
	}
	return "Offline"
}
