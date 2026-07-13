package owlvigilhttp

import "strings"

// Redact returns text with common credential values masked.
func Redact(text string, secrets ...string) string {
	redacted := text
	for _, secret := range secrets {
		if len(secret) < 6 {
			continue
		}
		redacted = strings.ReplaceAll(redacted, secret, mask(secret))
	}
	return redacted
}

func mask(secret string) string {
	if len(secret) <= 8 {
		return "****"
	}
	return secret[:4] + "..." + secret[len(secret)-4:]
}
