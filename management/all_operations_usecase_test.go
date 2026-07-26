package management_test

import (
	"context"
	"strings"
	"testing"

	owlvigil "github.com/Syrovex/owlvigil_sdk_go"
	"github.com/Syrovex/owlvigil_sdk_go/management"
)

// executableManagementUseCase binds one published Management API operation to
// an actual SDK call. Keeping the operation string beside the call makes the
// cross-repository alignment gate prove executable coverage instead of merely
// finding route names in comments or skip messages.
type executableManagementUseCase struct {
	name     string
	contract string
	call     func(context.Context, *management.Client) error
}

func TestAllPublishedManagementOperationsHaveExecutableSDKUseCases(t *testing.T) {
	t.Parallel()

	useCases := allExecutableManagementUseCases()
	if got, want := len(useCases), 141; got != want {
		t.Fatalf("executable Management SDK use cases = %d, want %d", got, want)
	}

	seen := make(map[string]struct{}, len(useCases))
	for _, useCase := range useCases {
		if _, duplicate := seen[useCase.contract]; duplicate {
			t.Fatalf("duplicate executable Management SDK use case %q", useCase.contract)
		}
		seen[useCase.contract] = struct{}{}
	}
	for _, useCase := range useCases {
		useCase := useCase
		t.Run(useCase.name, func(t *testing.T) {
			t.Parallel()

			client, requests := newManagementContractClient(t, `null`)
			if err := useCase.call(t.Context(), client); err != nil {
				t.Fatalf("%s SDK call error = %v", useCase.contract, err)
			}

			method, pattern, ok := strings.Cut(useCase.contract, " ")
			if !ok {
				t.Fatalf("invalid Management operation %q", useCase.contract)
			}
			got := <-requests
			if got.method != method {
				t.Errorf("%s SDK method = %q, want %q", useCase.contract, got.method, method)
			}
			if want := executableManagementPath(pattern); got.path != want {
				t.Errorf("%s SDK path = %q, want %q", useCase.contract, got.path, want)
			}
			select {
			case extra := <-requests:
				t.Fatalf("%s emitted an extra request: %s %s", useCase.contract, extra.method, extra.path)
			default:
			}
		})
	}
}

func executableManagementPath(pattern string) string {
	path := strings.TrimPrefix(pattern, "/v1")
	return strings.NewReplacer(
		":workspace_id", "7",
		":audit_log_id", "17",
		":team_id", "27",
		":member_id", "37",
		":invitation_id", "47",
		":role_id", "57",
		":scope_type", "team",
		":scope_id", "27",
		":user_id", "37",
		":provider_id", "67",
		":key_id", "77",
		":keyword_id", "137",
		":model_id", "gpt-4.1",
		":route_id", "route_87",
		":request_id", "req_97",
		":trace_id", "trace_107",
		":payload_id", "117",
		":policy_id", "127",
		":plan_id", "pro",
		":session_id", "cs_137",
		":topup_id", "147",
		":order_id", "157",
		":invoice_id", "167",
		":payment_method_id", "pm_177",
		":endpoint_id", "187",
		":event_id", "197",
	).Replace(path)
}

func allExecutableManagementUseCases() []executableManagementUseCase {
	var useCases []executableManagementUseCase
	useCases = append(useCases, accountAndWorkspaceUseCases()...)
	useCases = append(useCases, workspaceAccessUseCases()...)
	useCases = append(useCases, gatewayUseCases()...)
	useCases = append(useCases, billingUseCases()...)
	useCases = append(useCases, financialUseCases()...)
	useCases = append(useCases, webhookUseCases()...)
	return useCases
}

