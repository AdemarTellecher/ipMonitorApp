package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"runtime"
	"strings"
	"sync"
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
			IP          string `json:"ip"`
			Name        string `json:"name"`
			Method      string `json:"method"`
			ThresholdMs int    `json:"thresholdMs"`
			UUID        string `json:"uuid"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			_ = json.NewEncoder(w).Encode(Result{Success: false, Error: "corpo da requisição inválido"})
			return
		}
		res := s.AddIPWithDetails(req.IP, req.Name, req.Method, req.ThresholdMs, req.UUID)
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
			ID          int    `json:"id"`
			NewIP       string `json:"newIp"`
			Name        string `json:"name"`
			Method      string `json:"method"`
			ThresholdMs int    `json:"thresholdMs"`
			UUID        string `json:"uuid"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			_ = json.NewEncoder(w).Encode(Result{Success: false, Error: "corpo da requisição inválido"})
			return
		}
		res := s.EditIPWithDetails(req.ID, req.NewIP, req.Name, req.Method, req.ThresholdMs, req.UUID)
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

// AddIP valida, adiciona e testa imediatamente a conectividade do novo IP ou Hostname/URL
func (s *MonitorService) AddIP(ip string) Result {
	return s.AddIPWithDetails(ip, "", "PING", 2000, "")
}

// AddIPWithDetails valida, adiciona com metadados ricos e testa imediatamente a conectividade
func (s *MonitorService) AddIPWithDetails(ip, name, method string, thresholdMs int, uuid string) Result {
	cleaned := cleanHostOrIP(ip)
	if !isValidHostOrIP(cleaned) {
		return Result{Success: false, Error: fmt.Sprintf("Endereço IP ou Hostname inválido: %s", ip)}
	}
	if method == "" {
		method = "PING"
	}
	if thresholdMs <= 0 {
		thresholdMs = 2000
	}

	err := s.Repo.AddDevice(model.IPDevice{
		IP:          cleaned,
		Status:      "Desconhecido",
		Name:        name,
		Method:      method,
		ThresholdMs: thresholdMs,
		UUID:        uuid,
	})
	if err != nil {
		return Result{Success: false, Error: err.Error()}
	}

	// Executa checagem imediata de conectividade
	status := checkIPOnline(cleaned)
	devices, err := s.Repo.List()
	if err == nil {
		for _, d := range devices {
			if d.IP == cleaned {
				_ = s.Repo.UpdateStatus(d.ID, status)
				break
			}
		}
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
	return s.EditIPWithDetails(id, newIP, "", "PING", 2000, "")
}

// EditIPWithDetails valida, atualiza os dados completos do host e testa a conectividade imediatamente
func (s *MonitorService) EditIPWithDetails(id int, newIP, name, method string, thresholdMs int, uuid string) Result {
	cleaned := cleanHostOrIP(newIP)
	if !isValidHostOrIP(cleaned) {
		return Result{Success: false, Error: fmt.Sprintf("Endereço IP ou Hostname inválido: %s", newIP)}
	}
	if method == "" {
		method = "PING"
	}
	if thresholdMs <= 0 {
		thresholdMs = 2000
	}

	err := s.Repo.UpdateDevice(id, cleaned, name, method, thresholdMs, uuid)
	if err != nil {
		return Result{Success: false, Error: fmt.Sprintf("Erro ao atualizar host: %s", err.Error())}
	}
	// Executa checagem imediata de status após salvar
	status := checkIPOnline(cleaned)
	_ = s.Repo.UpdateStatus(id, status)

	return Result{Success: true, Message: "Host atualizado com sucesso"}
}

// UpdateAllStatuses executa a verificação em paralelo em todos os IPs cadastrados e retorna a lista atualizada
func (s *MonitorService) UpdateAllStatuses() ([]model.IPDevice, error) {
	devices, err := s.Repo.List()
	if err != nil {
		return nil, err
	}

	type statusUpdate struct {
		id     int
		status string
	}

	// Executa as checagens com concorrência controlada para respostas ultrarrápidas
	updatesChan := make(chan statusUpdate, len(devices))
	sem := make(chan struct{}, 20) // limite de 20 conexões simultâneas

	var wg sync.WaitGroup
	for _, d := range devices {
		wg.Add(1)
		go func(dev model.IPDevice) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			st := checkIPOnline(dev.IP)
			updatesChan <- statusUpdate{id: dev.ID, status: st}
		}(d)
	}

	wg.Wait()
	close(updatesChan)

	for up := range updatesChan {
		_ = s.Repo.UpdateStatus(up.id, up.status)
	}

	return s.Repo.List()
}

