package pipeline

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/alexandrmotologa/formpilot/internal/config"
	"github.com/alexandrmotologa/formpilot/internal/dedup"
	"github.com/alexandrmotologa/formpilot/internal/receiver"
	"github.com/alexandrmotologa/formpilot/internal/server"
	"github.com/alexandrmotologa/formpilot/internal/storage"
)

func TestPipelineEndToEnd(t *testing.T) {
	tempDir := t.TempDir()
	db, err := storage.Open(filepath.Join(tempDir, "test_pipeline.db"))
	if err != nil {
		t.Fatalf("open db failed: %v", err)
	}
	defer db.Close()

	cfg := config.DefaultConfig()
	cache, err := dedup.NewAccountCache(db)
	if err != nil {
		t.Fatalf("cache init failed: %v", err)
	}

	sse := server.NewSSEBroadcaster()
	pipe := NewPipeline(cfg, db, cache, sse)

	// 1. Process first lead (New Account)
	payload1 := &receiver.InboundPayload{
		Source:    "sendgrid",
		Sender:    "Marcus Vance <m.vance@apexclearing.com>",
		Recipient: "sales@formpilot.io",
		Subject:   "Enterprise RFP: Inbound Data Pipeline for 250 FinTech Seats",
		Body: `Hi Sales Team,
I am Marcus Vance, VP of Engineering at Apex Clearing (apexclearing.com).
We need 250 enterprise seats for our trading desk.
Allocated budget is $120,000 for annual licensing.
Timeline: Immediate Q4 initiative; procurement within 30 days.
Phone: +1 (415) 890-3421`,
		RawPayload: "raw text",
	}

	lead1, err := pipe.Process(context.Background(), payload1)
	if err != nil {
		t.Fatalf("pipeline process lead 1 failed: %v", err)
	}

	if lead1.CompanyName != "Apex Clearing" {
		t.Errorf("expected Apex Clearing, got %q", lead1.CompanyName)
	}
	if lead1.Domain != "apexclearing.com" {
		t.Errorf("expected apexclearing.com, got %q", lead1.Domain)
	}
	if lead1.ICPScore < 80 {
		t.Errorf("expected score >= 80, got %d", lead1.ICPScore)
	}
	if lead1.DedupStatus != "new_account" {
		t.Errorf("expected new_account, got %s", lead1.DedupStatus)
	}
	if lead1.LatencyMs > 3000 {
		t.Errorf("latency %dms exceeded target 3000ms", lead1.LatencyMs)
	}

	// 2. Process second lead from SAME company (should trigger deduplication match!)
	payload2 := &receiver.InboundPayload{
		Source:    "web_form",
		Sender:    "David Kim <dkim@apexclearing.com>",
		Recipient: "info@formpilot.io",
		Subject:   "Additional 50 seats for risk team",
		Body: `Hello, this is David Kim, Director of Risk at Apex Clearing.
We want to add 50 seats to the evaluation our VP Marcus Vance started.`,
		RawPayload: "raw text 2",
	}

	lead2, err := pipe.Process(context.Background(), payload2)
	if err != nil {
		t.Fatalf("pipeline process lead 2 failed: %v", err)
	}

	if lead2.DedupStatus != "existing_matched" {
		t.Errorf("expected existing_matched for second lead from apexclearing.com, got %s", lead2.DedupStatus)
	}
	if lead2.MatchedAccountName != "Apex Clearing" {
		t.Errorf("expected matched account name Apex Clearing, got %s", lead2.MatchedAccountName)
	}

	// 3. Verify SQLite stats
	stats, err := db.GetStatsSummary()
	if err != nil {
		t.Fatalf("get stats failed: %v", err)
	}
	if stats.TotalLeads != 2 {
		t.Errorf("expected 2 leads, got %d", stats.TotalLeads)
	}
	if stats.DeduplicatedAccounts != 1 {
		t.Errorf("expected 1 deduplicated account, got %d", stats.DeduplicatedAccounts)
	}
}
