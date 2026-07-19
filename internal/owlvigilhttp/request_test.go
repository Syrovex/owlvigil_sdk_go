package owlvigilhttp

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	owlvigil "github.com/owlvigil/owlvigil-go"
)

func TestClientDoHeadersEnvelopeAndRetry(t *testing.T) {
	t.Parallel()

	var attempts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempt := attempts.Add(1)
		if r.Method != http.MethodPost || r.URL.Path != "/v1/test" {
			t.Fatalf("request = %s %s", r.Method, r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer ov_sk_dynamic" {
			t.Fatalf("Authorization = %q", got)
		}
		if got := r.Header.Get("User-Agent"); got != "test-agent" {
			t.Fatalf("User-Agent = %q", got)
		}
		if got := r.Header.Get("Idempotency-Key"); got != "idem_123" {
			t.Fatalf("Idempotency-Key = %q", got)
		}
		if attempt == 1 {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"request_id":"req_retry","code":"unavailable","message":"try again"}`))
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"request_id": "req_ok",
			"code":       "ok",
			"message":    "ok",
			"data": map[string]any{
				"id": "done",
			},
		})
	}))
	defer server.Close()

	cfg := owlvigil.DefaultConfig(server.URL)
	cfg.UserAgent = "test-agent"
	cfg.RetryMax = 1
	cfg.RetryWait = time.Millisecond
	cfg.APIKeyProvider = func(context.Context) (string, error) {
		return "ov_sk_dynamic", nil
	}
	client := New(cfg)

	var out struct {
		ID string `json:"id"`
	}
	meta, err := client.Do(context.Background(), http.MethodPost, "/v1/test", nil, map[string]string{"ping": "pong"}, &out, owlvigil.WithIdempotencyKey("idem_123"))
	if err != nil {
		t.Fatalf("Do returned error: %v", err)
	}
	if attempts.Load() != 2 {
		t.Fatalf("attempts = %d, want 2", attempts.Load())
	}
	if out.ID != "done" || meta.RequestID != "req_ok" {
		t.Fatalf("out=%+v meta=%+v", out, meta)
	}
}

func TestClientDo_DoesNotRetryMutationWithoutIdempotencyKey(t *testing.T) {
	t.Parallel()

	var attempts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts.Add(1)
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte(`{"code":"unavailable","message":"try again"}`))
	}))
	defer server.Close()

	cfg := owlvigil.DefaultConfig(server.URL)
	cfg.RetryMax = 2
	cfg.RetryWait = time.Millisecond
	client := New(cfg)

	_, err := client.Do(context.Background(), http.MethodPost, "/billing/subscription/in-app", nil, map[string]string{"plan_id": "pro"}, nil)
	if err == nil {
		t.Fatal("Do() error = nil, want service unavailable")
	}
	if got, want := attempts.Load(), int32(1); got != want {
		t.Fatalf("attempts = %d, want %d for a mutation without an idempotency key", got, want)
	}
}

func TestClientDoAPIErrorRedactsSecret(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"request_id":"req_bad","code":"bad_secret","message":"token ov_sk_123456789 leaked"}`))
	}))
	defer server.Close()

	cfg := owlvigil.DefaultConfig(server.URL)
	cfg.APIKey = "ov_sk_123456789"
	client := New(cfg)

	_, err := client.Do(context.Background(), http.MethodGet, "/bad", nil, nil, nil)
	var apiErr *owlvigil.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("err = %T, want APIError", err)
	}
	if strings.Contains(apiErr.Message, cfg.APIKey) || strings.Contains(apiErr.Body, cfg.APIKey) {
		t.Fatalf("secret was not redacted: message=%q body=%q", apiErr.Message, apiErr.Body)
	}
	if apiErr.RequestID != "req_bad" || apiErr.Code != "bad_secret" {
		t.Fatalf("apiErr = %+v", apiErr)
	}
}

func TestClientDoContextCanceled(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer server.Close()

	cfg := owlvigil.DefaultConfig(server.URL)
	cfg.RetryMax = 0
	client := New(cfg)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := client.Do(ctx, http.MethodGet, "/slow", nil, nil, nil)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want context.Canceled", err)
	}
}

func TestClientDoProviderError(t *testing.T) {
	t.Parallel()

	cfg := owlvigil.DefaultConfig("https://example.com")
	cfg.AccessTokenProvider = func(context.Context) (string, error) {
		return "", errors.New("no token")
	}
	client := New(cfg)

	_, err := client.Do(context.Background(), http.MethodGet, "/test", nil, nil, nil)
	if err == nil || !strings.Contains(err.Error(), "access token provider") {
		t.Fatalf("err = %v", err)
	}
}

