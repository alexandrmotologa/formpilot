package crm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alexandrmotologa/formpilot/internal/config"
)

func TestMockCRM_Sync(t *testing.T) {
	mock := NewMockCRM()
	req := &CRMRequest{
		LeadID:            "lead_test_123",
		CompanyName:       "Cyberdyne Systems",
		Domain:            "cyberdyne.systems",
		ContactName:       "Sarah Connor",
		ContactEmail:      "sarah@cyberdyne.systems",
		RequestedSeats:    100,
		Budget:            50000,
		ICPScore:          88,
		QualificationTier: "Enterprise Hot Lead",
		DedupStatus:       "new_account",
	}

	resp, err := mock.Sync(context.Background(), req)
	if err != nil {
		t.Fatalf("mock sync failed: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success")
	}
	if resp.AccountID == "" || resp.ContactID == "" || resp.DealID == "" {
		t.Errorf("expected IDs to be generated, got account=%s, contact=%s, deal=%s", resp.AccountID, resp.ContactID, resp.DealID)
	}
}

func TestDispatcher_MockMode(t *testing.T) {
	cfg := config.CRMConfig{Mode: "mock"}
	dispatcher := NewDispatcher(cfg, nil)

	req := &CRMRequest{
		LeadID:      "lead_test_456",
		CompanyName: "Stark Industries",
		Domain:      "stark.com",
	}

	responses, err := dispatcher.Dispatch(context.Background(), req)
	if err != nil {
		t.Fatalf("dispatcher failed: %v", err)
	}
	if len(responses) == 0 {
		t.Fatalf("expected at least 1 response")
	}
	if !responses[0].Success {
		t.Errorf("expected response to be successful")
	}
}

func TestPipedriveClient_MockFallback(t *testing.T) {
	client := NewPipedriveClient(config.PipedriveConfig{
		Enabled:  true,
		APIToken: "",
	})

	req := &CRMRequest{
		LeadID:       "lead_pd_123",
		CompanyName:  "Acme Corp",
		ContactEmail: "buyer@acme.com",
		Budget:       25000,
	}

	resp, err := client.Sync(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.CRMType != "pipedrive" {
		t.Errorf("expected pipedrive, got %s", resp.CRMType)
	}
	if !resp.Success {
		t.Errorf("expected success")
	}
}

func TestPipedriveClient_LiveAPISimulated(t *testing.T) {
	// Create mock test server mimicking Pipedrive REST API responses
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("api_token")
		if token != "valid_token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)

		switch r.URL.Path {
		case "/v1/organizations":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": true,
				"data":    map[string]interface{}{"id": 1001},
			})
		case "/v1/persons":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": true,
				"data":    map[string]interface{}{"id": 2002},
			})
		case "/v1/deals":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": true,
				"data":    map[string]interface{}{"id": 3003},
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	client := NewPipedriveClient(config.PipedriveConfig{
		Enabled:  true,
		APIToken: "valid_token",
	})

	// Test sendPost helper directly against the test server
	orgID, err := client.sendPost(context.Background(), ts.URL+"/v1/organizations", map[string]string{"name": "Wayne Enterprises"})
	if err != nil {
		t.Fatalf("sendPost failed: %v", err)
	}
	if orgID != "1001" {
		t.Errorf("expected 1001, got %s", orgID)
	}
}
