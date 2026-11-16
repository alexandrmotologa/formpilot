package scorer

import (
	"strings"
	"testing"

	"github.com/alexandrmotologa/formpilot/internal/config"
	"github.com/alexandrmotologa/formpilot/internal/extractor"
)

func TestCalculateICP_EnterpriseHotLead(t *testing.T) {
	weights := config.DefaultConfig().ICP.Weights

	lead := &extractor.ExtractedLead{
		CompanyName:       "Apex Clearing",
		Domain:            "apexclearing.com",
		ContactName:       "Marcus Vance",
		JobTitle:          "VP of Engineering",
		Seniority:         "C-Level / VP / Founder",
		ContactEmail:      "m.vance@apexclearing.com",
		RequestedSeats:    250,
		Budget:            120000,
		Urgency:           "P0 Immediate (< 1 mo)",
		TechStack:         []string{"SOC2 Type II", "SAML SSO / Okta"},
		RequestedProducts: []string{"CRM Ingestion Pipeline"},
	}

	result := CalculateICP(lead, weights)
	if result.TotalScore < 80 {
		t.Errorf("expected score >= 80 for enterprise rfq, got %d", result.TotalScore)
	}
	if result.Tier != "Enterprise Hot Lead" {
		t.Errorf("expected Enterprise Hot Lead, got %s", result.Tier)
	}
	if len(result.PositiveFactors) < 3 {
		t.Errorf("expected at least 3 positive factors, got %d", len(result.PositiveFactors))
	}
}

func TestCalculateICP_DisposableSpam(t *testing.T) {
	weights := config.DefaultConfig().ICP.Weights

	lead := &extractor.ExtractedLead{
		CompanyName:  "SEO Blast",
		Domain:       "tempmail.com",
		ContactEmail: "user@tempmail.com",
		Urgency:      "P3 Exploratory (> 6 mos)",
	}

	result := CalculateICP(lead, weights)
	if result.TotalScore > 20 {
		t.Errorf("expected low score for tempmail, got %d", result.TotalScore)
	}
	if result.Tier != "Unqualified / Low Intent" {
		t.Errorf("expected Unqualified / Low Intent, got %s", result.Tier)
	}
}

func TestGenerateOutreachDraft(t *testing.T) {
	lead := &extractor.ExtractedLead{
		CompanyName:       "CloudScale Analytics",
		Domain:            "cloudscale.io",
		ContactName:       "Elena Rostova",
		JobTitle:          "Head of Product Operations",
		Seniority:         "Director / Head",
		RequestedSeats:    45,
		RequestedProducts: []string{"HubSpot CRM Ingestion"},
		Urgency:           "P1 High (1-3 mos)",
	}

	draft := GenerateOutreachDraft(lead, "Sarah Connor", "https://cal.com/formpilot/demo", ToneExecutiveDirect)
	if !strings.Contains(draft, "Hi Elena") {
		t.Errorf("expected draft to start with 'Hi Elena', got:\n%s", draft)
	}
	if !strings.Contains(draft, "CloudScale Analytics") {
		t.Errorf("expected draft to mention company name, got:\n%s", draft)
	}
	if !strings.Contains(draft, "45 enterprise seats") {
		t.Errorf("expected draft to mention 45 seats, got:\n%s", draft)
	}
	if !strings.Contains(draft, "https://cal.com/formpilot/demo") {
		t.Errorf("expected calendar link in draft, got:\n%s", draft)
	}
}
