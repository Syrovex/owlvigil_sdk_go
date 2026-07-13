package oauth2_test

import (
	"net/url"
	"testing"

	owlvigil "github.com/owlvigil/owlvigil-go"
	oauth2 "github.com/owlvigil/owlvigil-go/oauth2"
)

func TestAuthorizationURL(t *testing.T) {
	t.Parallel()

	client := oauth2.NewClient()
	if client.BaseURL() != owlvigil.DefaultOAuthBaseURL {
		t.Fatalf("BaseURL = %q", client.BaseURL())
	}
	rawURL, err := client.AuthorizationURL(oauth2.AuthCodeOptions{
		ClientID:             "client_123",
		RedirectURI:          "https://app.example.com/callback",
		Scopes:               []string{"workspace:read", "gateway:write"},
		State:                "state_123",
		CodeChallenge:        "challenge",
		CodeChallengeMethod:  "S256",
		AdditionalParameters: url.Values{"prompt": []string{"consent"}},
	})
	if err != nil {
		t.Fatalf("AuthorizationURL returned error: %v", err)
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		t.Fatalf("parse url: %v", err)
	}
	if parsed.Scheme+"://"+parsed.Host != owlvigil.DefaultOAuthBaseURL || parsed.Path != "/oauth/authorize" {
		t.Fatalf("url = %s", rawURL)
	}
	q := parsed.Query()
	if q.Get("response_type") != "code" || q.Get("client_id") != "client_123" || q.Get("scope") != "workspace:read gateway:write" {
		t.Fatalf("query = %s", parsed.RawQuery)
	}
	if q.Get("code_challenge") != "challenge" || q.Get("code_challenge_method") != "S256" || q.Get("prompt") != "consent" {
		t.Fatalf("query = %s", parsed.RawQuery)
	}
}
