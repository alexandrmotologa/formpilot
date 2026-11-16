package extractor

import (
	"context"
	"testing"

	"github.com/alexandrmotologa/formpilot/internal/config"
)

func TestExtractFinTechRFQ(t *testing.T) {
	sender := "Marcus Vance <m.vance@apexclearing.com>"
	subject := "Enterprise RFP: Inbound Data Pipeline for 250 FinTech Seats"
	body := `Hi Sales Team,
I am Marcus Vance, VP of Engineering at Apex Clearing (apexclearing.com).
We need 250 enterprise seats for our operations.
Allocated budget is $120,000 for annual licensing.
Timeline: This is an immediate Q4 initiative; procurement within 30 days.
Phone: +1 (415) 890-3421
`

	ext := NewExtractor(config.ExtractionConfig{Provider: "heuristic"})
	lead := ext.Extract(context.Background(), sender, subject, body)

	if lead.CompanyName != "Apex Clearing" {
		t.Errorf("expected Apex Clearing, got %q", lead.CompanyName)
	}
	if lead.Domain != "apexclearing.com" {
		t.Errorf("expected apexclearing.com, got %q", lead.Domain)
	}
	if lead.ContactName != "Marcus Vance" {
		t.Errorf("expected Marcus Vance, got %q", lead.ContactName)
	}
	if lead.RequestedSeats != 250 {
		t.Errorf("expected 250 seats, got %d", lead.RequestedSeats)
	}
	if lead.Budget != 120000 {
		t.Errorf("expected budget 120000, got %f", lead.Budget)
	}
	if lead.Seniority != "C-Level / VP / Founder" {
		t.Errorf("expected C-Level / VP / Founder, got %q", lead.Seniority)
	}
	if lead.Urgency != "P0 Immediate (< 1 mo)" {
		t.Errorf("expected P0 Immediate (< 1 mo), got %q", lead.Urgency)
	}
	if lead.ContactPhone != "+14158903421" {
		t.Errorf("expected +14158903421, got %q", lead.ContactPhone)
	}
}

func TestExtractHealthcareRFQ(t *testing.T) {
	sender := "Dr. Jonathan Hayes <jhayes@stjudehealthsystem.org>"
	subject := "HIPAA Compliant Intake Automation - St Jude Health System"
	body := `Dear Enterprise Inquiries,
I am Dr. Jonathan Hayes, Chief Medical Information Officer at St Jude Health System.
Estimated volume: 500 clinical coordinators.
Budget: Board approved $300,000 capital expenditure.
Compliance: Strict HIPAA BAA.
`
	ext := NewExtractor(config.ExtractionConfig{Provider: "heuristic"})
	lead := ext.Extract(context.Background(), sender, subject, body)

	if lead.Domain != "stjudehealthsystem.org" {
		t.Errorf("expected stjudehealthsystem.org, got %q", lead.Domain)
	}
	if lead.RequestedSeats != 500 {
		t.Errorf("expected 500 seats, got %d", lead.RequestedSeats)
	}
	if lead.Budget != 300000 {
		t.Errorf("expected 300000 budget, got %f", lead.Budget)
	}
	if lead.Seniority != "C-Level / VP / Founder" {
		t.Errorf("expected C-Level / VP / Founder, got %q", lead.Seniority)
	}
}

func TestValidationAndNormalization(t *testing.T) {
	if !IsFreeWebmail("gmail.com") {
		t.Errorf("expected gmail.com to be free webmail")
	}
	if !IsDisposableEmail("tempmail.com") {
		t.Errorf("expected tempmail.com to be disposable")
	}
	if IsFreeWebmail("apexclearing.com") {
		t.Errorf("expected apexclearing.com to NOT be free webmail")
	}

	norm := NormalizeDomain("https://www.cyberdyne.systems/quote/")
	if norm != "cyberdyne.systems" {
		t.Errorf("expected cyberdyne.systems, got %q", norm)
	}

	phone := NormalizePhone("(415) 555-0199")
	if phone != "+14155550199" {
		t.Errorf("expected +14155550199, got %q", phone)
	}
}
