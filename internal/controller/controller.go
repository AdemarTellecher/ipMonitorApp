package controller

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/AdemarTellecher/ipmonitorapp/internal/model"
	"github.com/go-ping/ping"
)

type IPController struct {
	Repo *model.IPRepository
}

func NewIPController(repo *model.IPRepository) *IPController {
	return &IPController{Repo: repo}
}

func (c *IPController) AddIP(ip string) error {
	cleaned := cleanHostOrIP(ip)
	if net.ParseIP(cleaned) == nil {
		return fmt.Errorf("IP inválido: %s", ip)
	}
	return c.Repo.Add(cleaned)
}

func cleanHostOrIP(raw string) string {
	raw = strings.TrimSpace(raw)
	if strings.HasPrefix(raw, "http://") || strings.HasPrefix(raw, "https://") {
		u, err := url.Parse(raw)
		if err == nil {
			raw = u.Hostname()
		}
	} else {
		// Se contiver porta ou caminho
		if strings.Contains(raw, "/") {
			parts := strings.Split(raw, "/")
			raw = parts[0]
		}
		if strings.Contains(raw, ":") {
			host, _, err := net.SplitHostPort(raw)
			if err == nil {
				raw = host
			}
		}
	}
	return strings.TrimSpace(raw)
}

// ImportFromJSON lê a estrutura JSON com lista de sites e insere no banco
func (c *IPController) ImportFromJSON(reader io.Reader) (int, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return 0, fmt.Errorf("erro ao ler arquivo: %w", err)
	}

	var config model.SitesConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return 0, fmt.Errorf("formato JSON inválido: %w", err)
	}

	if len(config.Sites) == 0 {
		return 0, fmt.Errorf("nenhum site encontrado no arquivo JSON")
	}

	var validIPs []string
	seen := make(map[string]bool)

	for _, site := range config.Sites {
		cleaned := cleanHostOrIP(site.URL)
		if cleaned == "" || seen[cleaned] {
			continue
		}
		// Verifica se é IP válido ou hostname
		if net.ParseIP(cleaned) != nil {
			validIPs = append(validIPs, cleaned)
			seen[cleaned] = true
		}
	}

	if len(validIPs) == 0 {
		return 0, fmt.Errorf("nenhum endereço IP válido encontrado para importar")
	}

	return c.Repo.AddMultiple(validIPs)
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
