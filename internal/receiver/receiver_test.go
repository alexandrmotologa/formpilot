package receiver

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestParseRawEmail(t *testing.T) {
	raw := `From: Marcus Vance <marcus@apexclearing.com>
To: sales@formpilot.io
Subject: Enterprise RFP 250 Seats
Date: Tue, 17 Nov 2026 09:15:00 +0000
Content-Type: text/plain; charset="utf-8"

Hi team, we need 250 enterprise seats with SOC2 compliance.
Budget is $120,000.
`
	parsed, err := ParseRawEmail([]byte(raw))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(parsed.Sender, "marcus@apexclearing.com") {
		t.Errorf("expected sender to contain marcus@apexclearing.com, got %q", parsed.Sender)
	}
	if parsed.Subject != "Enterprise RFP 250 Seats" {
		t.Errorf("expected subject 'Enterprise RFP 250 Seats', got %q", parsed.Subject)
	}
	if !strings.Contains(parsed.TextBody, "250 enterprise seats") {
		t.Errorf("expected body to contain '250 enterprise seats', got %q", parsed.TextBody)
	}
}

func TestStripHTML(t *testing.T) {
	html := `<div>Hello <b>World</b>!</div><p>We are interested in your &quot;Enterprise&quot; plan.<br/>Contact us soon.</p>`
	clean := StripHTML(html)
	if !strings.Contains(clean, `Hello World!`) {
		t.Errorf("expected 'Hello World!', got %q", clean)
	}
	if !strings.Contains(clean, `"Enterprise"`) {
		t.Errorf("expected decoded quotes, got %q", clean)
	}
}

func TestPostmarkWebhook(t *testing.T) {
	payload := `{
		"From": "elena@cloudscale.io",
		"To": "leads@formpilot.io",
		"Subject": "Demo request 45 seats",
		"TextBody": "We are looking for a demo for 45 seats.",
		"HtmlBody": "<p>We are looking for a demo for 45 seats.</p>"
	}`

	inbound, err := ParsePostmarkWebhook([]byte(payload))
	if err != nil {
		t.Fatalf("postmark parse error: %v", err)
	}
	if inbound.Source != "postmark" {
		t.Errorf("expected source postmark, got %s", inbound.Source)
	}
	if inbound.Sender != "elena@cloudscale.io" {
		t.Errorf("expected sender elena@cloudscale.io, got %s", inbound.Sender)
	}
	if !strings.Contains(inbound.Body, "45 seats") {
		t.Errorf("expected body to have '45 seats', got %s", inbound.Body)
	}
}

func TestPDFExtraction(t *testing.T) {
	// Construct a synthetic minimal PDF with a compressed stream containing Tj text
	var streamBuf bytes.Buffer
	zw := zlib.NewWriter(&streamBuf)
	_, _ = zw.Write([]byte("BT\n/F1 12 Tf\n(Enterprise RFQ 500 Seats) Tj\nET"))
	_ = zw.Close()

	compressed := streamBuf.Bytes()

	pdfContent := fmt.Sprintf("%%PDF-1.4\n1 0 obj\n<< /Length %d /Filter /FlateDecode >>\nstream\n%s\nendstream\nendobj\nxref\ntrailer\n<< /Root 1 0 R >>\n%%%%EOF", len(compressed), string(compressed))

	text, err := ExtractTextFromPDF([]byte(pdfContent))
	if err != nil {
		t.Fatalf("pdf extract failed: %v", err)
	}
	if !strings.Contains(text, "Enterprise RFQ 500 Seats") {
		t.Errorf("expected text 'Enterprise RFQ 500 Seats', got %q", text)
	}
}

func TestGenericWebhook(t *testing.T) {
	req := httptest.NewRequest("POST", "/api/v1/webhooks/generic", strings.NewReader(`{"sender":"test@company.com","subject":"New inquiry","body":"We need a quote"}`))
	req.Header.Set("Content-Type", "application/json")

	payload, err := ParseGenericWebhook(req)
	if err != nil {
		t.Fatalf("generic webhook error: %v", err)
	}
	if payload.Sender != "test@company.com" {
		t.Errorf("expected test@company.com, got %s", payload.Sender)
	}
}
