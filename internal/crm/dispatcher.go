package crm

import (
	"context"
	"fmt"
	"time"

	"github.com/alexandrmotologa/formpilot/internal/config"
	"github.com/alexandrmotologa/formpilot/internal/storage"
)

type Dispatcher struct {
	clients []Client
	db      *storage.DB
}

func NewDispatcher(cfg config.CRMConfig, db *storage.DB) *Dispatcher {
	var clients []Client

	if cfg.Mode == "mock" {
		clients = append(clients, NewMockCRM())
	} else {
		if cfg.HubSpot.Enabled {
			clients = append(clients, NewHubSpotClient(cfg.HubSpot))
		}
		if cfg.Salesforce.Enabled {
			clients = append(clients, NewSalesforceClient(cfg.Salesforce))
		}
		if cfg.Pipedrive.Enabled {
			clients = append(clients, NewPipedriveClient(cfg.Pipedrive))
		}
		// Fallback to mock if none enabled
		if len(clients) == 0 {
			clients = append(clients, NewMockCRM())
		}
	}

	return &Dispatcher{
		clients: clients,
		db:      db,
	}
}

// Dispatch sends lead records to all active CRM endpoints and logs results.
func (d *Dispatcher) Dispatch(ctx context.Context, req *CRMRequest) ([]*CRMResponse, error) {
	var responses []*CRMResponse

	for _, client := range d.clients {
		resp, err := client.Sync(ctx, req)
		if err != nil {
			resp = &CRMResponse{
				CRMType:      client.Name(),
				Success:      false,
				ErrorMessage: err.Error(),
				StatusCode:   500,
			}
		}

		responses = append(responses, resp)

		// Log into SQLite
		if d.db != nil {
			logStatus := "success"
			if !resp.Success {
				logStatus = "failed"
			}
			_ = d.db.InsertSyncLog(&storage.SyncLog{
				ID:              fmt.Sprintf("log_%s_%s_%d", client.Name(), req.LeadID, time.Now().UnixNano()),
				LeadID:          req.LeadID,
				CRMType:         resp.CRMType,
				Action:          "sync_lead_objects",
				Status:          logStatus,
				StatusCode:      resp.StatusCode,
				RequestPayload:  resp.RawRequest,
				ResponsePayload: resp.RawResponse,
				DurationMs:      resp.DurationMs,
				ErrorMessage:    resp.ErrorMessage,
			})
		}
	}

	return responses, nil
}
