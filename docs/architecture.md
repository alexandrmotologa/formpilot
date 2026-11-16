# System Architecture

## Overview

formpilot runs as a self-contained daemon that processes inbound unstructured sales inquiries (emails and PDF attachments), extracts structured B2B entities, deduplicates against existing CRM accounts, calculates qualification scores, and dispatches records to CRM systems in under 3 seconds.

## Subsystem Breakdown

### 1. Ingestion and Receiver Subsystem
- Located in `internal/receiver/`.
- Parses incoming HTTP webhook payloads from email service providers (SendGrid, Postmark, Mailgun, AWS SES).
- Decodes MIME multipart messages into headers, plaintext bodies, HTML bodies, and file attachments.
- Parses PDF attachments in native Go to extract unformatted RFQ text.

### 2. Entity Extraction Subsystem
- Located in `internal/extractor/`.
- Employs a dual-mode strategy:
  1. Primary: Structured JSON extraction using LLM providers when an API key is available.
  2. Fallback: High-precision deterministic regex and rule-based parser that operates locally without network calls.
- Normalizes corporate domains by stripping prefixes (`www.`, protocol headers, trailing paths) and isolates root domains.
- Validates work email addresses by rejecting personal webmail domains (gmail.com, yahoo.com, outlook.com) and disposable email services.

### 3. Deduplication and Account Matcher Subsystem
- Located in `internal/dedup/`.
- Maintains an in-memory indexed cache of CRM accounts, backed by SQLite.
- Executes two-tier matching:
  1. Exact Domain Match: Compares normalized domain with existing accounts.
  2. Fuzzy Name Match: Uses Levenshtein distance on company names to catch typographical variations.
- Employs conflict resolution rules:
  - If an account exists, associate the new contact and open deal with the existing account owner without creating duplicate account records.
  - If no account exists, create a new Account/Company, Contact, and Deal atomically.

### 4. ICP Qualification and Draft Synthesis Subsystem
- Located in `internal/scorer/`.
- Evaluates six firmographic dimensions to generate a 0 to 100 score:
  - Corporate Domain (0 or 20 points)
  - Seat Volume (up to 25 points)
  - Decision Maker Seniority (up to 20 points)
  - Timeline Urgency (up to 15 points)
  - Budget Viability (up to 10 points)
  - Product Specification Fit (up to 10 points)
- Generates a contextual three-paragraph response email draft addressing the customer's specific product requests and proposed meeting next steps.

### 5. Multi-CRM Dispatcher Subsystem
- Located in `internal/crm/`.
- Provides a unified interface for CRM targets:
  - HubSpot CRM v3 Objects API (Companies, Contacts, Deals, and associations).
  - Salesforce REST Composite API with idempotency headers.
  - Pipedrive API (Organizations, Persons, Deals).
  - Sandboxed Mock CRM: In-memory simulation engine that records all dispatch transactions, validates schemas, and returns realistic synthetic identifiers.

### 6. Storage and Audit Trail
- Located in `internal/storage/`.
- Built on pure-Go SQLite (`modernc.org/sqlite`).
- Stores raw payloads, extracted entities, deduplication decisions, ICP scores, generated drafts, and dispatch latency logs.

### 7. Real-Time Management Cockpit
- Located in `web/` and served via `internal/server/`.
- React 19 single-page application embedded in the Go binary using `//go:embed`.
- Uses Server-Sent Events (SSE) to push lead lifecycle events to the browser in real time.
