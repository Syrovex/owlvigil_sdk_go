package management_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"

	owlvigil "github.com/Syrovex/owlvigil_sdk_go"
	"github.com/Syrovex/owlvigil_sdk_go/management"
)

type recordedManagementRequest struct {
	method string
	path   string
	query  url.Values
	body   []byte
}

func newManagementContractClient(t *testing.T, responseData string) (*management.Client, <-chan recordedManagementRequest) {
	t.Helper()

	requests := make(chan recordedManagementRequest, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		requests <- recordedManagementRequest{
			method: r.Method,
			path:   r.URL.Path,
			query:  r.URL.Query(),
			body:   body,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"request_id": "req_contract",
			"code":       "ok",
			"message":    "ok",
			"data":       json.RawMessage(responseData),
		})
	}))
	t.Cleanup(server.Close)

	client := management.NewClient(
		owlvigil.WithBaseURL(server.URL),
		owlvigil.WithAPIKey("management_test_key"),
		owlvigil.WithoutRetry(),
	)
	return client, requests
}

func assertManagementRequest(
	t *testing.T,
	requests <-chan recordedManagementRequest,
	wantMethod string,
	wantPath string,
	wantQuery url.Values,
	wantBody string,
) {
	t.Helper()

	got := <-requests
	if got.method != wantMethod {
		t.Errorf("SDK request method = %q, want %q", got.method, wantMethod)
	}
	if got.path != wantPath {
		t.Errorf("SDK request path = %q, want %q", got.path, wantPath)
	}
	if !reflect.DeepEqual(got.query, wantQuery) {
		t.Errorf("SDK request query = %v, want %v", got.query, wantQuery)
	}

	if wantBody == "" {
		if len(strings.TrimSpace(string(got.body))) != 0 {
			t.Errorf("SDK request body = %s, want empty body", got.body)
		}
		return
	}

	var gotJSON any
	if err := json.Unmarshal(got.body, &gotJSON); err != nil {
		t.Fatalf("json.Unmarshal(SDK request body %q) error = %v", got.body, err)
	}
	var wantJSON any
	if err := json.Unmarshal([]byte(wantBody), &wantJSON); err != nil {
		t.Fatalf("json.Unmarshal(want request body %q) error = %v", wantBody, err)
	}
	if !reflect.DeepEqual(gotJSON, wantJSON) {
		t.Errorf("SDK request JSON body = %#v, want %#v", gotJSON, wantJSON)
	}
}

func TestClient_CreateWorkspace_UsesOpenAPIContract(t *testing.T) {
	t.Parallel()

	client, requests := newManagementContractClient(t, `{
		"id": 7,
		"workspace_id": 7,
		"name": "Acme",
		"slug": "acme",
		"description": "Main workspace",
		"type": "team",
		"status": "active",
		"region": "us",
		"role": "owner",
		"created_at": "2026-07-25T00:00:00Z",
		"updated_at": "2026-07-25T00:00:01Z"
	}`)
	description := "Main workspace"

	got, meta, err := client.CreateWorkspace(t.Context(), &management.CreateWorkspaceRequest{
		Name:        "Acme",
		Slug:        "acme",
		Description: &description,
		Type:        "team",
		Region:      "us",
	})
	if err != nil {
		t.Fatalf("CreateWorkspace() error = %v, want nil", err)
	}
	if got.ID != 7 || got.WorkspaceID != 7 || got.Role != "owner" || got.UpdatedAt == "" {
		t.Errorf("CreateWorkspace() = %+v, want complete workspace fields", got)
	}
	if meta.RequestID != "req_contract" {
		t.Errorf("CreateWorkspace() response request ID = %q, want %q", meta.RequestID, "req_contract")
	}
	assertManagementRequest(
		t,
		requests,
		http.MethodPost,
		"/workspaces",
		url.Values{},
		`{"name":"Acme","slug":"acme","description":"Main workspace","type":"team","region":"us"}`,
	)
}

func TestClient_DeleteWorkspace_UsesOpenAPIContract(t *testing.T) {
	t.Parallel()

	client, requests := newManagementContractClient(t, `{"deleted":true}`)

	meta, err := client.DeleteWorkspace(t.Context(), 7)
	if err != nil {
		t.Fatalf("DeleteWorkspace(7) error = %v, want nil", err)
	}
	if meta.RequestID != "req_contract" {
		t.Errorf("DeleteWorkspace(7) response request ID = %q, want %q", meta.RequestID, "req_contract")
	}
	assertManagementRequest(t, requests, http.MethodDelete, "/workspaces/7", url.Values{}, "")
}

func TestClient_GetWorkspaceOverview_UsesOpenAPIContract(t *testing.T) {
	t.Parallel()

	client, requests := newManagementContractClient(t, `{
		"range": "7d",
		"summary": {
			"cost": {"current": 1.25, "budget": 10, "percent": 12, "weekly_current": 0.5, "weekly_budget": 5},
			"token_usage": {"current": 1234, "label": "1.2K", "weekly_current": 386, "weekly_label": "386"},
			"resources": {"active_keys": 1, "members": 2, "providers": 3, "teams": 4},
			"system_status": {"health_pct": 100, "success_rate": 99.9, "avg_latency_ms": 3399, "tone": "healthy"}
		},
		"cost_analysis": {"total": [], "by_team": [], "by_member": [], "by_provider": [], "series": {"teams": [], "members": [], "providers": []}},
		"usage_analysis": {"total": [], "by_team": [], "by_member": [], "by_provider": [], "series": {"teams": [], "members": [], "providers": []}},
		"breakdown": {"by_team": [], "by_member": [], "by_provider": []},
		"recent_activity": {
			"request_logs": [{"id": 11, "created_at": "2026-07-25T00:00:00Z", "key_name": "prod", "provider": "openai", "model": "gpt-4o-mini", "status": "success", "status_code": 200, "input_tokens": 10, "output_tokens": 20, "total_tokens": 30, "total_cost": 0.01}],
			"audit_logs_available": true,
			"audit_logs": [{"id": 12, "created_at": "2026-07-25T00:00:00Z", "action": "workspace.updated", "resource_type": "workspace", "resource_name": "Acme", "user_name": "Owner", "status": "success"}]
		}
	}`)

	got, _, err := client.GetWorkspaceOverview(
		t.Context(),
		7,
		management.WorkspaceOverviewOptions{Range: "7d"},
	)
	if err != nil {
		t.Fatalf("GetWorkspaceOverview(7, 7d) error = %v, want nil", err)
	}
	if got.Range != "7d" ||
		got.Summary.Cost.Current != 1.25 ||
		got.Summary.TokenUsage.Current != 1234 ||
		got.Summary.Resources.Providers != 3 ||
		got.Summary.SystemStatus.AvgLatencyMS != 3399 ||
		len(got.RecentActivity.RequestLogs) != 1 ||
		len(got.RecentActivity.AuditLogs) != 1 {
		t.Errorf("GetWorkspaceOverview(7, 7d) = %+v, want complete overview data", got)
	}
	assertManagementRequest(
		t,
		requests,
		http.MethodGet,
		"/workspaces/7/overview",
		url.Values{"range": {"7d"}},
		"",
	)
}

func TestClient_ListAuditLogs_UsesOpenAPIContract(t *testing.T) {
	t.Parallel()

	client, requests := newManagementContractClient(t, `{
		"items": [{
			"id": 91,
			"workspace_id": 7,
			"user_id": 10,
			"action": "workspace.updated",
			"resource_type": "workspace",
			"resource_id": "7",
			"status": "success",
			"actor_label": "Owner",
			"target_label": "Acme",
			"before_after_summary": "name changed",
			"created_at": "2026-07-25T00:00:00Z"
		}],
		"page_info": {"next_cursor": "audit_next", "has_more": true}
	}`)
	opts := management.AuditLogListOptions{
		Cursor:       "audit_cur",
		Limit:        25,
		UserID:       "10",
		Action:       "workspace.updated",
		ResourceType: "workspace",
		ResourceID:   "7",
		Status:       "success",
		DateFrom:     "2026-07-01T00:00:00Z",
		DateTo:       "2026-07-31T23:59:59Z",
		Search:       "Acme",
	}

	got, _, err := client.ListAuditLogs(t.Context(), 7, opts)
	if err != nil {
		t.Fatalf("ListAuditLogs(7, %+v) error = %v, want nil", opts, err)
	}
	if len(got.Items) != 1 || got.Items[0].ID != 91 || got.PageInfo.NextCursor != "audit_next" {
		t.Errorf("ListAuditLogs(7, %+v) = %+v, want decoded audit log page", opts, got)
	}
	assertManagementRequest(
		t,
		requests,
		http.MethodGet,
		"/workspaces/7/audit-logs",
		url.Values{
			"cursor":        {"audit_cur"},
			"limit":         {"25"},
			"user_id":       {"10"},
			"action":        {"workspace.updated"},
			"resource_type": {"workspace"},
			"resource_id":   {"7"},
			"status":        {"success"},
			"date_from":     {"2026-07-01T00:00:00Z"},
			"date_to":       {"2026-07-31T23:59:59Z"},
			"search":        {"Acme"},
		},
		"",
	)
}

func TestClient_GetAuditLog_UsesOpenAPIContract(t *testing.T) {
	t.Parallel()

	client, requests := newManagementContractClient(t, `{
		"id": 91,
		"workspace_id": 7,
		"user_id": 10,
		"action": "workspace.updated",
		"resource_type": "workspace",
		"resource_id": "7",
		"status": "success",
		"actor_label": "Owner",
		"target_label": "Acme",
		"before_after_summary": "name changed",
		"created_at": "2026-07-25T00:00:00Z"
	}`)

	got, _, err := client.GetAuditLog(t.Context(), 7, 91)
	if err != nil {
		t.Fatalf("GetAuditLog(7, 91) error = %v, want nil", err)
	}
	if got.ID != 91 || got.WorkspaceID != 7 || got.Action != "workspace.updated" {
		t.Errorf("GetAuditLog(7, 91) = %+v, want audit log 91", got)
	}
	assertManagementRequest(t, requests, http.MethodGet, "/workspaces/7/audit-logs/91", url.Values{}, "")
}

func TestClient_GetLoggingSettings_UsesOpenAPIContract(t *testing.T) {
	t.Parallel()

	client, requests := newManagementContractClient(t, `{
		"enabled": true,
		"retention_days": 30,
		"capture_headers": true,
		"capture_request_body": true,
		"capture_response_body": false,
		"capture_stream_chunks": false,
		"mask_sensitive_headers": true,
		"max_body_size": 1048576,
		"redact_headers": ["authorization"],
		"log_data_quota": {"used_bytes": 100, "limit_bytes": 1000, "usage_percent": 10, "warning": false, "over_limit": false, "cleanup_recommended": false, "last_calculated_at": "2026-07-25T00:00:00Z"}
	}`)

	got, _, err := client.GetLoggingSettings(t.Context(), 7)
	if err != nil {
		t.Fatalf("GetLoggingSettings(7) error = %v, want nil", err)
	}
	if !got.Enabled || got.RetentionDays != 30 || got.LogDataQuota == nil || got.LogDataQuota.LimitBytes != 1000 {
		t.Errorf("GetLoggingSettings(7) = %+v, want complete logging settings", got)
	}
	assertManagementRequest(t, requests, http.MethodGet, "/workspaces/7/logging-settings", url.Values{}, "")
}

func TestClient_UpdateLoggingSettings_UsesOpenAPIContract(t *testing.T) {
	t.Parallel()

	client, requests := newManagementContractClient(t, `{
		"enabled": false,
		"retention_days": 14,
		"capture_headers": true,
		"capture_request_body": false,
		"capture_response_body": false,
		"capture_stream_chunks": false,
		"mask_sensitive_headers": true,
		"max_body_size": 2048,
		"redact_headers": ["authorization", "x-api-key"]
	}`)
	enabled := false
	retentionDays := 14
	maxBodySize := 2048
	redactHeaders := []string{"authorization", "x-api-key"}
	req := &management.UpdateLoggingSettingsRequest{
		Enabled:       &enabled,
		RetentionDays: &retentionDays,
		MaxBodySize:   &maxBodySize,
		RedactHeaders: &redactHeaders,
	}

	got, _, err := client.UpdateLoggingSettings(t.Context(), 7, req)
	if err != nil {
		t.Fatalf("UpdateLoggingSettings(7, %+v) error = %v, want nil", req, err)
	}
	if got.Enabled || got.RetentionDays != 14 || len(got.RedactHeaders) != 2 {
		t.Errorf("UpdateLoggingSettings(7, %+v) = %+v, want updated settings", req, got)
	}
	assertManagementRequest(
		t,
		requests,
		http.MethodPut,
		"/workspaces/7/logging-settings",
		url.Values{},
		`{"enabled":false,"retention_days":14,"max_body_size":2048,"redact_headers":["authorization","x-api-key"]}`,
	)
}

