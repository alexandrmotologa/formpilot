package crm

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type MockCRM struct {
	mu      sync.RWMutex
	records map[string]*CRMRequest
}

func NewMockCRM() *MockCRM {
	return &MockCRM{
		records: make(map[string]*CRMRequest),
	}
}

func (m *MockCRM) Name() string {
	return "mock"
}

func (m *MockCRM) Sync(ctx context.Context, req *CRMRequest) (*CRMResponse, error) {
	start := time.Now()

	// Simulate brief network latency (30-80ms)
	time.Sleep(time.Duration(30+rand.Intn(50)) * time.Millisecond)

	m.mu.Lock()
	m.records[req.LeadID] = req
	m.mu.Unlock()

	reqData, _ := json.MarshalIndent(req, "", "  ")

	accountID := req.MatchedAccountID
	if accountID == "" {
		accountID = fmt.Sprintf("mock_acc_%06d", rand.Intn(999999))
	}
	contactID := fmt.Sprintf("mock_cnt_%06d", rand.Intn(999999))
	dealID := fmt.Sprintf("mock_opp_%06d", rand.Intn(999999))

	respPayload := map[string]interface{}{
		"status":     "PROVISIONED",
		"account_id": accountID,
		"contact_id": contactID,
		"deal_id":    dealID,
		"stage":      "Inbound Qualified",
		"synced_at":  time.Now().UTC().Format(time.RFC3339),
		"attributes": map[string]interface{}{
			"company":        req.CompanyName,
			"domain":         req.Domain,
			"seats":          req.RequestedSeats,
			"budget":         req.Budget,
			"icp_fit":        req.ICPScore,
			"classification": req.QualificationTier,
		},
	}
	respData, _ := json.MarshalIndent(respPayload, "", "  ")

	duration := time.Since(start).Milliseconds()

	return &CRMResponse{
		CRMType:     "mock",
		Success:     true,
		AccountID:   accountID,
		ContactID:   contactID,
		DealID:      dealID,
		StatusCode:  200,
		DurationMs:  duration,
		RawRequest:  string(reqData),
		RawResponse: string(respData),
	}, nil
}
