package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/alexandrmotologa/formpilot/internal/config"
	"github.com/alexandrmotologa/formpilot/internal/extractor"
	"github.com/alexandrmotologa/formpilot/internal/scorer"
)

type SlackNotifier struct {
	cfg        config.SlackConfig
	httpClient *http.Client
}

func NewSlackNotifier(cfg config.SlackConfig) *SlackNotifier {
	return &SlackNotifier{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 4 * time.Second,
		},
	}
}

// SendLeadAlert sends a rich Slack Block Kit notification for high-priority inbound leads.
func (s *SlackNotifier) SendLeadAlert(ctx context.Context, lead *extractor.ExtractedLead, icp *scorer.ICPResult, leadID string) error {
	if !s.cfg.Enabled || s.cfg.WebhookURL == "" {
		return nil
	}

	payload := map[string]interface{}{
		"text": fmt.Sprintf("New Inbound Lead: %s (%s) - Score %d/100", lead.CompanyName, lead.Domain, icp.TotalScore),
		"blocks": []map[string]interface{}{
			{
				"type": "header",
				"text": map[string]string{
					"type": "plain_text",
					"text": fmt.Sprintf("🚨 Inbound Lead Alert: %s [%d/100]", lead.CompanyName, icp.TotalScore),
				},
			},
			{
				"type": "section",
				"fields": []map[string]string{
					{"type": "mrkdwn", "text": fmt.Sprintf("*Company:*\n%s (%s)", lead.CompanyName, lead.Domain)},
					{"type": "mrkdwn", "text": fmt.Sprintf("*Contact:*\n%s (%s)", lead.ContactName, lead.JobTitle)},
					{"type": "mrkdwn", "text": fmt.Sprintf("*Volume:*\n%d seats", lead.RequestedSeats)},
					{"type": "mrkdwn", "text": fmt.Sprintf("*Budget:*\n$%.0f", lead.Budget)},
					{"type": "mrkdwn", "text": fmt.Sprintf("*Urgency:*\n%s", lead.Urgency)},
					{"type": "mrkdwn", "text": fmt.Sprintf("*Tier:*\n%s", icp.Tier)},
				},
			},
			{
				"type": "context",
				"elements": []map[string]string{
					{"type": "mrkdwn", "text": fmt.Sprintf("Lead ID: `%s` | formpilot autonomous ingestion engine", leadID)},
				},
			},
		},
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.cfg.WebhookURL, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("slack webhook error: status %d", resp.StatusCode)
	}

	return nil
}
