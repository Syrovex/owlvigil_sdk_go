package oauth2_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	owlvigil "github.com/owlvigil/owlvigil-go"
	oauth2 "github.com/owlvigil/owlvigil-go/oauth2"
)

func TestTokenFlows(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("User-Agent"); got == "" {
			t.Fatalf("missing User-Agent")
		}
		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm: %v", err)
		}
		switch r.URL.Path {
		case "/oauth/token":
			switch r.Form.Get("grant_type") {
			case "authorization_code":
				if r.Form.Get("code") != "code_123" {
					t.Fatalf("form = %s", r.Form.Encode())
				}
				_, _ = w.Write([]byte(`{"access_token":"access_1","refresh_token":"refresh_1","token_type":"Bearer","expires_in":3600,"scope":"workspace:read"}`))
			case "client_credentials":
				if r.Form.Get("client_secret") != "secret_123" || r.Form.Get("scope") != "workspace:read gateway:read" {
					t.Fatalf("form = %s", r.Form.Encode())
				}
				_, _ = w.Write([]byte(`{"access_token":"machine_access_1","token_type":"Bearer","expires_in":3600,"scope":"workspace:read gateway:read"}`))
			default:
				t.Fatalf("form = %s", r.Form.Encode())
			}
		case "/oauth/token/refresh":
			if r.Form.Get("grant_type") != "refresh_token" || r.Form.Get("refresh_token") != "refresh_1" {
				t.Fatalf("form = %s", r.Form.Encode())
			}
			_, _ = w.Write([]byte(`{"access_token":"access_2","token_type":"Bearer","expires_in":1800}`))
		case "/oauth/revoke":
			if r.Form.Get("token") != "refresh_1" {
				t.Fatalf("form = %s", r.Form.Encode())
			}
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
	}))
	defer server.Close()

	client := oauth2.NewClient(owlvigil.WithBaseURL(server.URL))
	ctx := context.Background()
	token, err := client.Exchange(ctx, oauth2.TokenExchangeRequest{
		ClientID:    "client_123",
		Code:        "code_123",
		RedirectURI: "https://app.example.com/callback",
	})
	if err != nil {
		t.Fatalf("Exchange returned error: %v", err)
	}
	if token.AccessToken != "access_1" || token.ExpiresAt.IsZero() {
		t.Fatalf("token = %+v", token)
	}
	refreshed, err := client.Refresh(ctx, oauth2.RefreshTokenRequest{ClientID: "client_123", RefreshToken: "refresh_1"})
	if err != nil {
		t.Fatalf("Refresh returned error: %v", err)
	}
	if refreshed.AccessToken != "access_2" {
		t.Fatalf("refreshed = %+v", refreshed)
	}
	machineToken, err := client.ClientCredentials(ctx, oauth2.ClientCredentialsRequest{
		ClientID:     "client_123",
		ClientSecret: "secret_123",
		Scopes:       []string{"workspace:read", "gateway:read"},
	})
	if err != nil {
		t.Fatalf("ClientCredentials returned error: %v", err)
	}
	if machineToken.AccessToken != "machine_access_1" || machineToken.RefreshToken != "" {
		t.Fatalf("machine token = %+v", machineToken)
	}
	if err := client.Revoke(ctx, "refresh_1"); err != nil {
		t.Fatalf("Revoke returned error: %v", err)
	}
}

func TestTokenOAuthErrorRedactsSecrets(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"invalid_client","error_description":"secret client_secret_123456 is invalid"}`))
	}))
	defer server.Close()

	client := oauth2.NewClient(owlvigil.WithBaseURL(server.URL))
	_, err := client.Exchange(context.Background(), oauth2.TokenExchangeRequest{
		ClientID:     "client_123",
		ClientSecret: "client_secret_123456",
		Code:         "code_123",
		RedirectURI:  "https://app.example.com/callback",
	})
	var oauthErr *owlvigil.OAuthError
	if !errors.As(err, &oauthErr) {
		t.Fatalf("err = %T, want OAuthError", err)
	}
	if oauthErr.ErrorDescription == "secret client_secret_123456 is invalid" {
		t.Fatalf("secret was not redacted: %+v", oauthErr)
	}
}
