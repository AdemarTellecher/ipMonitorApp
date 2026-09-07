package controller_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/AdemarTellecher/ipmonitorapp/internal/controller"
	"github.com/AdemarTellecher/ipmonitorapp/internal/model"
)

func TestImportFromJSON(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	repo, err := model.NewRepository(dbPath)
	if err != nil {
		t.Fatalf("Erro ao criar repo de teste: %v", err)
	}
	defer repo.Close()

	ctrl := controller.NewIPController(repo)

	// Abre o arquivo JSON real do repositório
	f, err := os.Open("../../Ping-2025-04-16T005844Z.json")
	if err != nil {
		t.Fatalf("Erro ao abrir arquivo JSON: %v", err)
	}
	defer f.Close()

	count, err := ctrl.ImportFromJSON(f)
	if err != nil {
		t.Fatalf("ImportFromJSON retornou erro: %v", err)
	}

	if count == 0 {
		t.Errorf("Esperava importar ao menos 1 IP, importou %d", count)
	}

	devices, err := ctrl.ListIPs()
	if err != nil {
		t.Fatalf("ListIPs retornou erro: %v", err)
	}

	t.Logf("Importados com sucesso %d IPs. Total no banco: %d", count, len(devices))
	for _, d := range devices {
		t.Logf("IP no banco: %s (status: %s)", d.IP, d.Status)
	}
}