func TestClient_ListPayloadLogs_UsesOpenAPIContract(t *testing.T) {
	t.Parallel()

	client, requests := newManagementContractClient(t, `{
		"items": [{
			"id": 81,
			"request_id": "req_81",
			"workspace_id": 7,
			"trace_id": 71,
			"trace_key": "trace_71",
			"model": "gpt-4o-mini",
			"format": "openai",
			"status": "completed",
			"stream": false,
			"created_at": "2026-07-25T00:00:00Z"
		}],
		"page_info": {"next_cursor": "payload_next", "has_more": true}
	}`)
	opts := management.ListOptions{Cursor: "payload_cur", Limit: 20}

	got, _, err := client.ListPayloadLogs(t.Context(), 7, opts)
	if err != nil {
		t.Fatalf("ListPayloadLogs(7, %+v) error = %v, want nil", opts, err)
	}
	if len(got.Items) != 1 || got.Items[0].ID != 81 || got.Items[0].RequestID != "req_81" {
		t.Errorf("ListPayloadLogs(7, %+v) = %+v, want decoded payload log page", opts, got)
	}
	assertManagementRequest(
		t,
		requests,
		http.MethodGet,
		"/gateway/payload-logs",
		url.Values{
			"workspace_id": {"7"},
			"cursor":       {"payload_cur"},
			"limit":        {"20"},
		},
		"",
	)
}

func TestClient_VerifyProviderConnection_UsesOpenAPIContract(t *testing.T) {
	t.Parallel()

	client, requests := newManagementContractClient(t, `{
		"success": true,
		"message": "connection verified",
		"latency_ms": 87.5
	}`)
	providerID := 22
	req := &management.VerifyProviderConnectionRequest{
		WorkspaceID: 7,
		ProviderID:  &providerID,
	}

	got, _, err := client.VerifyProviderConnection(t.Context(), req)
	if err != nil {
		t.Fatalf("VerifyProviderConnection(%+v) error = %v, want nil", req, err)
	}
	if !got.Success || got.Message != "connection verified" || got.LatencyMS != 87.5 {
		t.Errorf("VerifyProviderConnection(%+v) = %+v, want successful verification", req, got)
	}
	assertManagementRequest(
		t,
		requests,
		http.MethodPost,
		"/gateway/providers/verify-connection",
		url.Values{},
		`{"workspace_id":7,"provider_id":22}`,
	)
}

func TestClient_UpdateProvider_IncludesWorkspaceInRefactoredOpenAPIBody(t *testing.T) {
	t.Parallel()

	client, requests := newManagementContractClient(t, `{
		"id": 22,
		"workspace_id": 7,
		"name": "Updated",
		"type": "openai",
		"provider_source": "byok",
		"base_url": "https://api.openai.com",
		"default_model": "gpt-4o-mini",
		"api_mode": "responses",
		"status": "active",
		"request_count": 0,
		"error_count": 0,
		"created_at": "2026-07-25T00:00:00Z",
		"updated_at": "2026-07-25T00:00:01Z"
	}`)
	name := "Updated"
	req := &management.UpdateProviderRequest{Name: &name}

	got, _, err := client.UpdateProvider(t.Context(), 7, 22, req)
	if err != nil {
		t.Fatalf("UpdateProvider(7, 22, %+v) error = %v, want nil", req, err)
	}
	if got.ID != 22 || got.WorkspaceID != 7 || got.Name != "Updated" {
		t.Errorf("UpdateProvider(7, 22, %+v) = %+v, want updated provider", req, got)
	}
	assertManagementRequest(
		t,
		requests,
		http.MethodPatch,
		"/gateway/providers/22",
		url.Values{"workspace_id": {"7"}},
		`{"workspace_id":7,"name":"Updated"}`,
	)
}

func TestClient_CreateGatewayKey_UsesRefactoredOpenAPIContract(t *testing.T) {
	t.Parallel()

	client, requests := newManagementContractClient(t, `{
		"id": 31,
		"workspace_id": 7,
		"name": "prod",
		"key": "gw_once",
		"masked_key": "gw_****once",
		"status": "enabled",
		"provider_source": "byok",
		"provider_platform": "openai",
		"team_id": 3,
		"assignee_user_id": 10,
		"rate_limit_rpm": 100,
		"usage_multiplier": 1.25,
		"created_at": "2026-07-25T00:00:00Z",
		"updated_at": "2026-07-25T00:00:01Z"
	}`)
	providerPlatform := "openai"
	teamID := int64(3)
	assigneeUserID := 10
	rateLimitRPM := 100
	budgetLimit := 50.0
	usageMultiplier := 1.25
	req := &management.CreateGatewayKeyRequest{
		WorkspaceID:      7,
		Name:             "prod",
		ProviderSource:   "byok",
		ProviderPlatform: &providerPlatform,
		TeamID:           &teamID,
		AssigneeUserID:   &assigneeUserID,
		RateLimitRPM:     &rateLimitRPM,
		BudgetLimit:      &budgetLimit,
		UsageMultiplier:  &usageMultiplier,
	}

	got, _, err := client.CreateGatewayKey(t.Context(), req)
	if err != nil {
		t.Fatalf("CreateGatewayKey(%+v) error = %v, want nil", req, err)
	}
	if got.ID != 31 || got.MaskedKey != "gw_****once" || got.ProviderPlatform == nil ||
		*got.ProviderPlatform != "openai" || got.UsageMultiplier != usageMultiplier {
		t.Errorf("CreateGatewayKey(%+v) = %+v, want complete gateway key", req, got)
	}
	assertManagementRequest(
		t,
		requests,
		http.MethodPost,
		"/gateway/keys",
		url.Values{},
		`{"workspace_id":7,"name":"prod","provider_source":"byok","provider_platform":"openai","team_id":3,"assignee_user_id":10,"rate_limit_rpm":100,"budget_limit":50,"usage_multiplier":1.25}`,
	)
}

func TestClient_UpdateGatewayKey_UsesRefactoredOpenAPIContract(t *testing.T) {
	t.Parallel()

	client, requests := newManagementContractClient(t, `{
		"id": 31,
		"workspace_id": 7,
		"name": "prod-updated",
		"masked_key": "gw_****once",
		"status": "enabled",
		"provider_source": "platform",
		"usage_multiplier": 1.5,
		"created_at": "2026-07-25T00:00:00Z",
		"updated_at": "2026-07-25T00:00:01Z"
	}`)
	name := "prod-updated"
	providerSource := "platform"
	clearTeamID := true
	clearAssigneeUserID := true
	rateLimitRPM := 200
	usageMultiplier := 1.5
	req := &management.UpdateGatewayKeyRequest{
		Name:                &name,
		ProviderSource:      &providerSource,
		ClearTeamID:         clearTeamID,
		ClearAssigneeUserID: clearAssigneeUserID,
		RateLimitRPM:        &rateLimitRPM,
		UsageMultiplier:     &usageMultiplier,
	}

	got, _, err := client.UpdateGatewayKey(t.Context(), 31, req)
	if err != nil {
		t.Fatalf("UpdateGatewayKey(31, %+v) error = %v, want nil", req, err)
	}
	if got.ID != 31 || got.Name != "prod-updated" || got.ProviderSource != "platform" ||
		got.UsageMultiplier != usageMultiplier {
		t.Errorf("UpdateGatewayKey(31, %+v) = %+v, want updated gateway key", req, got)
	}
	assertManagementRequest(
		t,
		requests,
		http.MethodPatch,
		"/gateway/keys/31",
		url.Values{},
		`{"name":"prod-updated","provider_source":"platform","clear_team_id":true,"clear_assignee_user_id":true,"rate_limit_rpm":200,"usage_multiplier":1.5}`,
	)
}

func TestClient_ListRequestLogs_ForwardsGatewayKeyFilter(t *testing.T) {
	t.Parallel()

	client, requests := newManagementContractClient(t, `{"items":[],"page_info":{"has_more":false}}`)

	_, _, err := client.ListRequestLogs(
		t.Context(),
		management.ListOptions{Limit: 25, Cursor: "100"},
		"31",
		owlvigil.WithQueryParam("workspace_id", "7"),
	)
	if err != nil {
		t.Fatalf("ListRequestLogs(gateway_key_id=31) error = %v, want nil", err)
	}
	assertManagementRequest(
		t,
		requests,
		http.MethodGet,
		"/gateway/request-logs",
		url.Values{
			"workspace_id":   {"7"},
			"limit":          {"25"},
			"cursor":         {"100"},
			"gateway_key_id": {"31"},
		},
		"",
	)
}

func TestClient_UpdateUserProfile_UsesRefactoredOpenAPIContract(t *testing.T) {
	t.Parallel()

	client, requests := newManagementContractClient(t, `{
		"id": 10,
		"user_id": 10,
		"username": "owner",
		"email": "owner@example.com",
		"name": "Owner",
		"status": "active",
		"default_workspace_id": 7,
		"balance_notify_enabled": true,
		"balance_notify_threshold": 5,
		"balance_notify_extra_emails": [{"email": "billing@example.com", "disabled": false, "verified": true}],
		"created_at": "2026-07-25T00:00:00Z",
		"updated_at": "2026-07-25T00:00:01Z"
	}`)
	username := "owner"
	defaultWorkspaceID := int64(7)
	notifyEnabled := true
	notifyThreshold := 5.0
	extraEmails := []management.NotifyEmailEntry{{Email: "billing@example.com"}}
	req := &management.UpdateUserProfileRequest{
		Username:                 &username,
		DefaultWorkspaceID:       &defaultWorkspaceID,
		BalanceNotifyEnabled:     &notifyEnabled,
		BalanceNotifyThreshold:   &notifyThreshold,
		BalanceNotifyExtraEmails: &extraEmails,
	}

	got, _, err := client.UpdateUserProfile(t.Context(), req)
	if err != nil {
		t.Fatalf("UpdateUserProfile(%+v) error = %v, want nil", req, err)
	}
	if got.ID != 10 || got.Username != "owner" || !got.BalanceNotifyEnabled || len(got.BalanceNotifyExtraEmails) != 1 {
		t.Errorf("UpdateUserProfile(%+v) = %+v, want complete profile", req, got)
	}
	assertManagementRequest(
		t,
		requests,
		http.MethodPut,
		"/user/profile",
		url.Values{},
		`{"username":"owner","default_workspace_id":7,"balance_notify_enabled":true,"balance_notify_threshold":5,"balance_notify_extra_emails":[{"email":"billing@example.com","disabled":false,"verified":false}]}`,
	)
}

func TestClient_UpdateUserProfile_CanClearNullableFields(t *testing.T) {
	t.Parallel()

	client, requests := newManagementContractClient(t, `{
		"id": 10,
		"user_id": 10,
		"username": "owner",
		"email": "owner@example.com",
		"status": "active",
		"avatar_url": null,
		"balance_notify_threshold": null,
		"created_at": "2026-07-25T00:00:00Z",
		"updated_at": "2026-07-25T00:00:01Z"
	}`)
	req := &management.UpdateUserProfileRequest{
		ClearAvatarURL:              true,
		ClearBalanceNotifyThreshold: true,
	}

	if _, _, err := client.UpdateUserProfile(t.Context(), req); err != nil {
		t.Fatalf("UpdateUserProfile(%+v) error = %v, want nil", req, err)
	}
	assertManagementRequest(
		t,
		requests,
		http.MethodPut,
		"/user/profile",
		url.Values{},
		`{"avatar_url":null,"balance_notify_threshold":null}`,
	)
}

func TestClient_UpdatePassword_UsesRefactoredOpenAPIContract(t *testing.T) {
	t.Parallel()

	client, requests := newManagementContractClient(t, `{"message":"password updated"}`)
	req := &management.UpdatePasswordRequest{
		OldPassword: "old-password",
		NewPassword: "new-password",
	}

	if _, err := client.UpdatePassword(t.Context(), req); err != nil {
		t.Fatalf("UpdatePassword(%+v) error = %v, want nil", req, err)
	}
	assertManagementRequest(
		t,
		requests,
		http.MethodPut,
		"/user/password",
		url.Values{},
		`{"old_password":"old-password","new_password":"new-password"}`,
	)
}

