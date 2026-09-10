package service_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/AdemarTellecher/ipmonitorapp/internal/model"
	"github.com/AdemarTellecher/ipmonitorapp/internal/service"
)

func TestMonitorService_ImportJSONContent(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	repo, err := model.NewRepository(dbPath)
	if err != nil {
		t.Fatalf("Erro ao criar repo de teste: %v", err)
	}
	defer repo.Close()

	svc := service.NewMonitorService(repo)

	// Mix de IPs, URLs completas e hostnames puros
	jsonSample := `{
		"sites": [
			{"url": "192.168.1.1"},
			{"url": "http://10.0.0.1:8080/path"},
			{"url": "https://google.com.br/search?q=test"},
			{"url": "github.com"},
			{"url": "192.168.1.1"}
		]
	}`

	res := svc.ImportJSONContent(jsonSample)
	if !res.Success {
		t.Fatalf("ImportJSONContent falhou: %s", res.Error)
	}

	// 192.168.1.1, 10.0.0.1, google.com.br, github.com (4 únicos)
	if res.Count != 4 {
		t.Errorf("Esperava 4 hosts inseridos (únicos), obteve %d", res.Count)
	}

	ips, err := svc.ListIPs()
	if err != nil {
		t.Fatalf("ListIPs falhou: %v", err)
	}

	if len(ips) != 4 {
		t.Errorf("Esperava 4 hosts listados, obteve %d", len(ips))
	}
}

func TestMonitorService_AddAndRemove(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	repo, err := model.NewRepository(dbPath)
	if err != nil {
		t.Fatalf("Erro ao criar repo de teste: %v", err)
	}
	defer repo.Close()

	svc := service.NewMonitorService(repo)

	// 1. Adicionar IP válido
	res := svc.AddIP("192.168.0.10")
	if !res.Success {
		t.Fatalf("Esperava sucesso ao adicionar IP válido: %s", res.Error)
	}

	// 2. Adicionar Hostname válido (ex: google.com.br)
	resHost := svc.AddIP("google.com.br")
	if !resHost.Success {
		t.Fatalf("Esperava sucesso ao adicionar Hostname google.com.br: %s", resHost.Error)
	}

	// 3. Adicionar URL válida com protocolo e path (deve sanitizar para domínio)
	resURL := svc.AddIP("https://brasil.gov.br/servicos")
	if !resURL.Success {
		t.Fatalf("Esperava sucesso ao adicionar URL sanitizada: %s", resURL.Error)
	}

	// Verificar se 'brasil.gov.br' foi inserido limpo
	devices, _ := svc.ListIPs()
	foundBrasil := false
	for _, d := range devices {
		if d.IP == "brasil.gov.br" {
			foundBrasil = true
			break
		}
	}
	if !foundBrasil {
		t.Errorf("Esperava encontrar 'brasil.gov.br' sanitizado na lista de dispositivos")
	}

	// 4. Tentar adicionar entrada inválida (contendo caracteres proibidos ou formato inválido)
	resInvalid := svc.AddIP("invalid_host!@#")
	if resInvalid.Success {
		t.Fatalf("Esperava falha ao adicionar host com caracteres especiais inválidos")
	}

	resInvalidEmpty := svc.AddIP("   ")
	if resInvalidEmpty.Success {
		t.Fatalf("Esperava falha ao adicionar string vazia")
	}

	// 5. Remover Host/IP
	resRemove := svc.RemoveIP("192.168.0.10")
	if !resRemove.Success {
		t.Fatalf("Esperava sucesso ao remover IP: %s", resRemove.Error)
	}

	ips, _ := svc.ListIPs()
	for _, item := range ips {
		if strings.Contains(item.IP, "192.168.0.10") {
			t.Errorf("IP deveria ter sido removido")
		}
	}
}

func TestMonitorService_EditIP(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	repo, err := model.NewRepository(dbPath)
	if err != nil {
		t.Fatalf("Erro ao criar repo de teste: %v", err)
	}
	defer repo.Close()

	svc := service.NewMonitorService(repo)

	// Adicionar um host inicial
	resAdd := svc.AddIP("192.168.1.50")
	if !resAdd.Success {
		t.Fatalf("Erro ao adicionar host inicial: %s", resAdd.Error)
	}

	devices, err := svc.ListIPs()
	if err != nil || len(devices) == 0 {
		t.Fatalf("Falha ao listar dispositivos após adicionar")
	}
	targetID := devices[0].ID

	// 1. Edição com sucesso para hostname
	resEdit := svc.EditIP(targetID, "google.com.br")
	if !resEdit.Success {
		t.Fatalf("Esperava sucesso ao editar para google.com.br: %s", resEdit.Error)
	}

	updatedDevices, _ := svc.ListIPs()
	if len(updatedDevices) != 1 || updatedDevices[0].IP != "google.com.br" {
		t.Errorf("Esperava host 'google.com.br', obteve %v", updatedDevices)
	}

	// 2. Edição com URL completa (deve sanitizar para domínio)
	resEditURL := svc.EditIP(targetID, "https://github.com/AdemarTellecher")
	if !resEditURL.Success {
		t.Fatalf("Esperava sucesso ao editar com URL: %s", resEditURL.Error)
	}

	devicesAfterURL, _ := svc.ListIPs()
	if len(devicesAfterURL) != 1 || devicesAfterURL[0].IP != "github.com" {
		t.Errorf("Esperava host 'github.com', obteve %v", devicesAfterURL)
	}

	// 3. Edição com valor inválido (deve rejeitar e não alterar)
	resEditInvalid := svc.EditIP(targetID, "invalido#host")
	if resEditInvalid.Success {
		t.Fatalf("Esperava erro ao tentar editar para host inválido")
	}

	// 4. Edição com valor duplicado
	_ = svc.AddIP("10.0.0.1")
	devicesList, _ := svc.ListIPs()
	var secondID int
	for _, d := range devicesList {
		if d.IP == "10.0.0.1" {
			secondID = d.ID
			break
		}
	}
	resEditDuplicate := svc.EditIP(secondID, "github.com")
	if resEditDuplicate.Success {
		t.Fatalf("Esperava falha ao tentar duplicar host existente")
	}
}

