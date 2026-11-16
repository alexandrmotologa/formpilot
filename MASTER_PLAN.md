# formpilot: Architecture and Engineering Master Plan
> Autonomous Unstructured-to-CRM Ingestion Agent & Reverse-ETL Pipeline for Enterprise Sales Operations
> Target: Single-Binary Go Engine (< 15MB RAM) with Embedded React 19 Sales Ops Studio & Multi-CRM Dispatcher

## 1. Problem Statement and Motivation

### 1.1 Inbound Lead Triage Latency
In high-velocity enterprise sales organizations, incoming business leads, Requests for Quotation (RFQs), and partnership inquiries arrive in messy, unstructured formats: long customer emails, PDF request attachments, contact form free-text textareas, and scanned purchase inquiries:
1. Lead Response Time: Research shows that contacting an inbound lead within 5 minutes results in a 391% higher qualification rate compared to waiting 30 minutes. When sales development representatives (SDRs) manually triage and type lead details into CRMs, average B2B enterprise response times exceed 4 hours.
2. Manual Data Entry Overhead: SDRs spend 20% to 30% of their working hours manually copy-pasting company names, contact titles, phone numbers, requested seat counts, and budget notes into Salesforce, HubSpot, or Pipedrive.
3. Database Duplication: Manual entry creates rampant CRM duplication (for example, creating "Acme Corp" when "Acme Corporation Inc." already exists with an assigned account executive). This causes multi-rep outreach collisions and corrupted pipeline reporting.
4. Inconsistent Lead Qualification: Human reps inconsistently score inbound urgency, frequently letting high-value urgent enterprise RFQs sit in shared inboxes while chasing small accounts.

### 1.2 The formpilot Solution
formpilot is an autonomous unstructured-to-CRM ingestion engine written in Go. Operating as a lightweight daemon receiving email webhooks and form submissions:
- Instant Ingestion: Receives inbound email webhooks (SendGrid, Postmark, Mailgun, AWS SES) and PDF quote requests in real time.
- Strict Entity Extraction: Employs schema-constrained language models and deterministic heuristic fallback parsers to extract verified B2B entities: Company Name, Corporate Domain, Contact Name, Seniority, Work Email, Requested Products, Estimated Seat Count, Budget, Urgency, and Tech Stack.
- Deduplication and Account Matching: Queries target CRMs and in-memory indexed caches to match existing company domains, parent corporate entities, or open opportunities before creating records.
- Lead Qualification Scoring (0 to 100): Calculates an objective Ideal Customer Profile (ICP) fit score and assigns priority routing tiers.
- Multi-CRM Dispatch: Injects normalized records into Salesforce (REST Composite API), HubSpot (CRM Objects API), or Pipedrive in under 3 seconds with strict idempotency keys.
- Automated Sales Outreach Draft: Synthesizes a personalized, contextual reply draft for the assigned sales representative, summarizing the customer requirements and proposing discovery next steps.
- Embedded React 19 Sales Ops Studio: Delivers RevOps teams an interactive cockpit to monitor ingestion velocity, review deduplication decisions, and fine-tune schema extraction rules.

---

## 2. High-Level System Architecture

