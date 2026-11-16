package crm

import (
	"context"
)

type CRMRequest struct {
	LeadID            string  `json:"lead_id"`
	CompanyName       string  `json:"company_name"`
	Domain            string  `json:"domain"`
	ContactName       string  `json:"contact_name"`
	ContactEmail      string  `json:"contact_email"`
	ContactPhone      string  `json:"contact_phone"`
	JobTitle          string  `json:"job_title"`
	RequestedSeats    int     `json:"requested_seats"`
	Budget            float64 `json:"budget"`
	Urgency           string  `json:"urgency"`
	ICPScore          int     `json:"icp_score"`
	QualificationTier string  `json:"qualification_tier"`
	DedupStatus       string  `json:"dedup_status"`
	MatchedAccountID  string  `json:"matched_account_id"`
	OutreachDraft     string  `json:"outreach_draft"`
}

type CRMResponse struct {
	CRMType      string `json:"crm_type"` // "hubspot", "salesforce", "pipedrive", "mock"
	Success      bool   `json:"success"`
	AccountID    string `json:"account_id"`
	ContactID    string `json:"contact_id"`
	DealID       string `json:"deal_id"`
	StatusCode   int    `json:"status_code"`
	DurationMs   int64  `json:"duration_ms"`
	ErrorMessage string `json:"error_message,omitempty"`
	RawRequest   string `json:"raw_request"`
	RawResponse  string `json:"raw_response"`
}

type Client interface {
	Name() string
	Sync(ctx context.Context, req *CRMRequest) (*CRMResponse, error)
}
