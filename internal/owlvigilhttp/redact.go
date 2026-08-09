// Package owlvigilhttp provides shared HTTP transport behavior for the SDK.
package owlvigilhttp

import (
	"encoding/json"
	"net/http"
	"strings"
)

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

func requestHeaderSecrets(req *http.Request) []string {
	if req == nil {
		return nil
	}
	var secrets []string
	for _, name := range []string{"Authorization", "X-API-Key", "API-Key"} {
		for _, value := range req.Header.Values(name) {
			value = strings.TrimSpace(value)
			if name == "Authorization" {
				if _, token, ok := strings.Cut(value, " "); ok {
					value = strings.TrimSpace(token)
				}
			}
			if value != "" {
				secrets = append(secrets, value)
			}
		}
	}
	return secrets
}

func sensitiveJSONValues(body []byte) []string {
	if len(body) == 0 {
		return nil
	}
	var value any
	if json.Unmarshal(body, &value) != nil {
		return nil
	}
	var secrets []string
	collectSensitiveJSONValues(value, "", &secrets)
	return secrets
}

func collectSensitiveJSONValues(value any, key string, secrets *[]string) {
	switch typed := value.(type) {
	case map[string]any:
		for childKey, child := range typed {
			collectSensitiveJSONValues(child, childKey, secrets)
		}
	case []any:
		for _, child := range typed {
			collectSensitiveJSONValues(child, key, secrets)
		}
	case string:
		if isSensitiveJSONKey(key) && typed != "" {
			*secrets = append(*secrets, typed)
		}
	}
}

func isSensitiveJSONKey(key string) bool {
	normalized := strings.NewReplacer("_", "", "-", "").Replace(strings.ToLower(key))
	switch normalized {
	case "apikey",
		"accesstoken",
		"refreshtoken",
		"clientsecret",
		"webhooksecret",
		"password",
		"oldpassword",
		"newpassword",
		"secret",
		"token",
		"codeverifier":
		return true
	default:
		return false
	}
}
