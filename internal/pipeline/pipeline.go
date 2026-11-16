package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/alexandrmotologa/formpilot/internal/config"
	"github.com/alexandrmotologa/formpilot/internal/crm"
	"github.com/alexandrmotologa/formpilot/internal/dedup"
	"github.com/alexandrmotologa/formpilot/internal/extractor"
	"github.com/alexandrmotologa/formpilot/internal/notify"
	"github.com/alexandrmotologa/formpilot/internal/receiver"
	"github.com/alexandrmotologa/formpilot/internal/scorer"
	"github.com/alexandrmotologa/formpilot/internal/server"
	"github.com/alexandrmotologa/formpilot/internal/storage"
)

type Pipeline struct {
	cfg        *config.Config
	db         *storage.DB
	extractor  *extractor.Extractor
	matcher    *dedup.Matcher
	cache      *dedup.AccountCache
	dispatcher *crm.Dispatcher
	slack      *notify.SlackNotifier
	sse        *server.SSEBroadcaster
}

func NewPipeline(
	cfg *config.Config,
	db *storage.DB,
	cache *dedup.AccountCache,
	broadcaster *server.SSEBroadcaster,
) *Pipeline {
	ext := extractor.NewExtractor(cfg.Extraction)
	mat := dedup.NewMatcher(cache)
	disp := crm.NewDispatcher(cfg.CRM, db)
	slk := notify.NewSlackNotifier(cfg.Slack)

	return &Pipeline{
		cfg:        cfg,
		db:         db,
		extractor:  ext,
		matcher:    mat,
		cache:      cache,
		dispatcher: disp,
		slack:      slk,
		sse:        broadcaster,
	}
}

// Process executes the complete autonomous unstructured-to-CRM ingestion pipeline.
func (p *Pipeline) Process(ctx context.Context, payload *receiver.InboundPayload) (*storage.Lead, error) {
	start := time.Now()
	leadID := fmt.Sprintf("lead_%d", time.Now().UnixNano())

	// Step 1: Entity Extraction
	extracted := p.extractor.Extract(ctx, payload.Sender, payload.Subject, payload.Body)

	// Step 2: Deduplication & Account Resolution
	dedupResult := p.matcher.Match(extracted)
	dedupStatus := "new_account"
	matchedID := ""
	matchedName := ""
	matchedOwner := ""

	if dedupResult.IsMatched && dedupResult.MatchedAccount != nil {
		dedupStatus = "existing_matched"
		matchedID = dedupResult.MatchedAccount.CRMID
		matchedName = dedupResult.MatchedAccount.CompanyName
		matchedOwner = dedupResult.MatchedAccount.OwnerName
	}

	// Step 3: ICP Fit Scoring
	icp := scorer.CalculateICP(extracted, p.cfg.ICP.Weights)

	// Step 4: Outreach Draft Synthesis
	draft := scorer.GenerateOutreachDraft(extracted, "", "", scorer.ToneExecutiveDirect)

	// Step 5: Multi-CRM Dispatch
	crmReq := &crm.CRMRequest{
		LeadID:            leadID,
		CompanyName:       extracted.CompanyName,
		Domain:            extracted.Domain,
		ContactName:       extracted.ContactName,
		ContactEmail:      extracted.ContactEmail,
		ContactPhone:      extracted.ContactPhone,
		JobTitle:          extracted.JobTitle,
		RequestedSeats:    extracted.RequestedSeats,
		Budget:            extracted.Budget,
		Urgency:           extracted.Urgency,
		ICPScore:          icp.TotalScore,
		QualificationTier: icp.Tier,
		DedupStatus:       dedupStatus,
		MatchedAccountID:  matchedID,
		OutreachDraft:     draft,
	}

	responses, err := p.dispatcher.Dispatch(ctx, crmReq)
	syncSuccess := true
	if err != nil {
		syncSuccess = false
	} else {
		for _, resp := range responses {
			if !resp.Success {
				syncSuccess = false
				break
			}
		}
	}

	// Step 6: Slack Emergency Notification for Enterprise Leads
	if icp.Tier == "Enterprise Hot Lead" {
		_ = p.slack.SendLeadAlert(ctx, extracted, icp, leadID)
	}

	// Step 7: Update Account Cache if new account provisioned
	if dedupStatus == "new_account" && extracted.Domain != "" && !extractor.IsFreeWebmail(extracted.Domain) {
		crmID := "mock_acc"
		if len(responses) > 0 && responses[0].AccountID != "" {
			crmID = responses[0].AccountID
		}
		_ = p.cache.Upsert(&storage.AccountRecord{
			ID:          fmt.Sprintf("acc_%s", extracted.Domain),
			Domain:      extracted.Domain,
			CompanyName: extracted.CompanyName,
			OwnerName:   "Enterprise AE Queue",
			OwnerEmail:  "sales@formpilot.io",
			CRMID:       crmID,
			CRMType:     "hubspot",
		})
	}

	// Step 8: Build Lead Record & Persist
	latency := time.Since(start).Milliseconds()

	status := "synced"
	if !syncSuccess {
		status = "failed"
	}

	extractedJSON, _ := json.Marshal(extracted)

	lead := &storage.Lead{
		ID:                  leadID,
		ReceivedAt:          time.Now().UTC(),
		Source:              payload.Source,
		Sender:              payload.Sender,
		Recipient:           payload.Recipient,
		Subject:             payload.Subject,
		RawPayload:          payload.RawPayload,
		ParsedBody:          payload.Body,
		ExtractedEntities:   string(extractedJSON),
		CompanyName:         extracted.CompanyName,
		Domain:              extracted.Domain,
		ContactName:         extracted.ContactName,
		ContactEmail:        extracted.ContactEmail,
		ContactPhone:        extracted.ContactPhone,
		RequestedSeats:      extracted.RequestedSeats,
		Budget:              extracted.Budget,
		Urgency:             extracted.Urgency,
		ICPScore:            icp.TotalScore,
		QualificationTier:   icp.Tier,
		DedupStatus:         dedupStatus,
		MatchedAccountID:    matchedID,
		MatchedAccountName:  matchedName,
		MatchedAccountOwner: matchedOwner,
		OutreachDraft:       draft,
		Status:              status,
		LatencyMs:           latency,
	}

	if p.db != nil {
		_ = p.db.InsertLead(lead)
		_ = p.db.InsertAuditEvent(&storage.AuditEvent{
			ID:      fmt.Sprintf("aud_%d", time.Now().UnixNano()),
			LeadID:  leadID,
			Stage:   "pipeline",
			Event:   "lead_processed",
			Details: fmt.Sprintf("Extracted %s, Score: %d, Tier: %s, Latency: %dms", lead.CompanyName, lead.ICPScore, lead.QualificationTier, latency),
		})
	}

	// Step 9: Broadcast real-time SSE event
	if p.sse != nil {
		p.sse.Broadcast("lead.processed", lead)
	}

	return lead, nil
}
