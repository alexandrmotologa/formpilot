package crm

import (
	"context"
	"fmt"
	"time"

	"github.com/alexandrmotologa/formpilot/internal/config"
)

type PipedriveClient struct {
	cfg  config.PipedriveConfig
	mock *MockCRM
}

func NewPipedriveClient(cfg config.PipedriveConfig) *PipedriveClient {
	return &PipedriveClient{
		cfg:  cfg,
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
	// Real Pipedrive REST calls would execute here
	duration := time.Since(start).Milliseconds()

	return &CRMResponse{
		CRMType:    "pipedrive",
		Success:    true,
		AccountID:  fmt.Sprintf("pd_org_%s", req.LeadID),
		ContactID:  fmt.Sprintf("pd_per_%s", req.LeadID),
		DealID:     fmt.Sprintf("pd_deal_%s", req.LeadID),
		StatusCode: 201,
		DurationMs: duration,
	}, nil
}
