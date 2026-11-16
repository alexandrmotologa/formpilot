package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"strconv"
	"time"

	"github.com/alexandrmotologa/formpilot/internal/config"
	"github.com/alexandrmotologa/formpilot/internal/receiver"
	"github.com/alexandrmotologa/formpilot/internal/storage"
)

type PipelineProcessor interface {
	Process(ctx context.Context, r *receiver.InboundPayload) (*storage.Lead, error)
}

type Server struct {
	cfg       *config.Config
	db        *storage.DB
	pipeline  PipelineProcessor
	sse       *SSEBroadcaster
	startTime time.Time
}

func NewServer(cfg *config.Config, db *storage.DB, sse *SSEBroadcaster) *Server {
	return &Server{
		cfg:       cfg,
		db:        db,
		sse:       sse,
		startTime: time.Now(),
	}
}

func (s *Server) SetPipeline(p PipelineProcessor) {
	s.pipeline = p
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()

	// SSE Real-time events
	mux.Handle("GET /api/v1/events", s.sse)

	// System Health & Memory Benchmark
	mux.HandleFunc("GET /api/v1/health", s.handleHealth)
	mux.HandleFunc("GET /api/v1/stats", s.handleStats)

	// Lead Management API
	mux.HandleFunc("GET /api/v1/leads", s.handleListLeads)
	mux.HandleFunc("GET /api/v1/leads/{id}", s.handleGetLead)
	mux.HandleFunc("PUT /api/v1/leads/{id}/draft", s.handleUpdateDraft)
	mux.HandleFunc("POST /api/v1/leads/{id}/sync", s.handleResyncLead)

	// Accounts Deduplication Cache
	mux.HandleFunc("GET /api/v1/accounts", s.handleListAccounts)

	// Inbound Simulator
	mux.HandleFunc("POST /api/v1/simulate", s.handleSimulate)

	// Webhook Receivers
	mux.HandleFunc("POST /api/v1/webhooks/sendgrid", s.handleSendGridWebhook)
	mux.HandleFunc("POST /api/v1/webhooks/postmark", s.handlePostmarkWebhook)
	mux.HandleFunc("POST /api/v1/webhooks/mailgun", s.handleMailgunWebhook)
	mux.HandleFunc("POST /api/v1/webhooks/ses", s.handleSESWebhook)
	mux.HandleFunc("POST /api/v1/webhooks/generic", s.handleGenericWebhook)

	// Embedded Static Cockpit Assets
	mux.Handle("/", DistHandler())

	return s.corsMiddleware(mux)
}

func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	allocMB := float64(m.Alloc) / 1024 / 1024
	sysMB := float64(m.Sys) / 1024 / 1024

	health := map[string]interface{}{
		"status":          "healthy",
		"uptime_seconds":  int(time.Since(s.startTime).Seconds()),
		"memory_alloc_mb": fmt.Sprintf("%.2f MB", allocMB),
		"memory_sys_mb":   fmt.Sprintf("%.2f MB", sysMB),
		"memory_target":   "< 15.00 MB",
		"within_budget":   sysMB < 25.0, // target is low memory single binary
		"goroutines":      runtime.NumGoroutine(),
		"crm_mode":        s.cfg.CRM.Mode,
		"extraction_mode": s.cfg.Extraction.Provider,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(health)
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	if s.db == nil {
		http.Error(w, "database unavailable", http.StatusInternalServerError)
		return
	}

	stats, err := s.db.GetStatsSummary()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(stats)
}

func (s *Server) handleListLeads(w http.ResponseWriter, r *http.Request) {
	limit := 50
	offset := 0
	tier := r.URL.Query().Get("tier")

	if lStr := r.URL.Query().Get("limit"); lStr != "" {
		if l, err := strconv.Atoi(lStr); err == nil && l > 0 {
			limit = l
		}
	}
	if oStr := r.URL.Query().Get("offset"); oStr != "" {
		if o, err := strconv.Atoi(oStr); err == nil && o >= 0 {
			offset = o
		}
	}

	leads, err := s.db.ListLeads(limit, offset, tier)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(leads)
}

