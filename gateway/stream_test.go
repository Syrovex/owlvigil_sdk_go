package gateway_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	owlvigil "github.com/owlvigil/owlvigil-go"
	"github.com/owlvigil/owlvigil-go/gateway"
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