func accountAndWorkspaceUseCases() []executableManagementUseCase {
	workspaceName := "SDK Contract Workspace"
	preference := true
	retentionDays := 30
	maxBodySize := 65536
	redactHeaders := []string{"authorization"}
	return []executableManagementUseCase{
		{
			name:     "get user profile",
			contract: "GET /v1/user/profile",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.UserProfile, *owlvigil.ResponseMeta, error) {
					return client.GetUserProfile(ctx)
				})
			},
		},
		{
			name:     "update user profile",
			contract: "PUT /v1/user/profile",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.UserProfile, *owlvigil.ResponseMeta, error) {
					return client.UpdateUserProfile(ctx, &management.UpdateUserProfileRequest{Username: &workspaceName})
				})
			},
		},
		{
			name:     "change password",
			contract: "PUT /v1/user/password",
			call: func(ctx context.Context, client *management.Client) error {
				return managementMeta(func() (*owlvigil.ResponseMeta, error) {
					return client.UpdatePassword(ctx, &management.UpdatePasswordRequest{
						OldPassword: "old-password",
						NewPassword: "new-password",
					})
				})
			},
		},
		{
			name:     "create support request",
			contract: "POST /v1/user/support-requests",
			call: func(ctx context.Context, client *management.Client) error {
				return managementMeta(func() (*owlvigil.ResponseMeta, error) {
					return client.CreateSupportRequest(ctx, &management.SupportRequest{
						Subject:     "SDK contract test",
						IssueType:   "technical",
						Description: "Executable SDK contract use case.",
					})
				})
			},
		},
		{
			name:     "get notification preferences",
			contract: "GET /v1/user/notification-preferences",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.NotificationPreferences, *owlvigil.ResponseMeta, error) {
					return client.GetNotificationPreferences(ctx)
				})
			},
		},
		{
			name:     "update notification preferences",
			contract: "PUT /v1/user/notification-preferences",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.NotificationPreferences, *owlvigil.ResponseMeta, error) {
					return client.UpdateNotificationPreferences(ctx, &management.UpdateNotificationPreferencesRequest{
						Budget:    &preference,
						Billing:   &preference,
						Reports:   &preference,
						Marketing: &preference,
					})
				})
			},
		},
		{
			name:     "get invite link",
			contract: "GET /v1/users/me/invite-link",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.InviteLink, *owlvigil.ResponseMeta, error) {
					return client.GetInviteLink(ctx)
				})
			},
		},
		{
			name:     "get invitation stats",
			contract: "GET /v1/users/me/invitation-stats",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.InviteStats, *owlvigil.ResponseMeta, error) {
					return client.GetInvitationStats(ctx)
				})
			},
		},
		{
			name:     "list user invitations",
			contract: "GET /v1/users/me/invitations",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.ListResponse[management.UserInvitation], *owlvigil.ResponseMeta, error) {
					return client.ListUserInvitations(ctx, management.ListOptions{})
				})
			},
		},
		{
			name:     "send invitation",
			contract: "POST /v1/users/me/send-invitation",
			call: func(ctx context.Context, client *management.Client) error {
				return managementMeta(func() (*owlvigil.ResponseMeta, error) {
					return client.SendInvitation(ctx, &management.SendInvitationRequest{
						Emails: []string{"invitee@example.com"},
					})
				})
			},
		},
		{
			name:     "list workspaces",
			contract: "GET /v1/workspaces",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.ListResponse[management.Workspace], *owlvigil.ResponseMeta, error) {
					return client.ListWorkspaces(ctx, management.ListOptions{})
				})
			},
		},
		{
			name:     "create workspace",
			contract: "POST /v1/workspaces",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.Workspace, *owlvigil.ResponseMeta, error) {
					return client.CreateWorkspace(ctx, &management.CreateWorkspaceRequest{Name: workspaceName, Type: "team"})
				})
			},
		},
		{
			name:     "get workspace",
			contract: "GET /v1/workspaces/:workspace_id",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.Workspace, *owlvigil.ResponseMeta, error) {
					return client.GetWorkspace(ctx, 7)
				})
			},
		},
		{
			name:     "get workspace overview",
			contract: "GET /v1/workspaces/:workspace_id/overview",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.WorkspaceOverview, *owlvigil.ResponseMeta, error) {
					return client.GetWorkspaceOverview(ctx, 7, management.WorkspaceOverviewOptions{})
				})
			},
		},
		{
			name:     "update workspace",
			contract: "PATCH /v1/workspaces/:workspace_id",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.Workspace, *owlvigil.ResponseMeta, error) {
					return client.UpdateWorkspace(ctx, 7, &management.UpdateWorkspaceRequest{Name: &workspaceName})
				})
			},
		},
		{
			name:     "delete workspace",
			contract: "DELETE /v1/workspaces/:workspace_id",
			call: func(ctx context.Context, client *management.Client) error {
				return managementMeta(func() (*owlvigil.ResponseMeta, error) {
					return client.DeleteWorkspace(ctx, 7)
				})
			},
		},
		{
			name:     "list workspace activity",
			contract: "GET /v1/workspaces/:workspace_id/activity",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.ListResponse[management.ActivityRecord], *owlvigil.ResponseMeta, error) {
					return client.ListWorkspaceActivity(ctx, 7, management.ListOptions{})
				})
			},
		},
		{
			name:     "list workspace audit logs",
			contract: "GET /v1/workspaces/:workspace_id/audit-logs",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.ListResponse[management.AuditLog], *owlvigil.ResponseMeta, error) {
					return client.ListAuditLogs(ctx, 7, management.AuditLogListOptions{})
				})
			},
		},
		{
			name:     "get workspace audit log",
			contract: "GET /v1/workspaces/:workspace_id/audit-logs/:audit_log_id",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.AuditLog, *owlvigil.ResponseMeta, error) {
					return client.GetAuditLog(ctx, 7, 17)
				})
			},
		},
		{
			name:     "get workspace logging settings",
			contract: "GET /v1/workspaces/:workspace_id/logging-settings",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.LoggingSettings, *owlvigil.ResponseMeta, error) {
					return client.GetLoggingSettings(ctx, 7)
				})
			},
		},
		{
			name:     "update workspace logging settings",
			contract: "PUT /v1/workspaces/:workspace_id/logging-settings",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.LoggingSettings, *owlvigil.ResponseMeta, error) {
					return client.UpdateLoggingSettings(ctx, 7, &management.UpdateLoggingSettingsRequest{
						Enabled:              &preference,
						RetentionDays:        &retentionDays,
						CaptureHeaders:       &preference,
						CaptureRequestBody:   &preference,
						CaptureResponseBody:  &preference,
						CaptureStreamChunks:  &preference,
						MaskSensitiveHeaders: &preference,
						MaxBodySize:          &maxBodySize,
						RedactHeaders:        &redactHeaders,
					})
				})
			},
		},
		{
			name:     "get workspace quota summary",
			contract: "GET /v1/workspaces/:workspace_id/quota-summary",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.QuotaSummary, *owlvigil.ResponseMeta, error) {
					return client.GetQuotaSummary(ctx, 7)
				})
			},
		},
		{
			name:     "get workspace quota usage",
			contract: "GET /v1/workspaces/:workspace_id/quota-usage",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.QuotaUsage, *owlvigil.ResponseMeta, error) {
					return client.GetQuotaUsage(ctx, 7)
				})
			},
		},
		{
			name:     "get workspace billing details",
			contract: "GET /v1/workspaces/:workspace_id/billing-details",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.BillingDetails, *owlvigil.ResponseMeta, error) {
					return client.GetBillingDetails(ctx, 7)
				})
			},
		},
		{
			name:     "update workspace billing details",
			contract: "PUT /v1/workspaces/:workspace_id/billing-details",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.BillingDetails, *owlvigil.ResponseMeta, error) {
					return client.UpdateBillingDetails(ctx, 7, &management.UpdateBillingDetailsRequest{
						BillingDetails: &management.BillingContact{
							Name:         "SDK Contract Customer",
							Email:        "billing@example.com",
							TaxID:        "tax_contract",
							Phone:        "+1-555-0100",
							Address:      "1 Contract Street",
							CCRecipients: []string{"finance@example.com"},
						},
					})
				})
			},
		},
	}
}

