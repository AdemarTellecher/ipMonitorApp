package model_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/AdemarTellecher/ipmonitorapp/internal/model"
)

func TestResolveDBPath_NonEmpty(t *testing.T) {
	path := model.ResolveDBPath("test.db")
	if path == "" {
		t.Fatalf("Esperava um caminho resolvido para test.db, mas retornou vazio")
	}

	if filepath.Base(path) != "test.db" {
		t.Errorf("Esperava nome base test.db, obteve %s", filepath.Base(path))
	}
}

func TestResolveDBPath_DefaultName(t *testing.T) {
	path := model.ResolveDBPath("")
	if filepath.Base(path) != "ipMonitorDB.db" {
		t.Errorf("Esperava nome padrão ipMonitorDB.db, obteve %s", filepath.Base(path))
	}
}

func TestResolveDBPath_DirectoryAccessible(t *testing.T) {
	path := model.ResolveDBPath("test_dir_access.db")
	dir := filepath.Dir(path)

	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("Diretório resolvido %s não está acessível: %v", dir, err)
	}

	if !info.IsDir() {
		t.Fatalf("Caminho retornado não é um diretório: %s", dir)
	}
}
