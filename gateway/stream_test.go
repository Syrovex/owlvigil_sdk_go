package gateway_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	owlvigil "github.com/Syrovex/owlvigil_sdk_go"
	"github.com/Syrovex/owlvigil_sdk_go/gateway"
)

func TestCreateChatCompletionStream(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		if got := r.Header.Get("Accept"); got != "text/event-stream" {
			t.Fatalf("Accept = %q", got)
		}
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("event: message\ndata: {\"delta\":\"hi\"}\n\n"))
	}))
	defer server.Close()

	client := gateway.NewClient(owlvigil.WithBaseURL(server.URL), owlvigil.WithAPIKey("ov_sk_test"))
	stream, err := client.CreateChatCompletionStream(context.Background(), &gateway.ChatCompletionRequest{
		Model:    "gpt-4o-mini",
		Messages: []gateway.Message{{Role: "user", Content: "hi"}},
	})
	if err != nil {
		t.Fatalf("CreateChatCompletionStream returned error: %v", err)
	}
	defer stream.Close()

	if !stream.Next() {
		t.Fatalf("Next = false, err = %v", stream.Err())
	}
	event := stream.Current()
	if event.Event != "message" || string(event.Data) != "{\"delta\":\"hi\"}" {
		t.Fatalf("event = %+v", event)
	}
	if stream.Next() {
		t.Fatalf("unexpected second event")
	}
	if err := stream.Err(); err != nil {
		t.Fatalf("Err = %v", err)
	}
}

func TestCreateStreamHTTPError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"request_id":"req_stream","code":"unauthorized","message":"bad key"}`))
	}))
	defer server.Close()

	client := gateway.NewClient(owlvigil.WithBaseURL(server.URL), owlvigil.WithAPIKey("ov_sk_test"))
	_, err := client.CreateResponseStream(context.Background(), &gateway.ResponseRequest{Model: "gpt-4o-mini", Input: "hi"})
	var apiErr *owlvigil.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("err = %T, want APIError", err)
	}
	if apiErr.StatusCode != http.StatusUnauthorized || apiErr.RequestID != "req_stream" {
		t.Fatalf("apiErr = %+v", apiErr)
	}
}

func TestCreateStreamHTTPError_RedactsDynamicCredential(t *testing.T) {
	t.Parallel()

	const dynamicKey = "dynamic_gateway_key_123456"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"code":"invalid_dynamic_gateway_key_123456","message":"bad dynamic_gateway_key_123456"}`))
	}))
	t.Cleanup(server.Close)

	client := gateway.NewClient(
		owlvigil.WithBaseURL(server.URL),
		owlvigil.WithAPIKeyProvider(func(context.Context) (string, error) {
			return dynamicKey, nil
		}),
	)
	_, err := client.CreateResponseStream(
		context.Background(),
		&gateway.ResponseRequest{Model: "gpt-4o-mini", Input: "hi"},
	)
	var apiErr *owlvigil.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("CreateResponseStream() error = %T, want *owlvigil.APIError", err)
	}
	if strings.Contains(apiErr.Code, dynamicKey) ||
		strings.Contains(apiErr.Message, dynamicKey) ||
		strings.Contains(apiErr.Body, dynamicKey) {
		t.Errorf("CreateResponseStream() error = %+v, want dynamic credential redacted", apiErr)
	}
}

func TestCreateStream_NilRequestReturnsError(t *testing.T) {
	t.Parallel()

	client := gateway.NewClient(owlvigil.WithBaseURL("https://example.com"))
	tests := []struct {
		name string
		call func() error
	}{
		{
			name: "chat completion",
			call: func() error {
				_, err := client.CreateChatCompletionStream(context.Background(), nil)
				return err
			},
		},
		{
			name: "response",
			call: func() error {
				_, err := client.CreateResponseStream(context.Background(), nil)
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.call(); err == nil {
				t.Errorf("%s stream call error = nil, want non-nil error", tt.name)
			}
		})
	}
}
