package management_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	owlvigil "github.com/owlvigil/owlvigil-go"
	"github.com/owlvigil/owlvigil-go/management"
)

func TestClientDefaults(t *testing.T) {
	t.Parallel()

	client := management.NewClient()

	if client.BaseURL() != owlvigil.DefaultManagementBaseURL {
		t.Fatalf("BaseURL = %q, want %q", client.BaseURL(), owlvigil.DefaultManagementBaseURL)
	}
}

func TestWorkspaceAndGatewayKeyRequests(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer management_api_key" {
			t.Fatalf("Authorization = %q", got)
		}
		switch r.URL.Path {
		case "/workspaces":
			if r.URL.Query().Get("cursor") != "cur_1" || r.URL.Query().Get("limit") != "50" {
				t.Fatalf("query = %s", r.URL.RawQuery)
			}
			_, _ = w.Write([]byte(`{"items":[{"id":1,"name":"Acme"}],"page_info":{"next_cursor":"cur_2","has_more":true}}`))
		case "/gateway/keys":
			if r.Method == http.MethodPost {
				if got := r.Header.Get("Idempotency-Key"); got != "idem_key" {
					t.Fatalf("Idempotency-Key = %q", got)
				}
				var payload management.CreateGatewayKeyRequest
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					t.Fatalf("decode request: %v", err)
				}
				if payload.Name != "prod" {
					t.Fatalf("payload = %+v", payload)
				}
				_, _ = w.Write([]byte(`{"id":9,"name":"prod","status":"enabled","secret":"ov_sk_once"}`))
				return
			}
			_, _ = w.Write([]byte(`{"items":[{"id":9,"name":"prod","status":"enabled"}],"page_info":{}}`))
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	client := management.NewClient(
		owlvigil.WithBaseURL(server.URL),
		owlvigil.WithAPIKey("management_api_key"),
		owlvigil.WithoutRetry(),
	)

	workspaces, _, err := client.ListWorkspaces(context.Background(), management.ListOptions{Cursor: "cur_1", Limit: 50})
	if err != nil {
		t.Fatalf("ListWorkspaces returned error: %v", err)
	}
	if len(workspaces.Items) != 1 || workspaces.PageInfo.NextCursor != "cur_2" {
		t.Fatalf("workspaces = %+v", workspaces)
	}

	key, _, err := client.CreateGatewayKey(
		context.Background(),
		&management.CreateGatewayKeyRequest{Name: "prod"},
		owlvigil.WithIdempotencyKey("idem_key"),
	)
	if err != nil {
		t.Fatalf("CreateGatewayKey returned error: %v", err)
	}
	if key.Secret != "ov_sk_once" {
		t.Fatalf("key = %+v", key)
	}

	keys, _, err := client.ListGatewayKeys(context.Background(), management.ListOptions{}, "")
	if err != nil {
		t.Fatalf("ListGatewayKeys returned error: %v", err)
	}
	if len(keys.Items) != 1 || keys.Items[0].ID != 9 {
		t.Fatalf("keys = %+v", keys)
	}
}

