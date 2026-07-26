package management_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	owlvigil "github.com/Syrovex/owlvigil_sdk_go"
	"github.com/Syrovex/owlvigil_sdk_go/management"
)

func TestRBACEndpoints(t *testing.T) {
	t.Parallel()

	expected := map[string]string{
		"GET /workspaces/1/roles":                        `{"items":[{"id":1,"name":"Admin","is_system":true}],"page_info":{}}`,
		"POST /workspaces/1/roles":                       `{"id":2,"name":"Custom Role","is_system":false}`,
		"GET /workspaces/1/roles/1":                      `{"id":1,"name":"Admin","is_system":true}`,
		"PATCH /workspaces/1/roles/2":                    `{"id":2,"name":"Updated Role","is_system":false}`,
		"DELETE /workspaces/1/roles/2":                   `{}`,
		"GET /workspaces/1/permissions":                  `{"groups":[{"id":"workspace_administration","label":"Workspace Administration","permissions":[{"id":"workspace.settings.manage","label":"Manage Settings","default":true,"effective":true}]}]}`,
		"GET /workspaces/1/members/1/permissions":        `{"workspace_id":1,"user_id":1,"role":"owner","effective":["workspace.settings.manage"],"groups":[{"id":"workspace_administration","label":"Workspace Administration","permissions":[{"id":"workspace.settings.manage","label":"Manage Settings","default":true,"effective":true}]}]}`,
		"PUT /workspaces/1/members/1/permissions":        `{"member_id":1,"override_permissions":["workspace.settings.manage"]}`,
		"POST /workspaces/1/members/1/permissions/reset": `{"member_id":1,"role_permissions":["workspace.dashboard.view"]}`,
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

	client := management.NewClient(owlvigil.WithBaseURL(server.URL), owlvigil.WithAccessToken("test"))
	ctx := context.Background()

	// Test ListRoles
	roles, _, err := client.ListRoles(ctx, 1, management.ListOptions{})
	if err != nil {
		t.Fatalf("ListRoles failed: %v", err)
	}
	if len(roles.Items) != 1 || roles.Items[0].Name != "Admin" {
		t.Fatalf("roles = %+v", roles)
	}

	// Test CreateRole
	role, _, err := client.CreateRole(ctx, 1, &management.CreateRoleRequest{
		Name:        "Custom Role",
		Permissions: []string{"workspace.dashboard.view"},
	})
	if err != nil {
		t.Fatalf("CreateRole failed: %v", err)
	}
	if role.Name != "Custom Role" {
		t.Fatalf("role = %+v", role)
	}

	// Test GetRole
	_, _, err = client.GetRole(ctx, 1, 1)
	if err != nil {
		t.Fatalf("GetRole failed: %v", err)
	}

	// Test UpdateRole
	name := "Updated Role"
	_, _, err = client.UpdateRole(ctx, 1, 2, &management.UpdateRoleRequest{
		Name: &name,
	})
	if err != nil {
		t.Fatalf("UpdateRole failed: %v", err)
	}

	// Test DeleteRole
	_, err = client.DeleteRole(ctx, 1, 2)
	if err != nil {
		t.Fatalf("DeleteRole failed: %v", err)
	}

	// Test ListPermissions
	permissions, _, err := client.ListPermissions(ctx, 1)
	if err != nil {
		t.Fatalf("ListPermissions failed: %v", err)
	}
	if len(permissions.Groups) != 1 || len(permissions.Groups[0].Permissions) != 1 {
		t.Fatalf("permissions = %+v", permissions)
	}

	// Test GetMemberPermissions
	memberPerms, _, err := client.GetMemberPermissions(ctx, 1, 1)
	if err != nil {
		t.Fatalf("GetMemberPermissions failed: %v", err)
	}
	if memberPerms.UserID != 1 || len(memberPerms.Effective) != 1 {
		t.Fatalf("memberPerms = %+v", memberPerms)
	}

	// Test UpdateMemberPermissions
	_, _, err = client.UpdateMemberPermissions(ctx, 1, 1, &management.UpdateMemberPermissionsRequest{
		Permissions: []string{"workspace.settings.manage"},
	})
	if err != nil {
		t.Fatalf("UpdateMemberPermissions failed: %v", err)
	}

	// Test ResetMemberPermissions
	_, _, err = client.ResetMemberPermissions(ctx, 1, 1)
	if err != nil {
		t.Fatalf("ResetMemberPermissions failed: %v", err)
	}
}

func TestFinancialControlEndpoints(t *testing.T) {
	t.Parallel()

	expected := map[string]string{
		"GET /workspaces/1/governance/financial":                           `{"thresholds":{"warning_percent":80,"critical_percent":95,"exceeded_action":"notify_only"}}`,
		"PUT /workspaces/1/governance/financial":                           `{"thresholds":{"warning_percent":75,"critical_percent":90,"exceeded_action":"block_request"}}`,
		"GET /workspaces/1/governance/financial/budget-caps":               `{"workspace":{"limit":1000,"used":200,"currency":"USD"}}`,
		"PUT /workspaces/1/governance/financial/budget-caps":               `{"workspace":{"limit":2000,"used":200,"currency":"USD"}}`,
		"PATCH /workspaces/1/governance/financial/budget-caps/workspace/1": `{"limit":1500,"used":200,"currency":"USD"}`,
		"GET /workspaces/1/governance/financial/spending-limits":           `{"items":[{"user_id":1,"daily_limit":100,"currency":"USD"}],"page_info":{}}`,
		"PUT /workspaces/1/governance/financial/spending-limits":           `{"items":[{"user_id":1,"daily_limit":150,"currency":"USD"}],"page_info":{}}`,
		"PATCH /workspaces/1/governance/financial/spending-limits/users/1": `{"user_id":1,"daily_limit":200,"currency":"USD"}`,
		"GET /workspaces/1/governance/financial/thresholds":                `{"warning_percent":80,"critical_percent":95,"exceeded_action":"notify_only"}`,
		"PUT /workspaces/1/governance/financial/thresholds":                `{"warning_percent":75,"critical_percent":90,"exceeded_action":"block_request"}`,
		"POST /workspaces/1/governance/financial/preview":                  `{"valid":true,"violations":[]}`,
		"GET /workspaces/1/governance/financial/spend-summary":             `{"workspace":{"spent":500,"limit":1000,"currency":"USD"}}`,
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

	client := management.NewClient(owlvigil.WithBaseURL(server.URL), owlvigil.WithAccessToken("test"))
	ctx := context.Background()

	// Test GetFinancialGovernance
	governance, _, err := client.GetFinancialGovernance(ctx, 1)
	if err != nil {
		t.Fatalf("GetFinancialGovernance failed: %v", err)
	}
	if governance.Thresholds == nil || governance.Thresholds.WarningPercent != 80 {
		t.Fatalf("governance = %+v", governance)
	}

	// Test UpdateFinancialGovernance
	_, _, err = client.UpdateFinancialGovernance(ctx, 1, &management.UpdateFinancialGovernanceRequest{
		Thresholds: &management.Thresholds{
			WarningPercent:  75,
			CriticalPercent: 90,
			ExceededAction:  "block_request",
		},
	})
	if err != nil {
		t.Fatalf("UpdateFinancialGovernance failed: %v", err)
	}

	// Test GetBudgetCaps
	caps, _, err := client.GetBudgetCaps(ctx, 1)
	if err != nil {
		t.Fatalf("GetBudgetCaps failed: %v", err)
	}
	if caps.Workspace == nil || caps.Workspace.Limit != 1000 {
		t.Fatalf("caps = %+v", caps)
	}

	// Test UpdateBudgetCaps
	_, _, err = client.UpdateBudgetCaps(ctx, 1, &management.UpdateBudgetCapsRequest{
		Workspace: &management.BudgetCap{Limit: 2000, Currency: "USD"},
	})
	if err != nil {
		t.Fatalf("UpdateBudgetCaps failed: %v", err)
	}

	// Test UpdateScopeBudgetCap
	_, _, err = client.UpdateScopeBudgetCap(ctx, 1, "workspace", "1", &management.UpdateScopeBudgetCapRequest{
		Limit: 1500,
	})
	if err != nil {
		t.Fatalf("UpdateScopeBudgetCap failed: %v", err)
	}

	// Test GetSpendingLimits
	limits, _, err := client.GetSpendingLimits(ctx, 1, management.ListOptions{})
	if err != nil {
		t.Fatalf("GetSpendingLimits failed: %v", err)
	}
	if len(limits.Items) != 1 {
		t.Fatalf("limits = %+v", limits)
	}

	// Test UpdateSpendingLimits
	_, _, err = client.UpdateSpendingLimits(ctx, 1, &management.UpdateSpendingLimitsRequest{
		Limits: []management.SpendingLimit{{UserID: 1, DailyLimit: 150, Currency: "USD"}},
	})
	if err != nil {
		t.Fatalf("UpdateSpendingLimits failed: %v", err)
	}

	// Test UpdateUserSpendingLimit
	daily := 200.0
	_, _, err = client.UpdateUserSpendingLimit(ctx, 1, 1, &management.UpdateUserSpendingLimitRequest{
		DailyLimit: &daily,
	})
	if err != nil {
		t.Fatalf("UpdateUserSpendingLimit failed: %v", err)
	}

	// Test GetFinancialThresholds
	thresholds, _, err := client.GetFinancialThresholds(ctx, 1)
	if err != nil {
		t.Fatalf("GetFinancialThresholds failed: %v", err)
	}
	if thresholds.WarningPercent != 80 {
		t.Fatalf("thresholds = %+v", thresholds)
	}

	// Test UpdateFinancialThresholds
	warning := 75.0
	_, _, err = client.UpdateFinancialThresholds(ctx, 1, &management.UpdateThresholdsRequest{
		WarningPercent: &warning,
	})
	if err != nil {
		t.Fatalf("UpdateFinancialThresholds failed: %v", err)
	}

	// Test PreviewFinancialChanges
	preview, _, err := client.PreviewFinancialChanges(ctx, 1, &management.PreviewFinancialChangesRequest{})
	if err != nil {
		t.Fatalf("PreviewFinancialChanges failed: %v", err)
	}
	if !preview.Valid {
		t.Fatalf("preview = %+v", preview)
	}

	// Test GetSpendSummary
	summary, _, err := client.GetSpendSummary(ctx, 1)
	if err != nil {
		t.Fatalf("GetSpendSummary failed: %v", err)
	}
	if summary.Workspace == nil || summary.Workspace.Spent != 500 {
		t.Fatalf("summary = %+v", summary)
	}
}