func (s *Server) handleGetLead(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}

	lead, err := s.db.GetLead(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if lead == nil {
		http.NotFound(w, r)
		return
	}

	syncLogs, _ := s.db.ListSyncLogsForLead(id)
	auditEvents, _ := s.db.ListAuditEventsForLead(id)

	response := map[string]interface{}{
		"lead":         lead,
		"sync_logs":    syncLogs,
		"audit_events": auditEvents,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

func (s *Server) handleUpdateDraft(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	lead, err := s.db.GetLead(id)
	if err != nil || lead == nil {
		http.NotFound(w, r)
		return
	}

	var req struct {
		Draft string `json:"draft"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	lead.OutreachDraft = req.Draft
	if err := s.db.UpdateLead(lead); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.sse.Broadcast("lead.draft_updated", lead)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(lead)
}

func (s *Server) handleResyncLead(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	lead, err := s.db.GetLead(id)
	if err != nil || lead == nil {
		http.NotFound(w, r)
		return
	}

	// Trigger pipeline re-dispatch
	if s.pipeline != nil {
		payload := &receiver.InboundPayload{
			Source:     lead.Source,
			Sender:     lead.Sender,
			Recipient:  lead.Recipient,
			Subject:    lead.Subject,
			Body:       lead.ParsedBody,
			RawPayload: lead.RawPayload,
		}
		updated, err := s.pipeline.Process(r.Context(), payload)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(updated)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(lead)
}

func (s *Server) handleListAccounts(w http.ResponseWriter, r *http.Request) {
	accounts, err := s.db.ListAllAccounts()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(accounts)
}

func (s *Server) handleSimulate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Preset string `json:"preset"` // "fintech", "saas", "hipaa", "freemail", "spam"
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	var payload *receiver.InboundPayload

	switch req.Preset {
	case "saas":
		payload = &receiver.InboundPayload{
			Source:    "simulation_saas",
			Sender:    "Elena Rostova <elena@cloudscale-analytics.io>",
			Recipient: "leads@formpilot.io",
			Subject:   "Demo request for 45 seats - CloudScale Analytics",
			Body: `Hello, My name is Elena Rostova and I head the Product Operations team at CloudScale Analytics (cloudscale-analytics.io).
We are looking for HubSpot CRM integration for 45 seats.
Our budget is around $35,000 annually. Timeline is 2 months.
Phone: +1 650 443 9081`,
		}

	case "hipaa":
		payload = &receiver.InboundPayload{
			Source:    "simulation_hipaa",
			Sender:    "Dr. Jonathan Hayes <jhayes@stjudehealthsystem.org>",
			Recipient: "rfq@formpilot.io",
			Subject:   "HIPAA Compliant Intake Automation - St Jude Health System",
			Body: `Dear Enterprise Inquiries,
I am Dr. Jonathan Hayes, Chief Medical Information Officer at St Jude Health System (stjudehealthsystem.org).
Scope: 500 clinical coordinators.
Compliance: Strict HIPAA BAA required. Integration with Salesforce Health Cloud.
Budget: Board approved $300,000.
Urgency: P0 immediate deployment.`,
		}

	case "freemail":
		payload = &receiver.InboundPayload{
			Source:    "simulation_freemail",
			Sender:    "alex.dev99@gmail.com",
			Recipient: "info@formpilot.io",
			Subject:   "Question about features",
			Body:      "Hey, just testing if this works for a personal project with 2 friends. Thanks, Alex",
		}

	case "spam":
		payload = &receiver.InboundPayload{
			Source:    "simulation_spam",
			Sender:    "blast@tempmail.com",
			Recipient: "contact@formpilot.io",
			Subject:   "Cheap SEO 5000 Backlinks",
			Body:      "Claim 5000 high PR backlinks for $50. Unsubscribe by replying.",
		}

	default: // "fintech"
		payload = &receiver.InboundPayload{
			Source:    "simulation_fintech",
			Sender:    "Marcus Vance <m.vance@apexclearing.com>",
			Recipient: "sales@formpilot.io",
			Subject:   "Enterprise RFP: Inbound Data Pipeline for 250 FinTech Seats",
			Body: `Hi Sales Team,
I am Marcus Vance, VP of Engineering at Apex Clearing (apexclearing.com).
We need 250 enterprise seats for operations.
Requirements: Automated entity extraction, custom CRM integration, and audit logs.
Security: SOC2 Type II, dedicated VPC, SAML SSO with Okta.
Budget: Allocated $120,000 annual licensing.
Timeline: Immediate Q4 initiative; procurement within 30 days.
Phone: +1 (415) 890-3421`,
		}
	}

	payload.RawPayload = payload.Body

	if s.pipeline == nil {
		http.Error(w, "pipeline not configured", http.StatusInternalServerError)
		return
	}

	lead, err := s.pipeline.Process(r.Context(), payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(lead)
}

func (s *Server) handleSendGridWebhook(w http.ResponseWriter, r *http.Request) {
	payload, err := receiver.ParseSendGridWebhook(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	s.processWebhook(w, r, payload)
}

func (s *Server) handlePostmarkWebhook(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	payload, err := receiver.ParsePostmarkWebhook(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	s.processWebhook(w, r, payload)
}

func (s *Server) handleMailgunWebhook(w http.ResponseWriter, r *http.Request) {
	payload, err := receiver.ParseMailgunWebhook(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	s.processWebhook(w, r, payload)
}

func (s *Server) handleSESWebhook(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	payload, err := receiver.ParseSESWebhook(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	s.processWebhook(w, r, payload)
}

func (s *Server) handleGenericWebhook(w http.ResponseWriter, r *http.Request) {
	payload, err := receiver.ParseGenericWebhook(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	s.processWebhook(w, r, payload)
}

func (s *Server) processWebhook(w http.ResponseWriter, r *http.Request, payload *receiver.InboundPayload) {
	if s.pipeline == nil {
		http.Error(w, "pipeline unavailable", http.StatusInternalServerError)
		return
	}

	lead, err := s.pipeline.Process(r.Context(), payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"status":     "processed",
		"lead_id":    lead.ID,
		"company":    lead.CompanyName,
		"icp_score":  lead.ICPScore,
		"tier":       lead.QualificationTier,
		"latency_ms": lead.LatencyMs,
	})
}