func workspaceAccessUseCases() []executableManagementUseCase {
	name := "SDK Contract Resource"
	status := "active"
	return []executableManagementUseCase{
		{
			name: "list teams", contract: "GET /v1/workspaces/:workspace_id/teams",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.ListResponse[management.Team], *owlvigil.ResponseMeta, error) {
					return client.ListTeams(ctx, 7, management.ListOptions{})
				})
			},
		},
		{
			name: "create team", contract: "POST /v1/workspaces/:workspace_id/teams",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.Team, *owlvigil.ResponseMeta, error) {
					return client.CreateTeam(ctx, 7, &management.CreateTeamRequest{Name: name})
				})
			},
		},
		{
			name: "get team", contract: "GET /v1/workspaces/:workspace_id/teams/:team_id",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.Team, *owlvigil.ResponseMeta, error) {
					return client.GetTeam(ctx, 7, 27)
				})
			},
		},
		{
			name: "update team", contract: "PATCH /v1/workspaces/:workspace_id/teams/:team_id",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.Team, *owlvigil.ResponseMeta, error) {
					return client.UpdateTeam(ctx, 7, 27, &management.UpdateTeamRequest{Name: &name})
				})
			},
		},
		{
			name: "delete team", contract: "DELETE /v1/workspaces/:workspace_id/teams/:team_id",
			call: func(ctx context.Context, client *management.Client) error {
				return managementMeta(func() (*owlvigil.ResponseMeta, error) { return client.DeleteTeam(ctx, 7, 27) })
			},
		},
		{
			name: "list members", contract: "GET /v1/workspaces/:workspace_id/members",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.ListResponse[management.Member], *owlvigil.ResponseMeta, error) {
					return client.ListMembersWithFilters(ctx, 7, management.MemberListOptions{})
				})
			},
		},
		{
			name: "get member role options", contract: "GET /v1/workspaces/:workspace_id/members/role-options",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.ListResponse[management.RoleOption], *owlvigil.ResponseMeta, error) {
					return client.ListRoleOptions(ctx, 7)
				})
			},
		},
		{
			name: "create member", contract: "POST /v1/workspaces/:workspace_id/members",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.Member, *owlvigil.ResponseMeta, error) {
					return client.CreateMember(ctx, 7, &management.CreateMemberRequest{Email: "member@example.com", Role: "member"})
				})
			},
		},
		{
			name: "get member", contract: "GET /v1/workspaces/:workspace_id/members/:member_id",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.Member, *owlvigil.ResponseMeta, error) {
					return client.GetMember(ctx, 7, 37)
				})
			},
		},
		{
			name: "update member", contract: "PATCH /v1/workspaces/:workspace_id/members/:member_id",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.Member, *owlvigil.ResponseMeta, error) {
					return client.UpdateMember(ctx, 7, 37, &management.UpdateMemberRequest{Role: "member", Status: &status})
				})
			},
		},
		{
			name: "delete member", contract: "DELETE /v1/workspaces/:workspace_id/members/:member_id",
			call: func(ctx context.Context, client *management.Client) error {
				return managementMeta(func() (*owlvigil.ResponseMeta, error) { return client.DeleteMember(ctx, 7, 37) })
			},
		},
		{
			name: "list workspace invitations", contract: "GET /v1/workspaces/:workspace_id/invitations",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.ListResponse[management.Invitation], *owlvigil.ResponseMeta, error) {
					return client.ListInvitations(ctx, 7, management.ListOptions{})
				})
			},
		},
		{
			name: "create workspace invitation", contract: "POST /v1/workspaces/:workspace_id/invitations",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.Invitation, *owlvigil.ResponseMeta, error) {
					return client.CreateInvitation(ctx, 7, &management.CreateInvitationRequest{Email: "invitee@example.com", Role: "member"})
				})
			},
		},
		{
			name: "resend workspace invitation", contract: "POST /v1/workspaces/:workspace_id/invitations/:invitation_id/resend",
			call: func(ctx context.Context, client *management.Client) error {
				return managementMeta(func() (*owlvigil.ResponseMeta, error) { return client.ResendInvitation(ctx, 7, 47) })
			},
		},
		{
			name: "revoke workspace invitation", contract: "POST /v1/workspaces/:workspace_id/invitations/:invitation_id/revoke",
			call: func(ctx context.Context, client *management.Client) error {
				return managementMeta(func() (*owlvigil.ResponseMeta, error) { return client.RevokeInvitation(ctx, 7, 47) })
			},
		},
		{
			name: "list roles", contract: "GET /v1/workspaces/:workspace_id/roles",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.ListResponse[management.Role], *owlvigil.ResponseMeta, error) {
					return client.ListRoles(ctx, 7, management.ListOptions{})
				})
			},
		},
		{
			name: "create role", contract: "POST /v1/workspaces/:workspace_id/roles",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.Role, *owlvigil.ResponseMeta, error) {
					return client.CreateRole(ctx, 7, &management.CreateRoleRequest{Name: name})
				})
			},
		},
		{
			name: "get role", contract: "GET /v1/workspaces/:workspace_id/roles/:role_id",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.Role, *owlvigil.ResponseMeta, error) {
					return client.GetRole(ctx, 7, 57)
				})
			},
		},
		{
			name: "update role", contract: "PATCH /v1/workspaces/:workspace_id/roles/:role_id",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.Role, *owlvigil.ResponseMeta, error) {
					return client.UpdateRole(ctx, 7, 57, &management.UpdateRoleRequest{Name: &name})
				})
			},
		},
		{
			name: "delete role", contract: "DELETE /v1/workspaces/:workspace_id/roles/:role_id",
			call: func(ctx context.Context, client *management.Client) error {
				return managementMeta(func() (*owlvigil.ResponseMeta, error) { return client.DeleteRole(ctx, 7, 57) })
			},
		},
		{
			name: "list permissions", contract: "GET /v1/workspaces/:workspace_id/permissions",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.ListResponse[management.Permission], *owlvigil.ResponseMeta, error) {
					return client.ListPermissions(ctx, 7)
				})
			},
		},
		{
			name: "get member permissions", contract: "GET /v1/workspaces/:workspace_id/members/:member_id/permissions",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.MemberPermissions, *owlvigil.ResponseMeta, error) {
					return client.GetMemberPermissions(ctx, 7, 37)
				})
			},
		},
		{
			name: "update member permissions", contract: "PUT /v1/workspaces/:workspace_id/members/:member_id/permissions",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.MemberPermissions, *owlvigil.ResponseMeta, error) {
					return client.UpdateMemberPermissions(ctx, 7, 37, &management.UpdateMemberPermissionsRequest{
						PermissionMap: map[string]bool{"workspace.read": true},
					})
				})
			},
		},
		{
			name: "reset member permissions", contract: "POST /v1/workspaces/:workspace_id/members/:member_id/permissions/reset",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.MemberPermissions, *owlvigil.ResponseMeta, error) {
					return client.ResetMemberPermissions(ctx, 7, 37)
				})
			},
		},
	}
}

