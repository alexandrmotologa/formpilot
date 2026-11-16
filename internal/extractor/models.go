package extractor

type ExtractedLead struct {
	CompanyName       string   `json:"company_name"`
	Domain            string   `json:"domain"`
	ContactName       string   `json:"contact_name"`
	JobTitle          string   `json:"job_title"`
	Seniority         string   `json:"seniority"` // "C-Level / VP / Founder", "Director / Head", "Manager / Lead", "Individual Contributor", "Unspecified"
	ContactEmail      string   `json:"contact_email"`
	ContactPhone      string   `json:"contact_phone"`
	RequestedProducts []string `json:"requested_products"`
	RequestedSeats    int      `json:"requested_seats"`
	Budget            float64  `json:"budget"`
	Currency          string   `json:"currency"`
	Urgency           string   `json:"urgency"` // "P0 Immediate (< 1 mo)", "P1 High (1-3 mos)", "P2 Medium (3-6 mos)", "P3 Exploratory (> 6 mos)"
	TechStack         []string `json:"tech_stack"`
	PainPoints        string   `json:"pain_points"`
	ConfidenceScore   float64  `json:"confidence_score"`
	ExtractionMethod  string   `json:"extraction_method"` // "llm" or "heuristic"
}
