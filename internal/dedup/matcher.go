package dedup

import (
	"strings"

	"github.com/alexandrmotologa/formpilot/internal/extractor"
	"github.com/alexandrmotologa/formpilot/internal/storage"
)

type MatchResult struct {
	IsMatched       bool                   `json:"is_matched"`
	MatchStrategy   string                 `json:"match_strategy"` // "exact_domain", "root_domain", "fuzzy_name", "none"
	MatchConfidence float64                `json:"match_confidence"`
	MatchedAccount  *storage.AccountRecord `json:"matched_account,omitempty"`
	Reason          string                 `json:"reason"`
}

type Matcher struct {
	cache *AccountCache
}

func NewMatcher(cache *AccountCache) *Matcher {
	return &Matcher{cache: cache}
}

// Match evaluates an extracted lead against known CRM accounts.
func (m *Matcher) Match(lead *extractor.ExtractedLead) *MatchResult {
	if lead == nil {
		return &MatchResult{MatchStrategy: "none", Reason: "empty lead"}
	}

	normDomain := extractor.NormalizeDomain(lead.Domain)

	// 1. Exact Domain Match (Confidence 1.0)
	if normDomain != "" && !extractor.IsFreeWebmail(normDomain) && !extractor.IsDisposableEmail(normDomain) {
		if acc := m.cache.GetByDomain(normDomain); acc != nil {
			return &MatchResult{
				IsMatched:       true,
				MatchStrategy:   "exact_domain",
				MatchConfidence: 1.0,
				MatchedAccount:  acc,
				Reason:          "Exact corporate domain match: " + normDomain,
			}
		}

		// Check root domain if subdomain was provided e.g. eu.apexclearing.com -> apexclearing.com
		parts := strings.Split(normDomain, ".")
		if len(parts) > 2 {
			rootDomain := strings.Join(parts[len(parts)-2:], ".")
			if acc := m.cache.GetByDomain(rootDomain); acc != nil {
				return &MatchResult{
					IsMatched:       true,
					MatchStrategy:   "root_domain",
					MatchConfidence: 0.95,
					MatchedAccount:  acc,
					Reason:          "Root domain match: " + rootDomain,
				}
			}
		}
	}

	// 2. Fuzzy Company Name Match
	cleanLeadCompany := extractor.CleanCompanyName(lead.CompanyName)
	if cleanLeadCompany != "" && len(cleanLeadCompany) >= 3 {
		allAccounts := m.cache.GetAll()
		var bestMatch *storage.AccountRecord
		bestScore := 0.0

		for _, acc := range allAccounts {
			cleanAccCompany := extractor.CleanCompanyName(acc.CompanyName)
			if cleanAccCompany == "" {
				continue
			}

			// Direct equality after cleaning
			if strings.EqualFold(cleanLeadCompany, cleanAccCompany) {
				return &MatchResult{
					IsMatched:       true,
					MatchStrategy:   "exact_name",
					MatchConfidence: 0.92,
					MatchedAccount:  &acc,
					Reason:          "Exact normalized company name match: " + acc.CompanyName,
				}
			}

			score := NameSimilarity(cleanLeadCompany, cleanAccCompany)
			if score > bestScore {
				bestScore = score
				accCopy := acc
				bestMatch = &accCopy
			}
		}

		if bestScore >= 0.80 && bestMatch != nil {
			return &MatchResult{
				IsMatched:       true,
				MatchStrategy:   "fuzzy_name",
				MatchConfidence: bestScore,
				MatchedAccount:  bestMatch,
				Reason:          "Fuzzy name match (" + bestMatch.CompanyName + ")",
			}
		}
	}

	return &MatchResult{
		IsMatched:       false,
		MatchStrategy:   "none",
		MatchConfidence: 0.0,
		Reason:          "No existing CRM account found; provision new account",
	}
}

// NameSimilarity calculates Levenshtein similarity ratio between 0.0 and 1.0.
func NameSimilarity(s1, s2 string) float64 {
	s1 = strings.ToLower(strings.TrimSpace(s1))
	s2 = strings.ToLower(strings.TrimSpace(s2))

	if s1 == s2 {
		return 1.0
	}
	if len(s1) == 0 || len(s2) == 0 {
		return 0.0
	}

	distance := LevenshteinDistance(s1, s2)
	maxLen := len(s1)
	if len(s2) > maxLen {
		maxLen = len(s2)
	}

	return 1.0 - (float64(distance) / float64(maxLen))
}

// LevenshteinDistance computes minimum edit distance.
func LevenshteinDistance(a, b string) int {
	ra, rb := []rune(a), []rune(b)
	la, lb := len(ra), len(rb)

	d := make([][]int, la+1)
	for i := range d {
		d[i] = make([]int, lb+1)
		d[i][0] = i
	}
	for j := 0; j <= lb; j++ {
		d[0][j] = j
	}

	for i := 1; i <= la; i++ {
		for j := 1; j <= lb; j++ {
			cost := 0
			if ra[i-1] != rb[j-1] {
				cost = 1
			}
			d[i][j] = min3(
				d[i-1][j]+1,      // deletion
				d[i][j-1]+1,      // insertion
				d[i-1][j-1]+cost, // substitution
			)
		}
	}

	return d[la][lb]
}

func min3(a, b, c int) int {
	if a <= b && a <= c {
		return a
	}
	if b <= a && b <= c {
		return b
	}
	return c
}
