package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/AdemarTellecher/ipmonitorapp/internal/model"
	"github.com/go-ping/ping"
)

type Result struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
	Count   int    `json:"count,omitempty"`
}

type MonitorService struct {
	Repo *model.IPRepository
}

func NewMonitorService(repo *model.IPRepository) *MonitorService {
	return &MonitorService{Repo: repo}
}

// ServeHTTP implementa http.Handler permitindo que o frontend faça chamadas fetch('/api/...') nativas
func (s *MonitorService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	path := strings.TrimPrefix(r.URL.Path, "/api")
	path = strings.TrimPrefix(path, "/")

	switch path {
	case "list":
		ips, err := s.ListIPs()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_ = json.NewEncoder(w).Encode(ips)

	case "add":
		var req struct {
			IP string `json:"ip"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			_ = json.NewEncoder(w).Encode(Result{Success: false, Error: "corpo da requisição inválido"})
			return
		}
		res := s.AddIP(req.IP)
		_ = json.NewEncoder(w).Encode(res)

	case "remove":
		var req struct {
			IP string `json:"ip"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			_ = json.NewEncoder(w).Encode(Result{Success: false, Error: "corpo da requisição inválido"})
			return
		}
		res := s.RemoveIP(req.IP)
		_ = json.NewEncoder(w).Encode(res)

	case "update":
		updated, err := s.UpdateAllStatuses()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_ = json.NewEncoder(w).Encode(updated)

	case "import":
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			_ = json.NewEncoder(w).Encode(Result{Success: false, Error: err.Error()})
			return
		}
		res := s.ImportJSONContent(string(bodyBytes))
		_ = json.NewEncoder(w).Encode(res)

	default:
		http.NotFound(w, r)
	}
}

// ListIPs retorna todos os IPs cadastrados
func (s *MonitorService) ListIPs() ([]model.IPDevice, error) {
	return s.Repo.List()
}

// AddIP valida e adiciona um novo IP
func (s *MonitorService) AddIP(ip string) Result {
	cleaned := cleanHostOrIP(ip)
	if net.ParseIP(cleaned) == nil {
		return Result{Success: false, Error: fmt.Sprintf("IP inválido: %s", ip)}
	}
	err := s.Repo.Add(cleaned)
	if err != nil {
		return Result{Success: false, Error: err.Error()}
	}
	return Result{Success: true, Message: "IP adicionado com sucesso"}
}

// RemoveIP remove o IP informado
func (s *MonitorService) RemoveIP(ip string) Result {
	err := s.Repo.Remove(ip)
	if err != nil {
		return Result{Success: false, Error: err.Error()}
	}
	return Result{Success: true, Message: "IP removido"}
}

// UpdateAllStatuses executa o ping em todos os IPs cadastrados e retorna a lista atualizada
func (s *MonitorService) UpdateAllStatuses() ([]model.IPDevice, error) {
	devices, err := s.Repo.List()
	if err != nil {
		return nil, err
	}
	for _, d := range devices {
		status := checkIPOnline(d.IP)
		_ = s.Repo.UpdateStatus(d.ID, status)
	}
	return s.Repo.List()
}

// ImportJSONContent processa o texto JSON recebido do frontend e insere os IPs
func (s *MonitorService) ImportJSONContent(content string) Result {
	var config model.SitesConfig
	if err := json.Unmarshal([]byte(content), &config); err != nil {
		return Result{Success: false, Error: "Formato JSON inválido"}
	}

	if len(config.Sites) == 0 {
		return Result{Success: false, Error: "Nenhum site encontrado no arquivo JSON"}
	}

	var validIPs []string
	seen := make(map[string]bool)

	for _, site := range config.Sites {
		cleaned := cleanHostOrIP(site.URL)
		if cleaned == "" || seen[cleaned] {
			continue
		}
		if net.ParseIP(cleaned) != nil {
			validIPs = append(validIPs, cleaned)
			seen[cleaned] = true
		}
	}

	if len(validIPs) == 0 {
		return Result{Success: false, Error: "Nenhum endereço IP válido encontrado"}
	}

	count, err := s.Repo.AddMultiple(validIPs)
	if err != nil {
		return Result{Success: false, Error: err.Error()}
	}

	return Result{Success: true, Count: count}
}

// ImportFromReader processa um reader direto
func (s *MonitorService) ImportFromReader(reader io.Reader) (int, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return 0, err
	}
	res := s.ImportJSONContent(string(data))
	if !res.Success {
		return 0, fmt.Errorf(res.Error)
	}
	return res.Count, nil
}

func cleanHostOrIP(raw string) string {
	raw = strings.TrimSpace(raw)
	if strings.HasPrefix(raw, "http://") || strings.HasPrefix(raw, "https://") {
		u, err := url.Parse(raw)
		if err == nil {
			raw = u.Hostname()
		}
	} else {
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
