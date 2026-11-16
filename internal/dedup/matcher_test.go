package dedup

import (
	"testing"

	"github.com/alexandrmotologa/formpilot/internal/extractor"
	"github.com/alexandrmotologa/formpilot/internal/storage"
)

func TestMatcher(t *testing.T) {
	cache, _ := NewAccountCache(nil)

	// Preload cache with known account
	_ = cache.Upsert(&storage.AccountRecord{
		ID:          "acc_cyberdyne",
		Domain:      "cyberdyne.systems",
		CompanyName: "Cyberdyne Systems Corp",
		OwnerName:   "John Connor",
		OwnerEmail:  "john@sales.formpilot.io",
		CRMID:       "hs_comp_123",
		CRMType:     "hubspot",
	})

	matcher := NewMatcher(cache)

	// 1. Exact Domain Match
	lead1 := &extractor.ExtractedLead{
		CompanyName: "Cyberdyne",
		Domain:      "cyberdyne.systems",
	}
	res1 := matcher.Match(lead1)
	if !res1.IsMatched || res1.MatchStrategy != "exact_domain" {
		t.Fatalf("expected exact_domain match, got %+v", res1)
	}
	if res1.MatchedAccount.CRMID != "hs_comp_123" {
		t.Errorf("expected crm id hs_comp_123, got %s", res1.MatchedAccount.CRMID)
	}

	// 2. Subdomain / Root Domain Match
	lead2 := &extractor.ExtractedLead{
		CompanyName: "Cyberdyne Systems EU",
		Domain:      "eu.cyberdyne.systems",
	}
	res2 := matcher.Match(lead2)
	if !res2.IsMatched || res2.MatchStrategy != "root_domain" {
		t.Fatalf("expected root_domain match, got %+v", res2)
	}

	// 3. Fuzzy Name Match without domain
	lead3 := &extractor.ExtractedLead{
		CompanyName: "Cyberdyne Systems Inc",
		Domain:      "",
	}
	res3 := matcher.Match(lead3)
	if !res3.IsMatched {
		t.Fatalf("expected fuzzy or exact_name match, got %+v", res3)
	}

	// 4. Completely New Account
	lead4 := &extractor.ExtractedLead{
		CompanyName: "Stark Industries",
		Domain:      "starkindustries.com",
	}
	res4 := matcher.Match(lead4)
	if res4.IsMatched {
		t.Fatalf("expected no match for stark industries, got %+v", res4)
	}
}
