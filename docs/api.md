# API Reference

The formpilot daemon exposes HTTP endpoints for webhook ingestion, lead management, and real-time streaming.

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
- Payload:
  ```json
  {
    "sender": "sarah.connor@cyberdyne.com",
    "subject": "Enterprise Quote Request for 150 Seats",
    "body": "Hi team, we are evaluating your platform for 150 engineers. We need SOC2 compliance and SSO. Budget is $80k. Looking to deploy by next month.",
    "source": "contact_form"
  }
  ```

## Management Endpoints

### 1. List Leads
- Method: `GET`
- Path: `/api/v1/leads`
- Query Parameters: `limit`, `offset`, `tier`, `status`
- Response: Array of lead records with extracted entities, scores, and sync statuses.

### 2. Get Lead Detail
- Method: `GET`
- Path: `/api/v1/leads/{id}`
- Response: Detailed lead record including raw payload, extraction confidence, and audit trail.

### 3. Update Outreach Draft
- Method: `PUT`
- Path: `/api/v1/leads/{id}/draft`
- Request: `{ "draft": "Updated email body text" }`

### 4. Trigger CRM Re-Sync
- Method: `POST`
- Path: `/api/v1/leads/{id}/sync`
- Request: `{ "crm": "hubspot" }`

### 5. Real-Time Event Stream
- Method: `GET`
- Path: `/api/v1/events`
- Protocol: Server-Sent Events (text/event-stream)
- Events: `lead.ingested`, `lead.extracted`, `lead.scored`, `lead.synced`

### 6. Inbound Simulator
- Method: `POST`
- Path: `/api/v1/simulate`
- Request: `{ "preset": "fintech" }`
- Presets: `fintech`, `saas_demo`, `healthcare`, `freemail_spam`

### 7. System Health and Metrics
- Method: `GET`
- Path: `/api/v1/health`
- Response: System uptime, memory allocation, total processed leads, and CRM status.
