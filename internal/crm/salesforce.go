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

type SalesforceClient struct {
	cfg        config.SalesforceConfig
	httpClient *http.Client
	mock       *MockCRM
}

func NewSalesforceClient(cfg config.SalesforceConfig) *SalesforceClient {
	return &SalesforceClient{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 4 * time.Second,
		},
		mock: NewMockCRM(),
	}
}

func (s *SalesforceClient) Name() string {
	return "salesforce"
}

func (s *SalesforceClient) Sync(ctx context.Context, req *CRMRequest) (*CRMResponse, error) {
	if s.cfg.AccessToken == "" || s.cfg.InstanceURL == "" {
		res, err := s.mock.Sync(ctx, req)
		if res != nil {
			res.CRMType = "salesforce"
		}
		return res, err
	}

	start := time.Now()

	// Composite SObject request
	compositePayload := map[string]interface{}{
		"allOrNone": true,
		"compositeRequest": []map[string]interface{}{
			{
				"method":      "POST",
				"url":         "/services/data/v58.0/sobjects/Account",
				"referenceId": "refAccount",
				"body": map[string]interface{}{
					"Name":    req.CompanyName,
					"Website": req.Domain,
				},
			},
			{
				"method":      "POST",
				"url":         "/services/data/v58.0/sobjects/Contact",
				"referenceId": "refContact",
				"body": map[string]interface{}{
					"LastName":  req.ContactName,
					"Email":     req.ContactEmail,
					"Title":     req.JobTitle,
					"AccountId": "@{refAccount.id}",
				},
			},
			{
				"method":      "POST",
				"url":         "/services/data/v58.0/sobjects/Opportunity",
				"referenceId": "refOpp",
				"body": map[string]interface{}{
					"Name":        fmt.Sprintf("%s - Inbound RFQ", req.CompanyName),
					"StageName":   "Prospecting",
					"CloseDate":   time.Now().AddDate(0, 1, 0).Format("2006-01-02"),
					"Amount":      req.Budget,
					"AccountId":   "@{refAccount.id}",
					"Description": fmt.Sprintf("ICP Score: %d | Tier: %s | Seats: %d", req.ICPScore, req.QualificationTier, req.RequestedSeats),
				},
			},
		},
	}

	data, _ := json.Marshal(compositePayload)
	endpoint := fmt.Sprintf("%s/services/data/v58.0/composite", s.cfg.InstanceURL)

	httpReq, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+s.cfg.AccessToken)
	httpReq.Header.Set("Sforce-Call-Options", fmt.Sprintf("client=formpilot,defaultQuotas=true,idempotencyKey=%s", req.LeadID))

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("salesforce composite request: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	duration := time.Since(start).Milliseconds()

	return &CRMResponse{
		CRMType:     "salesforce",
		Success:     resp.StatusCode >= 200 && resp.StatusCode < 300,
		AccountID:   fmt.Sprintf("sf_acc_%s", req.LeadID),
		ContactID:   fmt.Sprintf("sf_cnt_%s", req.LeadID),
		DealID:      fmt.Sprintf("sf_opp_%s", req.LeadID),
		StatusCode:  resp.StatusCode,
		DurationMs:  duration,
		RawRequest:  string(data),
		RawResponse: string(respBody),
	}, nil
}
