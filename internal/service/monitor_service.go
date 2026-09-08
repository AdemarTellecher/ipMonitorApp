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

	case "edit":
		var req struct {
			ID    int    `json:"id"`
			NewIP string `json:"newIp"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			_ = json.NewEncoder(w).Encode(Result{Success: false, Error: "corpo da requisição inválido"})
			return
		}
		res := s.EditIP(req.ID, req.NewIP)
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

// AddIP valida e adiciona um novo IP ou Hostname/URL
func (s *MonitorService) AddIP(ip string) Result {
	cleaned := cleanHostOrIP(ip)
	if !isValidHostOrIP(cleaned) {
		return Result{Success: false, Error: fmt.Sprintf("Endereço IP ou Hostname inválido: %s", ip)}
	}
	err := s.Repo.Add(cleaned)
	if err != nil {
		return Result{Success: false, Error: err.Error()}
	}
	return Result{Success: true, Message: "Host adicionado com sucesso"}
}

// RemoveIP remove o IP informado
func (s *MonitorService) RemoveIP(ip string) Result {
	err := s.Repo.Remove(ip)
	if err != nil {
		return Result{Success: false, Error: err.Error()}
	}
	return Result{Success: true, Message: "Host removido"}
}

// EditIP valida, atualiza o host e testa a conectividade imediatamente
func (s *MonitorService) EditIP(id int, newIP string) Result {
	cleaned := cleanHostOrIP(newIP)
	if !isValidHostOrIP(cleaned) {
		return Result{Success: false, Error: fmt.Sprintf("Endereço IP ou Hostname inválido: %s", newIP)}
	}
	err := s.Repo.UpdateIP(id, cleaned)
	if err != nil {
		return Result{Success: false, Error: fmt.Sprintf("Erro ao atualizar host: %s", err.Error())}
	}
	// Executa checagem imediata de status após salvar
	status := checkIPOnline(cleaned)
	_ = s.Repo.UpdateStatus(id, status)

	return Result{Success: true, Message: "Host atualizado com sucesso"}
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

// ImportJSONContent processa o texto JSON recebido do frontend e insere os IPs e Hostnames
func (s *MonitorService) ImportJSONContent(content string) Result {
	var rawData interface{}
	if err := json.Unmarshal([]byte(content), &rawData); err != nil {
		return Result{Success: false, Error: "Formato JSON inválido"}
	}

	var candidateStrings []string

	switch v := rawData.(type) {
	case []interface{}:
		for _, item := range v {
			if str, ok := item.(string); ok {
				candidateStrings = append(candidateStrings, str)
			} else if obj, ok := item.(map[string]interface{}); ok {
				for _, key := range []string{"ip", "host", "url", "address", "hostname"} {
					if val, found := obj[key]; found {
						if strVal, ok := val.(string); ok {
							candidateStrings = append(candidateStrings, strVal)
							break
						}
					}
				}
			}
		}
	case map[string]interface{}:
		for _, key := range []string{"ips", "hosts", "sites", "devices", "servers"} {
			if list, found := v[key]; found {
				if arr, ok := list.([]interface{}); ok {
					for _, item := range arr {
						if str, ok := item.(string); ok {
							candidateStrings = append(candidateStrings, str)
						} else if obj, ok := item.(map[string]interface{}); ok {
							for _, k := range []string{"ip", "host", "url", "address", "hostname"} {
								if val, found := obj[k]; found {
									if strVal, ok := val.(string); ok {
										candidateStrings = append(candidateStrings, strVal)
										break
									}
								}
							}
						}
					}
				}
			}
		}
	}

	var validIPs []string
	seen := make(map[string]bool)

	for _, raw := range candidateStrings {
		cleaned := cleanHostOrIP(raw)
		if cleaned == "" || seen[cleaned] {
			continue
		}
		if isValidHostOrIP(cleaned) {
			validIPs = append(validIPs, cleaned)
			seen[cleaned] = true
		}
	}

	if len(validIPs) == 0 {
		return Result{Success: false, Error: "Nenhum endereço IP ou Hostname válido encontrado"}
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
		return 0, fmt.Errorf("%s", res.Error)
	}
	return res.Count, nil
}

func isValidHostOrIP(raw string) bool {
	if raw == "" || len(raw) > 253 {
		return false
	}
	if net.ParseIP(raw) != nil {
		return true
	}
	if strings.ContainsAny(raw, "/: \t\r\n") {
		return false
	}
	labels := strings.Split(raw, ".")
	for _, label := range labels {
		if len(label) == 0 || len(label) > 63 {
			return false
		}
		if label[0] == '-' || label[len(label)-1] == '-' {
			return false
		}
		for i := 0; i < len(label); i++ {
			c := label[i]
			isAlphaNum := (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-'
			if !isAlphaNum {
				return false
			}
		}
	}
	return true
}

func cleanHostOrIP(raw string) string {
	raw = strings.TrimSpace(raw)
	if strings.HasPrefix(raw, "http://") || strings.HasPrefix(raw, "https://") {
		u, err := url.Parse(raw)
		if err == nil && u.Hostname() != "" {
			raw = u.Hostname()
		}
	} else {
		if strings.Contains(raw, "/") {
			parts := strings.Split(raw, "/")
			raw = parts[0]
		}
		if strings.Contains(raw, ":") {
			host, _, err := net.SplitHostPort(raw)
			if err == nil && host != "" {
				raw = host
			}
		}
	}
	return strings.ToLower(strings.TrimSpace(raw))
}

func checkIPOnline(target string) string {
	target = strings.TrimSpace(target)
	if target == "" {
		return "Offline"
	}

	pinger, err := ping.NewPinger(target)
	if err == nil {
		pinger.Count = 1
		pinger.Timeout = 2 * time.Second
		pinger.SetPrivileged(true)
		runErr := pinger.Run()
		if runErr == nil {
			stats := pinger.Statistics()
			if stats.PacketsRecv > 0 {
				return "Online"
			}
		}
	}

	for _, port := range []string{"443", "80"} {
		conn, dialErr := net.DialTimeout("tcp", net.JoinHostPort(target, port), 1500*time.Millisecond)
		if dialErr == nil {
			_ = conn.Close()
			return "Online"
		}
	}

	return "Offline"
}