func TestClient_CreateSupportRequest_UsesRefactoredOpenAPIContract(t *testing.T) {
	t.Parallel()

	client, requests := newManagementContractClient(t, `{"message":"submitted"}`)
	req := &management.SupportRequest{
		Subject:     "Need help",
		IssueType:   "Billing",
		Description: "Please review this account.",
	}

	if _, err := client.CreateSupportRequest(t.Context(), req); err != nil {
		t.Fatalf("CreateSupportRequest(%+v) error = %v, want nil", req, err)
	}
	assertManagementRequest(
		t,
		requests,
		http.MethodPost,
		"/user/support-requests",
		url.Values{},
		`{"subject":"Need help","issue_type":"Billing","description":"Please review this account."}`,
	)
}

func TestClient_UpdateNotificationPreferences_UsesRefactoredOpenAPIContract(t *testing.T) {
	t.Parallel()

	client, requests := newManagementContractClient(t, `{
		"budget": true,
		"billing": false,
		"reports": true,
		"marketing": false
	}`)
	budget, billing, reports, marketing := true, false, true, false
	req := &management.UpdateNotificationPreferencesRequest{
		Budget:    &budget,
		Billing:   &billing,
		Reports:   &reports,
		Marketing: &marketing,
	}

	got, _, err := client.UpdateNotificationPreferences(t.Context(), req)
	if err != nil {
		t.Fatalf("UpdateNotificationPreferences(%+v) error = %v, want nil", req, err)
	}
	if !got.Budget || got.Billing || !got.Reports || got.Marketing {
		t.Errorf("UpdateNotificationPreferences(%+v) = %+v, want refactored preference fields", req, got)
	}
	assertManagementRequest(
		t,
		requests,
		http.MethodPut,
		"/user/notification-preferences",
		url.Values{},
		`{"budget":true,"billing":false,"reports":true,"marketing":false}`,
	)
}

func TestUpdateNotificationPreferencesRequest_AlwaysEmitsPUTReplacement(t *testing.T) {
	t.Parallel()

	reports := true
	body, err := json.Marshal(management.UpdateNotificationPreferencesRequest{
		Reports: &reports,
	})
	if err != nil {
		t.Fatalf("json.Marshal(UpdateNotificationPreferencesRequest) error = %v", err)
	}
	assertJSONSemanticallyEqual(
		t,
		body,
		`{"budget":false,"billing":false,"reports":true,"marketing":false}`,
	)
}

func TestActionMethods_ValidatePublishedResponseShape(t *testing.T) {
	t.Parallel()

	workspace := owlvigil.WithWorkspaceID(7)
	tests := []struct {
		name string
		call func(*management.Client) error
	}{
		{
			name: "update password message",
			call: func(client *management.Client) error {
				_, err := client.UpdatePassword(t.Context(), &management.UpdatePasswordRequest{
					OldPassword: "old-password",
					NewPassword: "new-password",
				})
				return err
			},
		},
		{
			name: "support request message",
			call: func(client *management.Client) error {
				_, err := client.CreateSupportRequest(t.Context(), &management.SupportRequest{
					Subject:     "Help",
					IssueType:   "billing",
					Description: "Please help.",
				})
				return err
			},
		},
		{
			name: "send invitation result",
			call: func(client *management.Client) error {
				_, err := client.SendInvitation(t.Context(), &management.SendInvitationRequest{
					Emails: []string{"friend@example.com"},
				})
				return err
			},
		},
		{
			name: "enable gateway key",
			call: func(client *management.Client) error {
				_, err := client.EnableGatewayKey(t.Context(), 31, workspace)
				return err
			},
		},
		{
			name: "disable gateway key",
			call: func(client *management.Client) error {
				_, err := client.DisableGatewayKey(t.Context(), 31, workspace)
				return err
			},
		},
		{
			name: "delete gateway key",
			call: func(client *management.Client) error {
				_, err := client.DeleteGatewayKey(t.Context(), 31, workspace)
				return err
			},
		},
		{
			name: "delete provider",
			call: func(client *management.Client) error {
				_, err := client.DeleteProvider(t.Context(), 7, 41)
				return err
			},
		},
		{
			name: "resend workspace invitation",
			call: func(client *management.Client) error {
				_, err := client.ResendInvitation(t.Context(), 7, 51)
				return err
			},
		},
		{
			name: "revoke workspace invitation",
			call: func(client *management.Client) error {
				_, err := client.RevokeInvitation(t.Context(), 7, 51)
				return err
			},
		},
		{
			name: "delete member",
			call: func(client *management.Client) error {
				_, err := client.DeleteMember(t.Context(), 7, 61)
				return err
			},
		},
		{
			name: "delete role",
			call: func(client *management.Client) error {
				_, err := client.DeleteRole(t.Context(), 7, 71)
				return err
			},
		},
		{
			name: "delete team",
			call: func(client *management.Client) error {
				_, err := client.DeleteTeam(t.Context(), 7, 81)
				return err
			},
		},
		{
			name: "delete payment method",
			call: func(client *management.Client) error {
				_, err := client.DeletePaymentMethod(t.Context(), "pm_91")
				return err
			},
		},
		{
			name: "delete webhook endpoint",
			call: func(client *management.Client) error {
				_, err := client.DeleteWebhookEndpoint(t.Context(), 101, workspace)
				return err
			},
		},
		{
			name: "enable webhook endpoint",
			call: func(client *management.Client) error {
				_, err := client.EnableWebhookEndpoint(t.Context(), 101, workspace)
				return err
			},
		},
		{
			name: "disable webhook endpoint",
			call: func(client *management.Client) error {
				_, err := client.DisableWebhookEndpoint(t.Context(), 101, workspace)
				return err
			},
		},
		{
			name: "test webhook endpoint",
			call: func(client *management.Client) error {
				_, err := client.TestWebhookEndpoint(t.Context(), 101, workspace)
				return err
			},
		},
		{
			name: "retry webhook event",
			call: func(client *management.Client) error {
				_, err := client.RetryWebhookEvent(t.Context(), "111", workspace)
				return err
			},
		},
		{
			name: "redeliver webhook event",
			call: func(client *management.Client) error {
				_, err := client.RedeliverWebhookEvent(t.Context(), "111", workspace)
				return err
			},
		},
		{
			name: "bulk redeliver webhook events",
			call: func(client *management.Client) error {
				_, err := client.BulkRedeliverWebhookEvents(
					t.Context(),
					&management.BulkRedeliverRequest{WorkspaceID: 7, EventIDs: []int{111}},
				)
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			client, _ := newManagementContractClient(t, `"not-an-object"`)
			if err := tt.call(client); err == nil {
				t.Fatalf("%s error = nil, want response-shape decoding error", tt.name)
			}
		})
	}
}

func TestActionMethods_ExposePublishedResponseData(t *testing.T) {
	t.Parallel()

	t.Run("account actions", func(t *testing.T) {
		t.Parallel()

		client, _ := newManagementContractClient(t, `{"message":"password updated"}`)
		password, _, err := client.UpdatePasswordWithResult(
			t.Context(),
			&management.UpdatePasswordRequest{OldPassword: "old-password", NewPassword: "new-password"},
		)
		if err != nil || password.Message != "password updated" {
			t.Fatalf("UpdatePasswordWithResult() = %+v, %v", password, err)
		}

		client, _ = newManagementContractClient(t, `{"message":"support submitted"}`)
		support, _, err := client.CreateSupportRequestWithResult(
			t.Context(),
			&management.SupportRequest{Subject: "Help", IssueType: "billing", Description: "Please help."},
		)
		if err != nil || support.Message != "support submitted" {
			t.Fatalf("CreateSupportRequestWithResult() = %+v, %v", support, err)
		}

		client, _ = newManagementContractClient(t, `{"message":"sent","emails_sent":2}`)
		delivery, _, err := client.SendInvitationWithResult(
			t.Context(),
			&management.SendInvitationRequest{Emails: []string{"one@example.com", "two@example.com"}},
		)
		if err != nil || delivery.EmailsSent != 2 {
			t.Fatalf("SendInvitationWithResult() = %+v, %v", delivery, err)
		}
	})

	t.Run("gateway and workspace actions", func(t *testing.T) {
		t.Parallel()

		workspace := owlvigil.WithWorkspaceID(7)
		client, _ := newManagementContractClient(t, `{"id":31,"workspace_id":7,"status":"active"}`)
		key, _, err := client.EnableGatewayKeyWithResult(t.Context(), 31, workspace)
		if err != nil || key.ID != 31 || key.Status != "active" {
			t.Fatalf("EnableGatewayKeyWithResult() = %+v, %v", key, err)
		}
		client, _ = newManagementContractClient(t, `{"id":31,"workspace_id":7,"status":"inactive"}`)
		key, _, err = client.DisableGatewayKeyWithResult(t.Context(), 31, workspace)
		if err != nil || key.Status != "inactive" {
			t.Fatalf("DisableGatewayKeyWithResult() = %+v, %v", key, err)
		}

		deleteCalls := []struct {
			name string
			call func(*management.Client) (*management.DeleteResponse, error)
		}{
			{
				name: "gateway key",
				call: func(client *management.Client) (*management.DeleteResponse, error) {
					out, _, err := client.DeleteGatewayKeyWithResult(t.Context(), 31, workspace)
					return out, err
				},
			},
			{
				name: "provider",
				call: func(client *management.Client) (*management.DeleteResponse, error) {
					out, _, err := client.DeleteProviderWithResult(t.Context(), 7, 41)
					return out, err
				},
			},
			{
				name: "member",
				call: func(client *management.Client) (*management.DeleteResponse, error) {
					out, _, err := client.DeleteMemberWithResult(t.Context(), 7, 61)
					return out, err
				},
			},
			{
				name: "role",
				call: func(client *management.Client) (*management.DeleteResponse, error) {
					out, _, err := client.DeleteRoleWithResult(t.Context(), 7, 71)
					return out, err
				},
			},
			{
				name: "team",
				call: func(client *management.Client) (*management.DeleteResponse, error) {
					out, _, err := client.DeleteTeamWithResult(t.Context(), 7, 81)
					return out, err
				},
			},
			{
				name: "payment method",
				call: func(client *management.Client) (*management.DeleteResponse, error) {
					out, _, err := client.DeletePaymentMethodWithResult(t.Context(), "pm_91")
					return out, err
				},
			},
			{
				name: "webhook endpoint",
				call: func(client *management.Client) (*management.DeleteResponse, error) {
					out, _, err := client.DeleteWebhookEndpointWithResult(t.Context(), 101, workspace)
					return out, err
				},
			},
		}
		for _, item := range deleteCalls {
			client, _ = newManagementContractClient(t, `{"deleted":true}`)
			out, err := item.call(client)
			if err != nil || out == nil || !out.Deleted {
				t.Fatalf("%s delete result = %+v, %v", item.name, out, err)
			}
		}

		client, _ = newManagementContractClient(t, `{"id":51,"workspace_id":7,"status":"pending"}`)
		invitation, _, err := client.ResendInvitationWithResult(t.Context(), 7, 51)
		if err != nil || invitation.ID != 51 {
			t.Fatalf("ResendInvitationWithResult() = %+v, %v", invitation, err)
		}
		client, _ = newManagementContractClient(t, `{"id":51,"workspace_id":7,"status":"revoked"}`)
		invitation, _, err = client.RevokeInvitationWithResult(t.Context(), 7, 51)
		if err != nil || invitation.Status != "revoked" {
			t.Fatalf("RevokeInvitationWithResult() = %+v, %v", invitation, err)
		}
	})

	t.Run("webhook actions", func(t *testing.T) {
		t.Parallel()

		workspace := owlvigil.WithWorkspaceID(7)
		client, _ := newManagementContractClient(t, `{"id":101,"workspace_id":7,"status":"enabled"}`)
		endpoint, _, err := client.EnableWebhookEndpointWithResult(t.Context(), 101, workspace)
		if err != nil || endpoint.Status != "enabled" {
			t.Fatalf("EnableWebhookEndpointWithResult() = %+v, %v", endpoint, err)
		}
		client, _ = newManagementContractClient(t, `{"id":101,"workspace_id":7,"status":"disabled"}`)
		endpoint, _, err = client.DisableWebhookEndpointWithResult(t.Context(), 101, workspace)
		if err != nil || endpoint.Status != "disabled" {
			t.Fatalf("DisableWebhookEndpointWithResult() = %+v, %v", endpoint, err)
		}

		for _, item := range []struct {
			name string
			call func(*management.Client) (*management.WebhookEvent, error)
		}{
			{
				name: "test",
				call: func(client *management.Client) (*management.WebhookEvent, error) {
					out, _, err := client.TestWebhookEndpointWithResult(t.Context(), 101, workspace)
					return out, err
				},
			},
			{
				name: "retry",
				call: func(client *management.Client) (*management.WebhookEvent, error) {
					out, _, err := client.RetryWebhookEventWithResult(t.Context(), "111", workspace)
					return out, err
				},
			},
			{
				name: "redeliver",
				call: func(client *management.Client) (*management.WebhookEvent, error) {
					out, _, err := client.RedeliverWebhookEventWithResult(t.Context(), "111", workspace)
					return out, err
				},
			},
		} {
			client, _ = newManagementContractClient(t, `{"id":111,"workspace_id":7,"status":"pending"}`)
			event, err := item.call(client)
			if err != nil || event == nil || event.ID != "111" {
				t.Fatalf("%s webhook result = %+v, %v", item.name, event, err)
			}
		}

		client, _ = newManagementContractClient(t, `{"items":[{"id":111,"workspace_id":7,"status":"pending"}],"page_info":{"has_more":false}}`)
		events, _, err := client.BulkRedeliverWebhookEventsWithResult(
			t.Context(),
			&management.BulkRedeliverRequest{WorkspaceID: 7, EventIDs: []int{111}},
		)
		if err != nil || len(events.Items) != 1 || events.Items[0].ID != "111" {
			t.Fatalf("BulkRedeliverWebhookEventsWithResult() = %+v, %v", events, err)
		}
	})
}

