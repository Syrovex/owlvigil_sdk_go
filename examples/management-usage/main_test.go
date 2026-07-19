package main

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	owlvigil "github.com/Syrovex/owlvigil_sdk_go"
	"github.com/Syrovex/owlvigil_sdk_go/management"
)

func TestRunUsesConfiguredWorkspaceForUsageSummary(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/workspaces":
			if got, want := r.URL.Query().Get("limit"), "1"; got != want {
				t.Errorf("ListWorkspaces limit = %q, want %q", got, want)
			}
			_, _ = w.Write([]byte(`{"items":[{"id":7,"name":"default"}],"page_info":{}}`))
		case "/gateway/usage/summary":
			if got, want := r.URL.Query().Get("workspace_id"), "42"; got != want {
				t.Errorf("GetUsageSummary workspace_id = %q, want %q", got, want)
			}
			_, _ = w.Write([]byte(`{"requests":3,"tokens":9,"cost":1.5}`))
		default:
			t.Errorf("request path = %q, want a management usage path", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(server.Close)

	client := management.NewClient(
		owlvigil.WithBaseURL(server.URL),
		owlvigil.WithAPIKey("management_test_key"),
		owlvigil.WithoutRetry(),
	)
	var output bytes.Buffer
	if err := run(context.Background(), client, "42", &output); err != nil {
		t.Fatalf("run() error = %v", err)
	}
	if got, want := output.String(), "workspace_id=42 requests=3 tokens=9 cost=1.5\n"; got != want {
		t.Errorf("run() output = %q, want %q", got, want)
	}
}
