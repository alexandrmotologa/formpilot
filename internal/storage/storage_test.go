package storage

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestStorageLifecycle(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_formpilot.db")

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	// 1. Insert Lead
	lead := &Lead{
		ID:                "lead_1001",
		ReceivedAt:        time.Now().UTC(),
		Source:            "sendgrid",
		Sender:            "marcus@apexclearing.com",
		Recipient:         "sales@formpilot.io",
		Subject:           "Enterprise Quote",
		ParsedBody:        "Looking for 200 seats",
		CompanyName:       "Apex Clearing",
		Domain:            "apexclearing.com",
		ContactName:       "Marcus Vance",
		ContactEmail:      "marcus@apexclearing.com",
		RequestedSeats:    200,
		Budget:            100000,
		Urgency:           "Immediate",
		ICPScore:          92,
		QualificationTier: "Enterprise Hot Lead",
		DedupStatus:       "new_account",
		Status:            "synced",
		LatencyMs:         1450,
	}

	if err := db.InsertLead(lead); err != nil {
		t.Fatalf("insert lead failed: %v", err)
	}

	// 2. Query Lead
	fetched, err := db.GetLead("lead_1001")
	if err != nil {
		t.Fatalf("get lead failed: %v", err)
	}
	if fetched == nil || fetched.CompanyName != "Apex Clearing" {
		t.Fatalf("unexpected fetched lead: %+v", fetched)
	}

	// 3. Upsert Account
	acc := &AccountRecord{
		ID:          "acc_2001",
		Domain:      "apexclearing.com",
		CompanyName: "Apex Clearing LLC",
		OwnerName:   "Sarah Connor",
		OwnerEmail:  "sarah@ourcompany.com",
		CRMID:       "hs_comp_99",
		CRMType:     "hubspot",
	}
	if err := db.UpsertAccount(acc); err != nil {
		t.Fatalf("upsert account failed: %v", err)
	}

	// 4. Query Account
	fetchedAcc, err := db.GetAccountByDomain("apexclearing.com")
	if err != nil {
		t.Fatalf("get account failed: %v", err)
	}
	if fetchedAcc == nil || fetchedAcc.OwnerEmail != "sarah@ourcompany.com" {
		t.Fatalf("unexpected account: %+v", fetchedAcc)
	}

	// 5. Audit Event & Sync Log
	err = db.InsertAuditEvent(&AuditEvent{
		ID:      "aud_1",
		LeadID:  "lead_1001",
		Stage:   "receiver",
		Event:   "email_parsed",
		Details: "parsed 1 attachment",
	})
	if err != nil {
		t.Fatalf("insert audit event failed: %v", err)
	}

	err = db.InsertSyncLog(&SyncLog{
		ID:         "log_1",
		LeadID:     "lead_1001",
		CRMType:    "hubspot",
		Action:     "upsert_contact",
		Status:     "success",
		StatusCode: 200,
		DurationMs: 420,
	})
	if err != nil {
		t.Fatalf("insert sync log failed: %v", err)
	}

	// 6. Stats Summary
	stats, err := db.GetStatsSummary()
	if err != nil {
		t.Fatalf("get stats failed: %v", err)
	}
	if stats.TotalLeads != 1 || stats.EnterpriseHotLeads != 1 {
		t.Fatalf("unexpected stats: %+v", stats)
	}

	_ = os.Remove(dbPath)
}
