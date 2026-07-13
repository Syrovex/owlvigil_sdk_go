package owlvigil_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	owlvigil "github.com/owlvigil/owlvigil-go"
)

func TestDefaultConfig(t *testing.T) {
	t.Parallel()

	cfg := owlvigil.DefaultConfig(owlvigil.DefaultGatewayBaseURL)

	if cfg.BaseURL != owlvigil.DefaultGatewayBaseURL {
		t.Fatalf("BaseURL = %q, want %q", cfg.BaseURL, owlvigil.DefaultGatewayBaseURL)
	}
	if cfg.HTTPClient == nil {
		t.Fatalf("HTTPClient is nil")
	}
	if cfg.HTTPClient == http.DefaultClient {
		t.Fatalf("HTTPClient should not share http.DefaultClient")
	}
	if cfg.HTTPClient.Timeout != owlvigil.DefaultHTTPTimeout {
		t.Fatalf("HTTPClient.Timeout = %s, want %s", cfg.HTTPClient.Timeout, owlvigil.DefaultHTTPTimeout)
	}
	if cfg.UserAgent != owlvigil.UserAgent() {
		t.Fatalf("UserAgent = %q, want %q", cfg.UserAgent, owlvigil.UserAgent())
	}
}

func TestOptions(t *testing.T) {
	t.Parallel()

	cfg := owlvigil.DefaultConfig(owlvigil.DefaultGatewayBaseURL)
	client := &http.Client{Timeout: time.Second}
	apiKeyProvider := func(context.Context) (string, error) { return "dynamic_api_key", nil }
	accessTokenProvider := func(context.Context) (string, error) { return "dynamic_access_token", nil }

	owlvigil.WithBaseURL("https://private.example.com")(&cfg)
	owlvigil.WithHTTPClient(client)(&cfg)
	owlvigil.WithUserAgent("test-agent")(&cfg)
	owlvigil.WithAPIKey("ov_sk_test")(&cfg)
	owlvigil.WithAPIKeyProvider(apiKeyProvider)(&cfg)
	owlvigil.WithAccessToken("access_test")(&cfg)
	owlvigil.WithAccessTokenProvider(accessTokenProvider)(&cfg)
	owlvigil.WithRetry(3, 10*time.Millisecond)(&cfg)

	if cfg.BaseURL != "https://private.example.com" {
		t.Fatalf("BaseURL = %q", cfg.BaseURL)
	}
	if cfg.HTTPClient != client {
		t.Fatalf("HTTPClient was not configured")
	}
	if cfg.UserAgent != "test-agent" || cfg.APIKey != "ov_sk_test" || cfg.AccessToken != "access_test" {
		t.Fatalf("auth or user agent options were not configured: %+v", cfg)
	}
	if token, err := cfg.APIKeyProvider(context.Background()); err != nil || token != "dynamic_api_key" {
		t.Fatalf("APIKeyProvider = %q, %v", token, err)
	}
	if token, err := cfg.AccessTokenProvider(context.Background()); err != nil || token != "dynamic_access_token" {
		t.Fatalf("AccessTokenProvider = %q, %v", token, err)
	}
	if cfg.RetryMax != 3 || cfg.RetryWait != 10*time.Millisecond {
		t.Fatalf("retry = (%d, %s)", cfg.RetryMax, cfg.RetryWait)
	}
	owlvigil.WithoutRetry()(&cfg)
	if cfg.RetryMax != 0 {
		t.Fatalf("RetryMax = %d, want 0", cfg.RetryMax)
	}
}

func TestWithTimeoutClonesClient(t *testing.T) {
	t.Parallel()

	base := &http.Client{}
	cfg := owlvigil.DefaultConfig(owlvigil.DefaultGatewayBaseURL)
	owlvigil.WithHTTPClient(base)(&cfg)
	owlvigil.WithTimeout(5 * time.Second)(&cfg)

	if cfg.HTTPClient == base {
		t.Fatalf("WithTimeout should clone the configured client")
	}
	if cfg.HTTPClient.Timeout != 5*time.Second {
		t.Fatalf("Timeout = %s, want 5s", cfg.HTTPClient.Timeout)
	}
	if base.Timeout != 0 {
		t.Fatalf("base client was mutated")
	}
}

func TestRequestOptions(t *testing.T) {
	t.Parallel()

	var cfg owlvigil.RequestConfig
	owlvigil.WithIdempotencyKey("idem_1")(&cfg)
	owlvigil.WithHeader("X-Test", "value")(&cfg)
	owlvigil.WithQueryParam("expand", "usage")(&cfg)

	if cfg.IdempotencyKey != "idem_1" {
		t.Fatalf("IdempotencyKey = %q", cfg.IdempotencyKey)
	}
	if got := cfg.Headers.Get("X-Test"); got != "value" {
		t.Fatalf("header = %q", got)
	}
	if got := cfg.Query["expand"]; got != "usage" {
		t.Fatalf("query = %q", got)
	}
}
