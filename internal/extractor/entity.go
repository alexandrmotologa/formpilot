package extractor

import (
	"context"
	"time"

	"github.com/alexandrmotologa/formpilot/internal/config"
)

type Extractor struct {
	cfg config.ExtractionConfig
}

func NewExtractor(cfg config.ExtractionConfig) *Extractor {
	return &Extractor{cfg: cfg}
}

// Extract processes raw inbound lead text and returns structured B2B lead entities.
func (e *Extractor) Extract(ctx context.Context, sender, subject, body string) *ExtractedLead {
	// If provider is set to "llm" and API key is present, attempt LLM extraction
	if e.cfg.Provider == "llm" && e.cfg.APIKey != "" {
		llmCtx, cancel := context.WithTimeout(ctx, 4*time.Second)
		defer cancel()

		llmCfg := LLMConfig{
			Provider: e.cfg.Provider,
			Model:    e.cfg.Model,
			APIKey:   e.cfg.APIKey,
			Endpoint: e.cfg.Endpoint,
		}

		if lead, err := ExtractWithLLM(llmCtx, llmCfg, sender, subject, body); err == nil && lead != nil {
			lead.Domain = NormalizeDomain(lead.Domain)
			if lead.ContactPhone != "" {
				lead.ContactPhone = NormalizePhone(lead.ContactPhone)
			}
			return lead
		}
	}

	// Deterministic heuristic fallback
	lead := ExtractHeuristic(sender, subject, body)
	lead.Domain = NormalizeDomain(lead.Domain)
	if lead.ContactPhone != "" {
		lead.ContactPhone = NormalizePhone(lead.ContactPhone)
	}

	return lead
}