func TestClientDoEncodeAndRequestErrors(t *testing.T) {
	t.Parallel()

	client := New(owlvigil.DefaultConfig("://bad-url"))
	_, err := client.Do(context.Background(), http.MethodGet, "/test", nil, nil, nil)
	if err == nil {
		t.Fatalf("expected invalid URL error")
	}

	client = New(owlvigil.DefaultConfig("https://example.com"))
	_, err = client.Do(context.Background(), http.MethodPost, "/test", nil, func() {}, nil)
	if err == nil || !strings.Contains(err.Error(), "unsupported type") {
		t.Fatalf("err = %v", err)
	}
}

func TestStreamHTTPClientClearsTimeoutOnClone(t *testing.T) {
	t.Parallel()

	cfg := owlvigil.DefaultConfig("https://example.com")
	original := cfg.HTTPClient
	client := New(cfg)

	streamClient := client.StreamHTTPClient()
	if streamClient == original {
		t.Fatalf("StreamHTTPClient should clone clients with timeouts")
	}
	if streamClient.Timeout != 0 {
		t.Fatalf("stream timeout = %s, want 0", streamClient.Timeout)
	}
	if original.Timeout != owlvigil.DefaultHTTPTimeout {
		t.Fatalf("original timeout = %s, want %s", original.Timeout, owlvigil.DefaultHTTPTimeout)
	}
}

func TestClientAccessorsAndStreamRequest(t *testing.T) {
	t.Parallel()

	cfg := owlvigil.DefaultConfig("https://example.com/api")
	cfg.APIKey = "ov_sk_test"
	client := New(cfg)

	if client.BaseURL() != "https://example.com/api" {
		t.Fatalf("BaseURL = %q", client.BaseURL())
	}
	if client.Config().APIKey != "ov_sk_test" {
		t.Fatalf("Config APIKey = %q", client.Config().APIKey)
	}
	if client.HTTPClient() != cfg.HTTPClient {
		t.Fatalf("HTTPClient was not preserved")
	}

	req, err := client.NewStreamRequest(
		context.Background(),
		http.MethodPost,
		"/v1/stream",
		map[string]string{"model": "gpt-4o-mini"},
		owlvigil.WithHeader("X-Test", "value"),
		owlvigil.WithQueryParam("stream", "true"),
	)
	if err != nil {
		t.Fatalf("NewStreamRequest returned error: %v", err)
	}
	if req.URL.String() != "https://example.com/api/v1/stream?stream=true" {
		t.Fatalf("url = %s", req.URL.String())
	}
	if got := req.Header.Get("Authorization"); got != "Bearer ov_sk_test" {
		t.Fatalf("Authorization = %q", got)
	}
	if got := req.Header.Get("X-Test"); got != "value" {
		t.Fatalf("X-Test = %q", got)
	}
	if got := req.Header.Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type = %q", got)
	}
}

func TestClientNewDefaultsAndHelpers(t *testing.T) {
	t.Parallel()

	client := New(owlvigil.Config{BaseURL: "https://example.com/base"})
	if client.HTTPClient() == nil {
		t.Fatalf("HTTPClient is nil")
	}
	if client.Config().UserAgent == "" {
		t.Fatalf("UserAgent was not defaulted")
	}
	if client.StreamHTTPClient() != client.HTTPClient() {
		t.Fatalf("StreamHTTPClient should reuse clients without timeout")
	}

	body, err := encodeBody([]byte(`{"raw":true}`))
	if err != nil || string(body) != `{"raw":true}` {
		t.Fatalf("encodeBody bytes = %q, %v", body, err)
	}

	u, err := joinURL("https://example.com/base", "https://other.example.com/absolute")
	if err != nil || u.String() != "https://other.example.com/absolute" {
		t.Fatalf("joinURL absolute = %s, %v", u, err)
	}

	merged := mergeQuery(nil, map[string]string{"a": "b"})
	if merged.Get("a") != "b" {
		t.Fatalf("merged query = %s", merged.Encode())
	}

	base := url.Values{"keep": []string{"one", "two"}, "override": []string{"old"}}
	merged = mergeQuery(base, map[string]string{"override": "new"})
	if got := merged["keep"]; len(got) != 2 || got[0] != "one" || got[1] != "two" {
		t.Fatalf("keep values = %#v", got)
	}
	if got := merged.Get("override"); got != "new" {
		t.Fatalf("override = %q", got)
	}
	if got := base.Get("override"); got != "old" {
		t.Fatalf("base mutated, override = %q", got)
	}
}
