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
		LatencyMs:         480,
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
	if !strings.Contains(rec.Body.String(), "healthy") {
		t.Errorf("expected healthy in body, got %s", rec.Body.String())
	}

	// 2. Simulate FinTech lead
	simReq := httptest.NewRequest("POST", "/api/v1/simulate", strings.NewReader(`{"preset":"fintech"}`))
	simReq.Header.Set("Content-Type", "application/json")
	simRec := httptest.NewRecorder()
	handler.ServeHTTP(simRec, simReq)

	if simRec.Code != http.StatusOK {
		t.Errorf("simulate expected 200, got %d: %s", simRec.Code, simRec.Body.String())
	}
	if !strings.Contains(simRec.Body.String(), "Apex Clearing") {
		t.Errorf("expected Apex Clearing in simulated lead, got %s", simRec.Body.String())
	}

	// 3. Stats
	statsReq := httptest.NewRequest("GET", "/api/v1/stats", nil)
	statsRec := httptest.NewRecorder()
	handler.ServeHTTP(statsRec, statsReq)

	if statsRec.Code != http.StatusOK {
		t.Errorf("stats expected 200, got %d", statsRec.Code)
	}
}