func TestAccessRequestTypes_MarshalRefactoredOpenAPIFields(t *testing.T) {
	t.Parallel()

	monthlyBudget := 500.0
	teamID := int64(3)
	status := "active"
	tests := []struct {
		name  string
		value any
		want  string
	}{
		{
			name: "create team",
			value: management.CreateTeamRequest{
				Name:               "Engineering",
				Description:        "Builds the product",
				MonthlyBudgetLimit: &monthlyBudget,
			},
			want: `{"name":"Engineering","description":"Builds the product","monthly_budget_limit":500}`,
		},
		{
			name: "update team omits retired status field",
			value: management.UpdateTeamRequest{
				Status:             &status,
				MonthlyBudgetLimit: &monthlyBudget,
			},
			want: `{"monthly_budget_limit":500}`,
		},
		{
			name: "invite member",
			value: management.CreateMemberRequest{
				Email:  "member@example.com",
				Role:   "member",
				TeamID: &teamID,
			},
			want: `{"email":"member@example.com","role":"member","team_id":3}`,
		},
		{
			name: "update member",
			value: management.UpdateMemberRequest{
				Role:   "admin",
				Status: &status,
				TeamID: &teamID,
			},
			want: `{"role":"admin","status":"active","team_id":3}`,
		},
		{
			name: "create invitation",
			value: management.CreateInvitationRequest{
				Email:          "invite@example.com",
				Role:           "member",
				TeamID:         &teamID,
				ExpiresInHours: 72,
			},
			want: `{"email":"invite@example.com","role":"member","team_id":3,"expires_in_hours":72}`,
		},
		{
			name: "create role",
			value: management.CreateRoleRequest{
				Key:         "billing_viewer",
				Name:        "Billing Viewer",
				Permissions: []string{"billing.read"},
			},
			want: `{"key":"billing_viewer","name":"Billing Viewer","permissions":["billing.read"]}`,
		},
		{
			name:  "create role empty permissions",
			value: management.CreateRoleRequest{Name: "Empty Role"},
			want:  `{"name":"Empty Role","permissions":null}`,
		},
		{
			name: "update role clears permissions",
			value: management.UpdateRoleRequest{
				Permissions: &[]string{},
			},
			want: `{"permissions":[]}`,
		},
		{
			name: "update member permissions",
			value: management.UpdateMemberPermissionsRequest{
				PermissionMap: map[string]bool{
					"billing.read":  true,
					"billing.write": false,
				},
			},
			want: `{"permissions":{"billing.read":true,"billing.write":false}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.value)
			if err != nil {
				t.Fatalf("json.Marshal(%T) error = %v", tt.value, err)
			}
			assertJSONSemanticallyEqual(t, got, tt.want)
		})
	}
}

func TestClient_ListRoleOptions_DecodesRefactoredOpenAPIContract(t *testing.T) {
	t.Parallel()

	client, requests := newManagementContractClient(t, `{
		"invite_roles": [{"value":"member","label":"Member","enabled":true}],
		"edit_roles": [{"value":"admin","label":"Admin","enabled":true}]
	}`)

	got, _, err := client.ListRoleOptions(t.Context(), 7)
	if err != nil {
		t.Fatalf("ListRoleOptions(7) error = %v, want nil", err)
	}
	if len(got.InviteRoles) != 1 || got.InviteRoles[0].Value != "member" ||
		len(got.EditRoles) != 1 || got.EditRoles[0].Value != "admin" {
		t.Errorf("ListRoleOptions(7) = %+v, want invite and edit role groups", got)
	}
	assertManagementRequest(
		t,
		requests,
		http.MethodGet,
		"/workspaces/7/members/role-options",
		url.Values{},
		"",
	)
}

func TestBillingRequestTypes_MarshalRefactoredOpenAPIFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value any
		want  string
	}{
		{
			name: "subscription in app",
			value: management.CreateSubscriptionInAppRequest{
				PlanID:    "pro",
				Interval:  "monthly",
				ReturnURL: "https://example.com/return",
			},
			want: `{"plan_id":"pro","interval":"monthly","return_url":"https://example.com/return"}`,
		},
		{
			name: "confirm subscription",
			value: management.ConfirmSubscriptionInAppRequest{
				PlanID:               "pro",
				StripeSubscriptionID: "sub_123",
			},
			want: `{"plan_id":"pro","stripe_subscription_id":"sub_123"}`,
		},
		{
			name: "upgrade subscription",
			value: management.UpgradeSubscriptionRequest{
				PlanID:   "business",
				Interval: "yearly",
			},
			want: `{"plan_id":"business","interval":"yearly"}`,
		},
		{
			name: "downgrade subscription",
			value: management.DowngradeSubscriptionRequest{
				PlanID:   "starter",
				Interval: "monthly",
			},
			want: `{"plan_id":"starter","interval":"monthly"}`,
		},
		{
			name: "topup checkout",
			value: management.CreateTopupCheckoutRequest{
				WorkspaceID: 7,
				Amount:      100,
				SuccessURL:  "https://example.com/success",
				CancelURL:   "https://example.com/cancel",
			},
			want: `{"workspace_id":7,"amount":100,"success_url":"https://example.com/success","cancel_url":"https://example.com/cancel"}`,
		},
		{
			name: "topup in app",
			value: management.CreateTopupInAppRequest{
				WorkspaceID: 7,
				Amount:      100,
				ReturnURL:   "https://example.com/return",
			},
			want: `{"workspace_id":7,"amount":100,"return_url":"https://example.com/return"}`,
		},
		{
			name: "confirm topup",
			value: management.ConfirmTopupInAppRequest{
				PaymentIntentID: "pi_123",
				ClientSecret:    "pi_123_secret",
			},
			want: `{"payment_intent_id":"pi_123","client_secret":"pi_123_secret"}`,
		},
		{
			name: "billing details",
			value: management.UpdateBillingDetailsRequest{
				Name:         "Acme",
				Email:        "billing@example.com",
				TaxID:        "tax-123",
				Phone:        "+1-555-0100",
				AddressText:  "1 Main Street",
				CCRecipients: []string{"finance@example.com"},
			},
			want: `{"name":"Acme","email":"billing@example.com","tax_id":"tax-123","phone":"+1-555-0100","address":"1 Main Street","cc_recipients":["finance@example.com"]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.value)
			if err != nil {
				t.Fatalf("json.Marshal(%T) error = %v", tt.value, err)
			}
			assertJSONSemanticallyEqual(t, got, tt.want)
		})
	}
}