func gatewayUseCases() []executableManagementUseCase {
	workspace := owlvigil.WithWorkspaceID(7)
	name := "SDK Contract Gateway Resource"
	action := "allow"
	priority := 10
	enabled := true
	return []executableManagementUseCase{
		{
			name: "list providers", contract: "GET /v1/gateway/providers",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.ListResponse[management.Provider], *owlvigil.ResponseMeta, error) {
					return client.ListProviders(ctx, 7)
				})
			},
		},
		{
			name: "create provider", contract: "POST /v1/gateway/providers",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.Provider, *owlvigil.ResponseMeta, error) {
					return client.CreateProvider(ctx, &management.CreateProviderRequest{
						WorkspaceID: 7,
						Name:        name,
						Type:        "openai",
						APIKey:      "provider-test-key",
					})
				})
			},
		},
		{
			name: "verify provider connection", contract: "POST /v1/gateway/providers/verify-connection",
			call: func(ctx context.Context, client *management.Client) error {
				providerID := 67
				return managementResult(func() (*management.ProviderConnectionVerification, *owlvigil.ResponseMeta, error) {
					return client.VerifyProviderConnection(ctx, &management.VerifyProviderConnectionRequest{
						WorkspaceID: 7,
						ProviderID:  &providerID,
					})
				})
			},
		},
		{
			name: "get provider", contract: "GET /v1/gateway/providers/:provider_id",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.Provider, *owlvigil.ResponseMeta, error) {
					return client.GetProvider(ctx, 7, 67)
				})
			},
		},
		{
			name: "update provider", contract: "PATCH /v1/gateway/providers/:provider_id",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.Provider, *owlvigil.ResponseMeta, error) {
					return client.UpdateProvider(ctx, 7, 67, &management.UpdateProviderRequest{Name: &name})
				})
			},
		},
		{
			name: "delete provider", contract: "DELETE /v1/gateway/providers/:provider_id",
			call: func(ctx context.Context, client *management.Client) error {
				return managementMeta(func() (*owlvigil.ResponseMeta, error) {
					return client.DeleteProvider(ctx, 7, 67)
				})
			},
		},
		{
			name: "list gateway keys", contract: "GET /v1/gateway/keys",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.ListResponse[management.GatewayKey], *owlvigil.ResponseMeta, error) {
					return client.ListGatewayKeys(ctx, management.ListOptions{}, "", workspace)
				})
			},
		},
		{
			name: "create gateway key", contract: "POST /v1/gateway/keys",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.GatewayKey, *owlvigil.ResponseMeta, error) {
					return client.CreateGatewayKey(ctx, &management.CreateGatewayKeyRequest{
						WorkspaceID:    7,
						Name:           name,
						ProviderSource: "platform",
					})
				})
			},
		},
		{
			name: "get gateway key", contract: "GET /v1/gateway/keys/:key_id",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.GatewayKey, *owlvigil.ResponseMeta, error) {
					return client.GetGatewayKey(ctx, 77, workspace)
				})
			},
		},
		{
			name: "update gateway key", contract: "PATCH /v1/gateway/keys/:key_id",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.GatewayKey, *owlvigil.ResponseMeta, error) {
					return client.UpdateGatewayKey(ctx, 77, &management.UpdateGatewayKeyRequest{Name: &name}, workspace)
				})
			},
		},
		{
			name: "rotate gateway key", contract: "POST /v1/gateway/keys/:key_id/rotate",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.GatewayKey, *owlvigil.ResponseMeta, error) {
					return client.RotateGatewayKey(ctx, 77, workspace)
				})
			},
		},
		{
			name: "enable gateway key", contract: "POST /v1/gateway/keys/:key_id/enable",
			call: func(ctx context.Context, client *management.Client) error {
				return managementMeta(func() (*owlvigil.ResponseMeta, error) {
					return client.EnableGatewayKey(ctx, 77, workspace)
				})
			},
		},
		{
			name: "disable gateway key", contract: "POST /v1/gateway/keys/:key_id/disable",
			call: func(ctx context.Context, client *management.Client) error {
				return managementMeta(func() (*owlvigil.ResponseMeta, error) {
					return client.DisableGatewayKey(ctx, 77, workspace)
				})
			},
		},
		{
			name: "delete gateway key", contract: "DELETE /v1/gateway/keys/:key_id",
			call: func(ctx context.Context, client *management.Client) error {
				return managementMeta(func() (*owlvigil.ResponseMeta, error) {
					return client.DeleteGatewayKey(ctx, 77, workspace)
				})
			},
		},
		{
			name: "list models", contract: "GET /v1/gateway/models",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.ListResponse[management.Model], *owlvigil.ResponseMeta, error) {
					return client.ListModels(ctx, management.ListOptions{}, workspace)
				})
			},
		},
		{
			name: "get model", contract: "GET /v1/gateway/models/:model_id",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.Model, *owlvigil.ResponseMeta, error) {
					return client.GetModel(ctx, "gpt-4.1", workspace)
				})
			},
		},
		{
			name: "list routes", contract: "GET /v1/gateway/routes",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.ListResponse[management.Route], *owlvigil.ResponseMeta, error) {
					return client.ListRoutesWithFilters(ctx, management.RouteListOptions{}, workspace)
				})
			},
		},
		{
			name: "get route", contract: "GET /v1/gateway/routes/:route_id",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.Route, *owlvigil.ResponseMeta, error) {
					return client.GetRouteWithFilters(ctx, "route_87", management.RouteDetailOptions{}, workspace)
				})
			},
		},
		{
			name: "preview route", contract: "POST /v1/gateway/routes/preview",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.PreviewRouteResponse, *owlvigil.ResponseMeta, error) {
					return client.PreviewRoute(ctx, &management.PreviewRouteRequest{
						WorkspaceID: 7,
						Model:       "gpt-4.1",
					})
				})
			},
		},
		{
			name: "list usage", contract: "GET /v1/gateway/usage",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.ListResponse[management.UsageRecord], *owlvigil.ResponseMeta, error) {
					return client.ListUsage(ctx, management.ListOptions{}, workspace)
				})
			},
		},
		{
			name: "get usage summary", contract: "GET /v1/gateway/usage/summary",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.UsageSummary, *owlvigil.ResponseMeta, error) {
					return client.GetUsageSummary(ctx, workspace)
				})
			},
		},
		{
			name: "get gateway quota", contract: "GET /v1/gateway/quota",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.Quota, *owlvigil.ResponseMeta, error) {
					return client.GetQuota(ctx, workspace)
				})
			},
		},
		{
			name: "list request logs", contract: "GET /v1/gateway/request-logs",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.ListResponse[management.RequestLog], *owlvigil.ResponseMeta, error) {
					return client.ListRequestLogs(ctx, management.ListOptions{}, "", workspace)
				})
			},
		},
		{
			name: "get request log", contract: "GET /v1/gateway/request-logs/:request_id",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.RequestLog, *owlvigil.ResponseMeta, error) {
					return client.GetRequestLog(ctx, "req_97", workspace)
				})
			},
		},
		{
			name: "list traces", contract: "GET /v1/gateway/traces",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.ListResponse[management.Trace], *owlvigil.ResponseMeta, error) {
					return client.ListTraces(ctx, management.ListOptions{}, workspace)
				})
			},
		},
		{
			name: "get trace", contract: "GET /v1/gateway/traces/:trace_id",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.Trace, *owlvigil.ResponseMeta, error) {
					return client.GetTrace(ctx, "trace_107", workspace)
				})
			},
		},
		{
			name: "list payload logs", contract: "GET /v1/gateway/payload-logs",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.ListResponse[management.PayloadLogSummary], *owlvigil.ResponseMeta, error) {
					return client.ListPayloadLogs(ctx, 7, management.ListOptions{})
				})
			},
		},
		{
			name: "get payload log access", contract: "GET /v1/gateway/payload-logs/access",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.PayloadAccess, *owlvigil.ResponseMeta, error) {
					return client.GetPayloadAccess(ctx, workspace)
				})
			},
		},
		{
			name: "get payload log", contract: "GET /v1/gateway/payload-logs/:payload_id",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.PayloadLog, *owlvigil.ResponseMeta, error) {
					return client.GetPayloadLog(ctx, "117", workspace)
				})
			},
		},
		{
			name: "get gateway policies", contract: "GET /v1/gateway/policies",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.GatewayPolicy, *owlvigil.ResponseMeta, error) {
					return client.GetGatewayPolicies(ctx, 77, workspace)
				})
			},
		},
		{
			name: "preview gateway policies", contract: "POST /v1/gateway/policies/preview",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.PreviewPolicyResponse, *owlvigil.ResponseMeta, error) {
					return client.PreviewPolicyEffect(ctx, &management.PreviewPolicyRequest{
						WorkspaceID: 7,
						KeyID:       77,
						Model:       "gpt-4.1",
					})
				})
			},
		},
		{
			name: "add prompt keyword", contract: "POST /v1/gateway/policies/keywords",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.GatewayPolicy, *owlvigil.ResponseMeta, error) {
					return client.AddPromptKeyword(ctx, &management.AddPromptKeywordRequest{
						WorkspaceID: 7,
						Keyword:     "SensitiveToken",
					})
				})
			},
		},
		{
			name: "delete prompt keyword", contract: "DELETE /v1/gateway/policies/keywords/:keyword_id",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.GatewayPolicy, *owlvigil.ResponseMeta, error) {
					return client.DeletePromptKeyword(ctx, 7, 137)
				})
			},
		},
		{
			name: "update gateway policy", contract: "PATCH /v1/gateway/policies/:policy_id",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.GatewayPolicy, *owlvigil.ResponseMeta, error) {
					return client.UpdateGatewayPolicy(ctx, 127, &management.UpdateGatewayPolicyRequest{
						Action:   &action,
						Priority: &priority,
						Enabled:  &enabled,
					}, workspace)
				})
			},
		},
	}
}

