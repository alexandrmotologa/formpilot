package extractor

import (
	"regexp"
	"strconv"
	"strings"
)

// ExtractHeuristic runs deterministic regex and rule-based extraction on unstructured text.
func ExtractHeuristic(sender, subject, body string) *ExtractedLead {
	fullText := subject + "\n" + body
	lead := &ExtractedLead{
		Currency:         "USD",
		ExtractionMethod: "heuristic",
		ConfidenceScore:  0.85,
	}

	// 1. Extract Email and Contact Name from sender
	extractSenderDetails(sender, fullText, lead)

	// 2. Extract Domain & Company Name
	extractCompanyAndDomain(lead, fullText)

	// 3. Extract Job Title and Seniority
	extractTitleAndSeniority(lead, fullText)

	// 4. Extract Seat Volume
	extractSeats(lead, fullText)

	// 5. Extract Budget
	extractBudget(lead, fullText)

	// 6. Extract Phone
	extractPhone(lead, fullText)

	// 7. Extract Urgency / Timeline
	extractUrgency(lead, fullText)

	// 8. Extract Products, Tech Stack, and Compliance
	extractProductsAndStack(lead, fullText)

	// 9. Extract Pain Points summary
	extractPainPoints(lead, fullText)

	// Adjust confidence score based on completeness
	score := 0.50
	if lead.CompanyName != "" {
		score += 0.10
	}
	if lead.Domain != "" {
		score += 0.10
	}
	if lead.ContactEmail != "" {
		score += 0.10
	}
	if lead.RequestedSeats > 0 {
		score += 0.10
	}
	if lead.Budget > 0 {
		score += 0.05
	}
	if lead.Seniority != "Unspecified" {
		score += 0.05
	}
	lead.ConfidenceScore = score

	return lead
}

func extractSenderDetails(sender, fullText string, lead *ExtractedLead) {
	// From format: "Marcus Vance <m.vance@apexclearing.com>"
	reEmail := regexp.MustCompile(`([a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,})`)
	emailMatch := reEmail.FindString(sender)
	if emailMatch != "" {
		lead.ContactEmail = strings.ToLower(emailMatch)
	} else {
		// Try finding first email in body
		bodyEmail := reEmail.FindString(fullText)
		if bodyEmail != "" && !strings.Contains(bodyEmail, "sales@") && !strings.Contains(bodyEmail, "info@") {
			lead.ContactEmail = strings.ToLower(bodyEmail)
		}
	}

	// Extract display name
	if strings.Contains(sender, "<") {
		namePart := strings.TrimSpace(sender[:strings.Index(sender, "<")])
		namePart = strings.Trim(namePart, `"' `)
		if namePart != "" {
			lead.ContactName = namePart
		}
	} else if lead.ContactEmail != "" && sender == lead.ContactEmail {
		// Synthesize name from email prefix e.g. marcus.vance -> Marcus Vance
		prefix := strings.Split(lead.ContactEmail, "@")[0]
		parts := strings.FieldsFunc(prefix, func(r rune) bool {
			return r == '.' || r == '_' || r == '-'
		})
		if len(parts) >= 2 {
			lead.ContactName = strings.Title(parts[0]) + " " + strings.Title(parts[1])
		}
	}

	// Check if body starts with "I am [Name]" or "My name is [Name]"
	if lead.ContactName == "" {
		reName := regexp.MustCompile(`(?i)(?:I am|My name is|This is)\s+([A-Z][a-z]+(?:\s+[A-Z][a-z]+)?)`)
		m := reName.FindStringSubmatch(fullText)
		if len(m) > 1 {
			lead.ContactName = strings.TrimSpace(m[1])
		}
	}
}

func extractCompanyAndDomain(lead *ExtractedLead, fullText string) {
	if lead.ContactEmail != "" {
		rawDomain := ExtractDomainFromEmail(lead.ContactEmail)
		if !IsFreeWebmail(rawDomain) && !IsDisposableEmail(rawDomain) {
			lead.Domain = NormalizeDomain(rawDomain)
		}
	}

	// Search for explicit domain in text e.g. (apexclearing.com) or at apexclearing.com
	reDomainInText := regexp.MustCompile(`(?i)\b([a-zA-Z0-9-]+\.(?:com|org|io|net|co|ai|tech|us|ca|uk|de))\b`)
	domainMatches := reDomainInText.FindAllStringSubmatch(fullText, -1)
	for _, dm := range domainMatches {
		if len(dm) > 1 {
			cand := NormalizeDomain(dm[1])
			if !IsFreeWebmail(cand) && !IsDisposableEmail(cand) && !strings.Contains(cand, "formpilot") {
				if lead.Domain == "" {
					lead.Domain = cand
				}
				break
			}
		}
	}

	// Extract company name from patterns like "at [Company]", "from [Company]", "representing [Company]"
	reCompany := regexp.MustCompile(`(?i)(?:at|from|representing|Chief Medical Information Officer at|VP of Engineering at|Head of Product Operations at)\s+([A-Z][A-Za-z0-9&.\s]{2,30}?)(?:\s*\(|\s*,|\s*\.|\s*Phone|\s*We\b|\s*Our\b|\s*I\b|\n)`)
	m := reCompany.FindStringSubmatch(fullText)
	if len(m) > 1 {
		cand := strings.TrimSpace(m[1])
		cand = strings.TrimSuffix(cand, ".")
		cand = strings.TrimSuffix(cand, ",")
		if len(cand) > 2 && !strings.EqualFold(cand, "formpilot") {
			lead.CompanyName = cand
		}
	}

	// Fallback to domain prefix e.g. apexclearing.com -> Apex Clearing
	if lead.CompanyName == "" && lead.Domain != "" {
		prefix := strings.Split(lead.Domain, ".")[0]
		lead.CompanyName = strings.Title(prefix)
	}
}