func TestRequiredRequestFields_AreNotOmittedAtZeroValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value any
		want  string
	}{
		{
			name:  "route preview",
			value: management.PreviewRouteRequest{},
			want:  `{"workspace_id":0,"model":""}`,
		},
		{
			name:  "provider update",
			value: management.UpdateProviderRequest{},
			want:  `{"workspace_id":0}`,
		},
		{
			name:  "optional Stripe session",
			value: management.ConfirmStripeSessionRequest{},
			want:  `{}`,
		},
		{
			name:  "subscription confirmation",
			value: management.ConfirmSubscriptionInAppRequest{},
			want:  `{"plan_id":"","stripe_subscription_id":""}`,
		},
		{
			name:  "subscription upgrade",
			value: management.UpgradeSubscriptionRequest{},
			want:  `{"plan_id":"","interval":""}`,
		},
		{
			name:  "subscription downgrade",
			value: management.DowngradeSubscriptionRequest{},
			want:  `{"plan_id":"","interval":""}`,
		},
		{
			name:  "topup checkout",
			value: management.CreateTopupCheckoutRequest{},
			want:  `{"workspace_id":0,"amount":0,"success_url":"","cancel_url":""}`,
		},
		{
			name:  "topup in app",
			value: management.CreateTopupInAppRequest{},
			want:  `{"workspace_id":0,"amount":0,"return_url":""}`,
		},
		{
			name:  "webhook bulk redelivery",
			value: management.BulkRedeliverRequest{},
			want:  `{"workspace_id":0}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.value)
			if err != nil {
				t.Fatalf("json.Marshal(%T) error = %v", tt.value, err)
			}
			assertJSONSemanticallyEqual(t, got, tt.want)
		})
	}
}

func TestClient_CancelSubscription_SendsRequiredJSONBody(t *testing.T) {
	t.Parallel()

	client, requests := newManagementContractClient(t, `{
		"user_id": 10,
		"subscription_status": "active",
		"subscription_tier": "pro",
		"effective_subscription_tier": "pro",
		"max_workspaces": 2,
		"current_workspace_count": 1,
		"can_create_more_workspaces": true,
		"subscription_cancel_at_period_end": true,
		"is_trial_expired": false
	}`)

	if _, _, err := client.CancelSubscription(t.Context()); err != nil {
		t.Fatalf("CancelSubscription() error = %v, want nil", err)
	}
	assertManagementRequest(
		t,
		requests,
		http.MethodPost,
		"/billing/subscription/cancel",
		url.Values{},
		`{}`,
	)
}

func TestBillingWorkspaceMethods_UseRequiredQueryContract(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		response string
		call     func(*management.Client) error
		method   string
		path     string
	}{
		{
			name:     "overview",
			response: `{"workspace_id":7,"balance":10,"current_month_cost":1,"current_month_tokens":2,"current_month_requests":3,"billing_details":{},"recent_invoices":[],"recent_invoices_page_info":{"has_more":false}}`,
			call: func(client *management.Client) error {
				_, _, err := client.GetBillingOverviewForWorkspace(t.Context(), 7)
				return err
			},
			method: http.MethodGet,
			path:   "/billing/overview",
		},
		{
			name:     "balance",
			response: `{"workspace_id":7,"balance":10}`,
			call: func(client *management.Client) error {
				_, _, err := client.GetBalanceForWorkspace(t.Context(), 7)
				return err
			},
			method: http.MethodGet,
			path:   "/billing/balance",
		},
		{
			name:     "invoices",
			response: `{"items":[],"page_info":{"has_more":false}}`,
			call: func(client *management.Client) error {
				_, _, err := client.ListInvoicesForWorkspace(t.Context(), 7, management.ListOptions{Limit: 5})
				return err
			},
			method: http.MethodGet,
			path:   "/billing/invoices",
		},
		{
			name:     "invoice",
			response: `{"id":58,"workspace_id":7,"stripe_invoice_id":"in_58","amount_due":10,"amount_paid":10,"currency":"USD","status":"paid","created_at":"2026-07-25T00:00:00Z","updated_at":"2026-07-25T00:00:00Z"}`,
			call: func(client *management.Client) error {
				_, _, err := client.GetInvoiceForWorkspace(t.Context(), 7, "58")
				return err
			},
			method: http.MethodGet,
			path:   "/billing/invoices/58",
		},
		{
			name:     "payment methods",
			response: `{"items":[],"page_info":{"has_more":false}}`,
			call: func(client *management.Client) error {
				_, _, err := client.ListPaymentMethodsForWorkspace(t.Context(), 7)
				return err
			},
			method: http.MethodGet,
			path:   "/billing/payment-methods",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			client, requests := newManagementContractClient(t, tt.response)
			if err := tt.call(client); err != nil {
				t.Fatalf("%s call error = %v, want nil", tt.name, err)
			}
			query := url.Values{"workspace_id": {"7"}}
			if tt.name == "invoices" {
				query.Set("limit", "5")
			}
			assertManagementRequest(t, requests, tt.method, tt.path, query, "")
		})
	}
}

func TestClient_CancelSubscriptionWithRequest_UsesOpenAPIContract(t *testing.T) {
	t.Parallel()

	client, requests := newManagementContractClient(t, `{
		"user_id": 10,
		"subscription_status": "canceled",
		"subscription_tier": "pro",
		"effective_subscription_tier": "pro",
		"max_workspaces": 2,
		"current_workspace_count": 1,
		"can_create_more_workspaces": true,
		"subscription_cancel_at_period_end": false,
		"is_trial_expired": false
	}`)
	if _, _, err := client.CancelSubscriptionWithRequest(
		t.Context(),
		&management.CancelSubscriptionRequest{
			Reason:            "customer request",
			CancelImmediately: true,
		},
	); err != nil {
		t.Fatalf("CancelSubscriptionWithRequest() error = %v, want nil", err)
	}
	assertManagementRequest(
		t,
		requests,
		http.MethodPost,
		"/billing/subscription/cancel",
		url.Values{},
		`{"reason":"customer request","cancel_immediately":true}`,
	)
}

func TestBillingResponseTypes_UnmarshalRefactoredOpenAPIFields(t *testing.T) {
	t.Parallel()

	t.Run("plan", func(t *testing.T) {
		var got management.Plan
		mustUnmarshalJSON(t, `{
			"id": 2,
			"name": "Pro",
			"slug": "pro",
			"price_monthly": 20,
			"price_yearly": 200,
			"currency": "USD",
			"stripe_price_id_monthly": "price_monthly",
			"stripe_price_id_yearly": "price_yearly",
			"monthly_request_quota": 100000,
			"monthly_cost_quota": 500,
			"features": "routing,audit",
			"feature_items": ["routing", "audit"],
			"created_at": "2026-07-25T00:00:00Z",
			"updated_at": "2026-07-25T00:00:01Z"
		}`, &got)
		if got.ID != "2" || got.StripePriceIDMonthly == nil ||
			*got.StripePriceIDMonthly != "price_monthly" ||
			got.MonthlyRequestQuota == nil || *got.MonthlyRequestQuota != 100000 ||
			got.CreatedAt == "" || got.UpdatedAt == "" {
			t.Errorf("Plan = %+v, want current Open API fields", got)
		}
	})

	t.Run("subscription", func(t *testing.T) {
		var got management.Subscription
		mustUnmarshalJSON(t, `{
			"user_id": 10,
			"subscription_plan_id": 2,
			"subscription_status": "active",
			"subscription_tier": "pro",
			"effective_subscription_plan_id": 2,
			"effective_subscription_tier": "pro",
			"stripe_customer_id": "cus_123",
			"stripe_subscription_id": "sub_123",
			"current_period_start": "2026-07-01T00:00:00Z",
			"current_period_end": "2026-08-01T00:00:00Z",
			"max_workspaces": 5,
			"current_workspace_count": 2,
			"can_create_more_workspaces": true,
			"subscription_cancel_at_period_end": true,
			"pending_plan_id": 1,
			"is_trial_expired": false
		}`, &got)
		if got.UserID != 10 || got.SubscriptionPlanID == nil ||
			*got.SubscriptionPlanID != 2 || got.SubscriptionStatus != "active" ||
			got.Status != "active" || got.PlanID != "2" ||
			!got.SubscriptionCancelAtPeriodEnd || !got.CancelAtPeriodEnd ||
			got.PendingPlanID == nil || *got.PendingPlanID != 1 {
			t.Errorf("Subscription = %+v, want current and legacy aliases", got)
		}
	})

	t.Run("subscription checkout", func(t *testing.T) {
		var got management.CreateSubscriptionCheckoutResponse
		mustUnmarshalJSON(t, `{
			"session_id": "cs_123",
			"checkout_url": "https://checkout.example/123",
			"order_id": 59,
			"status": "open",
			"expires_at": "2026-07-25T01:00:00Z",
			"metadata": {"plan_id": "pro"}
		}`, &got)
		if got.OrderID != "59" || got.Status != "open" || got.ExpiresAt == "" ||
			got.Metadata["plan_id"] != "pro" {
			t.Errorf("CreateSubscriptionCheckoutResponse = %+v, want current Open API fields", got)
		}
	})

	t.Run("subscription in app", func(t *testing.T) {
		var got management.CreateSubscriptionInAppResponse
		mustUnmarshalJSON(t, `{
			"client_secret": "secret_123",
			"stripe_subscription_id": "sub_123",
			"payment_intent_id": "pi_123",
			"status": "requires_action",
			"amount": 20,
			"currency": "USD"
		}`, &got)
		if got.StripeSubscriptionID != "sub_123" || got.Amount != 20 || got.Currency != "USD" {
			t.Errorf("CreateSubscriptionInAppResponse = %+v, want current Open API fields", got)
		}
	})

	t.Run("topup plan", func(t *testing.T) {
		var got management.TopupPlan
		mustUnmarshalJSON(t, `{
			"id": "flex",
			"name": "Flexible",
			"currency": "USD",
			"min_amount": 10,
			"max_amount": 1000,
			"fee_rate": 0.03,
			"payment_channels": ["stripe"],
			"custom_amount": true
		}`, &got)
		if got.Name != "Flexible" || got.MinAmount != 10 || got.MaxAmount != 1000 ||
			got.FeeRate != 0.03 || len(got.PaymentChannels) != 1 || !got.CustomAmount {
			t.Errorf("TopupPlan = %+v, want current Open API fields", got)
		}
	})

	t.Run("checkout", func(t *testing.T) {
		var got management.CreateTopupCheckoutResponse
		mustUnmarshalJSON(t, `{
			"order_id": 55,
			"workspace_id": 7,
			"out_trade_no": "trade_55",
			"checkout_url": "https://checkout.example/55",
			"session_id": "cs_55",
			"amount": 100,
			"currency": "USD",
			"order_type": "topup",
			"status": "pending",
			"expires_at": "2026-07-25T01:00:00Z",
			"stripe_mode": "payment"
		}`, &got)
		if got.OrderID != "55" || got.WorkspaceID != 7 || got.OutTradeNo != "trade_55" ||
			got.OrderType != "topup" || got.StripeMode != "payment" {
			t.Errorf("CreateTopupCheckoutResponse = %+v, want current Open API fields", got)
		}
	})

	t.Run("in app payment", func(t *testing.T) {
		var got management.CreateTopupInAppResponse
		mustUnmarshalJSON(t, `{
			"order_id": 56,
			"workspace_id": 7,
			"status": "requires_action",
			"order_type": "topup",
			"amount": 100,
			"currency": "USD",
			"balance": 20,
			"total_recharged": 80,
			"payment_intent_id": "pi_56",
			"client_secret": "secret_56",
			"message": "confirm"
		}`, &got)
		if got.OrderID != "56" || got.WorkspaceID != 7 || got.OrderType != "topup" ||
			got.Balance != 20 || got.TotalRecharged != 80 || got.Message != "confirm" {
			t.Errorf("CreateTopupInAppResponse = %+v, want current Open API fields", got)
		}
	})

	t.Run("order", func(t *testing.T) {
		var got management.Order
		mustUnmarshalJSON(t, `{
			"id": 57,
			"workspace_id": 7,
			"user_id": 10,
			"amount": 100,
			"pay_amount": 103,
			"currency": "USD",
			"order_type": "topup",
			"status": "paid",
			"out_trade_no": "trade_57",
			"payment_type": "stripe",
			"stripe_session_id": "cs_57",
			"stripe_payment_intent_id": "pi_57",
			"expires_at": "2026-07-25T01:00:00Z",
			"created_at": "2026-07-25T00:00:00Z",
			"updated_at": "2026-07-25T00:01:00Z"
		}`, &got)
		if got.ID != "57" || got.WorkspaceID == nil || *got.WorkspaceID != 7 ||
			got.PayAmount != 103 || got.OrderType != "topup" || got.Type != "topup" ||
			got.OutTradeNo != "trade_57" || got.StripePaymentIntentID != "pi_57" {
			t.Errorf("Order = %+v, want current and legacy aliases", got)
		}
	})

	t.Run("payment method", func(t *testing.T) {
		var got management.PaymentMethod
		mustUnmarshalJSON(t, `{
			"id": "pm_1",
			"type": "card",
			"brand": "visa",
			"last4": "4242",
			"expiry_month": 12,
			"expiry_year": 2030,
			"is_default": true,
			"created_at": "2026-07-25T00:00:00Z",
			"billing_email": "billing@example.com"
		}`, &got)
		if got.ExpiryMonth != 12 || got.ExpiryYear != 2030 ||
			got.ExpMonth != 12 || got.ExpYear != 2030 ||
			got.BillingEmail != "billing@example.com" {
			t.Errorf("PaymentMethod = %+v, want current and legacy expiry fields", got)
		}
	})

	t.Run("balance", func(t *testing.T) {
		var got management.Balance
		mustUnmarshalJSON(t, `{"workspace_id":7,"balance":42.5}`, &got)
		if got.WorkspaceID != 7 || got.Balance != 42.5 || got.Amount != 42.5 {
			t.Errorf("Balance = %+v, want current and legacy amount fields", got)
		}
	})

	t.Run("invoice", func(t *testing.T) {
		var got management.Invoice
		mustUnmarshalJSON(t, `{
			"id": 58,
			"workspace_id": 7,
			"stripe_invoice_id": "in_58",
			"invoice_number": "INV-58",
			"amount_due": 100,
			"amount_paid": 100,
			"currency": "USD",
			"status": "paid",
			"invoice_pdf_url": "https://example.com/invoice.pdf",
			"hosted_invoice_url": "https://example.com/invoice",
			"period_start": "2026-07-01T00:00:00Z",
			"period_end": "2026-08-01T00:00:00Z",
			"created_at": "2026-07-25T00:00:00Z",
			"updated_at": "2026-07-25T00:01:00Z"
		}`, &got)
		if got.ID != "58" || got.WorkspaceID != 7 || got.StripeInvoiceID != "in_58" ||
			got.AmountDue != 100 || got.AmountPaid != 100 || got.Amount != 100 ||
			got.InvoicePDFURL != "https://example.com/invoice.pdf" {
			t.Errorf("Invoice = %+v, want current and legacy amount fields", got)
		}
	})

	t.Run("billing details wrapper", func(t *testing.T) {
		var got management.BillingDetails
		mustUnmarshalJSON(t, `{
			"workspace_id": 7,
			"details": {
				"name": "Acme",
				"email": "billing@example.com",
				"tax_id": "tax-123",
				"phone": "+1-555-0100",
				"address": "1 Main Street",
				"cc_recipients": ["finance@example.com"]
			}
		}`, &got)
		if got.WorkspaceID != 7 || got.Name != "Acme" || got.CompanyName != "Acme" ||
			got.AddressText != "1 Main Street" || len(got.Details) != 6 {
			t.Errorf("BillingDetails = %+v, want unwrapped current fields", got)
		}
	})

	t.Run("billing overview", func(t *testing.T) {
		var got management.BillingOverview
		mustUnmarshalJSON(t, `{
			"workspace_id": 7,
			"balance": 42.5,
			"monthly_budget_limit": 100,
			"current_month_cost": 10,
			"current_month_tokens": 200,
			"current_month_requests": 3,
			"billing_details": {"name": "Acme"},
			"recent_invoices": [],
			"recent_invoices_page_info": {"next_cursor": "next", "has_more": true}
		}`, &got)
		if got.WorkspaceID != 7 || got.BalanceAmount != 42.5 ||
			got.MonthlyBudgetLimit == nil || *got.MonthlyBudgetLimit != 100 ||
			got.BillingDetails["name"] != "Acme" ||
			!got.RecentInvoicesPageInfo.HasMore ||
			got.RecentInvoicesPageInfo.NextCursor != "next" {
			t.Errorf("BillingOverview = %+v, want current Open API fields", got)
		}
	})
}

func TestFinancialRequestTypes_MarshalRefactoredOpenAPIFields(t *testing.T) {
	t.Parallel()

	workspaceCap := management.BudgetCap{
		ScopeType:     "workspace",
		Enabled:       true,
		MonthlyAmount: 1000,
	}
	teamID := int64(3)
	teamCap := management.BudgetCap{
		ScopeType:     "team",
		ScopeID:       &teamID,
		TeamID:        &teamID,
		Name:          "Core",
		Enabled:       true,
		MonthlyAmount: 500,
	}
	exceededAction := "block"
	enabled := true
	monthlyAmount := 750.0
	warningPercent := 75.5
	criticalPercent := 90.5
	usageMultiplier := 1.25

	tests := []struct {
		name  string
		value any
		want  string
	}{
		{
			name: "governance",
			value: management.UpdateFinancialGovernanceRequest{
				WorkspaceCap: &workspaceCap,
				TeamCaps:     []management.BudgetCap{teamCap},
				MemberLimits: []management.SpendingLimit{{
					UserID:       10,
					DailyLimit:   20,
					WeeklyLimit:  100,
					MonthlyLimit: 300,
				}},
				Thresholds: &management.Thresholds{
					WarningPercent:  75.5,
					CriticalPercent: 90.5,
					ExceededAction:  "block",
				},
				ExceededAction:  &exceededAction,
				UsageMultiplier: &usageMultiplier,
			},
			want: `{
				"workspace_cap":{"scope_type":"workspace","enabled":true,"monthly_amount":1000},
				"team_caps":[{"scope_type":"team","scope_id":3,"team_id":3,"name":"Core","enabled":true,"monthly_amount":500}],
				"member_limits":[{"user_id":10,"daily_limit":20,"weekly_limit":100,"monthly_limit":300}],
					"thresholds":{"warning_percent":75.5,"critical_percent":90.5,"exceeded_action":"block"},
					"exceeded_action":"block",
					"usage_multiplier":1.25
				}`,
		},
		{
			name: "budget caps",
			value: management.UpdateBudgetCapsRequest{
				WorkspaceCap: &workspaceCap,
				TeamCaps:     []management.BudgetCap{teamCap},
			},
			want: `{
				"workspace_cap":{"scope_type":"workspace","enabled":true,"monthly_amount":1000},
				"team_caps":[{"scope_type":"team","scope_id":3,"team_id":3,"name":"Core","enabled":true,"monthly_amount":500}]
			}`,
		},
		{
			name: "scope cap",
			value: management.UpdateScopeBudgetCapRequest{
				Enabled:       &enabled,
				MonthlyAmount: &monthlyAmount,
			},
			want: `{"enabled":true,"monthly_amount":750}`,
		},
		{
			name: "thresholds",
			value: management.UpdateThresholdsRequest{
				WarningPercent:  &warningPercent,
				CriticalPercent: &criticalPercent,
				ExceededAction:  &exceededAction,
			},
			want: `{
				"warning_percent":75.5,
				"critical_percent":90.5,
				"exceeded_action":"block"
			}`,
		},
		{
			name: "preview",
			value: management.PreviewFinancialChangesRequest{
				WorkspaceCap: &workspaceCap,
				TeamCaps:     []management.BudgetCap{teamCap},
			},
			want: `{
				"workspace_cap":{"scope_type":"workspace","enabled":true,"monthly_amount":1000},
				"team_caps":[{"scope_type":"team","scope_id":3,"team_id":3,"name":"Core","enabled":true,"monthly_amount":500}]
			}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.value)
			if err != nil {
				t.Fatalf("json.Marshal(%T) error = %v", tt.value, err)
			}
			assertJSONSemanticallyEqual(t, got, tt.want)
		})
	}
}

