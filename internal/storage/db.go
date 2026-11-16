package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	_ "modernc.org/sqlite"
)

type DB struct {
	db *sql.DB
	mu sync.RWMutex
}

func Open(dbPath string) (*DB, error) {
	if dbPath == "" {
		dbPath = "formpilot.db"
	}

	dir := filepath.Dir(dbPath)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("create database directory: %w", err)
		}
	}

	// SQLite connection string with performance pragmas
	dsn := fmt.Sprintf("%s?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)&_pragma=foreign_keys(ON)&_pragma=cache_size(-2000)", dbPath)

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite db: %w", err)
	}

	// Limit to single connection for serialized write safety on SQLite
	db.SetMaxOpenConns(1)

	s := &DB{db: db}
	if err := s.migrate(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("run migrations: %w", err)
	}

	return s, nil
}

func (s *DB) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

func (s *DB) migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS leads (
		id TEXT PRIMARY KEY,
		received_at DATETIME NOT NULL,
		source TEXT NOT NULL,
		sender TEXT NOT NULL,
		recipient TEXT NOT NULL,
		subject TEXT NOT NULL,
		raw_payload TEXT,
		parsed_body TEXT,
		extracted_entities TEXT,
		company_name TEXT,
		domain TEXT,
		contact_name TEXT,
		contact_email TEXT,
		contact_phone TEXT,
		requested_seats INTEGER DEFAULT 0,
		budget REAL DEFAULT 0,
		urgency TEXT,
		icp_score INTEGER DEFAULT 0,
		qualification_tier TEXT,
		dedup_status TEXT,
		matched_account_id TEXT,
		matched_account_name TEXT,
		matched_account_owner TEXT,
		outreach_draft TEXT,
		status TEXT DEFAULT 'received',
		latency_ms INTEGER DEFAULT 0,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_leads_domain ON leads(domain);
	CREATE INDEX IF NOT EXISTS idx_leads_received_at ON leads(received_at DESC);
	CREATE INDEX IF NOT EXISTS idx_leads_qualification_tier ON leads(qualification_tier);

	CREATE TABLE IF NOT EXISTS accounts_cache (
		id TEXT PRIMARY KEY,
		domain TEXT UNIQUE NOT NULL,
		company_name TEXT NOT NULL,
		owner_name TEXT NOT NULL,
		owner_email TEXT NOT NULL,
		crm_id TEXT NOT NULL,
		crm_type TEXT NOT NULL,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_accounts_domain ON accounts_cache(domain);
	CREATE INDEX IF NOT EXISTS idx_accounts_company_name ON accounts_cache(company_name);

	CREATE TABLE IF NOT EXISTS sync_logs (
		id TEXT PRIMARY KEY,
		lead_id TEXT NOT NULL,
		crm_type TEXT NOT NULL,
		action TEXT NOT NULL,
		status TEXT NOT NULL,
		status_code INTEGER DEFAULT 0,
		request_payload TEXT,
		response_payload TEXT,
		duration_ms INTEGER DEFAULT 0,
		error_message TEXT,
		created_at DATETIME NOT NULL,
		FOREIGN KEY (lead_id) REFERENCES leads(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_sync_logs_lead_id ON sync_logs(lead_id);

	CREATE TABLE IF NOT EXISTS audit_events (
		id TEXT PRIMARY KEY,
		lead_id TEXT NOT NULL,
		stage TEXT NOT NULL,
		event TEXT NOT NULL,
		details TEXT,
		created_at DATETIME NOT NULL,
		FOREIGN KEY (lead_id) REFERENCES leads(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_audit_events_lead_id ON audit_events(lead_id);
	`

	_, err := s.db.Exec(schema)
	return err
}
