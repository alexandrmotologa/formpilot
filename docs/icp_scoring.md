# Ideal Customer Profile (ICP) Scoring

## Formula Overview

formpilot evaluates every incoming lead on a 0 to 100 scale across six weighted dimensions. The objective is to prioritize high-value commercial accounts and route them immediately to senior sales executives.

## Scoring Dimensions

### 1. Domain Reputation (Max: 20 points)
- Corporate Domain (e.g. `@cyberdyne.com`): +20 points
- Generic Free Webmail (`@gmail.com`, `@yahoo.com`, `@outlook.com`): +0 points
- Disposable Email (`@tempmail.com`, `@10minutemail.com`): -50 points

### 2. Seat Volume (Max: 25 points)
- 100+ Seats: +25 points
- 25 to 99 Seats: +18 points
- 10 to 24 Seats: +10 points
- 1 to 9 Seats: +4 points
- Unspecified: +0 points

### 3. Decision Maker Seniority (Max: 20 points)
- C-Level / VP / Founder / Partner: +20 points
- Director / Head of Department: +15 points
- Manager / Lead: +10 points
- Individual Contributor / Consultant: +5 points
- Unspecified: +0 points

### 4. Timeline Urgency (Max: 15 points)
- P0: Immediate / Within 30 days: +15 points
- P1: 1 to 3 months: +10 points
- P2: 3 to 6 months: +5 points
- P3: Exploratory / 6+ months: +2 points

### 5. Budget Viability (Max: 10 points)
- Explicit budget stated >= $50,000: +10 points
- Explicit budget stated < $50,000: +6 points
- Budget inquiry without stated amount: +3 points
- No budget mentioned: +0 points

### 6. Product Fit and Technical Specification (Max: 10 points)
- Specific enterprise features requested (SSO, SOC2, HIPAA, Dedicated VPC): +10 points
- Standard products identified: +5 points
- General or ambiguous inquiry: +0 points

## Qualification Tiers

| Score Range | Tier Name | Routing Action |
| :--- | :--- | :--- |
| 80 to 100 | Enterprise Hot Lead | Slack emergency alert, assign to Enterprise AE, instant draft |
| 60 to 79 | Mid-Market Qualified | Standard CRM assignment, standard AE notification |
| 40 to 59 | SMB / Self-Serve | Automated self-serve trial email dispatch |
| 0 to 39 | Unqualified / Low Intent | Low-priority queue, no immediate sales rep notification |
