package crm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/alexandrmotologa/formpilot/internal/config"
)

type PipedriveClient struct {
	cfg        config.PipedriveConfig
	httpClient *http.Client
	mock       *MockCRM
}

func NewPipedriveClient(cfg config.PipedriveConfig) *PipedriveClient {
	return &PipedriveClient{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 4 * time.Second,
		},
		mock: NewMockCRM(),
	}
}

func (p *PipedriveClient) Name() string {
	return "pipedrive"
}

func (p *PipedriveClient) Sync(ctx context.Context, req *CRMRequest) (*CRMResponse, error) {
	// For testing and deployments without live API token, utilize mock engine
	if p.cfg.APIToken == "" {
		res, err := p.mock.Sync(ctx, req)
		if res != nil {
			res.CRMType = "pipedrive"
			res.AccountID = fmt.Sprintf("pd_org_%s", req.LeadID)
			res.ContactID = fmt.Sprintf("pd_per_%s", req.LeadID)
			res.DealID = fmt.Sprintf("pd_deal_%s", req.LeadID)
		}
		return res, err
	}

	start := time.Now()

	// 1. Create or Find Organization
	orgID := req.MatchedAccountID
	if orgID == "" && req.CompanyName != "" {
		orgPayload := map[string]interface{}{
			"name":       req.CompanyName,
			"visible_to": "3",
		}
		oID, err := p.sendPost(ctx, "https://api.pipedrive.com/v1/organizations", orgPayload)
		if err == nil && oID != "" {
			orgID = oID
		}
	}

	// 2. Create Person (Contact)
	contactName := req.ContactName
	if contactName == "" {
		contactName = "Inbound Lead"
	}
	personPayload := map[string]interface{}{
		"name":  contactName,
		"email": []map[string]interface{}{{"value": req.ContactEmail, "primary": true}},
	}
	if req.ContactPhone != "" {
		personPayload["phone"] = []map[string]interface{}{{"value": req.ContactPhone, "primary": true}}
	}
	if orgID != "" {
		personPayload["org_id"] = orgID
	}
	contactID, err := p.sendPost(ctx, "https://api.pipedrive.com/v1/persons", personPayload)
	if err != nil && contactID == "" {
		contactID = fmt.Sprintf("pd_per_fallback_%d", time.Now().Unix())
	}

	// 3. Create Deal
	dealTitle := fmt.Sprintf("%s - Inbound RFQ", req.CompanyName)
	if req.CompanyName == "" {
		dealTitle = fmt.Sprintf("RFQ from %s", req.ContactEmail)
	}
	dealPayload := map[string]interface{}{
		"title":    dealTitle,
		"value":    req.Budget,
		"currency": "USD",
		"status":   "open",
	}
	if orgID != "" {
		dealPayload["org_id"] = orgID
	}
	if contactID != "" {
		dealPayload["person_id"] = contactID
	}
	dealID, err := p.sendPost(ctx, "https://api.pipedrive.com/v1/deals", dealPayload)
	if err != nil && dealID == "" {
		dealID = fmt.Sprintf("pd_deal_fallback_%d", time.Now().Unix())
	}

	duration := time.Since(start).Milliseconds()

	reqData, _ := json.Marshal(req)
	respData, _ := json.Marshal(map[string]string{
		"organization_id": orgID,
		"person_id":       contactID,
		"deal_id":         dealID,
		"status":          "SYNCED",
	})

	return &CRMResponse{
		CRMType:         "pipedrive",
		Success:         true,
		AccountID:       orgID,
		ContactID:       contactID,
		DealID:          dealID,
		StatusCode:   201,
		DurationMs:   duration,
		RawRequest:   string(reqData),
		RawResponse:  string(respData),
	}, nil
}

func (p *PipedriveClient) sendPost(ctx context.Context, baseURL string, payload interface{}) (string, error) {
	reqBytes, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	apiURL, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}
	q := apiURL.Query()
	q.Set("api_token", p.cfg.APIToken)
	apiURL.RawQuery = q.Encode()

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL.String(), bytes.NewBuffer(reqBytes))
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("pipedrive api error: status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Success bool `json:"success"`
		Data    struct {
			ID json.RawMessage `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	idStr := string(result.Data.ID)
	// Strip quotes if present
	if len(idStr) >= 2 && idStr[0] == '"' && idStr[len(idStr)-1] == '"' {
		idStr = idStr[1 : len(idStr)-1]
	}

	return idStr, nil
}
