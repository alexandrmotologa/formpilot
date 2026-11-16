package scorer

import (
	"strings"

	"github.com/alexandrmotologa/formpilot/internal/config"
	"github.com/alexandrmotologa/formpilot/internal/extractor"
)

type ScoreBreakdown struct {
	CorporateDomainScore int `json:"corporate_domain_score"`
	SeatVolumeScore      int `yaml:"seat_volume_score" json:"seat_volume_score"`
	SeniorityScore       int `json:"seniority_score"`
	UrgencyScore         int `json:"urgency_score"`
	BudgetScore          int `json:"budget_score"`
	ProductFitScore      int `json:"product_fit_score"`
}

type ICPResult struct {
	TotalScore             int            `json:"total_score"`
	Tier                   string         `json:"tier"` // "Enterprise Hot Lead", "Mid-Market Qualified", "SMB / Self-Serve", "Unqualified / Low Intent"
	Breakdown              ScoreBreakdown `json:"breakdown"`
	PositiveFactors        []string       `json:"positive_factors"`
	NegativeFactors        []string       `json:"negative_factors"`
	RoutingRecommendation string         `json:"routing_recommendation"`
}

// CalculateICP calculates the objective fit score from 0 to 100 based on extracted entities and weights.
func CalculateICP(lead *extractor.ExtractedLead, weights config.ICPWeights) *ICPResult {
	result := &ICPResult{
		PositiveFactors: []string{},
		NegativeFactors: []string{},
	}

	if lead == nil {
		result.Tier = "Unqualified / Low Intent"
		result.RoutingRecommendation = "Archive or discard empty submission"
		return result
	}

	normDomain := extractor.NormalizeDomain(lead.Domain)

	// 1. Corporate Domain vs Disposable/Webmail
	if extractor.IsDisposableEmail(normDomain) {
		result.NegativeFactors = append(result.NegativeFactors, "Disposable burner domain detected (-50)")
		result.Breakdown.CorporateDomainScore = -50
	} else if extractor.IsFreeWebmail(normDomain) {
		result.NegativeFactors = append(result.NegativeFactors, "Personal free webmail address (0)")
		result.Breakdown.CorporateDomainScore = 0
	} else if normDomain != "" {
		result.Breakdown.CorporateDomainScore = weights.CorporateDomain
		result.PositiveFactors = append(result.PositiveFactors, "Verified corporate root domain: "+normDomain)
	}

	// 2. Seat Volume
	seats := lead.RequestedSeats
	switch {
	case seats >= 100:
		result.Breakdown.SeatVolumeScore = weights.SeatVolume
		result.PositiveFactors = append(result.PositiveFactors, "High enterprise seat volume (100+ seats)")
	case seats >= 25:
		result.Breakdown.SeatVolumeScore = int(float64(weights.SeatVolume) * 0.72) // ~18
		result.PositiveFactors = append(result.PositiveFactors, "Mid-market seat volume (25-99 seats)")
	case seats >= 10:
		result.Breakdown.SeatVolumeScore = int(float64(weights.SeatVolume) * 0.40) // ~10
		result.PositiveFactors = append(result.PositiveFactors, "Growth team seat volume (10-24 seats)")
	case seats >= 1:
		result.Breakdown.SeatVolumeScore = int(float64(weights.SeatVolume) * 0.16) // ~4
	default:
		result.NegativeFactors = append(result.NegativeFactors, "Seat volume unspecified")
	}

	// 3. Seniority
	switch lead.Seniority {
	case "C-Level / VP / Founder":
		result.Breakdown.SeniorityScore = weights.Seniority
		result.PositiveFactors = append(result.PositiveFactors, "Key decision maker: C-Level / VP / Founder")
	case "Director / Head":
		result.Breakdown.SeniorityScore = int(float64(weights.Seniority) * 0.75) // ~15
		result.PositiveFactors = append(result.PositiveFactors, "Senior leadership: Director / Head")
	case "Manager / Lead":
		result.Breakdown.SeniorityScore = int(float64(weights.Seniority) * 0.50) // ~10
		result.PositiveFactors = append(result.PositiveFactors, "Department lead / Manager")
	case "Individual Contributor":
		result.Breakdown.SeniorityScore = int(float64(weights.Seniority) * 0.25) // ~5
	default:
		result.Breakdown.SeniorityScore = 0
	}

	// 4. Timeline Urgency
	switch {
	case strings.Contains(lead.Urgency, "P0"):
		result.Breakdown.UrgencyScore = weights.Urgency
		result.PositiveFactors = append(result.PositiveFactors, "Immediate deployment timeline (P0 < 30 days)")
	case strings.Contains(lead.Urgency, "P1"):
		result.Breakdown.UrgencyScore = int(float64(weights.Urgency) * 0.67) // ~10
		result.PositiveFactors = append(result.PositiveFactors, "High urgency timeline (1-3 months)")
	case strings.Contains(lead.Urgency, "P2"):
		result.Breakdown.UrgencyScore = int(float64(weights.Urgency) * 0.33) // ~5
	case strings.Contains(lead.Urgency, "P3"):
		result.Breakdown.UrgencyScore = 2
		result.NegativeFactors = append(result.NegativeFactors, "Exploratory / long-term timeline")
	default:
		result.Breakdown.UrgencyScore = 5
	}

	// 5. Budget Viability
	switch {
	case lead.Budget >= 50000:
		result.Breakdown.BudgetScore = weights.Budget
		result.PositiveFactors = append(result.PositiveFactors, "Substantial enterprise budget allocated ($50k+)")
	case lead.Budget > 0:
		result.Breakdown.BudgetScore = int(float64(weights.Budget) * 0.60) // ~6
		result.PositiveFactors = append(result.PositiveFactors, "Defined commercial budget stated")
	default:
		result.Breakdown.BudgetScore = 0
	}

	// 6. Product Fit / Enterprise Features
	fitPoints := 0
	if len(lead.TechStack) > 0 {
		for _, item := range lead.TechStack {
			if item == "SOC2 Type II" || item == "HIPAA / BAA" || item == "Dedicated VPC" || item == "SAML SSO / Okta" {
				fitPoints += 5
			}
		}
	}
	if len(lead.RequestedProducts) > 0 {
		fitPoints += 5
	}
	if fitPoints > weights.ProductFit {
		fitPoints = weights.ProductFit
	}
	result.Breakdown.ProductFitScore = fitPoints
	if fitPoints >= 5 {
		result.PositiveFactors = append(result.PositiveFactors, "High technical alignment with enterprise capabilities")
	}

	// Total calculation with bounds [0, 100]
	total := result.Breakdown.CorporateDomainScore +
		result.Breakdown.SeatVolumeScore +
		result.Breakdown.SeniorityScore +
		result.Breakdown.UrgencyScore +
		result.Breakdown.BudgetScore +
		result.Breakdown.ProductFitScore

	if total < 0 {
		total = 0
	}
	if total > 100 {
		total = 100
	}
	result.TotalScore = total

	// Tier classification
	switch {
	case total >= 80:
		result.Tier = "Enterprise Hot Lead"
		result.RoutingRecommendation = "Priority dispatch: Assign immediately to Enterprise AE and trigger Slack alert"
	case total >= 60:
		result.Tier = "Mid-Market Qualified"
		result.RoutingRecommendation = "Standard dispatch: Assign to Commercial AE queue"
	case total >= 40:
		result.Tier = "SMB / Self-Serve"
		result.RoutingRecommendation = "Automated onboarding: Send product tour and trial activation link"
	default:
		result.Tier = "Unqualified / Low Intent"
		result.RoutingRecommendation = "Low priority queue: Do not alert live sales representatives"
	}

	return result
}
