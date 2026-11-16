package storage

import (
	"database/sql"
	"fmt"
	"time"
)

// InsertLead saves a new lead record into the database.
func (s *DB) InsertLead(lead *Lead) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	if lead.CreatedAt.IsZero() {
		lead.CreatedAt = now
	}
	lead.UpdatedAt = now

	query := `
	INSERT INTO leads (
		id, received_at, source, sender, recipient, subject, raw_payload, parsed_body,
		extracted_entities, company_name, domain, contact_name, contact_email, contact_phone,
		requested_seats, budget, urgency, icp_score, qualification_tier, dedup_status,
		matched_account_id, matched_account_name, matched_account_owner, outreach_draft,
		status, latency_ms, latency_mime_ms, latency_extract_ms, latency_dedup_ms,
		latency_score_ms, latency_draft_ms, latency_crm_ms, latency_storage_ms, created_at, updated_at
	) VALUES (
		?, ?, ?, ?, ?, ?, ?, ?,
		?, ?, ?, ?, ?, ?,
		?, ?, ?, ?, ?, ?,
		?, ?, ?, ?,
		?, ?, ?, ?, ?,
		?, ?, ?, ?, ?, ?
	)
	`

	_, err := s.db.Exec(query,
		lead.ID, lead.ReceivedAt, lead.Source, lead.Sender, lead.Recipient, lead.Subject,
		lead.RawPayload, lead.ParsedBody, lead.ExtractedEntities, lead.CompanyName, lead.Domain,
		lead.ContactName, lead.ContactEmail, lead.ContactPhone, lead.RequestedSeats, lead.Budget,
		lead.Urgency, lead.ICPScore, lead.QualificationTier, lead.DedupStatus, lead.MatchedAccountID,
		lead.MatchedAccountName, lead.MatchedAccountOwner, lead.OutreachDraft, lead.Status,
		lead.LatencyMs, lead.LatencyMimeMs, lead.LatencyExtractMs, lead.LatencyDedupMs,
		lead.LatencyScoreMs, lead.LatencyDraftMs, lead.LatencyCrmMs, lead.LatencyStorageMs,
		lead.CreatedAt, lead.UpdatedAt,
	)
	return err
}

// UpdateLead updates an existing lead.
func (s *DB) UpdateLead(lead *Lead) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	lead.UpdatedAt = time.Now().UTC()

	query := `
	UPDATE leads SET
		source = ?, sender = ?, recipient = ?, subject = ?, parsed_body = ?,
		extracted_entities = ?, company_name = ?, domain = ?, contact_name = ?,
		contact_email = ?, contact_phone = ?, requested_seats = ?, budget = ?,
		urgency = ?, icp_score = ?, qualification_tier = ?, dedup_status = ?,
		matched_account_id = ?, matched_account_name = ?, matched_account_owner = ?,
		outreach_draft = ?, status = ?, latency_ms = ?, latency_mime_ms = ?,
		latency_extract_ms = ?, latency_dedup_ms = ?, latency_score_ms = ?,
		latency_draft_ms = ?, latency_crm_ms = ?, latency_storage_ms = ?, updated_at = ?
	WHERE id = ?
	`

	_, err := s.db.Exec(query,
		lead.Source, lead.Sender, lead.Recipient, lead.Subject, lead.ParsedBody,
		lead.ExtractedEntities, lead.CompanyName, lead.Domain, lead.ContactName,
		lead.ContactEmail, lead.ContactPhone, lead.RequestedSeats, lead.Budget,
		lead.Urgency, lead.ICPScore, lead.QualificationTier, lead.DedupStatus,
		lead.MatchedAccountID, lead.MatchedAccountName, lead.MatchedAccountOwner,
		lead.OutreachDraft, lead.Status, lead.LatencyMs, lead.LatencyMimeMs,
		lead.LatencyExtractMs, lead.LatencyDedupMs, lead.LatencyScoreMs,
		lead.LatencyDraftMs, lead.LatencyCrmMs, lead.LatencyStorageMs, lead.UpdatedAt, lead.ID,
	)
	return err
}

