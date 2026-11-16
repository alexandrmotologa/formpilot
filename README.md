<p align="center">
  <img src="docs/images/logo.png?raw=true" alt="formpilot Logo" width="130" style="border-radius: 24px;" />
</p>

<h1 align="center">formpilot</h1>

<p align="center">
  <strong>Autonomous unstructured-to-CRM ingestion engine and reverse-ETL pipeline for enterprise sales operations.</strong>
</p>

<p align="center">
  <a href="https://github.com/alexandrmotologa/formpilot/actions"><img src="https://img.shields.io/github/actions/workflow/status/alexandrmotologa/formpilot/ci.yml?branch=main&label=CI&style=flat-square" alt="CI Status" /></a>
  <a href="https://golang.org"><img src="https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat-square&logo=go&logoColor=white" alt="Go Version" /></a>
  <a href="https://react.dev"><img src="https://img.shields.io/badge/React-19-61DAFB?style=flat-square&logo=react&logoColor=black" alt="React 19" /></a>
  <a href="https://sqlite.org"><img src="https://img.shields.io/badge/SQLite-Pure--Go-003B57?style=flat-square&logo=sqlite&logoColor=white" alt="SQLite" /></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/License-ISC-blue?style=flat-square" alt="License" /></a>
  <img src="https://img.shields.io/badge/RAM-%3C%2015MB-emerald?style=flat-square" alt="RAM Baseline" />
  <img src="https://img.shields.io/badge/Latency-%3C%203s%20SLA-blueviolet?style=flat-square" alt="Latency SLA" />
</p>

---

<p align="center">
  <img src="docs/images/formpilot_demo.gif?raw=true" alt="formpilot Live Execution Demo" width="900" style="border-radius: 12px; border: 1px solid #1e293b;" />
</p>

---

## Executive Overview

formpilot transforms unstructured inbound RFQs, partner leads, and sales inquiries into structured CRM records in under 3 seconds. The daemon operates on less than 15MB RAM as a single statically compiled binary with zero CGO dependencies.

It listens to email webhooks from SendGrid, Postmark, Mailgun, and AWS SES, parses complex MIME bodies and PDF specification attachments, normalizes corporate domains, resolves account duplicates against CRM caches, scores Ideal Customer Profile (ICP) fit from 0 to 100, generates personalized executive outreach drafts, and synchronizes accounts, contacts, and deals into HubSpot, Salesforce, or Pipedrive.

An embedded Sales Ops Cockpit powered by React 19 provides real-time visibility through Server-Sent Events (SSE), with sub-millisecond stage-by-stage latency tracking, interactive deduplication override controls, and multi-tone outreach generation.

---

## Cockpit Interface

### Real-Time Inbound Cockpit & Pipeline Waterfall

The desktop cockpit visualizes inbound lead volume, qualification tiers, deduplication status, and execution latency. The sub-millisecond waterfall breaks down execution across MIME parsing, entity extraction, deduplication lookup, ICP scoring, draft synthesis, CRM dispatch, and SQLite persistence.

<p align="center">
  <img src="docs/images/cockpit_dashboard.png?raw=true" alt="formpilot Desktop Cockpit Dashboard" width="900" style="border-radius: 10px; border: 1px solid #1e293b;" />
</p>

### Sales Outreach Studio

The built-in outreach studio equips sales representatives to review, edit, and dispatch automated three-paragraph responses. Representatives can toggle between Executive Direct, Consultative Discovery, and Technical Architecture tones, insert dynamic token chips, and trigger direct email dispatch.

<p align="center">
  <img src="docs/images/outreach_studio.png?raw=true" alt="formpilot Sales Outreach Studio" width="900" style="border-radius: 10px; border: 1px solid #1e293b;" />
</p>

### RFQ Simulation & Responsive Multi-Device Design

The inbound simulator includes real-world enterprise test presets (FinTech Operations RFP, Healthcare HIPAA Intake, Mid-Market SaaS Demo, and Disposable Spam Backlinks) alongside a custom RFQ builder. The interface is responsive across desktop, tablet, and mobile devices.

<p align="center">
  <img src="docs/images/simulator_settings.png?raw=true" alt="formpilot Inbound Lead Simulator Modal" width="580" style="border-radius: 10px; border: 1px solid #1e293b; margin-right: 12px;" />
  <img src="docs/images/mobile_cockpit.png?raw=true" alt="formpilot Mobile Cockpit Layout" width="280" style="border-radius: 10px; border: 1px solid #1e293b;" />
</p>

---

## Core Capabilities

- Inbound Webhook Ingestion: Receives inbound email webhooks from SendGrid, Postmark, Mailgun, and AWS SES, as well as direct JSON and multipart REST submissions.
- Multipart MIME and PDF Parsing: Decodes RFC 2822 multipart messages, strips HTML tags, and extracts text directly from attached PDF RFQ specification sheets using a pure-Go stream parser.
- Dual Entity Extraction Engine: Uses schema-driven structured LLM extraction when an API key is present, with an automatic fallback to a deterministic rule-based extractor for offline and zero-dependency deployments.
- Corporate Domain Deduplication: Normalizes root domains and matches existing accounts using exact domain matching and Levenshtein distance to prevent duplicate account creation.
- 0 to 100 ICP Qualification Scoring: Evaluates firmographic fit across six dimensions: corporate domain reputation, requested seat volume, executive seniority, timeline urgency, allocated budget, and tech stack alignment.
- Automated Sales Outreach Drafts: Generates personalized, contextual three-paragraph response drafts tailored to specific product requirements and customer pain points.
- Multi-CRM Dispatch: Synchronizes leads, contacts, accounts, and opportunities into Salesforce Composite API, HubSpot CRM v3 API, and Pipedrive REST API with idempotency keys. Includes an integrated sandbox mock CRM for local testing without external credentials.
- Real-time React 19 Cockpit: Delivers a live RevOps feed, deduplication diff inspector, qualification score breakdown, outreach draft studio, and inbound simulator via Server-Sent Events (SSE).
- Compact Footprint: Operates under 15MB RAM baseline with zero CGO dependencies using pure-Go SQLite storage with WAL mode enabled.

---

## Architecture Pipeline

```
Inbound Email (SendGrid / Postmark / SES / PDF Attachment)
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

---

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
npm install --prefix web
npm run build --prefix web
go build -o bin/formpilot.exe ./cmd/formpilot

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

---

## Inbound Webhook Endpoints

formpilot provides dedicated webhook endpoints with signature validation:

- `POST /api/v1/webhooks/sendgrid`: SendGrid Inbound Parse webhook.
- `POST /api/v1/webhooks/postmark`: Postmark Inbound webhook.
- `POST /api/v1/webhooks/mailgun`: Mailgun Inbound webhook.
- `POST /api/v1/webhooks/ses`: AWS SES SNS notification webhook.
- `POST /api/v1/webhooks/generic`: Direct multipart or JSON lead submission.

---

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

---

## Benchmark Performance

| Dimension | Target Specification | Measured Performance |
| :--- | :--- | :--- |
| End-to-End Latency | Under 3.00 seconds | 35ms to 75ms (heuristic) / 1.4s (LLM) |
| Memory Consumption | Under 15MB RAM baseline | 6.27MB baseline / 11.27MB active load |
| Deduplication Precision | 100% root domain match | Sub-millisecond indexed lookup |
| Binary Size | Self-contained single binary | 12.03MB (Go runtime + embedded assets) |
| External Dependencies | Zero CGO requirement | Pure-Go modernc.org/sqlite |

---

## License

ISC License. See [LICENSE](LICENSE) for details.
