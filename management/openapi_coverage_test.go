package management_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	owlvigil "github.com/Syrovex/owlvigil_sdk_go"
	"github.com/Syrovex/owlvigil_sdk_go/management"
)

func TestOpenAPIContractCoverageForProvidersAndRoutes(t *testing.T) {
	t.Parallel()

	expected := map[string]string{
		"GET /gateway/providers":      `{"items":[{"id":9,"workspace_id":7,"name":"Primary","type":"openai","provider_source":"byok","base_url":"https://api.openai.com","default_model":"gpt-4o-mini","api_mode":"responses","status":"active"}],"page_info":{"has_more":false}}`,
		"POST /gateway/providers":     `{"id":9,"workspace_id":7,"name":"Primary","type":"openai","provider_source":"byok","status":"active"}`,
		"GET /gateway/providers/9":    `{"id":9,"workspace_id":7,"name":"Primary","type":"openai","provider_source":"byok","status":"active"}`,
		"PATCH /gateway/providers/9":  `{"id":9,"workspace_id":7,"name":"Renamed","type":"openai","provider_source":"byok","status":"inactive"}`,
		"DELETE /gateway/providers/9": `{}`,
		"GET /gateway/routes/route-1": `{"id":"route-1","model":"gpt-4o-mini","providers":["openai"],"priority":1,"fallback_enabled":true}`,
		"PATCH /gateway/policies/3":   `{"workspace_id":7,"key_id":9,"model_policies":{"action":"block"}}`,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/gateway/providers" && r.URL.Query().Get("workspace_id") != "7" {
			t.Errorf("providers workspace_id = %q, want %q", r.URL.Query().Get("workspace_id"), "7")
		}
		if r.URL.Path == "/gateway/routes/route-1" && r.URL.Query().Get("workspace_id") != "7" {
			t.Errorf("GetRoute workspace_id = %q, want %q", r.URL.Query().Get("workspace_id"), "7")
		}
		body, ok := expected[r.Method+" "+r.URL.Path]
		if !ok {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(server.Close)

	client := management.NewClient(owlvigil.WithBaseURL(server.URL), owlvigil.WithAPIKey("key_test"))
	ctx := context.Background()

	providers, _, err := client.ListProviders(ctx, 7)
	if err != nil || len(providers.Items) != 1 || providers.Items[0].Name != "Primary" {
		t.Fatalf("ListProviders(7) = %+v, %v, want provider Primary", providers, err)
	}
	provider, _, err := client.CreateProvider(ctx, &management.CreateProviderRequest{
		WorkspaceID: 7,
		Name:        "Primary",
		Type:        "openai",
		APIKey:      "upstream-key",
	})
	if err != nil || provider.ID != 9 {
		t.Fatalf("CreateProvider() = %+v, %v, want provider ID 9", provider, err)
	}
	provider, _, err = client.GetProvider(ctx, 7, 9)
	if err != nil || provider.Name != "Primary" {
		t.Fatalf("GetProvider(7, 9) = %+v, %v, want provider Primary", provider, err)
	}
	name, status := "Renamed", "inactive"
	provider, _, err = client.UpdateProvider(ctx, 7, 9, &management.UpdateProviderRequest{Name: &name, Status: &status})
	if err != nil || provider.Name != name || provider.Status != status {
		t.Fatalf("UpdateProvider(7, 9) = %+v, %v, want name %q and status %q", provider, err, name, status)
	}
	if _, err := client.DeleteProvider(ctx, 7, 9); err != nil {
		t.Fatalf("DeleteProvider(7, 9) error = %v, want nil", err)
	}

	route, _, err := client.GetRoute(ctx, "route-1", owlvigil.WithQueryParam("workspace_id", "7"))
	if err != nil || route.ID != "route-1" {
		t.Fatalf("GetRoute(route-1) = %+v, %v, want route ID route-1", route, err)
	}
	action := "block"
	policy, _, err := client.UpdateGatewayPolicy(ctx, 3, &management.UpdateGatewayPolicyRequest{Action: &action})
	if err != nil || policy.WorkspaceID != 7 {
		t.Fatalf("UpdateGatewayPolicy(3) = %+v, %v, want workspace ID 7", policy, err)
	}

}

func TestOpenAPIContractCoverageTypesMarshalProviderRequests(t *testing.T) {
	t.Parallel()

	apiKey := "upstream-key"
	body, err := json.Marshal(management.UpdateProviderRequest{APIKey: &apiKey})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(body), `{"api_key":"upstream-key"}`; got != want {
		t.Errorf("json.Marshal(UpdateProviderRequest) = %s, want %s", got, want)
	}
}
