package management_test

import (
	"context"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	owlvigil "github.com/owlvigil/owlvigil-go"
	"github.com/owlvigil/owlvigil-go/management"
)

func TestLiveOpenAPIReadOnly(t *testing.T) {
	if os.Getenv("OWLVIGIL_LIVE_TEST") != "1" {
		t.Skip("set OWLVIGIL_LIVE_TEST=1 with OWLVIGIL_API_KEY to run against the live Open API")
	}
	apiKey := strings.TrimSpace(os.Getenv("OWLVIGIL_API_KEY"))
	if apiKey == "" {
		t.Skip("OWLVIGIL_API_KEY is not configured")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)
	client := management.NewClient(owlvigil.WithAPIKey(apiKey), owlvigil.WithoutRetry())

	workspaces, _, err := client.ListWorkspaces(ctx, management.ListOptions{Limit: 1})
	if err != nil {
		t.Fatalf("ListWorkspaces() error = %v", err)
	}
	if len(workspaces.Items) == 0 {
		t.Fatal("ListWorkspaces() returned no accessible workspaces")
	}
	workspaceID := workspaces.Items[0].ID
	workspaceOpt := owlvigil.WithQueryParam("workspace_id", strconv.FormatInt(workspaceID, 10))

	if _, _, err := client.GetUsageSummary(ctx, workspaceOpt); err != nil {
		t.Errorf("GetUsageSummary(workspace_id=%d) error = %v", workspaceID, err)
	}
	if _, _, err := client.ListProviders(ctx, workspaceID); err != nil {
		t.Errorf("ListProviders(workspace_id=%d) error = %v", workspaceID, err)
	}
	routes, _, err := client.ListRoutes(ctx, management.ListOptions{Limit: 1}, workspaceOpt)
	if err != nil {
		t.Errorf("ListRoutes(workspace_id=%d) error = %v", workspaceID, err)
	} else if len(routes.Items) > 0 {
		if _, _, err := client.GetRoute(ctx, routes.Items[0].ID, workspaceOpt); err != nil {
			t.Errorf("GetRoute(%q, workspace_id=%d) error = %v", routes.Items[0].ID, workspaceID, err)
		}
	}

	if _, _, err := client.DocumentationNavigation(ctx); err != nil {
		t.Errorf("DocumentationNavigation() error = %v", err)
	}
	endpoints, _, err := client.ListDocumentedEndpoints(ctx, management.DocumentedEndpointListOptions{})
	if err != nil {
		t.Errorf("ListDocumentedEndpoints() error = %v", err)
	} else if len(endpoints.Items) > 0 {
		if _, _, err := client.GetDocumentedEndpoint(ctx, endpoints.Items[0].ID); err != nil {
			t.Errorf("GetDocumentedEndpoint(%q) error = %v", endpoints.Items[0].ID, err)
		}
	}
	if _, _, err := client.OpenAPISchema(ctx); err != nil {
		t.Errorf("OpenAPISchema() error = %v", err)
	}
	if _, _, err := client.SwaggerSchema(ctx); err != nil {
		t.Errorf("SwaggerSchema() error = %v", err)
	}
	if _, _, err := client.SDKPackages(ctx); err != nil {
		t.Errorf("SDKPackages() error = %v", err)
	}
}
