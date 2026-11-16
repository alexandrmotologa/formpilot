package crm

import (
	"context"
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