```
       +-------------------------------------------------------------+
       |               Inbound Unstructured Lead Sources             |
       |  (Inbound Emails via SendGrid/SES, PDF RFQs, Web Forms)     |
       +-------------------------------------------------------------+
                                      |
                         HTTP Webhook Payload / Email MIME
                                      v
       +-------------------------------------------------------------+
       |                     formpilot Core Daemon                   |
       |                   (High-Performance Go CLI)                |
       |                                                             |
       |  +---------------------+        +------------------------+  |
       |  | Ingestion Receiver  |------->| MIME & Attachment      |  |
       |  | & Webhook Verifier  |        | Text Extractor (PDF)   |  |
       |  +---------------------+        +------------------------+  |
       |                                              |              |
       |                                              v              |
       |  +-------------------------------------------------------+  |
       |  |          Structured B2B Entity Extraction Engine      |  |
       |  |  - Primary: Schema-Constrained LLM Engine             |  |
       |  |  - Fallback: Zero-Dependency Deterministic Heuristic  |  |
       |  +-------------------------------------------------------+  |
       |                               |                             |
       |                               v                             |
       |  +-------------------------------------------------------+  |
       |  |             CRM Deduplication & Account Matcher       |  |
       |  |  - Exact Root Domain Lookup & Fuzzy Levenshtein Match |  |
       |  |  - Associate with Existing Account OR Provision New   |  |
       |  +-------------------------------------------------------+  |
       |                               |                             |
       |                               v                             |
       |  +-------------------------------------------------------+  |
       |  |            ICP Scoring & Sales Draft Synthesizer      |  |
       |  |  - Fit Score (0-100) across 6 Firmographic Dimensions |  |
       |  |  - Contextual 1-Click Outreach Email Draft for Rep    |  |
       |  +-------------------------------------------------------+  |
       |                               |                             |
       |                               v                             |
       |  +-------------------------------------------------------+  |
       |  |               Pure-Go Embedded SQLite Storage         |  |
       |  |  - Leads, CRM Cache, Sync Logs, Full Audit Trail      |  |
       |  +-------------------------------------------------------+  |
       |                               |                             |
       +-------------------------------|-----------------------------+
                                       |
                                       v
       +-------------------------------------------------------------+
       |                  Enterprise Dispatch Surfaces               |
       |  1. Direct Multi-CRM Sync: Salesforce / HubSpot / Pipedrive |
       |  2. Slack Instant Lead Alert with Action Buttons            |
       |  3. Embedded React 19 Sales Studio (Local HTTP :9393)       |
       +-------------------------------------------------------------+
```

---

## 3. Entity Extraction & ICP Scoring Pipeline

| Processing Step | Technical Operation |
| :--- | :--- |
| 1. Payload Parsing | Parses multipart MIME emails, extracts body text, and decodes attached PDF inquiry sheets via native Go PDF parser. |
| 2. Schema-Constrained Extraction | Extracts JSON entities with strict type enforcement: `company`, `domain`, `contact_name`, `email`, `phone`, `seats`, `budget`, `urgency`. |
| 3. Domain Deduplication Check | Queries CRM account index using normalized root domain (`acme.com`). Prevents split accounts and associates with existing account owner. |
| 4. ICP Fit Scoring | Evaluates enterprise indicators: corporate email domain, seat count >= 25, executive seniority, explicit purchase timeline. |
| 5. CRM Object Provisioning | Executes atomic upsert on Contact, Account, and Deal objects with idempotency keys to prevent duplicates. |
| 6. Rep Outreach Synthesis | Generates a 3-paragraph executive reply addressing the lead's specific product inquiries, staged ready for rep approval. |

---

## 4. Repository Structure