func extractTitleAndSeniority(lead *ExtractedLead, fullText string) {
	lower := strings.ToLower(fullText)

	// Check explicit title patterns: "I am [Title] at" or signature blocks
	reTitle := regexp.MustCompile(`(?i)(?:I am|I'm|As)\s+(?:the\s+)?([A-Za-z\s]{3,35})\s+at\s+`)
	m := reTitle.FindStringSubmatch(fullText)
	if len(m) > 1 {
		lead.JobTitle = strings.TrimSpace(m[1])
	}

	// Seniority mapping
	switch {
	case strings.Contains(lower, "vp of") || strings.Contains(lower, "vice president") ||
		strings.Contains(lower, "cto") || strings.Contains(lower, "cio") || strings.Contains(lower, "ceo") ||
		strings.Contains(lower, "cfo") || strings.Contains(lower, "cmio") || strings.Contains(lower, "chief") ||
		strings.Contains(lower, "founder") || strings.Contains(lower, "partner") || strings.Contains(lower, "co-founder"):
		lead.Seniority = "C-Level / VP / Founder"
		if lead.JobTitle == "" {
			if strings.Contains(lower, "vp of engineering") {
				lead.JobTitle = "VP of Engineering"
			} else if strings.Contains(lower, "cmio") || strings.Contains(lower, "chief medical information officer") {
				lead.JobTitle = "Chief Medical Information Officer"
			} else {
				lead.JobTitle = "Executive / VP"
			}
		}

	case strings.Contains(lower, "director") || strings.Contains(lower, "head of") || strings.Contains(lower, "principal"):
		lead.Seniority = "Director / Head"
		if lead.JobTitle == "" {
			if strings.Contains(lower, "head of product operations") {
				lead.JobTitle = "Head of Product Operations"
			} else {
				lead.JobTitle = "Director"
			}
		}

	case strings.Contains(lower, "manager") || strings.Contains(lower, "lead") || strings.Contains(lower, "supervisor"):
		lead.Seniority = "Manager / Lead"
		if lead.JobTitle == "" {
			lead.JobTitle = "Team Lead / Manager"
		}

	case strings.Contains(lower, "engineer") || strings.Contains(lower, "specialist") || strings.Contains(lower, "developer") ||
		strings.Contains(lower, "coordinator") || strings.Contains(lower, "analyst"):
		lead.Seniority = "Individual Contributor"
		if lead.JobTitle == "" {
			lead.JobTitle = "Specialist / Engineer"
		}

	default:
		lead.Seniority = "Unspecified"
	}
}

