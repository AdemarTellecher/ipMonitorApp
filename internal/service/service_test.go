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
