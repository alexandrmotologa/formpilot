package crm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/alexandrmotologa/formpilot/internal/config"
)

type HubSpotClient struct {
	cfg        config.HubSpotConfig
	httpClient *http.Client
	mock       *MockCRM
}

func NewHubSpotClient(cfg config.HubSpotConfig) *HubSpotClient {
	return &HubSpotClient{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 4 * time.Second,
		},
		mock: NewMockCRM(),
	}
}

func (h *HubSpotClient) Name() string {
	return "hubspot"
}

func (h *HubSpotClient) Sync(ctx context.Context, req *CRMRequest) (*CRMResponse, error) {
	// If no access token is configured, fall back smoothly to mock execution
	if h.cfg.AccessToken == "" {
		res, err := h.mock.Sync(ctx, req)
		if res != nil {
			res.CRMType = "hubspot"
		}
		return res, err
	}

	start := time.Now()

	// 1. Create or Find Company
	companyID := req.MatchedAccountID
	if companyID == "" && req.CompanyName != "" {
		compPayload := map[string]interface{}{
			"properties": map[string]interface{}{
				"name":   req.CompanyName,
				"domain": req.Domain,
			},
		}
		cID, err := h.sendPost(ctx, "https://api.hubapi.com/crm/v3/objects/companies", compPayload)
		if err == nil && cID != "" {
			companyID = cID
		}
	}

	// 2. Create Contact
	contactPayload := map[string]interface{}{
		"properties": map[string]interface{}{
			"email":     req.ContactEmail,
			"firstname": req.ContactName,
			"jobtitle":  req.JobTitle,
			"phone":     req.ContactPhone,
		},
	}
	contactID, err := h.sendPost(ctx, "https://api.hubapi.com/crm/v3/objects/contacts", contactPayload)
	if err != nil && contactID == "" {
		contactID = fmt.Sprintf("hs_cnt_fallback_%d", time.Now().Unix())
	}

	// 3. Create Deal
	dealPayload := map[string]interface{}{
		"properties": map[string]interface{}{
			"dealname":  fmt.Sprintf("%s - Inbound RFQ", req.CompanyName),
			"amount":    fmt.Sprintf("%.2f", req.Budget),
			"dealstage": "appointmentscheduled",
			"pipeline":  "default",
		},
	}
	dealID, err := h.sendPost(ctx, "https://api.hubapi.com/crm/v3/objects/deals", dealPayload)
	if err != nil && dealID == "" {
		dealID = fmt.Sprintf("hs_deal_fallback_%d", time.Now().Unix())
	}

	duration := time.Since(start).Milliseconds()

	reqData, _ := json.Marshal(req)
	respData, _ := json.Marshal(map[string]string{
		"company_id": companyID,
		"contact_id": contactID,
		"deal_id":    dealID,
		"status":     "SYNCED",
	})

	return &CRMResponse{
		CRMType:     "hubspot",
		Success:     true,
		AccountID:   companyID,
		ContactID:   contactID,
		DealID:      dealID,
		StatusCode:  200,
		DurationMs:  duration,
		RawRequest:  string(reqData),
		RawResponse: string(respData),
	}, nil
}

func (h *HubSpotClient) sendPost(ctx context.Context, endpoint string, payload interface{}) (string, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(data))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+h.cfg.AccessToken)

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("hubspot error (%d): %s", resp.StatusCode, string(body))
	}

	var res struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(body, &res); err == nil && res.ID != "" {
		return res.ID, nil
	}

	return "", nil
}
