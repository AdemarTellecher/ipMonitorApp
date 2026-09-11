package service_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/AdemarTellecher/ipmonitorapp/internal/model"
	"github.com/AdemarTellecher/ipmonitorapp/internal/service"
)

func TestMonitorService_RESTEndpoints(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_api.db")
	repo, err := model.NewRepository(dbPath)
	if err != nil {
		t.Fatalf("Erro ao criar repo: %v", err)
	}
	defer repo.Close()

	svc := service.NewMonitorService(repo)

	// 1. Testar GET /api/version
	reqVer := httptest.NewRequest("GET", "/api/version", nil)
	recVer := httptest.NewRecorder()
	svc.ServeHTTP(recVer, reqVer)

	if recVer.Code != http.StatusOK {
		t.Errorf("Esperava 200 OK para /api/version, obteve %d", recVer.Code)
	}

	// 2. Testar POST /api/add
	addPayload := map[string]interface{}{
		"ip":          "1.1.1.1",
		"name":        "Cloudflare DNS",
		"method":      "PING",
		"thresholdMs": 1500,
	}
	bodyBytes, _ := json.Marshal(addPayload)
	reqAdd := httptest.NewRequest("POST", "/api/add", bytes.NewReader(bodyBytes))
	recAdd := httptest.NewRecorder()
	svc.ServeHTTP(recAdd, reqAdd)

	if recAdd.Code != http.StatusOK {
		t.Errorf("Esperava 200 OK para /api/add, obteve %d", recAdd.Code)
	}

	// 3. Testar GET /api/list
	reqList := httptest.NewRequest("GET", "/api/list", nil)
	recList := httptest.NewRecorder()
	svc.ServeHTTP(recList, reqList)

	if recList.Code != http.StatusOK {
		t.Errorf("Esperava 200 OK para /api/list, obteve %d", recList.Code)
	}

	var overview model.NetworkOverview
	if err := json.NewDecoder(recList.Body).Decode(&overview); err != nil {
		t.Fatalf("Erro ao decodificar /api/list: %v", err)
	}

	if overview.Total != 1 {
		t.Errorf("Esperava 1 host no overview, obteve %d", overview.Total)
	}
}