// ImportJSONContent processa o texto JSON recebido do frontend e insere os IPs, Hostnames e metadados
func (s *MonitorService) ImportJSONContent(content string) Result {
	var rawData interface{}
	if err := json.Unmarshal([]byte(content), &rawData); err != nil {
		return Result{Success: false, Error: "Formato JSON inválido"}
	}

	type Candidate struct {
		Raw         string
		Name        string
		Method      string
		ThresholdMs int
		UUID        string
	}

	var candidates []Candidate

	parseObj := func(obj map[string]interface{}) {
		var rawHost string
		for _, key := range []string{"url", "ip", "host", "address", "hostname"} {
			if val, found := obj[key]; found {
				if strVal, ok := val.(string); ok && strVal != "" {
					rawHost = strVal
					break
				}
			}
		}
		if rawHost == "" {
			return
		}

		c := Candidate{Raw: rawHost, Method: "PING", ThresholdMs: 2000}
		if nameVal, ok := obj["name"].(string); ok {
			c.Name = nameVal
		}
		if idVal, ok := obj["id"].(string); ok {
			c.UUID = idVal
		}
		if methodVal, ok := obj["method"].(string); ok && methodVal != "" {
			c.Method = methodVal
		}
		if thVal, ok := obj["thresholdMs"].(float64); ok && thVal > 0 {
			c.ThresholdMs = int(thVal)
		}

		candidates = append(candidates, c)
	}

	switch v := rawData.(type) {
	case []interface{}:
		for _, item := range v {
			if str, ok := item.(string); ok {
				candidates = append(candidates, Candidate{Raw: str, Method: "PING", ThresholdMs: 2000})
			} else if obj, ok := item.(map[string]interface{}); ok {
				parseObj(obj)
			}
		}
	case map[string]interface{}:
		for _, key := range []string{"sites", "ips", "hosts", "devices", "servers"} {
			if list, found := v[key]; found {
				if arr, ok := list.([]interface{}); ok {
					for _, item := range arr {
						if str, ok := item.(string); ok {
							candidates = append(candidates, Candidate{Raw: str, Method: "PING", ThresholdMs: 2000})
						} else if obj, ok := item.(map[string]interface{}); ok {
							parseObj(obj)
						}
					}
				}
			}
		}
	}

	var validDevices []model.IPDevice
	seen := make(map[string]bool)

	for _, c := range candidates {
		cleaned := cleanHostOrIP(c.Raw)
		if cleaned == "" || seen[cleaned] {
			continue
		}
		if isValidHostOrIP(cleaned) {
			validDevices = append(validDevices, model.IPDevice{
				IP:          cleaned,
				Status:      "Desconhecido",
				Name:        c.Name,
				Method:      c.Method,
				ThresholdMs: c.ThresholdMs,
				UUID:        c.UUID,
			})
			seen[cleaned] = true
		}
	}

	if len(validDevices) == 0 {
		return Result{Success: false, Error: "Nenhum endereço IP ou Hostname válido encontrado"}
	}

	count, err := s.Repo.AddMultipleDevices(validDevices)
	if err != nil {
		return Result{Success: false, Error: err.Error()}
	}

	// Dispara varredura dos hosts importados em segundo plano
	go func() {
		_, _ = s.UpdateAllStatuses()
	}()

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

	// 1. Tentativa via go-ping (ICMP direto)
	pinger, err := ping.NewPinger(target)
	if err == nil {
		pinger.Count = 1
		pinger.Timeout = 1500 * time.Millisecond
		// No Windows é obrigatório Privileged = true para sockets ICMP.
		// No macOS (Darwin) e Linux sem root, Privileged deve ser false (usa sockets UDP ICMP unprivileged).
		if runtime.GOOS == "windows" {
			pinger.SetPrivileged(true)
		} else {
			pinger.SetPrivileged(false)
		}

		runErr := pinger.Run()
		if runErr == nil {
			stats := pinger.Statistics()
			if stats.PacketsRecv > 0 {
				return "Online"
			}
		}
	}

	// 2. Fallback para ping nativo do sistema operacional (macOS / Linux / Windows)
	// Essencial no macOS onde o utilitário `/sbin/ping` possui setuid-root e sempre tem permissão ICMP
	if pingSystemCommand(target) {
		return "Online"
	}

	// 3. Fallback para portas TCP comuns de serviços (80, 443, 8080, 22)
	for _, port := range []string{"443", "80", "8080", "22"} {
		conn, dialErr := net.DialTimeout("tcp", net.JoinHostPort(target, port), 1000*time.Millisecond)
		if dialErr == nil {
			_ = conn.Close()
			return "Online"
		}
	}

	return "Offline"
}
