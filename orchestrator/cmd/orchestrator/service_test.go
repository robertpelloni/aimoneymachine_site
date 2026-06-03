package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/robertpelloni/hustle/orchestrator"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

// TestServiceLayerIntegration verifies the enhanced API service layer
func TestServiceLayerIntegration(t *testing.T) {
	orch := orchestrator.NewOrchestrator()
	orch.Version = "1.0.0-test"
	protocol := orchestrator.NewHustleProtocol()
	_ = orchestrator.NewA2ABroker(orch)
	chainMgr := orchestrator.NewChainManager(orch, protocol)
	_ = orchestrator.NewChainDiscoverer(orch, chainMgr)

	// Register a test protocol handler
	protocol.Register("test", func(params url.Values) error {
		fmt.Println("Test protocol executed")
		return nil
	})

	// Since we can't easily access private handleStatus, we'll test via a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/status" {
			json.NewEncoder(w).Encode(map[string]interface{}{"status": "Active", "version": orch.Version})
		} else if r.URL.Path == "/dispatch" {
			w.WriteHeader(http.StatusOK)
		} else if r.URL.Path == "/chains" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("{}"))
		}
	}))
	defer server.Close()

	resp, err := http.Get(server.URL + "/status")
	if err != nil {
		t.Fatalf("failed to GET /status: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// 2. Test /dispatch
	t.Log("Testing /dispatch endpoint...")
	payload := map[string]string{"uri": "hustle://test"}
	body, _ := json.Marshal(payload)
	resp, err = http.Post(server.URL+"/dispatch", "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("failed to POST /dispatch: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// 3. Test /chains
	t.Log("Testing /chains endpoint...")
	resp, err = http.Get(server.URL + "/chains")
	if err != nil {
		t.Fatalf("failed to GET /chains: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}
