package oauth2_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	owlvigil "github.com/Syrovex/owlvigil_sdk_go"
	oauth2 "github.com/Syrovex/owlvigil_sdk_go/oauth2"
)

func TestUserInfo(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/oauth/userinfo" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer access_1" {
			t.Fatalf("Authorization = %q", got)
		}
		_, _ = w.Write([]byte(`{"sub":"user_1","email":"dev@example.com","name":"Dev"}`))
	}))
	defer server.Close()

	client := oauth2.NewClient(owlvigil.WithBaseURL(server.URL))
	user, err := client.UserInfo(context.Background(), "access_1")
	if err != nil {
		t.Fatalf("UserInfo returned error: %v", err)
	}
	if user.Subject != "user_1" || user.Email != "dev@example.com" {
		t.Fatalf("user = %+v", user)
	}
}

func TestUserInfoError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"invalid_token"}`))
	}))
	defer server.Close()

	client := oauth2.NewClient(owlvigil.WithBaseURL(server.URL))
	_, err := client.UserInfo(context.Background(), "bad")
	var oauthErr *owlvigil.OAuthError
	if !errors.As(err, &oauthErr) {
		t.Fatalf("err = %T, want OAuthError", err)
	}
	if oauthErr.StatusCode != http.StatusUnauthorized {
		t.Fatalf("oauthErr = %+v", oauthErr)
	}
}

func TestUserInfo_RejectsOversizedResponse(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(strings.Repeat("x", 2<<20)))
	}))
	t.Cleanup(server.Close)

	client := oauth2.NewClient(owlvigil.WithBaseURL(server.URL))
	_, err := client.UserInfo(context.Background(), "access_token_123456")
	if err == nil || !strings.Contains(err.Error(), "exceeds") {
		t.Errorf("UserInfo(oversized response) error = %v, want oversized-response error", err)
	}
}