// GetLead retrieves a single lead by ID.
func (s *DB) GetLead(id string) (*Lead, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := `
	SELECT id, received_at, source, sender, recipient, subject, raw_payload, parsed_body,
		extracted_entities, company_name, domain, contact_name, contact_email, contact_phone,
		requested_seats, budget, urgency, icp_score, qualification_tier, dedup_status,
		matched_account_id, matched_account_name, matched_account_owner, outreach_draft,
		status, latency_ms, latency_mime_ms, latency_extract_ms, latency_dedup_ms,
		latency_score_ms, latency_draft_ms, latency_crm_ms, latency_storage_ms, created_at, updated_at
	FROM leads WHERE id = ?
	`

	row := s.db.QueryRow(query, id)
	l := &Lead{}
	err := row.Scan(
		&l.ID, &l.ReceivedAt, &l.Source, &l.Sender, &l.Recipient, &l.Subject, &l.RawPayload, &l.ParsedBody,
		&l.ExtractedEntities, &l.CompanyName, &l.Domain, &l.ContactName, &l.ContactEmail, &l.ContactPhone,
		&l.RequestedSeats, &l.Budget, &l.Urgency, &l.ICPScore, &l.QualificationTier, &l.DedupStatus,
		&l.MatchedAccountID, &l.MatchedAccountName, &l.MatchedAccountOwner, &l.OutreachDraft,
		&l.Status, &l.LatencyMs, &l.LatencyMimeMs, &l.LatencyExtractMs, &l.LatencyDedupMs,
		&l.LatencyScoreMs, &l.LatencyDraftMs, &l.LatencyCrmMs, &l.LatencyStorageMs, &l.CreatedAt, &l.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("scan lead: %w", err)
	}
	return l, nil
}

// ListLeads retrieves a list of leads sorted by received_at descending.
func (s *DB) ListLeads(limit, offset int, tier string) ([]Lead, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if limit <= 0 {
		limit = 50
	}

	query := `
	SELECT id, received_at, source, sender, recipient, subject, parsed_body,
		extracted_entities, company_name, domain, contact_name, contact_email, contact_phone,
		requested_seats, budget, urgency, icp_score, qualification_tier, dedup_status,
		matched_account_id, matched_account_name, matched_account_owner, outreach_draft,
		status, latency_ms, latency_mime_ms, latency_extract_ms, latency_dedup_ms,
		latency_score_ms, latency_draft_ms, latency_crm_ms, latency_storage_ms, created_at, updated_at
	FROM leads
	`
	args := []interface{}{}

	if tier != "" && tier != "all" {
		query += " WHERE qualification_tier = ?"
		args = append(args, tier)
	}

	query += " ORDER BY received_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("query leads: %w", err)
	}
	defer rows.Close()

	var leads []Lead
	for rows.Next() {
		var l Lead
		err := rows.Scan(
			&l.ID, &l.ReceivedAt, &l.Source, &l.Sender, &l.Recipient, &l.Subject, &l.ParsedBody,
			&l.ExtractedEntities, &l.CompanyName, &l.Domain, &l.ContactName, &l.ContactEmail, &l.ContactPhone,
			&l.RequestedSeats, &l.Budget, &l.Urgency, &l.ICPScore, &l.QualificationTier, &l.DedupStatus,
			&l.MatchedAccountID, &l.MatchedAccountName, &l.MatchedAccountOwner, &l.OutreachDraft,
			&l.Status, &l.LatencyMs, &l.LatencyMimeMs, &l.LatencyExtractMs, &l.LatencyDedupMs,
			&l.LatencyScoreMs, &l.LatencyDraftMs, &l.LatencyCrmMs, &l.LatencyStorageMs, &l.CreatedAt, &l.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan lead row: %w", err)
		}
		leads = append(leads, l)
	}
	return leads, nil
}

