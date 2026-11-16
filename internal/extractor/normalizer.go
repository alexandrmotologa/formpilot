package extractor

import (
	"net/url"
	"regexp"
	"strings"
)

// NormalizeDomain cleans URLs or domain strings into canonical root domains.
func NormalizeDomain(raw string) string {
	raw = strings.TrimSpace(strings.ToLower(raw))
	if raw == "" {
		return ""
	}

	// Remove common wrapping characters e.g. (example.com) or [example.com]
	raw = strings.Trim(raw, "()[]<>,;\"'")

	// Prepend http if missing to allow url.Parse to extract host properly
	if !strings.HasPrefix(raw, "http://") && !strings.HasPrefix(raw, "https://") {
		raw = "http://" + raw
	}

	u, err := url.Parse(raw)
	if err != nil {
		return ""
	}

	host := u.Hostname()
	host = strings.TrimPrefix(host, "www.")
	host = strings.TrimPrefix(host, "mail.")
	host = strings.TrimPrefix(host, "portal.")

	return strings.ToLower(host)
}

// CleanCompanyName standardizes company names for fuzzy comparison.
func CleanCompanyName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}

	// Remove common legal entity suffixes
	reSuffix := regexp.MustCompile(`(?i)\b(inc|incorporated|corp|corporation|llc|l\.l\.c\.|ltd|limited|gmbh|co|company|holdings|group)\b[.]?`)
	cleaned := reSuffix.ReplaceAllString(name, "")

	// Remove punctuation and extra whitespace
	rePunct := regexp.MustCompile(`[,\-_.]`)
	cleaned = rePunct.ReplaceAllString(cleaned, " ")

	reSpaces := regexp.MustCompile(`\s+`)
	cleaned = reSpaces.ReplaceAllString(cleaned, " ")

	return strings.TrimSpace(cleaned)
}

// NormalizePhone cleans phone strings to standardized E.164-like formatting.
func NormalizePhone(phone string) string {
	re := regexp.MustCompile(`[^\d+]`)
	digits := re.ReplaceAllString(phone, "")
	if len(digits) == 10 && !strings.HasPrefix(digits, "+") {
		return "+1" + digits
	}
	if len(digits) == 11 && strings.HasPrefix(digits, "1") {
		return "+" + digits
	}
	return digits
}
