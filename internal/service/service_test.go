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

	jsonSample := `{
		"sites": [
			{"url": "192.168.1.1"},
			{"url": "http://10.0.0.1:8080/path"},
			{"url": "192.168.1.1"}
		]
	}`

	res := svc.ImportJSONContent(jsonSample)
	if !res.Success {
		t.Fatalf("ImportJSONContent falhou: %s", res.Error)
	}

	if res.Count != 2 {
		t.Errorf("Esperava 2 IPs inseridos (únicos), obteve %d", res.Count)
	}

	ips, err := svc.ListIPs()
	if err != nil {
		t.Fatalf("ListIPs falhou: %v", err)
	}

	if len(ips) != 2 {
		t.Errorf("Esperava 2 IPs listados, obteve %d", len(ips))
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

	// Adicionar IP válido
	res := svc.AddIP("192.168.0.10")
	if !res.Success {
		t.Fatalf("Esperava sucesso ao adicionar IP válido: %s", res.Error)
	}

	// Tentar adicionar IP inválido
	resInvalid := svc.AddIP("invalido_ip")
	if resInvalid.Success {
		t.Fatalf("Esperava falha ao adicionar IP inválido")
	}

	// Remover IP
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