func billingUseCases() []executableManagementUseCase {
	return []executableManagementUseCase{
		{
			name: "get billing overview", contract: "GET /v1/billing/overview",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.BillingOverview, *owlvigil.ResponseMeta, error) {
					return client.GetBillingOverviewForWorkspace(ctx, 7)
				})
			},
		},
		{
			name: "get billing balance", contract: "GET /v1/billing/balance",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.Balance, *owlvigil.ResponseMeta, error) {
					return client.GetBalanceForWorkspace(ctx, 7)
				})
			},
		},
		{
			name: "list billing plans", contract: "GET /v1/billing/plans",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.ListResponse[management.Plan], *owlvigil.ResponseMeta, error) {
					return client.ListPlans(ctx, management.ListOptions{})
				})
			},
		},
		{
			name: "get billing plan", contract: "GET /v1/billing/plans/:plan_id",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.Plan, *owlvigil.ResponseMeta, error) {
					return client.GetPlan(ctx, "pro")
				})
			},
		},
		{
			name: "get subscription", contract: "GET /v1/billing/subscription",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.Subscription, *owlvigil.ResponseMeta, error) {
					return client.GetSubscription(ctx)
				})
			},
		},
		{
			name: "create subscription checkout", contract: "POST /v1/billing/subscription/checkout",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.CreateSubscriptionCheckoutResponse, *owlvigil.ResponseMeta, error) {
					return client.CreateSubscriptionCheckout(ctx, &management.CreateSubscriptionCheckoutRequest{
						PlanID:     "pro",
						Interval:   "monthly",
						SuccessURL: "https://example.com/success",
						CancelURL:  "https://example.com/cancel",
					})
				})
			},
		},
		{
			name: "create subscription in app", contract: "POST /v1/billing/subscription/in-app",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.CreateSubscriptionInAppResponse, *owlvigil.ResponseMeta, error) {
					return client.CreateSubscriptionInApp(ctx, &management.CreateSubscriptionInAppRequest{
						PlanID:    "pro",
						Interval:  "monthly",
						ReturnURL: "https://example.com/return",
					})
				})
			},
		},
		{
			name: "confirm subscription in app", contract: "POST /v1/billing/subscription/in-app/confirm",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.Subscription, *owlvigil.ResponseMeta, error) {
					return client.ConfirmSubscriptionInApp(ctx, &management.ConfirmSubscriptionInAppRequest{
						PlanID:               "pro",
						StripeSubscriptionID: "sub_contract",
					})
				})
			},
		},
		{
			name: "upgrade subscription", contract: "POST /v1/billing/subscription/upgrade",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.Subscription, *owlvigil.ResponseMeta, error) {
					return client.UpgradeSubscription(ctx, &management.UpgradeSubscriptionRequest{
						PlanID: "pro", Interval: "monthly",
					})
				})
			},
		},
		{
			name: "downgrade subscription", contract: "POST /v1/billing/subscription/downgrade",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.Subscription, *owlvigil.ResponseMeta, error) {
					return client.DowngradeSubscription(ctx, &management.DowngradeSubscriptionRequest{
						PlanID: "starter", Interval: "monthly",
					})
				})
			},
		},
		{
			name: "cancel subscription", contract: "POST /v1/billing/subscription/cancel",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.Subscription, *owlvigil.ResponseMeta, error) {
					return client.CancelSubscriptionWithRequest(ctx, &management.CancelSubscriptionRequest{
						Reason: "SDK contract test",
					})
				})
			},
		},
		{
			name: "reactivate subscription", contract: "POST /v1/billing/subscription/reactivate",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.Subscription, *owlvigil.ResponseMeta, error) {
					return client.ReactivateSubscription(ctx)
				})
			},
		},
		{
			name: "get subscription checkout session", contract: "GET /v1/billing/subscription/checkout-sessions/:session_id",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.CheckoutSession, *owlvigil.ResponseMeta, error) {
					return client.GetSubscriptionCheckoutSession(ctx, "cs_137")
				})
			},
		},
		{
			name: "sync latest subscription checkout", contract: "POST /v1/billing/subscription/checkout-sessions/sync-latest",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.CheckoutSession, *owlvigil.ResponseMeta, error) {
					return client.SyncLatestSubscriptionCheckout(ctx)
				})
			},
		},
		{
			name: "list topup plans", contract: "GET /v1/billing/topup-plans",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.ListResponse[management.TopupPlan], *owlvigil.ResponseMeta, error) {
					return client.ListTopupPlans(ctx, management.ListOptions{})
				})
			},
		},
		{
			name: "list topups", contract: "GET /v1/billing/topups",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.ListResponse[management.Topup], *owlvigil.ResponseMeta, error) {
					return client.ListTopupsWithFilters(ctx, management.OrderListOptions{})
				})
			},
		},
		{
			name: "create topup checkout", contract: "POST /v1/billing/topups/checkout",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.CreateTopupCheckoutResponse, *owlvigil.ResponseMeta, error) {
					return client.CreateTopupCheckout(ctx, &management.CreateTopupCheckoutRequest{
						WorkspaceID: 7,
						Amount:      100,
						SuccessURL:  "https://example.com/success",
						CancelURL:   "https://example.com/cancel",
					})
				})
			},
		},
		{
			name: "create topup in app", contract: "POST /v1/billing/topups/in-app",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.CreateTopupInAppResponse, *owlvigil.ResponseMeta, error) {
					return client.CreateTopupInApp(ctx, &management.CreateTopupInAppRequest{
						WorkspaceID: 7,
						Amount:      100,
						ReturnURL:   "https://example.com/return",
					})
				})
			},
		},
		{
			name: "confirm topup in app", contract: "POST /v1/billing/topups/in-app/confirm",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.Topup, *owlvigil.ResponseMeta, error) {
					return client.ConfirmTopupInApp(ctx, &management.ConfirmTopupInAppRequest{
						PaymentIntentID: "pi_contract",
						ClientSecret:    "secret_contract",
					})
				})
			},
		},
		{
			name: "get topup", contract: "GET /v1/billing/topups/:topup_id",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.Topup, *owlvigil.ResponseMeta, error) {
					return client.GetTopup(ctx, "147")
				})
			},
		},
		{
			name: "list billing orders", contract: "GET /v1/billing/orders",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.ListResponse[management.Order], *owlvigil.ResponseMeta, error) {
					return client.ListOrdersWithFilters(ctx, management.OrderListOptions{})
				})
			},
		},
		{
			name: "get billing order", contract: "GET /v1/billing/orders/:order_id",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.Order, *owlvigil.ResponseMeta, error) {
					return client.GetOrder(ctx, "157")
				})
			},
		},
		{
			name: "confirm billing order stripe session", contract: "POST /v1/billing/orders/:order_id/confirm-stripe-session",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.Order, *owlvigil.ResponseMeta, error) {
					return client.ConfirmStripeSession(ctx, "157", &management.ConfirmStripeSessionRequest{
						SessionID: "cs_contract",
					})
				})
			},
		},
		{
			name: "list invoices", contract: "GET /v1/billing/invoices",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.ListResponse[management.Invoice], *owlvigil.ResponseMeta, error) {
					return client.ListInvoicesForWorkspace(ctx, 7, management.ListOptions{})
				})
			},
		},
		{
			name: "get invoice", contract: "GET /v1/billing/invoices/:invoice_id",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.Invoice, *owlvigil.ResponseMeta, error) {
					return client.GetInvoiceForWorkspace(ctx, 7, "167")
				})
			},
		},
		{
			name: "list payment methods", contract: "GET /v1/billing/payment-methods",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.ListResponse[management.PaymentMethod], *owlvigil.ResponseMeta, error) {
					return client.ListPaymentMethodsForWorkspace(ctx, 7)
				})
			},
		},
		{
			name: "create payment method setup intent", contract: "POST /v1/billing/payment-methods/setup-intent",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.SetupIntent, *owlvigil.ResponseMeta, error) {
					return client.CreatePaymentMethodSetupIntentForWorkspace(ctx, 7)
				})
			},
		},
		{
			name: "save payment method", contract: "POST /v1/billing/payment-methods",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.PaymentMethod, *owlvigil.ResponseMeta, error) {
					return client.SavePaymentMethod(ctx, &management.SavePaymentMethodRequest{
						PaymentMethodID: "pm_177",
						SetAsDefault:    true,
					})
				})
			},
		},
		{
			name: "set default payment method", contract: "PUT /v1/billing/payment-methods/:payment_method_id/default",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.PaymentMethod, *owlvigil.ResponseMeta, error) {
					return client.SetDefaultPaymentMethod(ctx, "pm_177")
				})
			},
		},
		{
			name: "delete payment method", contract: "DELETE /v1/billing/payment-methods/:payment_method_id",
			call: func(ctx context.Context, client *management.Client) error {
				return managementMeta(func() (*owlvigil.ResponseMeta, error) {
					return client.DeletePaymentMethod(ctx, "pm_177")
				})
			},
		},
	}
}