func TestFinancialResponseTypes_UnmarshalRefactoredOpenAPIFields(t *testing.T) {
	t.Parallel()

	var governance management.FinancialGovernance
	mustUnmarshalJSON(t, `{
		"workspace_id": 7,
		"workspace_cap": {"scope_type":"workspace","enabled":true,"monthly_amount":1000,"current_spend":100},
		"team_caps": [{"scope_type":"team","scope_id":3,"team_id":3,"name":"Core","enabled":true,"monthly_amount":500,"current_spend":50}],
		"member_caps": [],
		"gateway_key_caps": [],
		"member_limits": [{"user_id":10,"email":"owner@example.com","name":"Owner","daily_limit":20,"weekly_limit":100,"monthly_limit":300,"daily_spend":2,"weekly_spend":10,"monthly_spend":30}],
			"thresholds": {"warning_percent":75.5,"critical_percent":90.5,"exceeded_action":"block"},
			"exceeded_action": "block",
			"usage_multiplier": 1.25,
			"currency": "USD",
		"notification_channels": ["email"]
	}`, &governance)
	if governance.WorkspaceID != 7 || governance.WorkspaceCap.MonthlyAmount != 1000 ||
		len(governance.TeamCaps) != 1 || governance.TeamCaps[0].TeamID == nil ||
		len(governance.MemberLimits) != 1 || governance.MemberLimits[0].MonthlySpend != 30 ||
		governance.Thresholds.WarningPercent != 75.5 ||
		governance.UsageMultiplier != 1.25 ||
		len(governance.NotificationChannels) != 1 {
		t.Errorf("FinancialGovernance = %+v, want current Open API fields", governance)
	}

	var caps management.BudgetCaps
	mustUnmarshalJSON(t, `{
		"workspace_id": 7,
		"workspace_cap": {"scope_type":"workspace","enabled":true,"monthly_amount":1000,"current_spend":100},
		"team_caps": [],
		"member_caps": [],
		"gateway_key_caps": [],
		"member_limits": [],
			"thresholds": {"warning_percent":75,"critical_percent":90,"exceeded_action":"notify"},
			"exceeded_action": "notify",
			"usage_multiplier": 1.25,
			"currency": "USD",
		"notification_channels": []
	}`, &caps)
	if caps.WorkspaceID != 7 || caps.WorkspaceCap.MonthlyAmount != 1000 ||
		caps.Workspace == nil || caps.Workspace.Limit != 1000 ||
		caps.UsageMultiplier != 1.25 {
		t.Errorf("BudgetCaps = %+v, want current and legacy aliases", caps)
	}

	t.Run("clears a decoded current cap instead of restoring the legacy limit", func(t *testing.T) {
		var cap management.BudgetCap
		mustUnmarshalJSON(t, `{
			"scope_type":"workspace",
			"enabled":true,
			"monthly_amount":100
		}`, &cap)
		cap.MonthlyAmount = 0

		got, err := json.Marshal(cap)
		if err != nil {
			t.Fatalf("json.Marshal(BudgetCap) error = %v", err)
		}
		assertJSONSemanticallyEqual(t, got, `{
			"scope_type":"workspace",
			"enabled":true,
			"monthly_amount":0
		}`)
	})

	var summary management.SpendSummary
	mustUnmarshalJSON(t, `{
		"workspace_id": 7,
		"current_period": "2026-07",
		"workspace_spend": {"scope_type":"workspace","scope_id":7,"name":"Acme","spend":100,"limit":1000,"percent":10,"status":"ok"},
		"team_spends": [{"scope_type":"team","scope_id":3,"name":"Core","spend":50,"limit":500,"percent":10,"status":"ok"}],
		"member_spends": [],
		"gateway_key_spends": []
	}`, &summary)
	if summary.WorkspaceID != 7 || summary.CurrentPeriod != "2026-07" ||
		summary.WorkspaceSpend.Spend != 100 || len(summary.TeamSpends) != 1 ||
		summary.TeamSpends[0].Status != "ok" {
		t.Errorf("SpendSummary = %+v, want current Open API fields", summary)
	}
}

func TestClient_GetSpendingLimits_UsesDeclaredQueryContract(t *testing.T) {
	t.Parallel()

	t.Run("legacy list options do not send unsupported pagination", func(t *testing.T) {
		t.Parallel()
		client, requests := newManagementContractClient(t, `{"items":[],"page_info":{"has_more":false}}`)
		if _, _, err := client.GetSpendingLimits(
			t.Context(),
			7,
			management.ListOptions{Limit: 5, Cursor: "ignored"},
		); err != nil {
			t.Fatalf("GetSpendingLimits() error = %v, want nil", err)
		}
		assertManagementRequest(
			t,
			requests,
			http.MethodGet,
			"/workspaces/7/governance/financial/spending-limits",
			url.Values{},
			"",
		)
	})

	t.Run("filters", func(t *testing.T) {
		t.Parallel()
		client, requests := newManagementContractClient(t, `{"items":[],"page_info":{"has_more":false}}`)
		if _, _, err := client.GetSpendingLimitsWithFilters(
			t.Context(),
			7,
			management.SpendingLimitOptions{TeamID: 3, UserID: 10},
		); err != nil {
			t.Fatalf("GetSpendingLimitsWithFilters() error = %v, want nil", err)
		}
		assertManagementRequest(
			t,
			requests,
			http.MethodGet,
			"/workspaces/7/governance/financial/spending-limits",
			url.Values{"team_id": {"3"}, "user_id": {"10"}},
			"",
		)
	})
}

func TestCatalogResponseTypes_UnmarshalRefactoredOpenAPIFields(t *testing.T) {
	t.Parallel()

	var model management.Model
	mustUnmarshalJSON(t, `{
		"id": 21,
		"model_id": "gpt-4.1",
		"developer": "openai",
		"type": "chat",
		"name": "GPT-4.1",
		"icon": "openai",
		"group": "gpt-4",
		"status": "active",
		"route_count": 2,
		"created_at": "2026-07-25T00:00:00Z",
		"updated_at": "2026-07-25T00:01:00Z",
		"model_card": {"context_window": 1000000},
		"routes": [{
			"id": 31,
			"route_id": "route_31",
			"workspace_id": 7,
			"model": "gpt-4.1",
			"actual_model": "gpt-4.1-2025-04-14",
			"match_source": "exact",
			"channel_id": 41,
			"channel_name": "OpenAI",
			"channel_type": "openai",
			"channel_status": "active",
			"provider_source": "byok",
			"provider_platform": "openai",
			"price": {"input": 2},
			"price_reference_id": "price_1"
		}]
	}`, &model)
	if model.ID != "gpt-4.1" || model.ModelID != "gpt-4.1" ||
		model.Developer != "openai" || model.Type != "chat" ||
		model.RouteCount != 2 || model.CreatedAt == "" || len(model.ModelCard) == 0 ||
		len(model.Routes) != 1 || model.Routes[0].ID != "31" ||
		model.Routes[0].RouteID != "route_31" || model.Routes[0].ActualModel == "" ||
		model.Routes[0].ProviderPlatform == nil {
		t.Errorf("Model = %+v, want current catalog detail fields", model)
	}

	var preview management.PreviewRouteResponse
	mustUnmarshalJSON(t, `{
		"workspace_id": 7,
		"model": "gpt-4.1",
		"candidate_count": 1,
		"candidates": [{
			"id": 31,
			"route_id": "route_31",
			"workspace_id": 7,
			"model": "gpt-4.1",
			"actual_model": "gpt-4.1-2025-04-14",
			"match_source": "exact",
			"channel_id": 41,
			"channel_name": "OpenAI",
			"channel_type": "openai",
			"channel_status": "active",
			"provider_source": "byok"
		}],
		"preview_metadata": {"key_id": 9}
	}`, &preview)
	if preview.WorkspaceID != 7 || preview.CandidateCount != 1 ||
		len(preview.Candidates) != 1 || preview.Candidates[0].ChannelName != "OpenAI" ||
		preview.PreviewMetadata["key_id"] == nil {
		t.Errorf("PreviewRouteResponse = %+v, want current preview fields", preview)
	}
}

func TestObservabilityResponseTypes_UnmarshalRefactoredOpenAPIFields(t *testing.T) {
	t.Parallel()

	var log management.RequestLog
	mustUnmarshalJSON(t, `{
		"id": 51,
		"request_id": "req_51",
		"trace_id": "trace_51",
		"workspace_id": 7,
		"provider": "OpenAI",
		"provider_id": 11,
		"provider_source": "byok",
		"model": "gpt-4.1",
		"format": "responses",
		"status": "success",
		"status_code": 200,
		"latency_ms": 1234,
		"input_tokens": 100,
		"output_tokens": 50,
		"total_tokens": 150,
		"total_cost": 0.01,
		"gateway_key_id": 9,
		"key_name": "prod",
		"user_id": 10,
		"member_name": "Owner",
		"team_id": 3,
		"team_name": "Core",
		"stream": true,
		"created_at": "2026-07-25T00:00:00Z"
	}`, &log)
	if log.ID != 51 || log.TraceID != "trace_51" || log.WorkspaceID != 7 ||
		log.ProviderID != 11 || log.LatencyMS == nil || *log.LatencyMS != 1234 ||
		log.TotalTokens != 150 || log.TotalCost != 0.01 ||
		log.Cost != 0.01 || log.Duration != 1234 || log.TeamID == nil {
		t.Errorf("RequestLog = %+v, want current and legacy fields", log)
	}

	var trace management.Trace
	mustUnmarshalJSON(t, `{
		"id": 61,
		"trace_id": "trace_61",
		"workspace_id": 7,
		"thread_id": 71,
		"created_at": "2026-07-25T00:00:00Z",
		"updated_at": "2026-07-25T00:01:00Z"
	}`, &trace)
	if trace.ID != 61 || trace.WorkspaceID != 7 || trace.ThreadID != "71" ||
		trace.UpdatedAt == "" {
		t.Errorf("Trace = %+v, want current observability fields", trace)
	}
}

