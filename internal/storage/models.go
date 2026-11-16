package storage

import (
	"time"
)

type Lead struct {
	ID                  string    `json:"id"`
	ReceivedAt          time.Time `json:"received_at"`
	Source              string    `json:"source"`
	Sender              string    `json:"sender"`
	Recipient           string    `json:"recipient"`
	Subject             string    `json:"subject"`
	RawPayload          string    `json:"raw_payload"`
	ParsedBody          string    `json:"parsed_body"`
	ExtractedEntities   string    `json:"extracted_entities"`
	CompanyName         string    `json:"company_name"`
	Domain              string    `json:"domain"`
	ContactName         string    `json:"contact_name"`
	ContactEmail        string    `json:"contact_email"`
	ContactPhone        string    `json:"contact_phone"`
	RequestedSeats      int       `json:"requested_seats"`
	Budget              float64   `json:"budget"`
	Urgency             string    `json:"urgency"`
	ICPScore            int       `json:"icp_score"`
	QualificationTier   string    `json:"qualification_tier"`
	DedupStatus         string    `json:"dedup_status"`
	MatchedAccountID    string    `json:"matched_account_id"`
	MatchedAccountName  string    `json:"matched_account_name"`
	MatchedAccountOwner string    `json:"matched_account_owner"`
	OutreachDraft       string    `json:"outreach_draft"`
	Status              string    `json:"status"`
	LatencyMs           int64     `json:"latency_ms"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

type AccountRecord struct {
	ID          string    `json:"id"`
	Domain      string    `json:"domain"`
	CompanyName string    `json:"company_name"`
	OwnerName   string    `json:"owner_name"`
	OwnerEmail  string    `json:"owner_email"`
	CRMID       string    `json:"crm_id"`
	CRMType     string    `json:"crm_type"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type SyncLog struct {
	ID              string    `json:"id"`
	LeadID          string    `json:"lead_id"`
	CRMType         string    `json:"crm_type"`
	Action          string    `json:"action"`
	Status          string    `json:"status"`
	StatusCode      int       `json:"status_code"`
	RequestPayload  string    `json:"request_payload"`
	ResponsePayload string    `json:"response_payload"`
	DurationMs      int64     `json:"duration_ms"`
	ErrorMessage    string    `json:"error_message"`
	CreatedAt       time.Time `json:"created_at"`
}

type AuditEvent struct {
	ID        string    `json:"id"`
	LeadID    string    `json:"lead_id"`
	Stage     string    `json:"stage"`
	Event     string    `json:"event"`
	Details   string    `json:"details"`
	CreatedAt time.Time `json:"created_at"`
}

type StatsSummary struct {
	TotalLeads           int     `json:"total_leads"`
	EnterpriseHotLeads   int     `json:"enterprise_hot_leads"`
	MidMarketLeads       int     `json:"mid_market_leads"`
	SMBLeads             int     `json:"smb_leads"`
	UnqualifiedLeads     int     `json:"unqualified_leads"`
	DeduplicatedAccounts int     `json:"deduplicated_accounts"`
	SuccessfulSyncs      int     `json:"successful_syncs"`
	AverageLatencyMs     float64 `json:"average_latency_ms"`
}