func TestManagementEndpointCoverage(t *testing.T) {
	t.Parallel()

	expected := map[string]string{
		"GET /workspaces/1":                       `{"id":1,"name":"Acme"}`,
		"GET /gateway/keys/9":                     `{"id":9,"name":"prod","status":"enabled"}`,
		"PATCH /gateway/keys/9":                   `{"id":9,"name":"new","status":"enabled"}`,
		"POST /gateway/keys/9/rotate":             `{"id":9,"name":"prod","status":"enabled","secret":"ov_sk_new"}`,
		"POST /gateway/keys/9/enable":             `{}`,
		"POST /gateway/keys/9/disable":            `{}`,
		"DELETE /gateway/keys/9":                  `{}`,
		"GET /gateway/usage/summary":              `{"requests":10,"tokens":20,"cost":1.2}`,
		"GET /gateway/quota":                      `{"limit":100,"used":20,"remaining":80}`,
		"GET /billing/balance":                    `{"amount":10,"currency":"USD"}`,
		"GET /billing/invoices":                   `{"items":[{"id":"inv_1","amount":10,"status":"paid"}],"page_info":{}}`,
		"GET /gateway/request-logs":               `{"items":[{"request_id":"req_1","model":"m","status":"ok"}],"page_info":{}}`,
		"GET /gateway/request-logs/req_1":         `{"request_id":"req_1","model":"m","status":"ok"}`,
		"GET /gateway/traces/trace_1":             `{"trace_id":"trace_1"}`,
		"GET /webhook-endpoints":                  `{"items":[{"id":1,"url":"https://example.com","events":["request.completed"],"status":"enabled"}],"page_info":{}}`,
		"POST /webhook-endpoints":                 `{"id":1,"url":"https://example.com","events":["request.completed"],"status":"enabled"}`,
		"GET /webhook-endpoints/1":                `{"id":1,"url":"https://example.com","events":["request.completed"],"status":"enabled"}`,
		"PATCH /webhook-endpoints/1":              `{"id":1,"url":"https://example.com/new","events":["request.completed"],"status":"enabled"}`,
		"DELETE /webhook-endpoints/1":             `{}`,
		"POST /webhook-endpoints/1/enable":        `{}`,
		"POST /webhook-endpoints/1/disable":       `{}`,
		"POST /webhook-endpoints/1/rotate-secret": `{"id":1,"url":"https://example.com","events":["request.completed"],"status":"enabled","secret":"whsec_new"}`,
		"POST /webhook-endpoints/1/test":          `{}`,
		"GET /webhook-event-types":                `{"items":[{"type":"request.completed","group":"request","description":"Request completed"}],"page_info":{}}`,
		"GET /webhook-events":                     `{"items":[{"id":"evt_1","type":"request.completed","status":"delivered"}],"page_info":{}}`,
		"GET /webhook-events/evt_1":               `{"id":"evt_1","type":"request.completed","status":"delivered"}`,
		"GET /webhook-endpoints/1/events":         `{"items":[{"id":"evt_1","type":"request.completed","status":"delivered"}],"page_info":{}}`,
		"POST /webhook-events/evt_1/retry":        `{}`,
		"POST /webhook-events/evt_1/redeliver":    `{}`,
		"POST /webhook-events/bulk-redeliver":     `{}`,
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Method + " " + r.URL.Path
		body, ok := expected[key]
		if !ok {
			t.Fatalf("unexpected request %s", key)
		}
		_, _ = w.Write([]byte(body))
	}))
	defer server.Close()

	client := management.NewClient(owlvigil.WithBaseURL(server.URL), owlvigil.WithAccessToken("access_test"))
	ctx := context.Background()

	if _, _, err := client.GetWorkspace(ctx, 1); err != nil {
		t.Fatal(err)
	}
	if _, _, err := client.GetGatewayKey(ctx, 9); err != nil {
		t.Fatal(err)
	}
	name := "new"
	if _, _, err := client.UpdateGatewayKey(ctx, 9, &management.UpdateGatewayKeyRequest{Name: &name}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := client.RotateGatewayKey(ctx, 9); err != nil {
		t.Fatal(err)
	}
	if _, err := client.EnableGatewayKey(ctx, 9); err != nil {
		t.Fatal(err)
	}
	if _, err := client.DisableGatewayKey(ctx, 9); err != nil {
		t.Fatal(err)
	}
	if _, err := client.DeleteGatewayKey(ctx, 9); err != nil {
		t.Fatal(err)
	}
	if _, _, err := client.GetUsageSummary(ctx); err != nil {
		t.Fatal(err)
	}
	if _, _, err := client.GetQuota(ctx); err != nil {
		t.Fatal(err)
	}
	if _, _, err := client.GetBalance(ctx); err != nil {
		t.Fatal(err)
	}
	if _, _, err := client.ListInvoices(ctx, management.ListOptions{}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := client.ListRequestLogs(ctx, management.ListOptions{}, ""); err != nil {
		t.Fatal(err)
	}
	if _, _, err := client.GetRequestLog(ctx, "req_1"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := client.GetTrace(ctx, "trace_1"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := client.ListWebhookEndpoints(ctx, management.ListOptions{}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := client.CreateWebhookEndpoint(ctx, &management.CreateWebhookEndpointRequest{URL: "https://example.com", Events: []string{"request.completed"}}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := client.GetWebhookEndpoint(ctx, 1); err != nil {
		t.Fatal(err)
	}
	webhookURL := "https://example.com/new"
	if _, _, err := client.UpdateWebhookEndpoint(ctx, 1, &management.UpdateWebhookEndpointRequest{URL: &webhookURL}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.DeleteWebhookEndpoint(ctx, 1); err != nil {
		t.Fatal(err)
	}
	if _, err := client.EnableWebhookEndpoint(ctx, 1); err != nil {
		t.Fatal(err)
	}
	if _, err := client.DisableWebhookEndpoint(ctx, 1); err != nil {
		t.Fatal(err)
	}
	if _, _, err := client.RotateWebhookSecret(ctx, 1); err != nil {
		t.Fatal(err)
	}
	if _, err := client.TestWebhookEndpoint(ctx, 1); err != nil {
		t.Fatal(err)
	}
	if _, _, err := client.ListWebhookEventTypes(ctx); err != nil {
		t.Fatal(err)
	}
	if _, _, err := client.ListWebhookEvents(ctx, management.ListOptions{}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := client.GetWebhookEvent(ctx, "evt_1"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := client.ListEndpointEvents(ctx, 1, management.ListOptions{}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.RetryWebhookEvent(ctx, "evt_1"); err != nil {
		t.Fatal(err)
	}
	if _, err := client.RedeliverWebhookEvent(ctx, "evt_1"); err != nil {
		t.Fatal(err)
	}
	if _, err := client.BulkRedeliverWebhookEvents(ctx, &management.BulkRedeliverRequest{EventIDs: []int{1}}); err != nil {
		t.Fatal(err)
	}
}
