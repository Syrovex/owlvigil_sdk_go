package main

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	owlvigil "github.com/Syrovex/owlvigil_sdk_go"
	"github.com/Syrovex/owlvigil_sdk_go/gateway"
)

func TestRunListsModelsAndWritesIDs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Method, http.MethodGet; got != want {
			t.Errorf("models request method = %q, want %q", got, want)
		}
		if got, want := r.URL.Path, "/v1/models"; got != want {
			t.Errorf("models request path = %q, want %q", got, want)
		}
		if got, want := r.Header.Get("Authorization"), "Bearer gateway_test_key"; got != want {
			t.Errorf("models request authorization = %q, want %q", got, want)
		}
		_, _ = w.Write([]byte(`{"object":"list","data":[{"id":"model-a"},{"id":"model-b"}]}`))
	}))
	t.Cleanup(server.Close)

	client := gateway.NewClient(
		owlvigil.WithBaseURL(server.URL),
		owlvigil.WithAPIKey("gateway_test_key"),
		owlvigil.WithoutRetry(),
	)
	var output bytes.Buffer
	if err := run(context.Background(), client, &output); err != nil {
		t.Fatalf("run() error = %v", err)
	}
	if got, want := output.String(), "model-a\nmodel-b\n"; got != want {
		t.Errorf("run() output = %q, want %q", got, want)
	}
}