// ListAllLeads retrieves all leads without limit for data export.
func (s *DB) ListAllLeads(tier string) ([]Lead, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := `
	SELECT id, received_at, source, sender, recipient, subject, parsed_body,
		extracted_entities, company_name, domain, contact_name, contact_email, contact_phone,
		requested_seats, budget, urgency, icp_score, qualification_tier, dedup_status,
		matched_account_id, matched_account_name, matched_account_owner, outreach_draft,
		status, latency_ms, latency_mime_ms, latency_extract_ms, latency_dedup_ms,
		latency_score_ms, latency_draft_ms, latency_crm_ms, latency_storage_ms, created_at, updated_at
	FROM leads
	`
	args := []interface{}{}

	if tier != "" && tier != "all" {
		query += " WHERE qualification_tier = ?"
		args = append(args, tier)
	}

	query += " ORDER BY received_at DESC"

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("query all leads: %w", err)
	}
	defer rows.Close()

	var leads []Lead
	for rows.Next() {
		var l Lead
		err := rows.Scan(
			&l.ID, &l.ReceivedAt, &l.Source, &l.Sender, &l.Recipient, &l.Subject, &l.ParsedBody,
			&l.ExtractedEntities, &l.CompanyName, &l.Domain, &l.ContactName, &l.ContactEmail, &l.ContactPhone,
			&l.RequestedSeats, &l.Budget, &l.Urgency, &l.ICPScore, &l.QualificationTier, &l.DedupStatus,
			&l.MatchedAccountID, &l.MatchedAccountName, &l.MatchedAccountOwner, &l.OutreachDraft,
			&l.Status, &l.LatencyMs, &l.LatencyMimeMs, &l.LatencyExtractMs, &l.LatencyDedupMs,
			&l.LatencyScoreMs, &l.LatencyDraftMs, &l.LatencyCrmMs, &l.LatencyStorageMs, &l.CreatedAt, &l.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan lead row: %w", err)
		}
		leads = append(leads, l)
	}
	return leads, nil
}

// GetAccountByDomain searches for an account in the cache by normalized domain.
func (s *DB) GetAccountByDomain(domain string) (*AccountRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := `
	SELECT id, domain, company_name, owner_name, owner_email, crm_id, crm_type, created_at, updated_at
	FROM accounts_cache WHERE domain = ? LIMIT 1
	`

	row := s.db.QueryRow(query, domain)
	acc := &AccountRecord{}
	err := row.Scan(
		&acc.ID, &acc.Domain, &acc.CompanyName, &acc.OwnerName, &acc.OwnerEmail,
		&acc.CRMID, &acc.CRMType, &acc.CreatedAt, &acc.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("scan account by domain: %w", err)
	}
	return acc, nil
}