func extractSeats(lead *ExtractedLead, fullText string) {
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)(?:need|looking for|evaluating for|volume|size|access|rough|roughly|around|for)?[:\s]*(\d{1,5})\s*(?:enterprise\s*|clinical\s*)?(?:seats?|licenses?|users?|coordinators?|engineers?|intake specialists?|specialists?)\b`),
		regexp.MustCompile(`(?i)(?:seat count|seats?|volume|licenses?)[:\s]+(\d{1,5})\b`),
		regexp.MustCompile(`(?i)\b(\d{1,5})\s+(?:enterprise\s*)?seats?\b`),
	}

	for _, re := range patterns {
		m := re.FindStringSubmatch(fullText)
		if len(m) > 1 {
			if val, err := strconv.Atoi(m[1]); err == nil && val > 0 {
				lead.RequestedSeats = val
				return
			}
		}
	}
}

func extractBudget(lead *ExtractedLead, fullText string) {
	reBudget := regexp.MustCompile(`(?i)(?:budget|approved|capital expenditure|expenditure|allocated|price range)?\s*(?:is|of|around|approx)?\s*\$([\d,]+(?:\.\d+)?)\s*(k|thousand|million|m)?`)
	matches := reBudget.FindAllStringSubmatch(fullText, -1)
	for _, m := range matches {
		if len(m) > 1 {
			cleanNum := strings.ReplaceAll(m[1], ",", "")
			val, err := strconv.ParseFloat(cleanNum, 64)
			if err == nil && val > 0 {
				multiplier := 1.0
				if len(m) > 2 {
					unit := strings.ToLower(m[2])
					if unit == "k" || unit == "thousand" {
						multiplier = 1000.0
					} else if unit == "m" || unit == "million" {
						multiplier = 1000000.0
					}
				}
				lead.Budget = val * multiplier
				return
			}
		}
	}
}

func extractPhone(lead *ExtractedLead, fullText string) {
	rePhone := regexp.MustCompile(`(?i)(?:Phone|Tel|Mobile|Call)?[:\s]*(\+?1?[-.\s]?\(?\d{3}\)?[-.\s]?\d{3}[-.\s]?\d{4})`)
	m := rePhone.FindStringSubmatch(fullText)
	if len(m) > 1 {
		lead.ContactPhone = NormalizePhone(m[1])
	}
}

func extractUrgency(lead *ExtractedLead, fullText string) {
	lower := strings.ToLower(fullText)
	switch {
	case strings.Contains(lower, "p0") || strings.Contains(lower, "immediate") || strings.Contains(lower, "30 days") ||
		strings.Contains(lower, "asap") || strings.Contains(lower, "next month") || strings.Contains(lower, "this week") ||
		strings.Contains(lower, "this thursday"):
		lead.Urgency = "P0 Immediate (< 1 mo)"

	case strings.Contains(lower, "1-3 months") || strings.Contains(lower, "2 months") || strings.Contains(lower, "next quarter") ||
		strings.Contains(lower, "q4") || strings.Contains(lower, "q1") || strings.Contains(lower, "q2") || strings.Contains(lower, "q3"):
		lead.Urgency = "P1 High (1-3 mos)"

	case strings.Contains(lower, "3-6 months") || strings.Contains(lower, "6 months"):
		lead.Urgency = "P2 Medium (3-6 mos)"

	case strings.Contains(lower, "curious") || strings.Contains(lower, "just checking") || strings.Contains(lower, "side project") ||
		strings.Contains(lower, "exploratory"):
		lead.Urgency = "P3 Exploratory (> 6 mos)"

	default:
		lead.Urgency = "P1 High (1-3 mos)"
	}
}

func extractProductsAndStack(lead *ExtractedLead, fullText string) {
	lower := strings.ToLower(fullText)

	// Check products
	products := []string{}
	if strings.Contains(lower, "entity extraction") || strings.Contains(lower, "unstructured") {
		products = append(products, "Automated Entity Extraction")
	}
	if strings.Contains(lower, "crm integration") || strings.Contains(lower, "hubspot") || strings.Contains(lower, "salesforce") {
		products = append(products, "CRM Ingestion Pipeline")
	}
	if strings.Contains(lower, "webhook") || strings.Contains(lower, "inbound") {
		products = append(products, "Real-time Webhook Receiver")
	}
	if strings.Contains(lower, "audit log") || strings.Contains(lower, "audit") {
		products = append(products, "Enterprise Audit Logs")
	}
	if strings.Contains(lower, "referral") || strings.Contains(lower, "intake") {
		products = append(products, "Clinical Intake Automation")
	}
	lead.RequestedProducts = products

	// Check tech stack & compliance
	stack := []string{}
	if strings.Contains(lower, "salesforce") {
		stack = append(stack, "Salesforce")
	}
	if strings.Contains(lower, "hubspot") {
		stack = append(stack, "HubSpot")
	}
	if strings.Contains(lower, "pipedrive") {
		stack = append(stack, "Pipedrive")
	}
	if strings.Contains(lower, "okta") || strings.Contains(lower, "saml") || strings.Contains(lower, "sso") {
		stack = append(stack, "SAML SSO / Okta")
	}
	if strings.Contains(lower, "soc2") {
		stack = append(stack, "SOC2 Type II")
	}
	if strings.Contains(lower, "hipaa") || strings.Contains(lower, "baa") {
		stack = append(stack, "HIPAA / BAA")
	}
	if strings.Contains(lower, "vpc") || strings.Contains(lower, "dedicated vpc") {
		stack = append(stack, "Dedicated VPC")
	}
	lead.TechStack = stack
}

func extractPainPoints(lead *ExtractedLead, fullText string) {
	// Look for requirements or scope lines
	lines := strings.Split(fullText, "\n")
	var notes []string
	for _, l := range lines {
		trimmed := strings.TrimSpace(l)
		if strings.HasPrefix(trimmed, "-") || strings.HasPrefix(trimmed, "*") {
			notes = append(notes, strings.TrimPrefix(strings.TrimPrefix(trimmed, "-"), "*"))
		}
	}
	if len(notes) > 0 {
		lead.PainPoints = strings.Join(notes, "; ")
	} else if len(lines) > 2 {
		lead.PainPoints = strings.TrimSpace(lines[1])
	}
}
