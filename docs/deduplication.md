# Account Deduplication Engine

## Problem Statement

When prospective customers contact an organization through multiple reps or public forms, CRM records often fracture into duplicates (e.g., "Stark Industries", "Stark Industries LLC", "stark.com"). Duplicate accounts cause conflicting sales outreach and inaccurate pipeline reporting.

## Deduplication Strategy

formpilot resolves account duplication through a two-stage matching pipeline:

### 1. Root Domain Normalization
The primary key for corporate entity resolution is the normalized internet domain:
- Strips URL schemes (`http://`, `https://`).
- Strips standard host prefixes (`www.`, `mail.`, `portal.`).
- Strips URL paths, query parameters, and port numbers.
- Converts to lowercase ASCII.
- Filters out public webmail providers (e.g. `gmail.com`, `yahoo.com`, `outlook.com`) to prevent clustering individual leads under a shared public domain.

### 2. Matching Hierarchy
1. Exact Normalized Domain:
   - Compares the inbound lead's domain against the indexed CRM account cache.
   - Match Confidence: 1.0.
   - Action: Link incoming Contact and Deal to the existing Account. Preserve existing Account Executive (AE) ownership.

2. Fuzzy Company Name (Levenshtein Distance):
   - Triggered when domain is absent or unrecognized.
   - Evaluates similarity score: `1 - (Distance / MaxLength)`.
   - Threshold: >= 0.85 indicates a probable match.
   - Action: Flag for review in the Sales Cockpit with visual diff.

3. Provisioning New Account:
   - When confidence is below threshold, creates a new Account record alongside Contact and Deal.
   - Updates local cache and persists to SQLite.