func TestPolicyContracts_UseRefactoredOpenAPIFields(t *testing.T) {
	t.Parallel()

	t.Run("typed policy configuration", func(t *testing.T) {
		var got management.GatewayPolicy
		mustUnmarshalJSON(t, `{
			"workspace_id":7,
			"key_id":77,
			"model_policies":[{
				"id":1,
				"provider":"openai",
				"model_id":"gpt-4.1",
				"action":"allow",
				"priority":10,
				"scope":"workspace",
				"applies_to":["chat"]
			}],
			"keyword_policies":[],
			"budget_policies":[{
				"id":2,
				"scope":"member",
				"scope_id":37,
				"monthly_limit":100,
				"current_spend":25,
				"enforcement_mode":"block"
			}],
			"log_policies":[{
				"id":3,
				"scope":"workspace",
				"capture_request":true,
				"capture_response":false,
				"retention_days":30
			}],
			"rate_limits":[]
		}`, &got)
		if len(got.ModelPolicies) != 1 ||
			got.ModelPolicies[0].ModelID != "gpt-4.1" ||
			len(got.BudgetPolicies) != 1 ||
			got.BudgetPolicies[0].ScopeID != 37 ||
			len(got.LogPolicies) != 1 ||
			!got.LogPolicies[0].CaptureRequest {
			t.Errorf("GatewayPolicy = %+v, want fully typed policy groups", got)
		}
	})

	t.Run("zero key ID is omitted", func(t *testing.T) {
		t.Parallel()
		client, requests := newManagementContractClient(t, `{
			"workspace_id":7,
			"model_policies":[],
			"keyword_policies":[],
			"budget_policies":[],
			"log_policies":[],
			"rate_limits":[]
		}`)
		got, _, err := client.GetGatewayPolicies(
			t.Context(),
			0,
			owlvigil.WithQueryParam("workspace_id", "7"),
		)
		if err != nil {
			t.Fatalf("GetGatewayPolicies() error = %v, want nil", err)
		}
		if got.WorkspaceID != 7 || got.KeywordPolicies == nil || got.RateLimits == nil {
			t.Errorf("GetGatewayPolicies() = %+v, want current response fields", got)
		}
		assertManagementRequest(
			t,
			requests,
			http.MethodGet,
			"/gateway/policies",
			url.Values{"workspace_id": {"7"}},
			"",
		)
	})

	t.Run("preview body", func(t *testing.T) {
		t.Parallel()
		keyID := int64(9)
		body, err := json.Marshal(management.PreviewPolicyRequest{
			WorkspaceID: 7,
			KeyID:       keyID,
			Model:       "gpt-4.1",
			Request:     map[string]any{"input": "hello"},
		})
		if err != nil {
			t.Fatalf("json.Marshal(PreviewPolicyRequest) error = %v", err)
		}
		assertJSONSemanticallyEqual(
			t,
			body,
			`{"workspace_id":7,"key_id":9,"model":"gpt-4.1","request":{"input":"hello"}}`,
		)
	})

	t.Run("preview response", func(t *testing.T) {
		t.Parallel()
		var got management.PreviewPolicyResponse
		mustUnmarshalJSON(t, `{
			"allowed": true,
			"blocked_by": [],
			"modified_by": ["policy_1"],
			"redirected_to": "gpt-4.1-mini",
			"budget_check": "ok",
			"rate_limit_check": "ok",
			"warnings": []
		}`, &got)
		if !got.Allowed || len(got.ModifiedBy) != 1 || got.RedirectedTo == nil ||
			got.BudgetCheck != "ok" || got.RateLimitCheck != "ok" {
			t.Errorf("PreviewPolicyResponse = %+v, want current fields", got)
		}
	})

	t.Run("add prompt keyword", func(t *testing.T) {
		t.Parallel()
		client, requests := newManagementContractClient(t, `{
			"workspace_id":7,
			"model_policies":[],
			"keyword_policies":[{
				"id":31,
				"keyword":"SensitiveToken",
				"action":"block",
				"enabled":true,
				"scope":"workspace",
				"description":"SDK test"
			}],
			"budget_policies":[],
			"log_policies":[],
			"rate_limits":[]
		}`)
		enabled := true
		req := &management.AddPromptKeywordRequest{
			WorkspaceID: 7,
			Keyword:     "SensitiveToken",
			Description: "SDK test",
			Enabled:     &enabled,
		}
		got, _, err := client.AddPromptKeyword(t.Context(), req)
		if err != nil {
			t.Fatalf("AddPromptKeyword(%+v) error = %v, want nil", req, err)
		}
		if len(got.KeywordPolicies) != 1 || got.KeywordPolicies[0].ID != 31 ||
			got.KeywordPolicies[0].Keyword != "SensitiveToken" {
			t.Errorf("AddPromptKeyword(%+v) = %+v, want created keyword policy", req, got)
		}
		assertManagementRequest(
			t,
			requests,
			http.MethodPost,
			"/gateway/policies/keywords",
			url.Values{},
			`{"workspace_id":7,"keyword":"SensitiveToken","description":"SDK test","enabled":true}`,
		)
	})

	t.Run("delete prompt keyword", func(t *testing.T) {
		t.Parallel()
		client, requests := newManagementContractClient(t, `{
			"workspace_id":7,
			"model_policies":[],
			"keyword_policies":[],
			"budget_policies":[],
			"log_policies":[],
			"rate_limits":[]
		}`)
		got, _, err := client.DeletePromptKeyword(t.Context(), 7, 31)
		if err != nil {
			t.Fatalf("DeletePromptKeyword(7, 31) error = %v, want nil", err)
		}
		if got.WorkspaceID != 7 || len(got.KeywordPolicies) != 0 {
			t.Errorf("DeletePromptKeyword(7, 31) = %+v, want no keyword policies", got)
		}
		assertManagementRequest(
			t,
			requests,
			http.MethodDelete,
			"/gateway/policies/keywords/31",
			url.Values{"workspace_id": {"7"}},
			"",
		)
	})
}

func TestWebhookContracts_UseRefactoredOpenAPIFields(t *testing.T) {
	t.Parallel()

	t.Run("create legacy events alias", func(t *testing.T) {
		t.Parallel()
		got, err := json.Marshal(management.CreateWebhookEndpointRequest{
			WorkspaceID: 7,
			URL:         "https://example.com/hooks/owlvigil",
			Events:      []string{"gateway.key.updated"},
		})
		if err != nil {
			t.Fatalf("json.Marshal(CreateWebhookEndpointRequest) error = %v", err)
		}
		assertJSONSemanticallyEqual(
			t,
			got,
			`{"workspace_id":7,"url":"https://example.com/hooks/owlvigil","event_types":["gateway.key.updated"]}`,
		)
	})

	t.Run("update empty event types", func(t *testing.T) {
		t.Parallel()
		got, err := json.Marshal(management.UpdateWebhookEndpointRequest{
			EventTypes: []string{},
		})
		if err != nil {
			t.Fatalf("json.Marshal(UpdateWebhookEndpointRequest) error = %v", err)
		}
		assertJSONSemanticallyEqual(t, got, `{"event_types":[]}`)
	})

	t.Run("bulk request omits removed time filters", func(t *testing.T) {
		t.Parallel()
		start, end := "2026-07-01T00:00:00Z", "2026-07-31T23:59:59Z"
		got, err := json.Marshal(management.BulkRedeliverRequest{
			WorkspaceID: 7,
			EventIDs:    []int{91},
			StartTime:   &start,
			EndTime:     &end,
			Limit:       10,
		})
		if err != nil {
			t.Fatalf("json.Marshal(BulkRedeliverRequest) error = %v", err)
		}
		assertJSONSemanticallyEqual(
			t,
			got,
			`{"workspace_id":7,"event_ids":[91],"limit":10}`,
		)
	})

	t.Run("response fields", func(t *testing.T) {
		t.Parallel()
		var endpoint management.WebhookEndpoint
		mustUnmarshalJSON(t, `{
			"id":81,
			"workspace_id":7,
			"url":"https://example.com/hooks/owlvigil",
			"event_types":["gateway.key.updated"],
			"status":"enabled",
			"secret":"whsec_123",
			"created_at":"2026-07-25T00:00:00Z",
			"updated_at":"2026-07-25T00:01:00Z"
		}`, &endpoint)
		if endpoint.ID != 81 || endpoint.UpdatedAt == "" ||
			len(endpoint.EventTypes) != 1 || len(endpoint.Events) != 1 {
			t.Errorf("WebhookEndpoint = %+v, want current and legacy event fields", endpoint)
		}

		var event management.WebhookEvent
		mustUnmarshalJSON(t, `{
			"id":91,
			"endpoint_id":81,
			"workspace_id":7,
			"event_type":"gateway.key.updated",
			"status":"delivered",
			"payload":{"key_id":9},
			"attempts":2,
			"last_error":"",
			"delivered_at":"2026-07-25T00:02:00Z",
			"created_at":"2026-07-25T00:00:00Z",
			"updated_at":"2026-07-25T00:02:00Z"
		}`, &event)
		if event.ID != "91" || event.WorkspaceID != 7 || event.Attempts != 2 ||
			event.UpdatedAt == "" || event.Type != "gateway.key.updated" {
			t.Errorf("WebhookEvent = %+v, want current and legacy fields", event)
		}
	})
}

func TestStrictQueryContracts_OmitRemovedFilters(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		response string
		call     func(*management.Client) error
		path     string
		query    url.Values
	}{
		{
			name:     "user invitations ignore pagination",
			response: `[]`,
			call: func(client *management.Client) error {
				_, _, err := client.ListUserInvitations(t.Context(), management.ListOptions{Limit: 5, Cursor: "old"})
				return err
			},
			path:  "/users/me/invitations",
			query: url.Values{},
		},
		{
			name:     "invitations ignore pagination",
			response: `{"items":[],"page_info":{"has_more":false}}`,
			call: func(client *management.Client) error {
				_, _, err := client.ListInvitations(t.Context(), 7, management.ListOptions{Limit: 5, Cursor: "old"})
				return err
			},
			path:  "/workspaces/7/invitations",
			query: url.Values{},
		},
		{
			name:     "roles ignore pagination",
			response: `{"items":[],"page_info":{"has_more":false}}`,
			call: func(client *management.Client) error {
				_, _, err := client.ListRoles(t.Context(), 7, management.ListOptions{Limit: 5, Cursor: "old"})
				return err
			},
			path:  "/workspaces/7/roles",
			query: url.Values{},
		},
		{
			name:     "members omit unsupported cursor",
			response: `{"items":[],"page_info":{"has_more":false}}`,
			call: func(client *management.Client) error {
				_, _, err := client.ListMembers(t.Context(), 7, management.ListOptions{Limit: 5, Cursor: "old"})
				return err
			},
			path:  "/workspaces/7/members",
			query: url.Values{"limit": {"5"}},
		},
		{
			name:     "activity omits unsupported cursor",
			response: `[]`,
			call: func(client *management.Client) error {
				_, _, err := client.ListWorkspaceActivity(t.Context(), 7, management.ListOptions{Limit: 5, Cursor: "old"})
				return err
			},
			path:  "/workspaces/7/activity",
			query: url.Values{"limit": {"5"}},
		},
		{
			name:     "gateway keys omit unsupported status",
			response: `{"items":[],"page_info":{"has_more":false}}`,
			call: func(client *management.Client) error {
				_, _, err := client.ListGatewayKeys(
					t.Context(),
					management.ListOptions{Limit: 5},
					"active",
					owlvigil.WithQueryParam("workspace_id", "7"),
				)
				return err
			},
			path:  "/gateway/keys",
			query: url.Values{"workspace_id": {"7"}, "limit": {"5"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			client, requests := newManagementContractClient(t, tt.response)
			if err := tt.call(client); err != nil {
				t.Fatalf("%s error = %v, want nil", tt.name, err)
			}
			assertManagementRequest(t, requests, http.MethodGet, tt.path, tt.query, "")
		})
	}
}

