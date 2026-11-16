package extractor

import (
	"strings"
)

var freeWebmailDomains = map[string]bool{
	"gmail.com":       true,
	"googlemail.com":  true,
	"yahoo.com":       true,
	"yahoo.co.uk":     true,
	"hotmail.com":     true,
	"outlook.com":     true,
	"live.com":        true,
	"msn.com":         true,
	"icloud.com":      true,
	"me.com":          true,
	"mac.com":         true,
	"aol.com":         true,
	"proton.me":       true,
	"protonmail.com":  true,
	"zoho.com":        true,
	"mail.com":        true,
	"gmx.com":         true,
	"yandex.com":      true,
	"yandex.ru":       true,
	"mail.ru":         true,
}

var disposableDomains = map[string]bool{
	"tempmail.com":         true,
	"temp-mail.org":        true,
	"10minutemail.com":     true,
	"guerrillamail.com":    true,
	"mailinator.com":       true,
	"throwawaymail.com":    true,
	"yopmail.com":          true,
	"trashmail.com":        true,
	"sharklasers.com":      true,
	"dispostable.com":      true,
	"fakeinbox.com":        true,
	"burnermail.io":        true,
	"getnada.com":          true,
}

// IsFreeWebmail checks if the domain belongs to a consumer email provider.
func IsFreeWebmail(domain string) bool {
	return freeWebmailDomains[strings.ToLower(strings.TrimSpace(domain))]
}

// IsDisposableEmail checks if the domain belongs to a temporary or burner email provider.
func IsDisposableEmail(domain string) bool {
	return disposableDomains[strings.ToLower(strings.TrimSpace(domain))]
}

// ExtractDomainFromEmail returns the domain part of an email address.
func ExtractDomainFromEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) == 2 {
		return strings.ToLower(strings.TrimSpace(parts[1]))
	}
	return ""
}
