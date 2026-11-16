# API Reference

The formpilot daemon exposes HTTP endpoints for webhook ingestion, lead management, real-time streaming, configuration, and data exports.

## Webhook Endpoints

### 1. SendGrid Inbound Parse
- Method: `POST`
- Path: `/api/v1/webhooks/sendgrid`
- Content-Type: `multipart/form-data`
- Fields: `from`, `to`, `subject`, `text`, `html`, `attachments`

### 2. Postmark Inbound
- Method: `POST`
- Path: `/api/v1/webhooks/postmark`
- Content-Type: `application/json`
- Fields: `From`, `To`, `Subject`, `TextBody`, `HtmlBody`, `Attachments`

### 3. Mailgun Inbound
- Method: `POST`
- Path: `/api/v1/webhooks/mailgun`
- Content-Type: `multipart/form-data`
- Fields: `sender`, `recipient`, `subject`, `body-plain`, `body-html`, `attachment-count`

### 4. AWS SES SNS Notification
- Method: `POST`
- Path: `/api/v1/webhooks/ses`
- Content-Type: `application/json`
- Fields: `Type`, `Message` (JSON containing mail headers and body content)

### 5. Generic Lead Ingestion
- Method: `POST`
- Path: `/api/v1/webhooks/generic`
- Content-Type: `application/json` or `multipart/form-data`

### 6. Webhook Gateway Info & cURL Playground
- Method: `GET`
- Path: `/api/v1/webhooks/info`
- Response: Returns all active endpoints, supported formats, and ready-to-run cURL commands.

## Management Endpoints

### 1. List Leads
- Method: `GET`
- Path: `/api/v1/leads`
- Query Parameters: `limit`, `offset`, `tier`

### 2. Export Leads (CSV & JSON)
- Method: `GET`
- Path: `/api/v1/leads/export`
- Query Parameters: `format` (`csv` or `json`), `tier` (optional filter)
- Response: Downloadable RFC 4180 CSV file or JSON array.

### 3. Get Lead Detail
- Method: `GET`
- Path: `/api/v1/leads/{id}`
- Response: Detailed lead record including extracted entities, latency breakdown, CRM logs, and audit trail.

### 4. Update Outreach Draft
- Method: `PUT`
- Path: `/api/v1/leads/{id}/draft`
- Request: `{ "draft": "Updated text" }`

### 5. Regenerate Outreach Draft
- Method: `POST`
- Path: `/api/v1/leads/{id}/regenerate-draft`
- Request: `{ "tone": "executive_direct | consultative | technical_value", "rep_name": "...", "booking_link": "..." }`

### 6. 1-Click Send Outreach Email
- Method: `POST`
- Path: `/api/v1/leads/{id}/send-email`
- Response: Updates status to `contacted`, inserts audit trail, and broadcasts SSE event.

### 7. Override Account Deduplication
- Method: `POST`
- Path: `/api/v1/leads/{id}/override-dedup`
- Request: `{ "action": "split_new" | "link_account", "account_id": "...", "name": "...", "owner": "..." }`

### 8. Trigger CRM Re-Sync
- Method: `POST`
- Path: `/api/v1/leads/{id}/sync`

### 9. ICP Configuration
- Method: `GET` / `PUT`
- Path: `/api/v1/config/icp`
- Request (PUT): `{ "min_qualification_score": 60, "weights": { "corporate_domain": 20, "seat_volume": 25, "seniority": 20, "urgency": 15, "budget": 10, "product_fit": 10 } }`

### 10. Recalculate All Leads
- Method: `POST`
- Path: `/api/v1/leads/recalculate-all`
- Response: Re-evaluates all existing leads against active weights and updates database.

### 11. Real-Time Event Stream
- Method: `GET`
- Path: `/api/v1/events`
- Protocol: Server-Sent Events (`text/event-stream`)
- Events: `lead.processed`, `lead.draft_updated`, `stats.updated`

### 12. Inbound Simulator
- Method: `POST`
- Path: `/api/v1/simulate`
- Request: `{ "preset": "fintech" | "saas" | "hipaa" | "freemail" | "spam" }`

### 13. System Health and Metrics
- Method: `GET`
- Path: `/api/v1/health`