func TestStrictQueryContracts_ExposeAllPublishedFilters(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		response string
		call     func(*management.Client) error
		path     string
		query    url.Values
	}{
		{
			name:     "workspace activity",
			response: `[]`,
			call: func(client *management.Client) error {
				_, _, err := client.ListWorkspaceActivityWithFilters(
					t.Context(),
					7,
					management.WorkspaceActivityOptions{
						Limit:  20,
						Offset: 40,
						Tab:    "audit",
						Search: "gateway",
					},
				)
				return err
			},
			path: "/workspaces/7/activity",
			query: url.Values{
				"limit":  {"20"},
				"offset": {"40"},
				"tab":    {"audit"},
				"search": {"gateway"},
			},
		},
		{
			name:     "workspace members",
			response: `{"items":[],"page_info":{"has_more":false}}`,
			call: func(client *management.Client) error {
				_, _, err := client.ListMembersWithFilters(
					t.Context(),
					7,
					management.MemberListOptions{
						Limit:  25,
						Offset: 50,
						Search: "owner",
						Role:   "admin",
						Status: "active",
						Team:   "Core",
					},
				)
				return err
			},
			path: "/workspaces/7/members",
			query: url.Values{
				"limit":  {"25"},
				"offset": {"50"},
				"search": {"owner"},
				"role":   {"admin"},
				"status": {"active"},
				"team":   {"Core"},
			},
		},
		{
			name:     "gateway routes",
			response: `{"items":[],"page_info":{"has_more":false}}`,
			call: func(client *management.Client) error {
				_, _, err := client.ListRoutesWithFilters(
					t.Context(),
					management.RouteListOptions{
						Cursor: "next",
						Limit:  10,
						KeyID:  31,
						Model:  "gpt-4.1",
					},
					owlvigil.WithWorkspaceID(7),
				)
				return err
			},
			path: "/gateway/routes",
			query: url.Values{
				"workspace_id": {"7"},
				"cursor":       {"next"},
				"limit":        {"10"},
				"key_id":       {"31"},
				"model":        {"gpt-4.1"},
			},
		},
		{
			name:     "gateway route detail",
			response: `{"id":41,"model":"gpt-4.1"}`,
			call: func(client *management.Client) error {
				_, _, err := client.GetRouteWithFilters(
					t.Context(),
					"41",
					management.RouteDetailOptions{
						KeyID: 31,
						Model: "gpt-4.1",
					},
					owlvigil.WithWorkspaceID(7),
				)
				return err
			},
			path: "/gateway/routes/41",
			query: url.Values{
				"workspace_id": {"7"},
				"key_id":       {"31"},
				"model":        {"gpt-4.1"},
			},
		},
		{
			name:     "billing topups",
			response: `{"items":[],"page_info":{"has_more":false}}`,
			call: func(client *management.Client) error {
				_, _, err := client.ListTopupsWithFilters(
					t.Context(),
					management.OrderListOptions{
						Cursor:    "next",
						Limit:     10,
						OrderType: "topup",
						Status:    "paid",
					},
				)
				return err
			},
			path: "/billing/topups",
			query: url.Values{
				"cursor":     {"next"},
				"limit":      {"10"},
				"order_type": {"topup"},
				"status":     {"paid"},
			},
		},
		{
			name:     "billing orders",
			response: `{"items":[],"page_info":{"has_more":false}}`,
			call: func(client *management.Client) error {
				_, _, err := client.ListOrdersWithFilters(
					t.Context(),
					management.OrderListOptions{
						Cursor:    "next",
						Limit:     10,
						OrderType: "subscription",
						Status:    "pending",
					},
				)
				return err
			},
			path: "/billing/orders",
			query: url.Values{
				"cursor":     {"next"},
				"limit":      {"10"},
				"order_type": {"subscription"},
				"status":     {"pending"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			client, requests := newManagementContractClient(t, tt.response)
			if err := tt.call(client); err != nil {
				t.Fatalf("%s error = %v, want nil", tt.name, err)
			}
			assertManagementRequest(t, requests, http.MethodGet, tt.path, tt.query, "")
		})
	}
}

func TestAccountResponseTypes_UnmarshalRefactoredOpenAPIFields(t *testing.T) {
	t.Parallel()

	var link management.InviteLink
	mustUnmarshalJSON(t, `{
		"invite_code":"OWL123",
		"invite_link":"https://app.owlvigil.com/invite/OWL123",
		"total_invitations":10,
		"converted_invitations":4,
		"pending_invitations":6,
		"conversion_rate":40
	}`, &link)
	if link.InviteCode != "OWL123" ||
		link.InviteLink != "https://app.owlvigil.com/invite/OWL123" ||
		link.InviteURL != link.InviteLink ||
		link.TotalInvitations != 10 || link.ConvertedInvitations != 4 ||
		link.Stats == nil || link.Stats.AcceptedInvites != 4 {
		t.Errorf("InviteLink = %+v, want current and legacy invitation stats", link)
	}

	var invitation management.UserInvitation
	mustUnmarshalJSON(t, `{
		"id":71,
		"invite_code":"OWL123",
		"inviter_user_id":10,
		"invited_user_id":11,
		"invited_email":"friend@example.com",
		"status":"converted",
		"created_at":"2026-07-25T00:00:00Z",
		"converted_at":"2026-07-25T01:00:00Z"
	}`, &invitation)
	if invitation.ID != 71 || invitation.InviteCode != "OWL123" ||
		invitation.InviterUserID != 10 || invitation.InvitedUserID == nil ||
		invitation.InvitedEmail != "friend@example.com" ||
		invitation.Email != "friend@example.com" || invitation.CreatedAt == "" {
		t.Errorf("UserInvitation = %+v, want current and legacy fields", invitation)
	}
}

func TestLogsResponseTypes_UnmarshalRefactoredOpenAPIFields(t *testing.T) {
	t.Parallel()

	var access management.PayloadAccess
	mustUnmarshalJSON(t, `{
		"enabled":true,
		"capture_request_body":true,
		"capture_response_body":false,
		"capture_stream_chunks":true,
		"message":"enabled",
		"response_disabled_message":"response capture disabled"
	}`, &access)
	if !access.Enabled || !access.Allowed || !access.CaptureRequestBody ||
		access.CaptureResponseBody || !access.CaptureStreamChunks ||
		access.Message != "enabled" {
		t.Errorf("PayloadAccess = %+v, want current and legacy fields", access)
	}

	var payload management.PayloadLog
	mustUnmarshalJSON(t, `{
		"id":81,
		"request_id":"req_81",
		"workspace_id":7,
		"api_key_id":9,
		"channel_id":11,
		"trace_id":13,
		"trace_key":"trace_81",
		"model":"gpt-4.1",
		"format":"responses",
		"status":"success",
		"stream":true,
		"client_ip":"127.0.0.1",
		"external_id":"ext_81",
		"duration_ms":1000,
		"first_token_ms":100,
		"reasoning_duration_ms":200,
		"request_body":{"input":"hello"},
		"response_body":{"output":"world"},
		"response_chunks":[{"delta":"world"}],
		"executions":[{
			"id":91,
			"channel_id":11,
			"model":"gpt-4.1",
			"format":"responses",
			"status":"success",
			"status_code":200,
			"stream":true,
			"created_at":"2026-07-25T00:00:00Z",
			"updated_at":"2026-07-25T00:00:01Z"
		}],
		"created_at":"2026-07-25T00:00:00Z",
		"updated_at":"2026-07-25T00:00:01Z"
	}`, &payload)
	if payload.ID != 81 || payload.PayloadID != "81" || payload.WorkspaceID != 7 ||
		payload.APIKeyID == nil || payload.DurationMS == nil ||
		payload.RequestBody == nil || payload.Request == nil ||
		payload.ResponseBody == nil || payload.Response == nil ||
		len(payload.Executions) != 1 || payload.Executions[0].StatusCode == nil {
		t.Errorf("PayloadLog = %+v, want current and legacy fields", payload)
	}
}

func TestProviderResponseType_UnmarshalRefactoredOpenAPIFields(t *testing.T) {
	t.Parallel()

	var provider management.Provider
	mustUnmarshalJSON(t, `{
		"id":22,
		"workspace_id":7,
		"name":"Primary",
		"type":"openai",
		"provider_source":"byok",
		"provider_platform":"openai",
		"is_platform_managed":false,
		"base_url":"https://api.openai.com",
		"default_model":"gpt-4.1",
		"api_mode":"responses",
		"status":"active",
		"request_count":10,
		"error_count":1,
		"last_used_at":"2026-07-25T00:00:00Z",
		"created_at":"2026-07-24T00:00:00Z",
		"updated_at":"2026-07-25T00:00:00Z"
	}`, &provider)
	if provider.ID != 22 || provider.ProviderPlatform == nil ||
		*provider.ProviderPlatform != "openai" || provider.IsPlatformManaged ||
		provider.RequestCount != 10 || provider.LastUsedAt == "" {
		t.Errorf("Provider = %+v, want current provider fields", provider)
	}
}

func TestRemainingResponseAliases_UnmarshalRefactoredOpenAPIFields(t *testing.T) {
	t.Parallel()

	t.Run("checkout confirmation", func(t *testing.T) {
		var got management.Order
		mustUnmarshalJSON(t, `{
			"order_id":101,
			"workspace_id":7,
			"status":"completed",
			"order_type":"topup",
			"balance":50,
			"total_recharged":150
		}`, &got)
		if got.ID != "101" || got.WorkspaceID == nil || *got.WorkspaceID != 7 ||
			got.OrderType != "topup" || got.Balance != 50 || got.TotalRecharged != 150 {
			t.Errorf("Order checkout confirmation = %+v, want current confirmation fields", got)
		}
	})

	t.Run("updated model policy", func(t *testing.T) {
		var got management.GatewayPolicy
		mustUnmarshalJSON(t, `{
			"id":111,
			"provider":"openai",
			"model_id":"gpt-4.1",
			"action":"redirect",
			"redirect_to":"gpt-4.1-mini",
			"priority":1,
			"scope":"workspace",
			"applies_to":["chat"],
			"reason":"cost"
		}`, &got)
		if got.ID != 111 || got.Provider != "openai" || got.ModelID != "gpt-4.1" ||
			got.Action != "redirect" || got.RedirectTo == nil || got.Scope != "workspace" {
			t.Errorf("GatewayPolicy model update = %+v, want current ModelPolicy fields", got)
		}
	})

	t.Run("quota usage aliases", func(t *testing.T) {
		var got management.QuotaUsage
		mustUnmarshalJSON(t, `{
			"teams":{"used":1,"limit":5},
			"members":{"used":2,"limit":10},
			"api_keys":{"used":3,"limit":20}
		}`, &got)
		if got.Teams.Used != 1 || got.Members.Limit != 10 ||
			got.APIKeys.Used != 3 || got.GatewayKeys.Limit != 20 {
			t.Errorf("QuotaUsage = %+v, want typed counters and gateway legacy alias", got)
		}
	})

	t.Run("quota summary", func(t *testing.T) {
		var got management.QuotaSummary
		mustUnmarshalJSON(t, `{
			"workspace_id":7,
			"plan":{"name":"Pro","slug":"pro","quotas":{"members":10}},
			"items":[{"key":"members","label":"Members","used":2,"limit":10,"unit":"members"}]
		}`, &got)
		if got.Plan.Slug != "pro" || got.Plan.Quotas["members"] != 10 ||
			len(got.Items) != 1 || got.Items[0].Unit != "members" {
			t.Errorf("QuotaSummary = %+v, want typed plan and quota items", got)
		}
	})

	t.Run("role", func(t *testing.T) {
		var got management.Role
		mustUnmarshalJSON(t, `{
			"id":121,
			"key":"billing_viewer",
			"name":"Billing Viewer",
			"description":"Read billing",
			"scope_type":"workspace",
			"scope_id":7,
			"built_in":true,
			"permissions":["billing.read"],
			"created_at":"2026-07-25T00:00:00Z",
			"updated_at":"2026-07-25T00:01:00Z"
		}`, &got)
		if got.ID != 121 || got.Key != "billing_viewer" || got.ScopeType != "workspace" ||
			got.ScopeID != 7 || !got.BuiltIn || !got.IsSystem || got.UpdatedAt == "" {
			t.Errorf("Role = %+v, want current and legacy role fields", got)
		}
	})
}

func mustUnmarshalJSON(t *testing.T, data string, out any) {
	t.Helper()
	if err := json.Unmarshal([]byte(data), out); err != nil {
		t.Fatalf("json.Unmarshal(%T) error = %v", out, err)
	}
}

func assertJSONSemanticallyEqual(t *testing.T, got []byte, want string) {
	t.Helper()
	var gotValue any
	if err := json.Unmarshal(got, &gotValue); err != nil {
		t.Fatalf("json.Unmarshal(got %q) error = %v", got, err)
	}
	var wantValue any
	if err := json.Unmarshal([]byte(want), &wantValue); err != nil {
		t.Fatalf("json.Unmarshal(want %q) error = %v", want, err)
	}
	if !reflect.DeepEqual(gotValue, wantValue) {
		t.Errorf("JSON value = %#v, want %#v", gotValue, wantValue)
	}
}