func financialUseCases() []executableManagementUseCase {
	enabled := true
	monthlyAmount := 1000.0
	warningPercent := 75.5
	criticalPercent := 90.5
	exceededAction := "block"
	workspaceCap := management.BudgetCap{
		ScopeType:     "workspace",
		Enabled:       true,
		MonthlyAmount: monthlyAmount,
	}
	memberLimit := management.SpendingLimit{
		UserID:       37,
		DailyLimit:   50,
		WeeklyLimit:  250,
		MonthlyLimit: 1000,
	}
	return []executableManagementUseCase{
		{
			name: "get financial governance", contract: "GET /v1/workspaces/:workspace_id/governance/financial",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.FinancialGovernance, *owlvigil.ResponseMeta, error) {
					return client.GetFinancialGovernance(ctx, 7)
				})
			},
		},
		{
			name: "update financial governance", contract: "PUT /v1/workspaces/:workspace_id/governance/financial",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.FinancialGovernance, *owlvigil.ResponseMeta, error) {
					return client.UpdateFinancialGovernance(ctx, 7, &management.UpdateFinancialGovernanceRequest{
						WorkspaceCap:   &workspaceCap,
						MemberLimits:   []management.SpendingLimit{memberLimit},
						ExceededAction: &exceededAction,
					})
				})
			},
		},
		{
			name: "get budget caps", contract: "GET /v1/workspaces/:workspace_id/governance/financial/budget-caps",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.BudgetCaps, *owlvigil.ResponseMeta, error) {
					return client.GetBudgetCaps(ctx, 7)
				})
			},
		},
		{
			name: "update budget caps", contract: "PUT /v1/workspaces/:workspace_id/governance/financial/budget-caps",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.BudgetCaps, *owlvigil.ResponseMeta, error) {
					return client.UpdateBudgetCaps(ctx, 7, &management.UpdateBudgetCapsRequest{
						WorkspaceCap: &workspaceCap,
					})
				})
			},
		},
		{
			name: "update scope budget cap", contract: "PATCH /v1/workspaces/:workspace_id/governance/financial/budget-caps/:scope_type/:scope_id",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.BudgetCap, *owlvigil.ResponseMeta, error) {
					return client.UpdateScopeBudgetCap(ctx, 7, "team", "27", &management.UpdateScopeBudgetCapRequest{
						Enabled:       &enabled,
						MonthlyAmount: &monthlyAmount,
					})
				})
			},
		},
		{
			name: "get spending limits", contract: "GET /v1/workspaces/:workspace_id/governance/financial/spending-limits",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.ListResponse[management.SpendingLimit], *owlvigil.ResponseMeta, error) {
					return client.GetSpendingLimitsWithFilters(ctx, 7, management.SpendingLimitOptions{})
				})
			},
		},
		{
			name: "update spending limits", contract: "PUT /v1/workspaces/:workspace_id/governance/financial/spending-limits",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.ListResponse[management.SpendingLimit], *owlvigil.ResponseMeta, error) {
					return client.UpdateSpendingLimits(ctx, 7, &management.UpdateSpendingLimitsRequest{
						Limits: []management.SpendingLimit{memberLimit},
					})
				})
			},
		},
		{
			name: "update user spending limit", contract: "PATCH /v1/workspaces/:workspace_id/governance/financial/spending-limits/users/:user_id",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.SpendingLimit, *owlvigil.ResponseMeta, error) {
					return client.UpdateUserSpendingLimit(ctx, 7, 37, &management.UpdateUserSpendingLimitRequest{
						DailyLimit:   &memberLimit.DailyLimit,
						WeeklyLimit:  &memberLimit.WeeklyLimit,
						MonthlyLimit: &memberLimit.MonthlyLimit,
					})
				})
			},
		},
		{
			name: "get financial thresholds", contract: "GET /v1/workspaces/:workspace_id/governance/financial/thresholds",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.Thresholds, *owlvigil.ResponseMeta, error) {
					return client.GetFinancialThresholds(ctx, 7)
				})
			},
		},
		{
			name: "update financial thresholds", contract: "PUT /v1/workspaces/:workspace_id/governance/financial/thresholds",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.Thresholds, *owlvigil.ResponseMeta, error) {
					return client.UpdateFinancialThresholds(ctx, 7, &management.UpdateThresholdsRequest{
						WarningPercent:  &warningPercent,
						CriticalPercent: &criticalPercent,
						ExceededAction:  &exceededAction,
					})
				})
			},
		},
		{
			name: "preview financial changes", contract: "POST /v1/workspaces/:workspace_id/governance/financial/preview",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.PreviewFinancialChangesResponse, *owlvigil.ResponseMeta, error) {
					return client.PreviewFinancialChanges(ctx, 7, &management.PreviewFinancialChangesRequest{
						WorkspaceCap: &workspaceCap,
					})
				})
			},
		},
		{
			name: "get spend summary", contract: "GET /v1/workspaces/:workspace_id/governance/financial/spend-summary",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.SpendSummary, *owlvigil.ResponseMeta, error) {
					return client.GetSpendSummary(ctx, 7)
				})
			},
		},
	}
}

