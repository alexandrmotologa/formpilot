package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/alexandrmotologa/formpilot/internal/config"
	"github.com/alexandrmotologa/formpilot/internal/receiver"
	"github.com/alexandrmotologa/formpilot/internal/storage"
)

type mockPipeline struct{}

func (m *mockPipeline) Process(ctx context.Context, p *receiver.InboundPayload) (*storage.Lead, error) {
	return &storage.Lead{
		ID:                "lead_sim_1",
		ReceivedAt:        time.Now().UTC(),
		Source:            p.Source,
		Sender:            p.Sender,
		Subject:           p.Subject,
		ParsedBody:        p.Body,
		CompanyName:       "Apex Clearing",
		Domain:            "apexclearing.com",
		ICPScore:          92,
		QualificationTier: "Enterprise Hot Lead",
		Status:            "synced",
		LatencyMs:         48,
	}, nil
}

func TestServerEndpoints(t *testing.T) {
	tempDir := t.TempDir()
	db, err := storage.Open(filepath.Join(tempDir, "test.db"))
	if err != nil {
		t.Fatalf("open db failed: %v", err)
	}
	defer db.Close()

	cfg := config.DefaultConfig()
	sse := NewSSEBroadcaster()
	srv := NewServer(cfg, db, sse)
	srv.SetPipeline(&mockPipeline{})
	handler := srv.Routes()

	// 1. Health check
	req := httptest.NewRequest("GET", "/api/v1/health", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("health expected 200, got %d", rec.Code)
	}

	// 2. Simulate FinTech lead
	simReq := httptest.NewRequest("POST", "/api/v1/simulate", strings.NewReader(`{"preset":"fintech"}`))
	simReq.Header.Set("Content-Type", "application/json")
	simRec := httptest.NewRecorder()
	handler.ServeHTTP(simRec, simReq)
	if simRec.Code != http.StatusOK {
		t.Errorf("simulate expected 200, got %d: %s", simRec.Code, simRec.Body.String())
	}

	// 3. Stats
	statsReq := httptest.NewRequest("GET", "/api/v1/stats", nil)
	statsRec := httptest.NewRecorder()
	handler.ServeHTTP(statsRec, statsReq)
	if statsRec.Code != http.StatusOK {
		t.Errorf("stats expected 200, got %d", statsRec.Code)
	}

	// 4. Webhook Info
	infoReq := httptest.NewRequest("GET", "/api/v1/webhooks/info", nil)
	infoRec := httptest.NewRecorder()
	handler.ServeHTTP(infoRec, infoReq)
	if infoRec.Code != http.StatusOK || !strings.Contains(infoRec.Body.String(), "curl -X POST") {
		t.Errorf("expected webhook info with curl samples, got %s", infoRec.Body.String())
	}

	// 5. CSV Export
	csvReq := httptest.NewRequest("GET", "/api/v1/leads/export?format=csv", nil)
	csvRec := httptest.NewRecorder()
	handler.ServeHTTP(csvRec, csvReq)
	if csvRec.Code != http.StatusOK || !strings.Contains(csvRec.Body.String(), "Company,Domain") {
		t.Errorf("expected csv headers, got: %s", csvRec.Body.String())
	}

	// 6. JSON Export
	jsonReq := httptest.NewRequest("GET", "/api/v1/leads/export?format=json", nil)
	jsonRec := httptest.NewRecorder()
	handler.ServeHTTP(jsonRec, jsonReq)
	if jsonRec.Code != http.StatusOK {
		t.Errorf("json export expected 200, got %d", jsonRec.Code)
	}

	// 7. ICP Config GET and PUT
	getIcpReq := httptest.NewRequest("GET", "/api/v1/config/icp", nil)
	getIcpRec := httptest.NewRecorder()
	handler.ServeHTTP(getIcpRec, getIcpReq)
	if getIcpRec.Code != http.StatusOK {
		t.Errorf("get icp expected 200, got %d", getIcpRec.Code)
	}

	putIcpReq := httptest.NewRequest("PUT", "/api/v1/config/icp", strings.NewReader(`{"min_qualification_score":70,"weights":{"corporate_domain":25,"seat_volume":30,"seniority":20,"urgency":15,"budget":5,"product_fit":5}}`))
	putIcpReq.Header.Set("Content-Type", "application/json")
	putIcpRec := httptest.NewRecorder()
	handler.ServeHTTP(putIcpRec, putIcpReq)
	if putIcpRec.Code != http.StatusOK {
		t.Errorf("put icp expected 200, got %d", putIcpRec.Code)
	}

	// 8. Recalculate all leads
	recalcReq := httptest.NewRequest("POST", "/api/v1/leads/recalculate-all", nil)
	recalcRec := httptest.NewRecorder()
	handler.ServeHTTP(recalcRec, recalcReq)
	if recalcRec.Code != http.StatusOK {
		t.Errorf("recalculate expected 200, got %d", recalcRec.Code)
	}
}
