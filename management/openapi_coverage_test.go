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

func TestOpenAPIContractCoverageForProvidersRoutesAndDocumentation(t *testing.T) {
	t.Parallel()

	expected := map[string]string{
		"GET /gateway/providers":                `{"items":[{"id":9,"workspace_id":7,"name":"Primary","type":"openai","provider_source":"byok","base_url":"https://api.openai.com","default_model":"gpt-4o-mini","api_mode":"responses","status":"active"}],"page_info":{"has_more":false}}`,
		"POST /gateway/providers":               `{"id":9,"workspace_id":7,"name":"Primary","type":"openai","provider_source":"byok","status":"active"}`,
		"GET /gateway/providers/9":              `{"id":9,"workspace_id":7,"name":"Primary","type":"openai","provider_source":"byok","status":"active"}`,
		"PATCH /gateway/providers/9":            `{"id":9,"workspace_id":7,"name":"Renamed","type":"openai","provider_source":"byok","status":"inactive"}`,
		"DELETE /gateway/providers/9":           `{}`,
		"GET /gateway/routes/route-1":           `{"id":"route-1","model":"gpt-4o-mini","providers":["openai"],"priority":1,"fallback_enabled":true}`,
		"PATCH /gateway/policies/3":             `{"workspace_id":7,"key_id":9,"model_policies":{"action":"block"}}`,
		"GET /docs/navigation":                  `{"groups":[{"id":"gateway","title":"Gateway"}]}`,
		"GET /docs/endpoints":                   `{"items":[{"id":"gateway-providers","group":"gateway","method":"GET","path":"/open/v1/gateway/providers","scope":"gateway:read","status":"active","description":"List providers"}],"page_info":{"has_more":false}}`,
		"GET /docs/endpoints/gateway-providers": `{"id":"gateway-providers","group":"gateway","method":"GET","path":"/open/v1/gateway/providers","scope":"gateway:read","status":"active","description":"List providers"}`,
		"GET /openapi.json":                     `{"openapi":"3.1.0","info":{"title":"OwlVigil Open API"}}`,
		"GET /swagger.json":                     `{"openapi":"3.1.0","info":{"title":"OwlVigil Open API"}}`,
		"GET /sdk/packages":                     `{"items":[{"language":"go","package":"owlvigil-go","status":"available"}]}`,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/gateway/providers" && r.URL.Query().Get("workspace_id") != "7" {
			t.Errorf("providers workspace_id = %q, want %q", r.URL.Query().Get("workspace_id"), "7")
		}
		if r.URL.Path == "/gateway/routes/route-1" && r.URL.Query().Get("workspace_id") != "7" {
			t.Errorf("GetRoute workspace_id = %q, want %q", r.URL.Query().Get("workspace_id"), "7")
		}
		if r.URL.Path == "/docs/endpoints" {
			if got, want := r.URL.Query().Get("group"), "gateway"; got != want {
				t.Errorf("ListDocumentedEndpoints group = %q, want %q", got, want)
			}
			if got, want := r.URL.Query().Get("scope"), "gateway:read"; got != want {
				t.Errorf("ListDocumentedEndpoints scope = %q, want %q", got, want)
			}
			if got, want := r.URL.Query().Get("status"), "active"; got != want {
				t.Errorf("ListDocumentedEndpoints status = %q, want %q", got, want)
			}
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

	navigation, _, err := client.DocumentationNavigation(ctx)
	if err != nil || len(navigation.Groups) != 1 || navigation.Groups[0].ID != "gateway" {
		t.Fatalf("DocumentationNavigation() = %+v, %v, want gateway group", navigation, err)
	}
	documentation, _, err := client.ListDocumentedEndpoints(ctx, management.DocumentedEndpointListOptions{
		Group:  "gateway",
		Scope:  "gateway:read",
		Status: "active",
	})
	if err != nil || len(documentation.Items) != 1 || documentation.Items[0].ID != "gateway-providers" {
		t.Fatalf("ListDocumentedEndpoints() = %+v, %v, want gateway-providers", documentation, err)
	}
	endpoint, _, err := client.GetDocumentedEndpoint(ctx, "gateway-providers")
	if err != nil || endpoint.Path != "/open/v1/gateway/providers" {
		t.Fatalf("GetDocumentedEndpoint(gateway-providers) = %+v, %v, want provider path", endpoint, err)
	}

	for _, fetch := range []struct {
		name string
		call func(context.Context) (map[string]any, *owlvigil.ResponseMeta, error)
	}{
		{name: "OpenAPISchema", call: func(ctx context.Context) (map[string]any, *owlvigil.ResponseMeta, error) {
			return client.OpenAPISchema(ctx)
		}},
		{name: "SwaggerSchema", call: func(ctx context.Context) (map[string]any, *owlvigil.ResponseMeta, error) {
			return client.SwaggerSchema(ctx)
		}},
	} {
		schema, _, err := fetch.call(ctx)
		if err != nil || schema["openapi"] != "3.1.0" {
			t.Fatalf("%s() = %+v, %v, want OpenAPI 3.1.0", fetch.name, schema, err)
		}
	}
	packages, _, err := client.SDKPackages(ctx)
	if err != nil || len(packages.Items) != 1 || packages.Items[0].Language != "go" {
		t.Fatalf("SDKPackages() = %+v, %v, want Go package", packages, err)
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