// ListAllAccounts returns all cached accounts for deduplication fuzzy matching.
func (s *DB) ListAllAccounts() ([]AccountRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := `
	SELECT id, domain, company_name, owner_name, owner_email, crm_id, crm_type, created_at, updated_at
	FROM accounts_cache ORDER BY company_name ASC
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("query all accounts: %w", err)
	}
	defer rows.Close()

	var accounts []AccountRecord
	for rows.Next() {
		var a AccountRecord
		if err := rows.Scan(
			&a.ID, &a.Domain, &a.CompanyName, &a.OwnerName, &a.OwnerEmail,
			&a.CRMID, &a.CRMType, &a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan account row: %w", err)
		}
		accounts = append(accounts, a)
	}
	return accounts, nil
}

// UpsertAccount updates or inserts an account in the cache.
func (s *DB) UpsertAccount(acc *AccountRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	if acc.CreatedAt.IsZero() {
		acc.CreatedAt = now
	}
	acc.UpdatedAt = now

	query := `
	INSERT INTO accounts_cache (id, domain, company_name, owner_name, owner_email, crm_id, crm_type, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(domain) DO UPDATE SET
		company_name = excluded.company_name,
		owner_name = excluded.owner_name,
		owner_email = excluded.owner_email,
		crm_id = excluded.crm_id,
		crm_type = excluded.crm_type,
		updated_at = excluded.updated_at
	`

	_, err := s.db.Exec(query,
		acc.ID, acc.Domain, acc.CompanyName, acc.OwnerName, acc.OwnerEmail,
		acc.CRMID, acc.CRMType, acc.CreatedAt, acc.UpdatedAt,
	)
	return err
}

// InsertSyncLog logs a CRM synchronization event.
func (s *DB) InsertSyncLog(log *SyncLog) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if log.CreatedAt.IsZero() {
		log.CreatedAt = time.Now().UTC()
	}

	query := `
	INSERT INTO sync_logs (
		id, lead_id, crm_type, action, status, status_code,
		request_payload, response_payload, duration_ms, error_message, created_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.Exec(query,
		log.ID, log.LeadID, log.CRMType, log.Action, log.Status, log.StatusCode,
		log.RequestPayload, log.ResponsePayload, log.DurationMs, log.ErrorMessage, log.CreatedAt,
	)
	return err
}

// ListSyncLogsForLead returns all CRM dispatch logs for a given lead.
func (s *DB) ListSyncLogsForLead(leadID string) ([]SyncLog, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := `
	SELECT id, lead_id, crm_type, action, status, status_code,
		request_payload, response_payload, duration_ms, error_message, created_at
	FROM sync_logs WHERE lead_id = ? ORDER BY created_at ASC
	`

	rows, err := s.db.Query(query, leadID)
	if err != nil {
		return nil, fmt.Errorf("query sync logs: %w", err)
	}
	defer rows.Close()

	var logs []SyncLog
	for rows.Next() {
		var l SyncLog
		if err := rows.Scan(
			&l.ID, &l.LeadID, &l.CRMType, &l.Action, &l.Status, &l.StatusCode,
			&l.RequestPayload, &l.ResponsePayload, &l.DurationMs, &l.ErrorMessage, &l.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan sync log row: %w", err)
		}
		logs = append(logs, l)
	}
	return logs, nil
}

// InsertAuditEvent records an audit trail step for a lead.
func (s *DB) InsertAuditEvent(event *AuditEvent) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now().UTC()
	}

	query := `
	INSERT INTO audit_events (id, lead_id, stage, event, details, created_at)
	VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.Exec(query, event.ID, event.LeadID, event.Stage, event.Event, event.Details, event.CreatedAt)
	return err
}

// ListAuditEventsForLead returns all lifecycle events for a lead.
func (s *DB) ListAuditEventsForLead(leadID string) ([]AuditEvent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := `
	SELECT id, lead_id, stage, event, details, created_at
	FROM audit_events WHERE lead_id = ? ORDER BY created_at ASC
	`

	rows, err := s.db.Query(query, leadID)
	if err != nil {
		return nil, fmt.Errorf("query audit events: %w", err)
	}
	defer rows.Close()

	var events []AuditEvent
	for rows.Next() {
		var e AuditEvent
		if err := rows.Scan(&e.ID, &e.LeadID, &e.Stage, &e.Event, &e.Details, &e.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan audit event row: %w", err)
		}
		events = append(events, e)
	}
	return events, nil
}

// GetStatsSummary aggregates operational metrics across all leads.
func (s *DB) GetStatsSummary() (*StatsSummary, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := &StatsSummary{}

	query := `
	SELECT
		COUNT(*) as total_leads,
		COALESCE(SUM(CASE WHEN qualification_tier = 'Enterprise Hot Lead' THEN 1 ELSE 0 END), 0) as enterprise_leads,
		COALESCE(SUM(CASE WHEN qualification_tier = 'Mid-Market Qualified' THEN 1 ELSE 0 END), 0) as midmarket_leads,
		COALESCE(SUM(CASE WHEN qualification_tier = 'SMB / Self-Serve' THEN 1 ELSE 0 END), 0) as smb_leads,
		COALESCE(SUM(CASE WHEN qualification_tier = 'Unqualified / Low Intent' THEN 1 ELSE 0 END), 0) as unqualified_leads,
		COALESCE(SUM(CASE WHEN dedup_status = 'existing_matched' THEN 1 ELSE 0 END), 0) as dedup_accounts,
		COALESCE(SUM(CASE WHEN status = 'synced' OR status = 'contacted' THEN 1 ELSE 0 END), 0) as successful_syncs,
		COALESCE(SUM(CASE WHEN status = 'contacted' THEN 1 ELSE 0 END), 0) as contacted_leads,
		COALESCE(AVG(latency_ms), 0) as avg_latency
	FROM leads
	`

	row := s.db.QueryRow(query)
	err := row.Scan(
		&stats.TotalLeads,
		&stats.EnterpriseHotLeads,
		&stats.MidMarketLeads,
		&stats.SMBLeads,
		&stats.UnqualifiedLeads,
		&stats.DeduplicatedAccounts,
		&stats.SuccessfulSyncs,
		&stats.ContactedLeads,
		&stats.AverageLatencyMs,
	)
	if err != nil {
		return nil, fmt.Errorf("aggregate stats: %w", err)
	}
	return stats, nil
}