```
formpilot/
├── cmd/
│   └── formpilot/
│       └── main.go                 # CLI entry point, subcommands: serve, ingest, simulate, status
├── internal/
│   ├── config/
│   │   ├── config.go               # YAML configuration for CRM API keys, webhooks, ICP rules
│   │   └── schema.go               # Target CRM field mappings, custom fields, and thresholds
│   ├── storage/
│   │   ├── db.go                   # Pure-Go SQLite connection and migration engine
│   │   ├── leads.go                # Lead entity storage and query operations
│   │   ├── accounts.go             # Account deduplication cache repository
│   │   └── logs.go                 # CRM sync log and audit trail repository
│   ├── receiver/
│   │   ├── webhook.go              # SendGrid, Postmark, Mailgun, and AWS SES webhook handler
│   │   ├── mime_parser.go          # Multipart email body and attachment parser
│   │   └── pdf_extract.go          # PDF text extraction for RFQ attachment documents
│   ├── extractor/
│   │   ├── models.go               # Structured entity data structures
│   │   ├── entity.go               # Extraction orchestrator with fallback routing
│   │   ├── heuristic.go            # High-precision deterministic regex and rule extractor
│   │   ├── llm.go                  # Schema-driven LLM extractor (OpenAI / Gemini / Anthropic)
│   │   ├── normalizer.go           # Corporate domain cleaner and phone normalizer
│   │   └── validator.go            # Verifies work email validity (rejects personal/disposable domains)
│   ├── dedup/
│   │   ├── matcher.go              # Domain and fuzzy name account deduplication logic
│   │   └── cache.go                # In-memory indexed cache of active CRM accounts and domains
│   ├── scorer/
│   │   ├── icp.go                  # 0-100 ICP fit scoring algorithm with 6-dimension breakdown
│   │   └── draft.go                # Personalized sales rep outreach email generator
│   ├── crm/
│   │   ├── types.go                # Unified CRM client interface and result models
│   │   ├── hubspot.go              # HubSpot CRM Objects v3 API client
│   │   ├── salesforce.go           # Salesforce REST Composite API client
│   │   ├── pipedrive.go            # Pipedrive Leads and Deals API client
│   │   ├── mock.go                 # Sandboxed local CRM simulator for testing
│   │   └── dispatcher.go           # Multi-CRM dispatch orchestrator with retry tracking
│   ├── notify/
│   │   └── slack.go                # Slack webhook alert with rich lead card and action buttons
│   ├── pipeline/
│   │   └── pipeline.go             # End-to-end ingestion pipeline tying all stages together (< 3s target)
│   └── server/
│       ├── api.go                  # Management REST endpoints for leads, stats, simulator, logs
│       ├── embed.go                # Embedded React 19 studio static assets
│       └── sse.go                  # Real-time event stream broadcaster for incoming leads
├── web/
│   ├── src/
│   │   ├── App.tsx                 # React 19 root application
│   │   ├── components/
│   │   │   ├── LiveLeadFeed.tsx    # Real-time card feed of ingested inbound leads
│   │   │   ├── LeadDetailModal.tsx # Full entity inspection, raw MIME, and audit trail
│   │   │   ├── DedupInspector.tsx  # Side-by-side view of matched existing vs new accounts
│   │   │   ├── LeadScoreGauge.tsx  # Radial ICP fit score with qualification breakdown
│   │   │   ├── OutreachDraft.tsx   # Editable email draft viewer with 1-click send button
│   │   │   ├── CrmSyncStatus.tsx   # Synchronization status indicators (HubSpot / Salesforce / Pipedrive)
│   │   │   ├── InboundSimulator.tsx# Interactive RFQ sender with presets (FinTech, SaaS, HIPAA, Spam)
│   │   │   └── Header.tsx          # System status header with latency metrics
│   │   ├── types/
│   │   │   └── index.ts            # Frontend TypeScript models matching Go API
│   │   └── pages/
│   │       └── Dashboard.tsx       # Main RevOps monitoring cockpit
│   ├── package.json
│   ├── tsconfig.json
│   ├── vite.config.ts
│   └── tailwind.config.js
├── docs/
│   ├── architecture.md             # Detailed architectural specifications and data flow
│   ├── api.md                      # REST and Webhook API reference documentation
│   ├── icp_scoring.md              # Qualification formula and scoring rules reference
│   └── deduplication.md            # Domain normalization and fuzzy match algorithms
├── examples/
│   ├── formpilot.yaml              # Production configuration sample
│   └── sample_emails/              # Realistic raw inbound emails for testing
│       ├── enterprise_fintech_rfq.eml
│       ├── midmarket_saas_demo.txt
│       ├── healthcare_hipaa_rfq.txt
│       ├── ambiguous_inquiry.txt
│       └── disposable_webmail_spam.txt
├── Makefile
├── go.mod
├── go.sum
├── LICENSE
└── README.md
```

---

## 5. Implementation Roadmap and Checklist

- [x] Phase 0: Master Plan enhancement, GitHub repository initialization, and core documentation.
- [ ] Phase 1: Go Engine foundation, YAML config parser, and pure-Go SQLite storage engine.
- [ ] Phase 2: Inbound webhook receivers (SendGrid, Postmark, Mailgun, SES, Generic) and MIME/PDF parsers.
- [ ] Phase 3: Schema-driven entity extraction engine (LLM + zero-dependency deterministic heuristic fallback), domain normalizer, and work email validator.
- [ ] Phase 4: CRM deduplication, account matching engine (exact domain + Levenshtein fuzzy match), and account cache.
- [ ] Phase 5: ICP fit scoring algorithm (0 to 100 with itemized breakdown) and sales outreach email draft synthesizer.
- [ ] Phase 6: Multi-CRM dispatch engine (HubSpot, Salesforce, Pipedrive, Sandboxed Mock CRM) and Slack Block Kit alerts.
- [ ] Phase 7: REST API server, SSE real-time event broadcaster, and pipeline orchestrator.
- [ ] Phase 8: Embedded React 19 Sales Ops Cockpit (Tailwind CSS, Lucide icons, Live Feed, Dedup Inspector, Draft Studio, Inbound Simulator).
- [ ] Phase 9: CLI entry point (`cmd/formpilot`), test datasets, end-to-end integration test suite, and performance verification (< 15MB RAM, < 3s latency).
- [ ] Phase 10: Final documentation review and GitHub synchronization.
