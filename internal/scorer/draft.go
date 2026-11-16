package scorer

import (
	"fmt"
	"strings"

	"github.com/alexandrmotologa/formpilot/internal/extractor"
)

type DraftTone string

const (
	ToneExecutiveDirect DraftTone = "executive_direct"
	ToneConsultative    DraftTone = "consultative"
	ToneTechnicalValue  DraftTone = "technical_value"
)

// GenerateOutreachDraft creates a personalized three-paragraph sales rep email draft.
func GenerateOutreachDraft(lead *extractor.ExtractedLead, repName, bookingLink string, tone DraftTone) string {
	if lead == nil {
		return ""
	}

	if repName == "" {
		repName = "Alex Motologa"
	}
	if bookingLink == "" {
		bookingLink = "https://cal.com/formpilot-enterprise/discovery"
	}

	contactFirst := extractFirstName(lead.ContactName)
	company := lead.CompanyName
	if company == "" {
		company = "your organization"
	}

	var sb strings.Builder

	// Salutation
	if contactFirst != "" {
		sb.WriteString(fmt.Sprintf("Hi %s,\n\n", contactFirst))
	} else {
		sb.WriteString(fmt.Sprintf("Hello %s Team,\n\n", company))
	}

	// Paragraph 1: Contextual Greeting & Role Alignment
	switch tone {
	case ToneConsultative:
		sb.WriteString(fmt.Sprintf("Thank you for reaching out regarding %s's inbound data workflows. ", company))
		if lead.JobTitle != "" {
			sb.WriteString(fmt.Sprintf("Given your role as %s, I appreciate the operational focus on eliminating manual pipeline bottlenecks.\n\n", lead.JobTitle))
		} else {
			sb.WriteString("We regularly help fast-moving teams automate unstructured customer inquiries directly into their CRM records.\n\n")
		}

	case ToneTechnicalValue:
		sb.WriteString(fmt.Sprintf("Thanks for connecting with us regarding %s's pipeline requirements. ", company))
		if len(lead.TechStack) > 0 {
			sb.WriteString(fmt.Sprintf("Our architecture is designed specifically for environments utilizing %s with sub-3 second latency.\n\n", strings.Join(lead.TechStack, " and ")))
		} else {
			sb.WriteString("Our single-binary engine provides deterministic entity extraction and real-time reverse-ETL.\n\n")
		}

	default: // ToneExecutiveDirect
		sb.WriteString(fmt.Sprintf("Thank you for contacting formpilot regarding %s. ", company))
		if lead.JobTitle != "" {
			sb.WriteString(fmt.Sprintf("I noted your inquiry as %s, and our enterprise sales engineering team is prepared to assist immediately.\n\n", lead.JobTitle))
		} else {
			sb.WriteString("We are prepared to review your deployment scope immediately.\n\n")
		}
	}

	// Paragraph 2: Core Requirements, Volume, and Compliance
	requirements := []string{}
	if lead.RequestedSeats > 0 {
		requirements = append(requirements, fmt.Sprintf("%d enterprise seats", lead.RequestedSeats))
	}
	if len(lead.RequestedProducts) > 0 {
		requirements = append(requirements, strings.Join(lead.RequestedProducts, ", "))
	}
	if len(lead.TechStack) > 0 {
		requirements = append(requirements, fmt.Sprintf("integrations with %s", strings.Join(lead.TechStack, ", ")))
	}

	if len(requirements) > 0 {
		sb.WriteString(fmt.Sprintf("We reviewed your requirements for %s. ", strings.Join(requirements, " alongside ")))
	} else {
		sb.WriteString("We reviewed your inquiry details and system volume specifications. ")
	}

	if strings.Contains(lead.Urgency, "P0") || strings.Contains(lead.Urgency, "Immediate") {
		sb.WriteString("Given your target timeline to deploy within the month, we can expedite standard enterprise security questionnaires and provide a dedicated sandbox tenant today.\n\n")
	} else {
		sb.WriteString("We can align our implementation schedule with your target milestone dates.\n\n")
	}

	// Paragraph 3: Clear Call to Action
	sb.WriteString(fmt.Sprintf("Would you or your technical team be open to a 20-minute architecture overview this week? You can pick a convenient time directly on my calendar here: %s\n\n", bookingLink))
	sb.WriteString(fmt.Sprintf("Best regards,\n%s\nEnterprise Sales & Solutions Engineering\nformpilot", repName))

	return sb.String()
}

func extractFirstName(fullName string) string {
	parts := strings.Fields(fullName)
	if len(parts) > 0 {
		first := parts[0]
		// Skip title prefixes like Dr., Mr., Ms.
		if (first == "Dr." || first == "Mr." || first == "Ms." || first == "Mrs.") && len(parts) > 1 {
			return parts[1]
		}
		return first
	}
	return ""
}