func webhookUseCases() []executableManagementUseCase {
	workspace := owlvigil.WithWorkspaceID(7)
	webhookURL := "https://example.com/owlvigil/webhook"
	return []executableManagementUseCase{
		{
			name: "list webhook endpoints", contract: "GET /v1/webhook-endpoints",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.ListResponse[management.WebhookEndpoint], *owlvigil.ResponseMeta, error) {
					return client.ListWebhookEndpoints(ctx, management.ListOptions{}, workspace)
				})
			},
		},
		{
			name: "create webhook endpoint", contract: "POST /v1/webhook-endpoints",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.WebhookEndpoint, *owlvigil.ResponseMeta, error) {
					return client.CreateWebhookEndpoint(ctx, &management.CreateWebhookEndpointRequest{
						WorkspaceID: 7,
						URL:         webhookURL,
						EventTypes:  []string{"gateway.key.updated"},
					})
				})
			},
		},
		{
			name: "get webhook endpoint", contract: "GET /v1/webhook-endpoints/:endpoint_id",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.WebhookEndpoint, *owlvigil.ResponseMeta, error) {
					return client.GetWebhookEndpoint(ctx, 187, workspace)
				})
			},
		},
		{
			name: "update webhook endpoint", contract: "PATCH /v1/webhook-endpoints/:endpoint_id",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.WebhookEndpoint, *owlvigil.ResponseMeta, error) {
					return client.UpdateWebhookEndpoint(ctx, 187, &management.UpdateWebhookEndpointRequest{
						URL: &webhookURL,
					}, workspace)
				})
			},
		},
		{
			name: "delete webhook endpoint", contract: "DELETE /v1/webhook-endpoints/:endpoint_id",
			call: func(ctx context.Context, client *management.Client) error {
				return managementMeta(func() (*owlvigil.ResponseMeta, error) {
					return client.DeleteWebhookEndpoint(ctx, 187, workspace)
				})
			},
		},
		{
			name: "enable webhook endpoint", contract: "POST /v1/webhook-endpoints/:endpoint_id/enable",
			call: func(ctx context.Context, client *management.Client) error {
				return managementMeta(func() (*owlvigil.ResponseMeta, error) {
					return client.EnableWebhookEndpoint(ctx, 187, workspace)
				})
			},
		},
		{
			name: "disable webhook endpoint", contract: "POST /v1/webhook-endpoints/:endpoint_id/disable",
			call: func(ctx context.Context, client *management.Client) error {
				return managementMeta(func() (*owlvigil.ResponseMeta, error) {
					return client.DisableWebhookEndpoint(ctx, 187, workspace)
				})
			},
		},
		{
			name: "rotate webhook secret", contract: "POST /v1/webhook-endpoints/:endpoint_id/rotate-secret",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.WebhookEndpoint, *owlvigil.ResponseMeta, error) {
					return client.RotateWebhookSecret(ctx, 187, workspace)
				})
			},
		},
		{
			name: "test webhook endpoint", contract: "POST /v1/webhook-endpoints/:endpoint_id/test",
			call: func(ctx context.Context, client *management.Client) error {
				return managementMeta(func() (*owlvigil.ResponseMeta, error) {
					return client.TestWebhookEndpoint(ctx, 187, workspace)
				})
			},
		},
		{
			name: "list endpoint webhook events", contract: "GET /v1/webhook-endpoints/:endpoint_id/events",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.ListResponse[management.WebhookEvent], *owlvigil.ResponseMeta, error) {
					return client.ListEndpointEvents(ctx, 187, management.ListOptions{}, workspace)
				})
			},
		},
		{
			name: "list webhook event types", contract: "GET /v1/webhook-event-types",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.ListResponse[management.WebhookEventType], *owlvigil.ResponseMeta, error) {
					return client.ListWebhookEventTypes(ctx)
				})
			},
		},
		{
			name: "list webhook events", contract: "GET /v1/webhook-events",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.ListResponse[management.WebhookEvent], *owlvigil.ResponseMeta, error) {
					return client.ListWebhookEvents(ctx, management.ListOptions{}, workspace)
				})
			},
		},
		{
			name: "get webhook event", contract: "GET /v1/webhook-events/:event_id",
			call: func(ctx context.Context, client *management.Client) error {
				return managementResult(func() (*management.WebhookEvent, *owlvigil.ResponseMeta, error) {
					return client.GetWebhookEvent(ctx, "197", workspace)
				})
			},
		},
		{
			name: "retry webhook event", contract: "POST /v1/webhook-events/:event_id/retry",
			call: func(ctx context.Context, client *management.Client) error {
				return managementMeta(func() (*owlvigil.ResponseMeta, error) {
					return client.RetryWebhookEvent(ctx, "197", workspace)
				})
			},
		},
		{
			name: "redeliver webhook event", contract: "POST /v1/webhook-events/:event_id/redeliver",
			call: func(ctx context.Context, client *management.Client) error {
				return managementMeta(func() (*owlvigil.ResponseMeta, error) {
					return client.RedeliverWebhookEvent(ctx, "197", workspace)
				})
			},
		},
		{
			name: "bulk redeliver webhook events", contract: "POST /v1/webhook-events/bulk-redeliver",
			call: func(ctx context.Context, client *management.Client) error {
				return managementMeta(func() (*owlvigil.ResponseMeta, error) {
					return client.BulkRedeliverWebhookEvents(ctx, &management.BulkRedeliverRequest{
						WorkspaceID: 7,
						EventIDs:    []int{197},
					})
				})
			},
		},
	}
}

func managementResult[T any](call func() (*T, *owlvigil.ResponseMeta, error)) error {
	_, _, err := call()
	return err
}

func managementMeta(call func() (*owlvigil.ResponseMeta, error)) error {
	_, err := call()
	return err
}
