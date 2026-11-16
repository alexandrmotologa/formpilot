# formpilot

Autonomous unstructured-to-CRM ingestion engine and reverse-ETL pipeline for enterprise sales operations. Built as a single Go binary with an embedded React 19 cockpit, real-time deduplication, ICP qualification scoring, and multi-CRM dispatch.

## Core Capabilities

- Inbound Webhook Ingestion: Receives inbound email webhooks from SendGrid, Postmark, Mailgun, and AWS SES, as well as direct REST submissions.
- Multipart MIME and PDF Parsing: Decodes RFC 2822 multipart messages, extracts plaintext and HTML bodies, and reads attached PDF RFQ sheets.
- Dual Entity Extraction Engine: Uses schema-driven structured LLM extraction when an API key is present, with an automatic fallback to a deterministic rule-based extractor for offline and zero-dependency deployments.
- Corporate Domain Deduplication: Normalizes root domains and matches existing accounts using exact domain matching and Levenshtein distance to prevent duplicate account creation.
- 0 to 100 ICP Qualification Scoring: Evaluates firmographic fit across six dimensions including domain reputation, seat volume, executive seniority, timeline urgency, budget viability, and tech stack alignment.
- Automated Sales Outreach Drafts: Generates personalized, contextual three-paragraph response drafts tailored to specific product requirements and customer pain points.
- Multi-CRM Dispatch: Synchronizes leads, contacts, accounts, and opportunities into Salesforce Composite API, HubSpot CRM v3 API, and Pipedrive API with idempotency keys. Includes an integrated sandbox mock CRM for local testing without external credentials.
- Real-time React 19 Cockpit: Delivers a live RevOps feed, deduplication diff inspector, qualification score breakdown, outreach draft studio, and inbound simulator via Server-Sent Events (SSE).
- Compact Footprint: Operates under 15MB RAM baseline with zero CGO dependencies using pure-Go SQLite storage.

## Architecture

```
Inbound Email (SendGrid / Postmark / SES / PDF)
                    │
                    ▼
       ┌───────────────────────────┐
       │ Inbound Webhook Receiver  │
       │  (Signature Verification) │
       └────────────┬──────────────┘
                    ▼
       ┌───────────────────────────┐
       │  MIME & PDF Text Parser   │
       └────────────┬──────────────┘
                    ▼
       ┌───────────────────────────┐
       │   B2B Entity Extractor    │
       │ (LLM + Heuristic Fallback)│
       └────────────┬──────────────┘
                    ▼
       ┌───────────────────────────┐
       │ CRM Deduplication Matcher │
       │ (Domain Index + Levenshtein│
       └────────────┬──────────────┘
                    ▼
       ┌───────────────────────────┐
       │ ICP Fit Scoring & Drafts  │
       │   (0-100 Score + 3-Para)  │
       └────────────┬──────────────┘
                    │
        ┌───────────┴───────────┐
        ▼                       ▼
┌──────────────────┐    ┌──────────────────┐
│ SQLite Datastore │    │ Real-time SSE    │
│  (Audit Trail)   │    │ (React 19 Studio)│
└──────────────────┘    └──────────────────┘
        │
        ▼
┌──────────────────────────────────────────┐
│ Enterprise CRM Dispatcher                │
│ (Salesforce / HubSpot / Pipedrive / Mock)│
└──────────────────────────────────────────┘
```

## Quick Start

### Prerequisites

- Go 1.22 or higher
- Node.js 20 or higher (for building the web cockpit)

### Build and Run

```bash
# Clone the repository
git clone https://github.com/alexandrmotologa/formpilot.git
cd formpilot

# Build web frontend and Go daemon
make build

# Run the daemon on default port :9393
./bin/formpilot serve --config examples/formpilot.yaml
```

The embedded sales cockpit opens at `http://localhost:9393`.

### CLI Commands

```bash
# Start daemon with API and embedded studio
formpilot serve --port 9393 --config config.yaml

# Ingest an email or PDF file directly via CLI
formpilot ingest --file examples/sample_emails/enterprise_fintech_rfq.eml

# Send a simulated test lead to the local daemon
formpilot simulate --preset fintech

# Check system health, storage statistics, and CRM connectivity
formpilot status
```

## Inbound Webhooks

formpilot provides dedicated webhook endpoints with signature validation:

- `POST /api/v1/webhooks/sendgrid`: SendGrid Inbound Parse webhook.
- `POST /api/v1/webhooks/postmark`: Postmark Inbound webhook.
- `POST /api/v1/webhooks/mailgun`: Mailgun Inbound webhook.
- `POST /api/v1/webhooks/ses`: AWS SES SNS notification webhook.
- `POST /api/v1/webhooks/generic`: Direct multipart or JSON lead submission.

## Configuration

Configuration is managed via YAML or environment variables:

```yaml
server:
  port: 9393
  host: "0.0.0.0"

database:
  path: "formpilot.db"

extraction:
  provider: "heuristic" # "heuristic" or "openai"
  model: "gpt-4o-mini"
  api_key: "${LLM_API_KEY}"

crm:
  mode: "mock" # "mock" or "live"
  hubspot:
    enabled: true
    access_token: "${HUBSPOT_ACCESS_TOKEN}"
  salesforce:
    enabled: false
    instance_url: "${SALESFORCE_INSTANCE_URL}"
    access_token: "${SALESFORCE_ACCESS_TOKEN}"
  pipedrive:
    enabled: false
    api_token: "${PIPEDRIVE_API_TOKEN}"

slack:
  enabled: false
  webhook_url: "${SLACK_WEBHOOK_URL}"

icp:
  min_qualification_score: 60
  weights:
    corporate_domain: 20
    seat_volume: 25
    seniority: 20
    urgency: 15
    budget: 10
    product_fit: 10
```

## License

ISC License. See `LICENSE` for details.
