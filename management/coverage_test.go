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

func TestRemainingManagementEndpoints(t *testing.T) {
	t.Parallel()

	expected := map[string]string{
		"GET /user/profile":                   `{"user_id":10,"email":"user@example.com","status":"active"}`,
		"PUT /user/profile":                   `{"user_id":10,"email":"new@example.com","name":"New Name","status":"active"}`,
		"PUT /user/password":                  `{}`,
		"POST /user/support-requests":         `{}`,
		"GET /user/notification-preferences":  `{"budget_alerts":true,"billing_alerts":true,"report_emails":false,"marketing_emails":false}`,
		"PUT /user/notification-preferences":  `{"budget_alerts":false,"billing_alerts":true,"report_emails":true,"marketing_emails":false}`,
		"GET /users/me/invite-link":           `{"invite_code":"abc","invite_url":"https://example.com/i/abc","stats":{"total_invites":2,"accepted_invites":1,"pending_invites":1,"conversion_rate":0.5}}`,
		"GET /users/me/invitation-stats":      `{"total_invites":2,"accepted_invites":1,"pending_invites":1,"conversion_rate":0.5}`,
		"GET /users/me/invitations":           `{"items":[{"id":1,"email":"friend@example.com","status":"pending"}],"page_info":{}}`,
		"POST /users/me/send-invitation":      `{}`,
		"GET /gateway/models":                 `{"items":[{"id":123,"name":"GPT","developer":"openai","status":"active"}],"page_info":{}}`,
		"GET /gateway/models/gpt-4o%2Fmini":   `{"model_id":"gpt-4o-mini","name":"GPT","developer":"openai","status":"active"}`,
		"GET /gateway/routes":                 `{"items":[{"id":456,"model":"gpt-4o-mini","providers":["openai"],"priority":1,"fallback_enabled":true}],"page_info":{}}`,
		"POST /gateway/routes/preview":        `{"provider":"openai","channel":"primary","fallbacks":["anthropic"]}`,
		"GET /gateway/traces":                 `{"items":[{"trace_id":"trace_1","thread_id":"thread_1"}],"page_info":{}}`,
		"GET /gateway/payload-logs/access":    `{"allowed":true}`,
		"GET /gateway/payload-logs/payload_1": `{"payload_id":"payload_1","request_id":"req_1","request":{"model":"gpt-4o-mini"}}`,
		"GET /gateway/policies":               `{"workspace_id":1,"key_id":9,"model_policies":{"allowed":["gpt-4o-mini"]}}`,
		"POST /gateway/policies/preview":      `{"allowed":true,"modified_by":["model_policies"]}`,
		"GET /gateway/usage":                  `{"items":[{"id":"u_1","timestamp":"2026-07-08T00:00:00Z","model":"gpt-4o-mini","requests":1,"tokens":2,"cost":0.01}],"page_info":{}}`,
		"GET /workspaces/1/quota-summary":     `{"workspace_id":1,"plan":"pro","items":[{"key":"tokens","used":2,"limit":100}]}`,
		"GET /workspaces/1/quota-usage":       `{"teams":{"1":2},"members":{"10":1},"gateway_keys":{"9":2}}`,
		"PATCH /workspaces/1":                 `{"id":1,"name":"Renamed","status":"active"}`,
		"GET /workspaces/1/activity":          `{"items":[{"id":123,"workspace_id":1,"actor_id":10,"what":"updated","resource":"workspace","created_at":"2026-07-08T00:00:00Z"}],"page_info":{}}`,
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/gateway/policies" && r.URL.Query().Get("key_id") != "9" {
			t.Fatalf("key_id query = %q", r.URL.Query().Get("key_id"))
		}
		if r.URL.Path == "/users/me/invitations" && r.URL.Query().Get("limit") != "10" {
			t.Fatalf("limit query = %q", r.URL.Query().Get("limit"))
		}
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

	name := "New Name"
	if got, _, err := client.GetUserProfile(ctx); err != nil || got.UserID != 10 {
		t.Fatalf("GetUserProfile = %+v, %v", got, err)
	}
	if got, _, err := client.UpdateUserProfile(ctx, &management.UpdateUserProfileRequest{Name: &name}); err != nil || got.Name != name {
		t.Fatalf("UpdateUserProfile = %+v, %v", got, err)
	}
	if _, err := client.UpdatePassword(ctx, &management.UpdatePasswordRequest{CurrentPassword: "old", NewPassword: "new"}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.CreateSupportRequest(ctx, &management.SupportRequest{Title: "Help", Type: "billing", Description: "Need help"}); err != nil {
		t.Fatal(err)
	}
	if got, _, err := client.GetNotificationPreferences(ctx); err != nil || !got.BudgetAlerts {
		t.Fatalf("GetNotificationPreferences = %+v, %v", got, err)
	}
	reportEmails := true
	if got, _, err := client.UpdateNotificationPreferences(ctx, &management.UpdateNotificationPreferencesRequest{ReportEmails: &reportEmails}); err != nil || !got.ReportEmails {
		t.Fatalf("UpdateNotificationPreferences = %+v, %v", got, err)
	}
	if got, _, err := client.GetInviteLink(ctx); err != nil || got.InviteCode != "abc" {
		t.Fatalf("GetInviteLink = %+v, %v", got, err)
	}
	if got, _, err := client.GetInvitationStats(ctx); err != nil || got.TotalInvites != 2 {
		t.Fatalf("GetInvitationStats = %+v, %v", got, err)
	}
	if got, _, err := client.ListUserInvitations(ctx, management.ListOptions{Limit: 10}); err != nil || len(got.Items) != 1 {
		t.Fatalf("ListUserInvitations = %+v, %v", got, err)
	}
	if _, err := client.SendInvitation(ctx, &management.SendInvitationRequest{Emails: []string{"friend@example.com"}}); err != nil {
		t.Fatal(err)
	}

	if got, _, err := client.ListModels(ctx, management.ListOptions{}); err != nil || got.Items[0].ID != "123" || got.Items[0].Provider != "openai" {
		t.Fatalf("ListModels = %+v, %v", got, err)
	}
	if got, _, err := client.GetModel(ctx, "gpt-4o/mini"); err != nil || got.ID != "gpt-4o-mini" || got.Provider != "openai" {
		t.Fatalf("GetModel = %+v, %v", got, err)
	}
	if got, _, err := client.ListRoutes(ctx, management.ListOptions{}); err != nil || got.Items[0].ID != "456" || got.Items[0].ModelID != "gpt-4o-mini" {
		t.Fatalf("ListRoutes = %+v, %v", got, err)
	}
	if got, _, err := client.PreviewRoute(ctx, &management.PreviewRouteRequest{Model: "gpt-4o-mini"}); err != nil || got.Provider != "openai" {
		t.Fatalf("PreviewRoute = %+v, %v", got, err)
	}

	if got, _, err := client.ListTraces(ctx, management.ListOptions{}); err != nil || len(got.Items) != 1 {
		t.Fatalf("ListTraces = %+v, %v", got, err)
	}
	if got, _, err := client.GetPayloadAccess(ctx); err != nil || !got.Allowed {
		t.Fatalf("GetPayloadAccess = %+v, %v", got, err)
	}
	if got, _, err := client.GetPayloadLog(ctx, "payload_1"); err != nil || got.PayloadID != "payload_1" {
		t.Fatalf("GetPayloadLog = %+v, %v", got, err)
	}
	if got, _, err := client.GetGatewayPolicies(ctx, 9); err != nil || got.KeyID != 9 {
		t.Fatalf("GetGatewayPolicies = %+v, %v", got, err)
	}
	if got, _, err := client.PreviewPolicyEffect(ctx, &management.PreviewPolicyRequest{Model: "gpt-4o-mini"}); err != nil || !got.Allowed {
		t.Fatalf("PreviewPolicyEffect = %+v, %v", got, err)
	}
	if got, _, err := client.ListUsage(ctx, management.ListOptions{}); err != nil || len(got.Items) != 1 {
		t.Fatalf("ListUsage = %+v, %v", got, err)
	}
	if got, _, err := client.GetQuotaSummary(ctx, 1); err != nil || got.WorkspaceID != 1 {
		t.Fatalf("GetQuotaSummary = %+v, %v", got, err)
	}
	if got, _, err := client.GetQuotaUsage(ctx, 1); err != nil || got.Teams == nil {
		t.Fatalf("GetQuotaUsage = %+v, %v", got, err)
	}

	workspaceName := "Renamed"
	if got, _, err := client.UpdateWorkspace(ctx, 1, &management.UpdateWorkspaceRequest{Name: &workspaceName}); err != nil || got.Name != workspaceName {
		t.Fatalf("UpdateWorkspace = %+v, %v", got, err)
	}
	if got, _, err := client.ListWorkspaceActivity(ctx, 1, management.ListOptions{}); err != nil || got.Items[0].ID != "123" || got.Items[0].Action != "updated" {
		t.Fatalf("ListWorkspaceActivity = %+v, %v", got, err)
	}
}

func TestManagementCompatibilityUnmarshal(t *testing.T) {
	t.Parallel()

	var overview management.BillingOverview
	if err := json.Unmarshal([]byte(`{"balance":12.5}`), &overview); err != nil {
		t.Fatal(err)
	}
	if overview.Balance == nil || overview.BalanceAmount != 12.5 {
		t.Fatalf("overview = %+v", overview)
	}

	var plan management.Plan
	if err := json.Unmarshal([]byte(`{"slug":"starter","name":"Starter","features":"one, two; three","currency":"USD","interval":"monthly"}`), &plan); err != nil {
		t.Fatal(err)
	}
	if plan.ID != "starter" || len(plan.Features) != 3 || plan.FeatureText == "" {
		t.Fatalf("plan = %+v", plan)
	}

	var activity management.ActivityRecord
	if err := json.Unmarshal([]byte(`{"id":123,"what":"created","created_at":"2026-07-08T00:00:00Z"}`), &activity); err != nil {
		t.Fatal(err)
	}
	if activity.ID != "123" || activity.Action != "created" || activity.Timestamp == "" {
		t.Fatalf("activity = %+v", activity)
	}

	var list management.ListResponse[management.Workspace]
	if err := json.Unmarshal([]byte(`[{"id":1,"name":"Acme"}]`), &list); err != nil {
		t.Fatal(err)
	}
	if len(list.Items) != 1 || list.Items[0].ID != 1 {
		t.Fatalf("list = %+v", list)
	}
}