func TestMonitorService_UpdateAllStatuses(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	repo, err := model.NewRepository(dbPath)
	if err != nil {
		t.Fatalf("Erro ao criar repo de teste: %v", err)
	}
	defer repo.Close()

	svc := service.NewMonitorService(repo)

	_ = svc.AddIP("127.0.0.1")
	_ = svc.AddIP("192.0.2.1") // IP RFC 5737 de teste não roteável (deve ficar offline)

	overview, err := svc.UpdateAllStatuses()
	if err != nil {
		t.Fatalf("UpdateAllStatuses falhou: %v", err)
	}

	if overview == nil {
		t.Fatalf("Overview não deveria ser nil")
	}

	if overview.Total != 2 {
		t.Errorf("Esperava 2 hosts no total, obteve %d", overview.Total)
	}

	if len(overview.Devices) != 2 {
		t.Errorf("Esperava 2 devices no overview, obteve %d", len(overview.Devices))
	}
}

func TestMonitorService_OrderingAndOverviewMetrics(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_order.db")

	repo, err := model.NewRepository(dbPath)
	if err != nil {
		t.Fatalf("Erro ao criar repo de teste: %v", err)
	}
	defer repo.Close()

	// Inserir dispositivos com status variados
	_ = repo.AddDevice(model.IPDevice{IP: "10.0.0.1", Name: "Host Online"})
	_ = repo.AddDevice(model.IPDevice{IP: "10.0.0.2", Name: "Host Offline"})
	_ = repo.AddDevice(model.IPDevice{IP: "10.0.0.3", Name: "Host Desconhecido"})

	devices, _ := repo.List()
	for _, d := range devices {
		if d.IP == "10.0.0.1" {
			_ = repo.UpdateStatus(d.ID, "Online")
		} else if d.IP == "10.0.0.2" {
			_ = repo.UpdateStatus(d.ID, "Offline")
		} else if d.IP == "10.0.0.3" {
			_ = repo.UpdateStatus(d.ID, "Desconhecido")
		}
	}

	svc := service.NewMonitorService(repo)
	overview, err := svc.GetNetworkOverview()
	if err != nil {
		t.Fatalf("GetNetworkOverview falhou: %v", err)
	}

	// 1. Validar métricas de rede consolidadas
	if overview.Total != 3 {
		t.Errorf("Esperava total 3, obteve %d", overview.Total)
	}
	if overview.Online != 1 {
		t.Errorf("Esperava online 1, obteve %d", overview.Online)
	}
	if overview.Offline != 1 {
		t.Errorf("Esperava offline 1, obteve %d", overview.Offline)
	}

	// 2. Validar ordem canônica de prioridade: Offline (0) > Desconhecido (1) > Online (2)
	if len(overview.Devices) != 3 {
		t.Fatalf("Esperava 3 dispositivos, obteve %d", len(overview.Devices))
	}

	if overview.Devices[0].Status != "Offline" {
		t.Errorf("Primeiro dispositivo deveria ser 'Offline', obteve %s (%s)", overview.Devices[0].Status, overview.Devices[0].IP)
	}
	if overview.Devices[1].Status != "Desconhecido" {
		t.Errorf("Segundo dispositivo deveria ser 'Desconhecido', obteve %s (%s)", overview.Devices[1].Status, overview.Devices[1].IP)
	}
	if overview.Devices[2].Status != "Online" {
		t.Errorf("Terceiro dispositivo deveria ser 'Online', obteve %s (%s)", overview.Devices[2].Status, overview.Devices[2].IP)
	}
}

func TestMonitorService_GetVersion(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_ver.db")
	repo, err := model.NewRepository(dbPath)
	if err != nil {
		t.Fatalf("Erro ao criar repo: %v", err)
	}
	defer repo.Close()

	svc := service.NewMonitorService(repo)
	ver := svc.GetVersion()
	if !strings.HasPrefix(ver, "v") {
		t.Errorf("Versão esperada iniciando com 'v', obteve %s", ver)
	}
}

