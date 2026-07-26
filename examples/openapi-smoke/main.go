package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	owlvigil "github.com/Syrovex/owlvigil_sdk_go"
	"github.com/Syrovex/owlvigil_sdk_go/examples/internal/envfile"
	"github.com/Syrovex/owlvigil_sdk_go/management"
	oauth2 "github.com/Syrovex/owlvigil_sdk_go/oauth2"
)

const (
	defaultScope             = "workspace:read workspace:write gateway:read gateway:write usage:read billing:read billing:write webhook:read webhook:write profile:read profile:write support:write notifications:read notifications:write invites:read invites:write audit_logs:read teams:read teams:write members:read members:write rbac:read rbac:write financial:read financial:write policies:read policies:write"
	cleanupTimeout           = 15 * time.Second
	managementOperationCount = 141
)

type step struct {
	Name     string
	Contract string
	Status   string
	Error    string
}

type runner struct {
	ctx          context.Context
	oauth        *oauth2.Client
	management   *management.Client
	accessToken  string
	refreshToken string
	clientID     string
	clientSecret string
	workspaceID  int64
	userID       int64
	writes       bool
	requireAll   bool
	steps        []step
}

func main() {
	if err := envfile.Load(); err != nil {
		log.Fatal(err)
	}
	apiKey, err := envfile.Required("OWLVIGIL_API_KEY")
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	clientID := os.Getenv("OWLVIGIL_CLIENT_ID")
	clientSecret := os.Getenv("OWLVIGIL_CLIENT_SECRET")
	accessToken := os.Getenv("OWLVIGIL_ACCESS_TOKEN")
	refreshToken := os.Getenv("OWLVIGIL_REFRESH_TOKEN")
	scope := envDefault("OWLVIGIL_SCOPE", defaultScope)

	oauthClient := oauth2.NewClient(owlvigil.WithEnvironmentFromEnv())
	if oauthEnabled(accessToken, clientID, clientSecret) && accessToken == "" {
		token, err := oauthClient.ClientCredentials(ctx, oauth2.ClientCredentialsRequest{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			Scopes:       strings.Fields(scope),
		})
		if err != nil {
			log.Fatalf("client_credentials failed: %v", err)
		}
		accessToken = token.AccessToken
		fmt.Printf("client_credentials: ok token_type=%s expires_in=%d scope=%q access_token=%s\n", token.TokenType, token.ExpiresIn, token.Scope, maskToken(accessToken))
	}

	managementClient := management.NewClient(
		owlvigil.WithEnvironmentFromEnv(),
		owlvigil.WithAPIKey(apiKey),
		owlvigil.WithoutRetry(),
	)
	r := &runner{
		ctx:          ctx,
		oauth:        oauthClient,
		management:   managementClient,
		accessToken:  accessToken,
		refreshToken: refreshToken,
		clientID:     clientID,
		clientSecret: clientSecret,
		writes:       writeSmokeEnabled(os.Getenv("OWLVIGIL_SMOKE_WRITES")),
		requireAll:   writeSmokeEnabled(os.Getenv("OWLVIGIL_SMOKE_REQUIRE_ALL")),
	}
	r.runAll()
	r.print()
	if !r.complete() {
		os.Exit(1)
	}
}

func (r *runner) runAll() {
	r.runOAuth()
	r.runUserAndWorkspace()
	r.runWorkspaceAccess()
	r.runGateway()
	r.runBilling()
	r.runFinancial()
	r.runPolicies()
	r.runWebhooks()
}

func (r *runner) runOAuth() {
	if !oauthEnabled(r.accessToken, r.clientID, r.clientSecret) {
		r.skip("oauth token", "POST /oauth/token", "OAuth credentials are not configured")
		r.skip("oauth userinfo", "GET /oauth/userinfo", "OAuth credentials are not configured")
		r.skip("oauth refresh", "POST /oauth/token/refresh", "OAuth credentials are not configured")
		r.skip("oauth revoke", "POST /oauth/revoke", "OAuth credentials are not configured")
		r.skip("oauth authorize", "GET /oauth/authorize", "browser/dashboard-login flow is covered by examples/oauth2-callback-server")
		return
	}
	r.pass("oauth token", "POST /oauth/token")
	r.call("oauth userinfo", "GET /oauth/userinfo", func() error {
		info, err := r.oauth.UserInfo(r.ctx, r.accessToken)
		if err == nil && info.Subject == "" {
			return errors.New("userinfo subject is empty")
		}
		return err
	})
	if r.refreshToken == "" {
		r.skip("oauth refresh", "POST /oauth/token/refresh", "set OWLVIGIL_REFRESH_TOKEN from an authorization_code flow")
	} else {
		r.call("oauth refresh", "POST /oauth/token/refresh", func() error {
			_, err := r.oauth.Refresh(r.ctx, oauth2.RefreshTokenRequest{
				ClientID:     r.clientID,
				ClientSecret: r.clientSecret,
				RefreshToken: r.refreshToken,
			})
			return err
		})
	}
	r.skip("oauth revoke", "POST /oauth/revoke", "not exercised by default; set up a disposable token before revocation testing")
	r.skip("oauth authorize", "GET /oauth/authorize", "browser/dashboard-login flow is covered by examples/oauth2-callback-server")
}

func (r *runner) runUserAndWorkspace() {
	r.call("get user profile", "GET /v1/user/profile", func() error {
		profile, _, err := r.management.GetUserProfile(r.ctx)
		if err != nil {
			return err
		}
		r.userID = profile.UserID
		if r.userID <= 0 {
			return errors.New("profile user_id is empty")
		}
		return nil
	})
	r.write("update user profile", "PUT /v1/user/profile", "user write endpoint not exercised in smoke", func() error {
		profile, _, err := r.management.GetUserProfile(r.ctx)
		if err != nil {
			return err
		}
		_, _, err = r.management.UpdateUserProfile(r.ctx, &management.UpdateUserProfileRequest{
			Name:               &profile.Name,
			AvatarURL:          &profile.AvatarURL,
			DefaultWorkspaceID: &profile.DefaultWorkspaceID,
		})
		return err
	})
	r.configuredWrite(
		"change user password",
		"PUT /v1/user/password",
		[]string{"OWLVIGIL_SMOKE_CURRENT_PASSWORD", "OWLVIGIL_SMOKE_NEW_PASSWORD"},
		func(values []string) error {
			currentPassword, newPassword := values[0], values[1]
			if currentPassword == newPassword {
				return errors.New("OWLVIGIL_SMOKE_NEW_PASSWORD must differ from OWLVIGIL_SMOKE_CURRENT_PASSWORD")
			}
			if _, err := r.management.UpdatePassword(r.ctx, &management.UpdatePasswordRequest{
				OldPassword: currentPassword,
				NewPassword: newPassword,
			}); err != nil {
				return err
			}
			cleanupCtx, cancel := r.cleanupContext()
			defer cancel()
			_, err := r.management.UpdatePassword(cleanupCtx, &management.UpdatePasswordRequest{
				OldPassword: newPassword,
				NewPassword: currentPassword,
			})
			return err
		},
	)
	r.write("submit support request", "POST /v1/user/support-requests", "support write endpoint not exercised unless OWLVIGIL_SMOKE_WRITES=1", func() error {
		_, err := r.management.CreateSupportRequest(r.ctx, &management.SupportRequest{
			Subject:     "Disposable SDK full-smoke request",
			IssueType:   "technical",
			Description: "Automated disposable-account contract verification. No response is required.",
		})
		return err
	})
	r.call("get notification preferences", "GET /v1/user/notification-preferences", func() error {
		_, _, err := r.management.GetNotificationPreferences(r.ctx)
		return err
	})
	r.write("update notification preferences", "PUT /v1/user/notification-preferences", "user write endpoint not exercised in smoke", func() error {
		preferences, _, err := r.management.GetNotificationPreferences(r.ctx)
		if err != nil {
			return err
		}
		_, _, err = r.management.UpdateNotificationPreferences(r.ctx, &management.UpdateNotificationPreferencesRequest{
			BudgetAlerts:    &preferences.BudgetAlerts,
			BillingAlerts:   &preferences.BillingAlerts,
			ReportEmails:    &preferences.ReportEmails,
			MarketingEmails: &preferences.MarketingEmails,
		})
		return err
	})
	r.call("get invite link", "GET /v1/users/me/invite-link", func() error {
		_, _, err := r.management.GetInviteLink(r.ctx)
		return err
	})
	r.call("get invitation stats", "GET /v1/users/me/invitation-stats", func() error {
		_, _, err := r.management.GetInvitationStats(r.ctx)
		return err
	})
	r.call("list user invitations", "GET /v1/users/me/invitations", func() error {
		_, _, err := r.management.ListUserInvitations(r.ctx, management.ListOptions{Limit: 5})
		return err
	})
	r.configuredWrite(
		"send invitation",
		"POST /v1/users/me/send-invitation",
		[]string{"OWLVIGIL_SMOKE_INVITE_EMAIL"},
		func(values []string) error {
			_, err := r.management.SendInvitation(r.ctx, &management.SendInvitationRequest{
				Emails:  []string{values[0]},
				Message: "Disposable SDK full-smoke invitation.",
			})
			return err
		},
	)

	r.call("list workspaces", "GET /v1/workspaces", func() error {
		workspaces, _, err := r.management.ListWorkspaces(r.ctx, management.ListOptions{Limit: 5})
		if err != nil {
			return err
		}
		if envWorkspace := os.Getenv("OWLVIGIL_WORKSPACE_ID"); envWorkspace != "" {
			id, parseErr := strconv.ParseInt(envWorkspace, 10, 64)
			if parseErr != nil {
				return parseErr
			}
			r.workspaceID = id
			return nil
		}
		if len(workspaces.Items) == 0 {
			return errors.New("no workspace returned")
		}
		r.workspaceID = workspaces.Items[0].ID
		if r.workspaceID <= 0 {
			return errors.New("workspace id is empty")
		}
		return nil
	})
	var temporaryWorkspaceID int64
	r.writeControlled(
		"create workspace",
		"POST /v1/workspaces",
		"workspace creation is disabled unless OWLVIGIL_SMOKE_WRITES=1",
		[]int{http.StatusBadRequest, http.StatusForbidden, http.StatusConflict, http.StatusUnprocessableEntity},
		func() error {
			workspace, _, err := r.management.CreateWorkspace(r.ctx, &management.CreateWorkspaceRequest{
				Name: smokeName("SDK Smoke Workspace"),
				Type: "team",
			})
			if err == nil {
				temporaryWorkspaceID = workspace.ID
			}
			return err
		},
	)
	if temporaryWorkspaceID > 0 {
		r.cleanup("delete workspace", "DELETE /v1/workspaces/:workspace_id", func(ctx context.Context) error {
			_, err := r.management.DeleteWorkspace(ctx, temporaryWorkspaceID)
			return err
		})
	} else {
		r.skip("delete workspace", "DELETE /v1/workspaces/:workspace_id", "temporary workspace was not created")
	}
	r.call("get workspace", "GET /v1/workspaces/:workspace_id", func() error {
		_, _, err := r.management.GetWorkspace(r.ctx, r.workspaceID)
		return err
	})
	r.call("get workspace overview", "GET /v1/workspaces/:workspace_id/overview", func() error {
		_, _, err := r.management.GetWorkspaceOverview(
			r.ctx,
			r.workspaceID,
			management.WorkspaceOverviewOptions{Range: "7d"},
		)
		return err
	})
	r.write("update workspace", "PATCH /v1/workspaces/:workspace_id", "workspace write endpoint not exercised in smoke", func() error {
		workspace, _, err := r.management.GetWorkspace(r.ctx, r.workspaceID)
		if err != nil {
			return err
		}
		_, _, err = r.management.UpdateWorkspace(r.ctx, r.workspaceID, &management.UpdateWorkspaceRequest{
			Name:        &workspace.Name,
			Description: &workspace.Description,
		})
		return err
	})
}

func (r *runner) runWorkspaceAccess() {
	r.call("get workspace activity", "GET /v1/workspaces/:workspace_id/activity", func() error {
		_, _, err := r.management.ListWorkspaceActivity(r.ctx, r.workspaceID, management.ListOptions{Limit: 5})
		return err
	})
	r.call("list workspace audit logs", "GET /v1/workspaces/:workspace_id/audit-logs", func() error {
		logs, _, err := r.management.ListAuditLogs(r.ctx, r.workspaceID, management.AuditLogListOptions{Limit: 5})
		if err != nil {
			return err
		}
		if len(logs.Items) == 0 {
			r.configuredCall(
				"get workspace audit log",
				"GET /v1/workspaces/:workspace_id/audit-logs/:audit_log_id",
				[]string{"OWLVIGIL_SMOKE_AUDIT_LOG_ID"},
				func(values []string) error {
					auditLogID, parseErr := strconv.ParseInt(values[0], 10, 64)
					if parseErr != nil || auditLogID <= 0 {
						return errors.New("OWLVIGIL_SMOKE_AUDIT_LOG_ID must be a positive integer")
					}
					_, _, err := r.management.GetAuditLog(r.ctx, r.workspaceID, auditLogID)
					return err
				},
			)
			return nil
		}
		_, _, err = r.management.GetAuditLog(r.ctx, r.workspaceID, logs.Items[0].ID)
		r.recordErr("get workspace audit log", "GET /v1/workspaces/:workspace_id/audit-logs/:audit_log_id", err)
		return nil
	})
	r.call("get workspace logging settings", "GET /v1/workspaces/:workspace_id/logging-settings", func() error {
		_, _, err := r.management.GetLoggingSettings(r.ctx, r.workspaceID)
		return err
	})
	r.write("update workspace logging settings", "PUT /v1/workspaces/:workspace_id/logging-settings", "logging write endpoint not exercised in smoke", func() error {
		settings, _, err := r.management.GetLoggingSettings(r.ctx, r.workspaceID)
		if err != nil {
			return err
		}
		redactHeaders := append([]string(nil), settings.RedactHeaders...)
		_, _, err = r.management.UpdateLoggingSettings(r.ctx, r.workspaceID, &management.UpdateLoggingSettingsRequest{
			Enabled:              &settings.Enabled,
			RetentionDays:        &settings.RetentionDays,
			CaptureHeaders:       &settings.CaptureHeaders,
			CaptureRequestBody:   &settings.CaptureRequestBody,
			CaptureResponseBody:  &settings.CaptureResponseBody,
			CaptureStreamChunks:  &settings.CaptureStreamChunks,
			MaskSensitiveHeaders: &settings.MaskSensitiveHeaders,
			MaxBodySize:          &settings.MaxBodySize,
			RedactHeaders:        &redactHeaders,
		})
		return err
	})
	r.call("get workspace quota summary", "GET /v1/workspaces/:workspace_id/quota-summary", func() error {
		_, _, err := r.management.GetQuotaSummary(r.ctx, r.workspaceID)
		return err
	})
	r.call("get workspace quota usage", "GET /v1/workspaces/:workspace_id/quota-usage", func() error {
		_, _, err := r.management.GetQuotaUsage(r.ctx, r.workspaceID)
		return err
	})
	var teamForGetID int64
	r.call("list workspace teams", "GET /v1/workspaces/:workspace_id/teams", func() error {
		teams, _, err := r.management.ListTeams(r.ctx, r.workspaceID, management.ListOptions{Limit: 5})
		if err != nil {
			return err
		}
		if len(teams.Items) > 0 {
			teamForGetID = teams.Items[0].ID
		}
		return nil
	})
	var createdTeamID int64
	r.write("create workspace team", "POST /v1/workspaces/:workspace_id/teams", "workspace write endpoint not exercised in smoke", func() error {
		team, _, err := r.management.CreateTeam(r.ctx, r.workspaceID, &management.CreateTeamRequest{Name: smokeName("SDK Smoke Team")})
		if err == nil {
			createdTeamID = team.ID
			if teamForGetID == 0 {
				teamForGetID = team.ID
			}
		}
		return err
	})
	if teamForGetID > 0 {
		r.call("get workspace team", "GET /v1/workspaces/:workspace_id/teams/:team_id", func() error {
			_, _, err := r.management.GetTeam(r.ctx, r.workspaceID, teamForGetID)
			return err
		})
	} else {
		r.skip("get workspace team", "GET /v1/workspaces/:workspace_id/teams/:team_id", "no existing or temporary team is available")
	}
	if createdTeamID > 0 {
		updatedName := smokeName("SDK Smoke Team Updated")
		r.call("update workspace team", "PATCH /v1/workspaces/:workspace_id/teams/:team_id", func() error {
			_, _, err := r.management.UpdateTeam(r.ctx, r.workspaceID, createdTeamID, &management.UpdateTeamRequest{Name: &updatedName})
			return err
		})
		r.cleanup("delete workspace team", "DELETE /v1/workspaces/:workspace_id/teams/:team_id", func(ctx context.Context) error {
			_, err := r.management.DeleteTeam(ctx, r.workspaceID, createdTeamID)
			return err
		})
	} else {
		r.skip("update workspace team", "PATCH /v1/workspaces/:workspace_id/teams/:team_id", "temporary team was not created")
		r.skip("delete workspace team", "DELETE /v1/workspaces/:workspace_id/teams/:team_id", "temporary team was not created")
	}

	r.call("list workspace members", "GET /v1/workspaces/:workspace_id/members", func() error {
		members, _, err := r.management.ListMembers(r.ctx, r.workspaceID, management.ListOptions{Limit: 5})
		if err != nil {
			return err
		}
		if len(members.Items) > 0 {
			userID := members.Items[0].UserID
			_, _, err = r.management.GetMember(r.ctx, r.workspaceID, userID)
			r.recordErr("get workspace member", "GET /v1/workspaces/:workspace_id/members/:member_id", err)
			_, _, err = r.management.GetMemberPermissions(r.ctx, r.workspaceID, userID)
			r.recordErr("get workspace member permissions", "GET /v1/workspaces/:workspace_id/members/:member_id/permissions", err)
		} else {
			r.skip("get workspace member", "GET /v1/workspaces/:workspace_id/members/:member_id", "members list returned no members")
			r.skip("get workspace member permissions", "GET /v1/workspaces/:workspace_id/members/:member_id/permissions", "members list returned no members")
		}
		return nil
	})
	r.callSkipKnown("get member role options", "GET /v1/workspaces/:workspace_id/members/role-options", []string{"feature.rbac is not included"}, func() error {
		_, _, err := r.management.ListRoleOptions(r.ctx, r.workspaceID)
		return err
	})
	var disposableMember *management.Member
	r.configuredWrite(
		"add workspace member",
		"POST /v1/workspaces/:workspace_id/members",
		[]string{"OWLVIGIL_SMOKE_MEMBER_EMAIL"},
		func(values []string) error {
			member, _, err := r.management.CreateMember(r.ctx, r.workspaceID, &management.CreateMemberRequest{
				Email: values[0],
				Role:  "member",
			})
			if err == nil {
				disposableMember = member
			}
			return err
		},
	)
	if disposableMember == nil || disposableMember.UserID <= 0 {
		r.skip("update workspace member", "PATCH /v1/workspaces/:workspace_id/members/:member_id", "disposable registered member was not created")
	} else {
		r.call("update workspace member", "PATCH /v1/workspaces/:workspace_id/members/:member_id", func() error {
			_, _, err := r.management.UpdateMember(
				r.ctx,
				r.workspaceID,
				disposableMember.UserID,
				&management.UpdateMemberRequest{
					Role:   disposableMember.Role,
					Status: &disposableMember.Status,
					TeamID: disposableMember.TeamID,
				},
			)
			return err
		})
	}

	r.call("list workspace invitations", "GET /v1/workspaces/:workspace_id/invitations", func() error {
		_, _, err := r.management.ListInvitations(r.ctx, r.workspaceID, management.ListOptions{Limit: 5})
		return err
	})
	var disposableInvitationID int64
	r.configuredWrite(
		"invite workspace member",
		"POST /v1/workspaces/:workspace_id/invitations",
		[]string{"OWLVIGIL_SMOKE_WORKSPACE_INVITE_EMAIL"},
		func(values []string) error {
			invitation, _, err := r.management.CreateInvitation(r.ctx, r.workspaceID, &management.CreateInvitationRequest{
				Email:          values[0],
				Role:           "member",
				ExpiresInHours: 24,
			})
			if err == nil {
				disposableInvitationID = invitation.ID
			}
			return err
		},
	)
	if disposableInvitationID <= 0 {
		r.skip("resend workspace invitation", "POST /v1/workspaces/:workspace_id/invitations/:invitation_id/resend", "disposable invitation was not created")
		r.skip("revoke workspace invitation", "POST /v1/workspaces/:workspace_id/invitations/:invitation_id/revoke", "disposable invitation was not created")
	} else {
		r.call("resend workspace invitation", "POST /v1/workspaces/:workspace_id/invitations/:invitation_id/resend", func() error {
			_, err := r.management.ResendInvitation(r.ctx, r.workspaceID, disposableInvitationID)
			return err
		})
		r.call("revoke workspace invitation", "POST /v1/workspaces/:workspace_id/invitations/:invitation_id/revoke", func() error {
			_, err := r.management.RevokeInvitation(r.ctx, r.workspaceID, disposableInvitationID)
			return err
		})
	}

	var roleForGetID int64
	r.callSkipKnown("list workspace roles", "GET /v1/workspaces/:workspace_id/roles", []string{"feature.rbac is not included"}, func() error {
		roles, _, err := r.management.ListRoles(r.ctx, r.workspaceID, management.ListOptions{Limit: 5})
		if err != nil {
			return err
		}
		if len(roles.Items) > 0 {
			roleForGetID = roles.Items[0].ID
		}
		return nil
	})
	var createdRoleID int64
	r.write("create workspace role", "POST /v1/workspaces/:workspace_id/roles", "workspace write endpoint not exercised in smoke", func() error {
		role, _, err := r.management.CreateRole(r.ctx, r.workspaceID, &management.CreateRoleRequest{Name: smokeName("SDK Smoke Role")})
		if err == nil {
			createdRoleID = role.ID
			if roleForGetID == 0 {
				roleForGetID = role.ID
			}
		}
		return err
	})
	if roleForGetID > 0 {
		r.callSkipKnown("get workspace role", "GET /v1/workspaces/:workspace_id/roles/:role_id", []string{"feature.rbac is not included"}, func() error {
			_, _, err := r.management.GetRole(r.ctx, r.workspaceID, roleForGetID)
			return err
		})
	} else {
		r.skip("get workspace role", "GET /v1/workspaces/:workspace_id/roles/:role_id", "no existing or temporary role is available")
	}
	if createdRoleID > 0 {
		updatedName := smokeName("SDK Smoke Role Updated")
		r.call("update workspace role", "PATCH /v1/workspaces/:workspace_id/roles/:role_id", func() error {
			_, _, err := r.management.UpdateRole(r.ctx, r.workspaceID, createdRoleID, &management.UpdateRoleRequest{Name: &updatedName})
			return err
		})
		r.cleanup("delete workspace role", "DELETE /v1/workspaces/:workspace_id/roles/:role_id", func(ctx context.Context) error {
			_, err := r.management.DeleteRole(ctx, r.workspaceID, createdRoleID)
			return err
		})
	} else {
		r.skip("update workspace role", "PATCH /v1/workspaces/:workspace_id/roles/:role_id", "temporary role was not created")
		r.skip("delete workspace role", "DELETE /v1/workspaces/:workspace_id/roles/:role_id", "temporary role was not created")
	}
	r.callSkipKnown("list workspace permissions", "GET /v1/workspaces/:workspace_id/permissions", []string{"feature.rbac is not included"}, func() error {
		_, _, err := r.management.ListPermissions(r.ctx, r.workspaceID)
		return err
	})
	if disposableMember == nil || disposableMember.UserID <= 0 {
		r.skip("update workspace member permissions", "PUT /v1/workspaces/:workspace_id/members/:member_id/permissions", "disposable registered member was not created")
		r.skip("reset workspace member permissions", "POST /v1/workspaces/:workspace_id/members/:member_id/permissions/reset", "disposable registered member was not created")
		r.skip("remove workspace member", "DELETE /v1/workspaces/:workspace_id/members/:member_id", "disposable registered member was not created")
	} else {
		r.writeSkipKnown(
			"update workspace member permissions",
			"PUT /v1/workspaces/:workspace_id/members/:member_id/permissions",
			"permission write endpoint not exercised in smoke",
			[]string{"feature.rbac is not included"},
			func() error {
				current, _, err := r.management.GetMemberPermissions(r.ctx, r.workspaceID, disposableMember.UserID)
				if err != nil {
					return err
				}
				overrides := map[string]bool{}
				for _, group := range current.Groups {
					for _, permission := range group.Permissions {
						switch permission.Override {
						case "allow":
							overrides[permission.ID] = true
						case "deny":
							overrides[permission.ID] = false
						}
					}
				}
				_, _, err = r.management.UpdateMemberPermissions(
					r.ctx,
					r.workspaceID,
					disposableMember.UserID,
					&management.UpdateMemberPermissionsRequest{PermissionMap: overrides},
				)
				return err
			},
		)
		r.writeSkipKnown(
			"reset workspace member permissions",
			"POST /v1/workspaces/:workspace_id/members/:member_id/permissions/reset",
			"permission write endpoint not exercised in smoke",
			[]string{"feature.rbac is not included"},
			func() error {
				_, _, err := r.management.ResetMemberPermissions(r.ctx, r.workspaceID, disposableMember.UserID)
				return err
			},
		)
		r.cleanup("remove workspace member", "DELETE /v1/workspaces/:workspace_id/members/:member_id", func(ctx context.Context) error {
			_, err := r.management.DeleteMember(ctx, r.workspaceID, disposableMember.UserID)
			return err
		})
	}
}

func (r *runner) runGateway() {
	workspaceOpt := owlvigil.WithWorkspaceID(r.workspaceID)
	var providerID int64
	r.call("list gateway providers", "GET /v1/gateway/providers", func() error {
		providers, _, err := r.management.ListProviders(r.ctx, r.workspaceID)
		if err != nil {
			return err
		}
		if len(providers.Items) > 0 {
			providerID = providers.Items[0].ID
		}
		return nil
	})
	var temporaryProviderID int64
	r.configuredWrite(
		"create gateway provider",
		"POST /v1/gateway/providers",
		[]string{"OWLVIGIL_SMOKE_PROVIDER_API_KEY"},
		func(values []string) error {
			provider, _, err := r.management.CreateProvider(r.ctx, &management.CreateProviderRequest{
				WorkspaceID: r.workspaceID,
				Name:        smokeName("SDK Smoke Provider"),
				Type:        envDefault("OWLVIGIL_SMOKE_PROVIDER_TYPE", "openai"),
				APIKey:      values[0],
				BaseURL:     strings.TrimSpace(os.Getenv("OWLVIGIL_SMOKE_PROVIDER_BASE_URL")),
			})
			if err == nil {
				temporaryProviderID = provider.ID
				providerID = provider.ID
			}
			return err
		},
	)
	if providerID > 0 {
		r.call("get gateway provider", "GET /v1/gateway/providers/:provider_id", func() error {
			_, _, err := r.management.GetProvider(r.ctx, r.workspaceID, providerID)
			return err
		})
	} else {
		r.skip("get gateway provider", "GET /v1/gateway/providers/:provider_id", "no existing or temporary provider is available")
	}
	if temporaryProviderID > 0 {
		updatedName := smokeName("SDK Smoke Provider Updated")
		r.call("update gateway provider", "PATCH /v1/gateway/providers/:provider_id", func() error {
			_, _, err := r.management.UpdateProvider(
				r.ctx,
				r.workspaceID,
				temporaryProviderID,
				&management.UpdateProviderRequest{Name: &updatedName},
			)
			return err
		})
	} else {
		r.skip("update gateway provider", "PATCH /v1/gateway/providers/:provider_id", "temporary provider was not created")
	}
	verifyProvider := func() error {
		providerIDValue := int(providerID)
		request := &management.VerifyProviderConnectionRequest{
			WorkspaceID: int(r.workspaceID),
		}
		if providerIDValue > 0 {
			request.ProviderID = &providerIDValue
		} else {
			request.Type = envDefault("OWLVIGIL_SMOKE_PROVIDER_TYPE", "openai")
			request.APIKey = envDefault("OWLVIGIL_SMOKE_PROVIDER_API_KEY", "sdk-smoke-invalid-provider-key")
			request.BaseURL = strings.TrimSpace(os.Getenv("OWLVIGIL_SMOKE_PROVIDER_BASE_URL"))
		}
		_, _, err := r.management.VerifyProviderConnection(r.ctx, request)
		return err
	}
	if providerID > 0 {
		r.write("verify gateway provider connection", "POST /v1/gateway/providers/verify-connection", "provider verification is disabled unless OWLVIGIL_SMOKE_WRITES=1", verifyProvider)
	} else {
		r.configuredWrite(
			"verify gateway provider connection",
			"POST /v1/gateway/providers/verify-connection",
			[]string{"OWLVIGIL_SMOKE_PROVIDER_API_KEY"},
			func(values []string) error {
				return verifyProvider()
			},
		)
	}
	if temporaryProviderID > 0 {
		r.cleanup("delete gateway provider", "DELETE /v1/gateway/providers/:provider_id", func(ctx context.Context) error {
			_, err := r.management.DeleteProvider(ctx, r.workspaceID, temporaryProviderID)
			return err
		})
	} else {
		r.skip("delete gateway provider", "DELETE /v1/gateway/providers/:provider_id", "temporary provider was not created")
	}

	r.call("list gateway keys", "GET /v1/gateway/keys", func() error {
		_, _, err := r.management.ListGatewayKeys(r.ctx, management.ListOptions{Limit: 5}, "", workspaceOpt)
		return err
	})
	var keyID int64
	r.writeSkipKnown("create gateway key", "POST /v1/gateway/keys", "gateway key write endpoint not exercised in smoke", []string{"quota.gateway_keys limit exceeded"}, func() error {
		key, _, err := r.management.CreateGatewayKey(r.ctx, &management.CreateGatewayKeyRequest{
			WorkspaceID:    r.workspaceID,
			Name:           smokeName("SDK Smoke Key"),
			ProviderSource: "byok",
		}, owlvigil.WithIdempotencyKey(smokeName("sdk-smoke-key")))
		if err != nil {
			return err
		}
		keyID = key.ID
		return nil
	})
	if keyID > 0 {
		r.call("get gateway key", "GET /v1/gateway/keys/:key_id", func() error {
			_, _, err := r.management.GetGatewayKey(r.ctx, keyID, workspaceOpt)
			return err
		})
		updatedName := smokeName("SDK Smoke Key Updated")
		r.write("update gateway key", "PATCH /v1/gateway/keys/:key_id", "gateway key write endpoint not exercised in smoke", func() error {
			_, _, err := r.management.UpdateGatewayKey(r.ctx, keyID, &management.UpdateGatewayKeyRequest{Name: &updatedName}, workspaceOpt)
			return err
		})
		r.write("rotate gateway key", "POST /v1/gateway/keys/:key_id/rotate", "gateway key write endpoint not exercised in smoke", func() error {
			_, _, err := r.management.RotateGatewayKey(r.ctx, keyID, workspaceOpt)
			return err
		})
		r.write("disable gateway key", "POST /v1/gateway/keys/:key_id/disable", "gateway key write endpoint not exercised in smoke", func() error {
			_, err := r.management.DisableGatewayKey(r.ctx, keyID, workspaceOpt)
			return err
		})
		r.write("enable gateway key", "POST /v1/gateway/keys/:key_id/enable", "gateway key write endpoint not exercised in smoke", func() error {
			_, err := r.management.EnableGatewayKey(r.ctx, keyID, workspaceOpt)
			return err
		})
		r.cleanup("delete gateway key", "DELETE /v1/gateway/keys/:key_id", func(ctx context.Context) error {
			_, err := r.management.DeleteGatewayKey(ctx, keyID, workspaceOpt)
			return err
		})
	} else {
		r.skip("get gateway key", "GET /v1/gateway/keys/:key_id", "gateway key was not created")
		r.skip("update gateway key", "PATCH /v1/gateway/keys/:key_id", "gateway key was not created")
		r.skip("rotate gateway key", "POST /v1/gateway/keys/:key_id/rotate", "gateway key was not created")
		r.skip("disable gateway key", "POST /v1/gateway/keys/:key_id/disable", "gateway key was not created")
		r.skip("enable gateway key", "POST /v1/gateway/keys/:key_id/enable", "gateway key was not created")
		r.skip("delete gateway key", "DELETE /v1/gateway/keys/:key_id", "gateway key was not created")
	}

	var modelID string
	r.call("list models", "GET /v1/gateway/models", func() error {
		models, _, err := r.management.ListModels(r.ctx, management.ListOptions{Limit: 5}, workspaceOpt)
		if err != nil {
			return err
		}
		if len(models.Items) > 0 {
			modelID = models.Items[0].ID
		}
		if modelID == "" {
			modelID = strings.TrimSpace(os.Getenv("OWLVIGIL_SMOKE_MODEL_ID"))
		}
		return nil
	})
	if modelID != "" {
		r.call("get model", "GET /v1/gateway/models/:model_id", func() error {
			_, _, err := r.management.GetModel(r.ctx, modelID, workspaceOpt)
			return err
		})
		r.call("preview route", "POST /v1/gateway/routes/preview", func() error {
			_, _, err := r.management.PreviewRoute(r.ctx, &management.PreviewRouteRequest{WorkspaceID: r.workspaceID, Model: modelID})
			return err
		})
	} else {
		r.skip("get model", "GET /v1/gateway/models/:model_id", "models list returned no models")
		r.skip("preview route", "POST /v1/gateway/routes/preview", "models list returned no models")
	}
	r.call("list routes", "GET /v1/gateway/routes", func() error {
		routes, _, err := r.management.ListRoutes(r.ctx, management.ListOptions{Limit: 5}, workspaceOpt)
		if err != nil {
			return err
		}
		if len(routes.Items) == 0 || routes.Items[0].ID == "" {
			r.configuredCall(
				"get route",
				"GET /v1/gateway/routes/:route_id",
				[]string{"OWLVIGIL_SMOKE_ROUTE_ID"},
				func(values []string) error {
					_, _, err := r.management.GetRoute(r.ctx, values[0], workspaceOpt)
					return err
				},
			)
			return nil
		}
		_, _, err = r.management.GetRoute(r.ctx, routes.Items[0].ID, workspaceOpt)
		r.recordErr("get route", "GET /v1/gateway/routes/:route_id", err)
		return nil
	})
	r.call("gateway usage", "GET /v1/gateway/usage", func() error {
		_, _, err := r.management.ListUsage(r.ctx, management.ListOptions{Limit: 5}, workspaceOpt)
		return err
	})
	r.call("gateway usage summary", "GET /v1/gateway/usage/summary", func() error {
		_, _, err := r.management.GetUsageSummary(r.ctx, workspaceOpt)
		return err
	})
	r.call("gateway quota", "GET /v1/gateway/quota", func() error {
		_, _, err := r.management.GetQuota(r.ctx, workspaceOpt)
		return err
	})
	r.call("list request logs", "GET /v1/gateway/request-logs", func() error {
		logs, _, err := r.management.ListRequestLogs(r.ctx, management.ListOptions{Limit: 5}, "", workspaceOpt)
		if err != nil {
			return err
		}
		if len(logs.Items) > 0 && logs.Items[0].RequestID != "" {
			_, _, err = r.management.GetRequestLog(r.ctx, logs.Items[0].RequestID, workspaceOpt)
			r.recordErr("get request log", "GET /v1/gateway/request-logs/:request_id", err)
		} else {
			r.configuredCall(
				"get request log",
				"GET /v1/gateway/request-logs/:request_id",
				[]string{"OWLVIGIL_SMOKE_REQUEST_ID"},
				func(values []string) error {
					_, _, err := r.management.GetRequestLog(r.ctx, values[0], workspaceOpt)
					return err
				},
			)
		}
		return nil
	})
	r.call("list traces", "GET /v1/gateway/traces", func() error {
		traces, _, err := r.management.ListTraces(r.ctx, management.ListOptions{Limit: 5}, workspaceOpt)
		if err != nil {
			return err
		}
		if len(traces.Items) > 0 && traces.Items[0].TraceID != "" {
			_, _, err = r.management.GetTrace(r.ctx, traces.Items[0].TraceID, workspaceOpt)
			r.recordErr("get trace", "GET /v1/gateway/traces/:trace_id", err)
		} else {
			r.configuredCall(
				"get trace",
				"GET /v1/gateway/traces/:trace_id",
				[]string{"OWLVIGIL_SMOKE_TRACE_ID"},
				func(values []string) error {
					_, _, err := r.management.GetTrace(r.ctx, values[0], workspaceOpt)
					return err
				},
			)
		}
		return nil
	})
	r.call("payload log access", "GET /v1/gateway/payload-logs/access", func() error {
		_, _, err := r.management.GetPayloadAccess(r.ctx, workspaceOpt)
		return err
	})
	r.call("list payload logs", "GET /v1/gateway/payload-logs", func() error {
		logs, _, err := r.management.ListPayloadLogs(r.ctx, r.workspaceID, management.ListOptions{Limit: 5})
		if err != nil {
			return err
		}
		if len(logs.Items) == 0 {
			r.configuredCall(
				"get payload log",
				"GET /v1/gateway/payload-logs/:payload_id",
				[]string{"OWLVIGIL_SMOKE_PAYLOAD_ID"},
				func(values []string) error {
					_, _, err := r.management.GetPayloadLog(r.ctx, values[0], workspaceOpt)
					return err
				},
			)
			return nil
		}
		_, _, err = r.management.GetPayloadLog(r.ctx, strconv.Itoa(logs.Items[0].ID), workspaceOpt)
		r.recordErr("get payload log", "GET /v1/gateway/payload-logs/:payload_id", err)
		return nil
	})
}

func (r *runner) runBilling() {
	stripePaymentMethodID := stripePaymentMethodID(os.Getenv("OWLVIGIL_STRIPE_TEST_PAYMENT_METHOD_ID"))
	r.call("billing overview", "GET /v1/billing/overview", func() error {
		_, _, err := r.management.GetBillingOverviewForWorkspace(r.ctx, r.workspaceID)
		return err
	})
	r.call("billing balance", "GET /v1/billing/balance", func() error {
		_, _, err := r.management.GetBalanceForWorkspace(r.ctx, r.workspaceID)
		return err
	})
	var planID, planInterval string
	r.call("billing plans", "GET /v1/billing/plans", func() error {
		plans, _, err := r.management.ListPlans(r.ctx, management.ListOptions{})
		if err != nil {
			return err
		}
		for _, plan := range plans.Items {
			if plan.ForSale && plan.ID != "" && plan.Interval != "" {
				planID = plan.ID
				planInterval = plan.Interval
				break
			}
		}
		if planID == "" && len(plans.Items) > 0 {
			planID = plans.Items[0].ID
			planInterval = plans.Items[0].Interval
		}
		if planID == "" {
			planID = strings.TrimSpace(os.Getenv("OWLVIGIL_SMOKE_PLAN_ID"))
			planInterval = strings.TrimSpace(os.Getenv("OWLVIGIL_SMOKE_PLAN_INTERVAL"))
		}
		return nil
	})
	if planID != "" {
		r.call("billing plan", "GET /v1/billing/plans/:plan_id", func() error {
			_, _, err := r.management.GetPlan(r.ctx, planID)
			return err
		})
	} else {
		r.skip("billing plan", "GET /v1/billing/plans/:plan_id", "plans list returned no plans")
	}
	r.call("billing subscription", "GET /v1/billing/subscription", func() error {
		_, _, err := r.management.GetSubscription(r.ctx)
		return err
	})
	r.call("billing topup plans", "GET /v1/billing/topup-plans", func() error {
		_, _, err := r.management.ListTopupPlans(r.ctx, management.ListOptions{})
		return err
	})
	r.call("billing topups", "GET /v1/billing/topups", func() error {
		topups, _, err := r.management.ListTopups(r.ctx, management.ListOptions{Limit: 5})
		if err != nil {
			return err
		}
		if len(topups.Items) > 0 && topups.Items[0].ID != "" {
			_, _, err = r.management.GetTopup(r.ctx, topups.Items[0].ID)
			r.recordErr("billing topup", "GET /v1/billing/topups/:topup_id", err)
		} else {
			r.configuredCall(
				"billing topup",
				"GET /v1/billing/topups/:topup_id",
				[]string{"OWLVIGIL_SMOKE_TOPUP_ID"},
				func(values []string) error {
					_, _, err := r.management.GetTopup(r.ctx, values[0])
					return err
				},
			)
		}
		return nil
	})
	r.call("billing invoices", "GET /v1/billing/invoices", func() error {
		invoices, _, err := r.management.ListInvoicesForWorkspace(r.ctx, r.workspaceID, management.ListOptions{Limit: 5})
		if err != nil {
			return err
		}
		if len(invoices.Items) > 0 && invoices.Items[0].ID != "" {
			_, _, err = r.management.GetInvoiceForWorkspace(r.ctx, r.workspaceID, invoices.Items[0].ID)
			r.recordErr("billing invoice", "GET /v1/billing/invoices/:invoice_id", err)
		} else {
			r.configuredCall(
				"billing invoice",
				"GET /v1/billing/invoices/:invoice_id",
				[]string{"OWLVIGIL_SMOKE_INVOICE_ID"},
				func(values []string) error {
					_, _, err := r.management.GetInvoiceForWorkspace(r.ctx, r.workspaceID, values[0])
					return err
				},
			)
		}
		return nil
	})
	r.call("billing orders", "GET /v1/billing/orders", func() error {
		orders, _, err := r.management.ListOrders(r.ctx, management.ListOptions{Limit: 5}, "")
		if err != nil {
			return err
		}
		if len(orders.Items) > 0 && orders.Items[0].ID != "" {
			_, _, err = r.management.GetOrder(r.ctx, orders.Items[0].ID)
			r.recordErr("billing order", "GET /v1/billing/orders/:order_id", err)
		} else {
			r.configuredCall(
				"billing order",
				"GET /v1/billing/orders/:order_id",
				[]string{"OWLVIGIL_SMOKE_ORDER_ID"},
				func(values []string) error {
					_, _, err := r.management.GetOrder(r.ctx, values[0])
					return err
				},
			)
		}
		return nil
	})
	r.call("billing details", "GET /v1/workspaces/:workspace_id/billing-details", func() error {
		_, _, err := r.management.GetBillingDetails(r.ctx, r.workspaceID)
		return err
	})
	var originalDefaultPaymentMethodID string
	r.call("list payment methods", "GET /v1/billing/payment-methods", func() error {
		methods, _, err := r.management.ListPaymentMethodsForWorkspace(r.ctx, r.workspaceID)
		if err != nil {
			return err
		}
		for _, method := range methods.Items {
			if method.IsDefault {
				originalDefaultPaymentMethodID = method.ID
				break
			}
		}
		return err
	})
	r.write("update workspace billing details", "PUT /v1/workspaces/:workspace_id/billing-details", "billing write endpoint not exercised in smoke", func() error {
		details, _, err := r.management.GetBillingDetails(r.ctx, r.workspaceID)
		if err != nil {
			return err
		}
		_, _, err = r.management.UpdateBillingDetails(r.ctx, r.workspaceID, &management.UpdateBillingDetailsRequest{
			BillingDetails: &management.BillingContact{
				Name:         details.Name,
				Email:        details.Email,
				TaxID:        details.TaxID,
				Phone:        details.Phone,
				Address:      details.AddressText,
				CCRecipients: append([]string(nil), details.CCRecipients...),
			},
		})
		return err
	})
	r.configuredWrite(
		"billing checkout",
		"POST /v1/billing/topups/checkout",
		[]string{"OWLVIGIL_SMOKE_TOPUP_AMOUNT", "OWLVIGIL_SMOKE_SUCCESS_URL", "OWLVIGIL_SMOKE_CANCEL_URL"},
		func(values []string) error {
			amount, err := positiveAmount(values[0])
			if err != nil {
				return err
			}
			_, _, err = r.management.CreateTopupCheckout(r.ctx, &management.CreateTopupCheckoutRequest{
				WorkspaceID: r.workspaceID,
				Amount:      amount,
				SuccessURL:  values[1],
				CancelURL:   values[2],
			})
			return err
		},
	)
	r.configuredWrite(
		"billing confirm stripe session",
		"POST /v1/billing/orders/:order_id/confirm-stripe-session",
		[]string{"OWLVIGIL_SMOKE_STRIPE_ORDER_ID", "OWLVIGIL_SMOKE_STRIPE_SESSION_ID"},
		func(values []string) error {
			_, _, err := r.management.ConfirmStripeSession(
				r.ctx,
				values[0],
				&management.ConfirmStripeSessionRequest{SessionID: values[1]},
			)
			return err
		},
	)
	var subscriptionCheckoutSessionID string
	r.configuredWrite(
		"create subscription checkout",
		"POST /v1/billing/subscription/checkout",
		[]string{"OWLVIGIL_SMOKE_SUCCESS_URL", "OWLVIGIL_SMOKE_CANCEL_URL"},
		func(values []string) error {
			if planID == "" || planInterval == "" {
				return errors.New("no saleable subscription plan is available")
			}
			checkout, _, err := r.management.CreateSubscriptionCheckout(r.ctx, &management.CreateSubscriptionCheckoutRequest{
				PlanID:     planID,
				Interval:   planInterval,
				SuccessURL: values[0],
				CancelURL:  values[1],
			})
			if err == nil {
				subscriptionCheckoutSessionID = checkout.SessionID
				if subscriptionCheckoutSessionID == "" {
					return errors.New("subscription checkout session ID is empty")
				}
			}
			return err
		},
	)
	r.write("create payment method setup intent", "POST /v1/billing/payment-methods/setup-intent", "billing write endpoint not exercised in smoke", func() error {
		setupIntent, _, err := r.management.CreatePaymentMethodSetupIntentForWorkspace(r.ctx, r.workspaceID)
		if err == nil && (setupIntent.SetupIntentID == "" || setupIntent.ClientSecret == "") {
			return errors.New("payment method setup intent is incomplete")
		}
		return err
	})

	var savedPaymentMethodID string
	if stripePaymentMethodID == "" {
		r.skip("save Stripe test payment method", "POST /v1/billing/payment-methods", "set OWLVIGIL_STRIPE_TEST_PAYMENT_METHOD_ID to a confirmed Stripe test payment method")
		r.skip("set Stripe test payment method default", "PUT /v1/billing/payment-methods/:payment_method_id/default", "no Stripe test payment method is configured")
	} else {
		r.write("save Stripe test payment method", "POST /v1/billing/payment-methods", "billing write endpoint not exercised in smoke", func() error {
			method, _, err := r.management.SavePaymentMethod(r.ctx, &management.SavePaymentMethodRequest{
				PaymentMethodID: stripePaymentMethodID,
				SetAsDefault:    true,
			})
			if err == nil {
				savedPaymentMethodID = method.ID
			}
			return err
		})
		if savedPaymentMethodID == "" {
			r.skip("set Stripe test payment method default", "PUT /v1/billing/payment-methods/:payment_method_id/default", "Stripe test payment method was not saved")
		} else {
			r.call("set Stripe test payment method default", "PUT /v1/billing/payment-methods/:payment_method_id/default", func() error {
				_, _, err := r.management.SetDefaultPaymentMethod(r.ctx, savedPaymentMethodID)
				return err
			})
		}
	}

	var topupPaymentIntentID, topupClientSecret string
	r.configuredWrite(
		"billing in-app",
		"POST /v1/billing/topups/in-app",
		[]string{"OWLVIGIL_SMOKE_TOPUP_AMOUNT", "OWLVIGIL_SMOKE_RETURN_URL"},
		func(values []string) error {
			amount, err := positiveAmount(values[0])
			if err != nil {
				return err
			}
			payment, _, err := r.management.CreateTopupInApp(r.ctx, &management.CreateTopupInAppRequest{
				WorkspaceID: r.workspaceID,
				Amount:      amount,
				ReturnURL:   values[1],
			})
			if err == nil {
				topupPaymentIntentID = payment.PaymentIntentID
				topupClientSecret = payment.ClientSecret
			}
			return err
		},
	)
	if topupPaymentIntentID != "" && topupClientSecret != "" {
		r.call("confirm topup in-app", "POST /v1/billing/topups/in-app/confirm", func() error {
			_, _, err := r.management.ConfirmTopupInApp(r.ctx, &management.ConfirmTopupInAppRequest{
				PaymentIntentID: topupPaymentIntentID,
				ClientSecret:    topupClientSecret,
			})
			return err
		})
	} else {
		r.configuredWrite(
			"confirm topup in-app",
			"POST /v1/billing/topups/in-app/confirm",
			[]string{"OWLVIGIL_SMOKE_TOPUP_PAYMENT_INTENT_ID", "OWLVIGIL_SMOKE_TOPUP_CLIENT_SECRET"},
			func(values []string) error {
				_, _, err := r.management.ConfirmTopupInApp(r.ctx, &management.ConfirmTopupInAppRequest{
					PaymentIntentID: values[0],
					ClientSecret:    values[1],
				})
				return err
			},
		)
	}

	var stripeSubscriptionID string
	subscriptionReturnURL := strings.TrimSpace(os.Getenv("OWLVIGIL_SMOKE_RETURN_URL"))
	if savedPaymentMethodID == "" || planID == "" || planInterval == "" || subscriptionReturnURL == "" {
		r.skip("create subscription in-app", "POST /v1/billing/subscription/in-app", "requires a saved Stripe test payment method and a saleable plan")
		r.skip("confirm subscription in-app", "POST /v1/billing/subscription/in-app/confirm", "subscription payment was not created")
	} else {
		r.write("create subscription in-app", "POST /v1/billing/subscription/in-app", "subscription write endpoint not exercised in smoke", func() error {
			payment, _, err := r.management.CreateSubscriptionInApp(r.ctx, &management.CreateSubscriptionInAppRequest{
				PlanID:    planID,
				Interval:  planInterval,
				ReturnURL: subscriptionReturnURL,
			})
			if err == nil {
				stripeSubscriptionID = payment.StripeSubscriptionID
				if stripeSubscriptionID == "" {
					return errors.New("Stripe subscription ID is empty")
				}
			}
			return err
		})
		if stripeSubscriptionID == "" {
			r.skip("confirm subscription in-app", "POST /v1/billing/subscription/in-app/confirm", "subscription payment was not created")
		} else {
			r.call("confirm subscription in-app", "POST /v1/billing/subscription/in-app/confirm", func() error {
				_, _, err := r.management.ConfirmSubscriptionInApp(r.ctx, &management.ConfirmSubscriptionInAppRequest{
					PlanID:               planID,
					StripeSubscriptionID: stripeSubscriptionID,
				})
				return err
			})
		}
	}
	r.configuredWrite(
		"upgrade subscription",
		"POST /v1/billing/subscription/upgrade",
		[]string{"OWLVIGIL_SMOKE_UPGRADE_PLAN_ID", "OWLVIGIL_SMOKE_UPGRADE_INTERVAL"},
		func(values []string) error {
			_, _, err := r.management.UpgradeSubscription(r.ctx, &management.UpgradeSubscriptionRequest{
				PlanID:   values[0],
				Interval: values[1],
			})
			return err
		},
	)
	r.configuredWrite(
		"downgrade subscription",
		"POST /v1/billing/subscription/downgrade",
		[]string{"OWLVIGIL_SMOKE_DOWNGRADE_PLAN_ID", "OWLVIGIL_SMOKE_DOWNGRADE_INTERVAL"},
		func(values []string) error {
			_, _, err := r.management.DowngradeSubscription(r.ctx, &management.DowngradeSubscriptionRequest{
				PlanID:   values[0],
				Interval: values[1],
			})
			return err
		},
	)
	subscriptionCanceled := false
	if stripeSubscriptionID == "" {
		r.skip("cancel subscription", "POST /v1/billing/subscription/cancel", "subscription payment was not created")
	} else {
		r.call("cancel subscription", "POST /v1/billing/subscription/cancel", func() error {
			_, _, err := r.management.CancelSubscriptionWithRequest(r.ctx, &management.CancelSubscriptionRequest{
				CancelImmediately: false,
			})
			subscriptionCanceled = err == nil
			return err
		})
	}
	if !subscriptionCanceled {
		r.skip("reactivate subscription", "POST /v1/billing/subscription/reactivate", "subscription was not canceled by this smoke run")
	} else {
		r.call("reactivate subscription", "POST /v1/billing/subscription/reactivate", func() error {
			_, _, err := r.management.ReactivateSubscription(r.ctx)
			return err
		})
	}
	if subscriptionCheckoutSessionID == "" {
		r.skip("get subscription checkout session", "GET /v1/billing/subscription/checkout-sessions/:session_id", "subscription checkout session was not created")
		r.skip("sync latest subscription checkout", "POST /v1/billing/subscription/checkout-sessions/sync-latest", "subscription checkout session was not created")
	} else {
		r.call("get subscription checkout session", "GET /v1/billing/subscription/checkout-sessions/:session_id", func() error {
			_, _, err := r.management.GetSubscriptionCheckoutSession(r.ctx, subscriptionCheckoutSessionID)
			return err
		})
		r.call("sync latest subscription checkout", "POST /v1/billing/subscription/checkout-sessions/sync-latest", func() error {
			_, _, err := r.management.SyncLatestSubscriptionCheckout(r.ctx)
			return err
		})
	}
	if savedPaymentMethodID == "" {
		r.skip("restore default payment method", "PUT /v1/billing/payment-methods/:payment_method_id/default", "Stripe test payment method was not saved")
		r.skip("delete Stripe test payment method", "DELETE /v1/billing/payment-methods/:payment_method_id", "Stripe test payment method was not saved")
	} else {
		if originalDefaultPaymentMethodID != "" && originalDefaultPaymentMethodID != savedPaymentMethodID {
			r.cleanup("restore default payment method", "PUT /v1/billing/payment-methods/:payment_method_id/default", func(ctx context.Context) error {
				_, _, err := r.management.SetDefaultPaymentMethod(ctx, originalDefaultPaymentMethodID)
				return err
			})
		} else {
			r.skip("restore default payment method", "PUT /v1/billing/payment-methods/:payment_method_id/default", "no previous default payment method exists")
		}
		r.cleanup("delete Stripe test payment method", "DELETE /v1/billing/payment-methods/:payment_method_id", func(ctx context.Context) error {
			_, err := r.management.DeletePaymentMethod(ctx, savedPaymentMethodID)
			return err
		})
	}
}

func (r *runner) runFinancial() {
	r.call("get financial governance", "GET /v1/workspaces/:workspace_id/governance/financial", func() error {
		_, _, err := r.management.GetFinancialGovernance(r.ctx, r.workspaceID)
		return err
	})
	r.call("get budget caps", "GET /v1/workspaces/:workspace_id/governance/financial/budget-caps", func() error {
		_, _, err := r.management.GetBudgetCaps(r.ctx, r.workspaceID)
		return err
	})
	r.call("get spending limits", "GET /v1/workspaces/:workspace_id/governance/financial/spending-limits", func() error {
		_, _, err := r.management.GetSpendingLimits(r.ctx, r.workspaceID, management.ListOptions{})
		return err
	})
	r.call("get financial thresholds", "GET /v1/workspaces/:workspace_id/governance/financial/thresholds", func() error {
		_, _, err := r.management.GetFinancialThresholds(r.ctx, r.workspaceID)
		return err
	})
	r.call("preview financial changes", "POST /v1/workspaces/:workspace_id/governance/financial/preview", func() error {
		_, _, err := r.management.PreviewFinancialChanges(r.ctx, r.workspaceID, &management.PreviewFinancialChangesRequest{})
		return err
	})
	r.call("get spend summary", "GET /v1/workspaces/:workspace_id/governance/financial/spend-summary", func() error {
		_, _, err := r.management.GetSpendSummary(r.ctx, r.workspaceID)
		return err
	})
	r.write("update financial governance", "PUT /v1/workspaces/:workspace_id/governance/financial", "financial write endpoint not exercised in smoke", func() error {
		governance, _, err := r.management.GetFinancialGovernance(r.ctx, r.workspaceID)
		if err != nil {
			return err
		}
		_, _, err = r.management.UpdateFinancialGovernance(r.ctx, r.workspaceID, &management.UpdateFinancialGovernanceRequest{
			WorkspaceCap:   &governance.WorkspaceCap,
			TeamCaps:       governance.TeamCaps,
			MemberCaps:     governance.MemberCaps,
			GatewayKeyCaps: governance.GatewayKeyCaps,
			MemberLimits:   governance.MemberLimits,
			Thresholds:     governance.Thresholds,
			ExceededAction: &governance.ExceededAction,
		})
		return err
	})
	r.write("update budget caps", "PUT /v1/workspaces/:workspace_id/governance/financial/budget-caps", "financial write endpoint not exercised in smoke", func() error {
		caps, _, err := r.management.GetBudgetCaps(r.ctx, r.workspaceID)
		if err != nil {
			return err
		}
		_, _, err = r.management.UpdateBudgetCaps(r.ctx, r.workspaceID, &management.UpdateBudgetCapsRequest{
			WorkspaceCap: &caps.WorkspaceCap, TeamCaps: caps.TeamCaps, MemberCaps: caps.MemberCaps, GatewayKeyCaps: caps.GatewayKeyCaps,
		})
		return err
	})
	if !r.writes {
		r.skip("update scope budget cap", "PATCH /v1/workspaces/:workspace_id/governance/financial/budget-caps/:scope_type/:scope_id", "financial write endpoint not exercised in smoke")
	} else if caps, _, err := r.management.GetBudgetCaps(r.ctx, r.workspaceID); err != nil {
		r.recordErr("update scope budget cap", "PATCH /v1/workspaces/:workspace_id/governance/financial/budget-caps/:scope_type/:scope_id", err)
	} else if caps.Workspace == nil || caps.Workspace.ScopeType == "" || caps.Workspace.ScopeID == nil {
		r.configuredWrite(
			"update scope budget cap",
			"PATCH /v1/workspaces/:workspace_id/governance/financial/budget-caps/:scope_type/:scope_id",
			[]string{
				"OWLVIGIL_SMOKE_BUDGET_SCOPE_TYPE",
				"OWLVIGIL_SMOKE_BUDGET_SCOPE_ID",
				"OWLVIGIL_SMOKE_BUDGET_ENABLED",
				"OWLVIGIL_SMOKE_BUDGET_MONTHLY_AMOUNT",
			},
			func(values []string) error {
				enabled, parseErr := strconv.ParseBool(values[2])
				if parseErr != nil {
					return errors.New("OWLVIGIL_SMOKE_BUDGET_ENABLED must be true or false")
				}
				monthlyAmount, parseErr := positiveAmount(values[3])
				if parseErr != nil {
					return fmt.Errorf("OWLVIGIL_SMOKE_BUDGET_MONTHLY_AMOUNT: %w", parseErr)
				}
				_, _, err := r.management.UpdateScopeBudgetCap(
					r.ctx,
					r.workspaceID,
					values[0],
					values[1],
					&management.UpdateScopeBudgetCapRequest{
						Enabled:       &enabled,
						MonthlyAmount: &monthlyAmount,
					},
				)
				return err
			},
		)
	} else {
		r.write("update scope budget cap", "PATCH /v1/workspaces/:workspace_id/governance/financial/budget-caps/:scope_type/:scope_id", "financial write endpoint not exercised in smoke", func() error {
			_, _, err := r.management.UpdateScopeBudgetCap(
				r.ctx,
				r.workspaceID,
				caps.Workspace.ScopeType,
				strconv.FormatInt(*caps.Workspace.ScopeID, 10),
				&management.UpdateScopeBudgetCapRequest{
					Enabled:       &caps.Workspace.Enabled,
					MonthlyAmount: &caps.Workspace.MonthlyAmount,
				},
			)
			return err
		})
	}
	r.write("update spending limits", "PUT /v1/workspaces/:workspace_id/governance/financial/spending-limits", "financial write endpoint not exercised in smoke", func() error {
		limits, _, err := r.management.GetSpendingLimits(r.ctx, r.workspaceID, management.ListOptions{})
		if err != nil {
			return err
		}
		_, _, err = r.management.UpdateSpendingLimits(r.ctx, r.workspaceID, &management.UpdateSpendingLimitsRequest{Limits: limits.Items})
		return err
	})
	if !r.writes {
		r.skip("update user spending limit", "PATCH /v1/workspaces/:workspace_id/governance/financial/spending-limits/users/:user_id", "financial write endpoint not exercised in smoke")
	} else if limits, _, err := r.management.GetSpendingLimits(r.ctx, r.workspaceID, management.ListOptions{}); err != nil {
		r.recordErr("update user spending limit", "PATCH /v1/workspaces/:workspace_id/governance/financial/spending-limits/users/:user_id", err)
	} else if len(limits.Items) == 0 {
		r.configuredWrite(
			"update user spending limit",
			"PATCH /v1/workspaces/:workspace_id/governance/financial/spending-limits/users/:user_id",
			[]string{
				"OWLVIGIL_SMOKE_SPENDING_USER_ID",
				"OWLVIGIL_SMOKE_SPENDING_DAILY_LIMIT",
				"OWLVIGIL_SMOKE_SPENDING_WEEKLY_LIMIT",
				"OWLVIGIL_SMOKE_SPENDING_MONTHLY_LIMIT",
			},
			func(values []string) error {
				userID, parseErr := strconv.ParseInt(values[0], 10, 64)
				if parseErr != nil || userID <= 0 {
					return errors.New("OWLVIGIL_SMOKE_SPENDING_USER_ID must be a positive integer")
				}
				dailyLimit, parseErr := positiveAmount(values[1])
				if parseErr != nil {
					return fmt.Errorf("OWLVIGIL_SMOKE_SPENDING_DAILY_LIMIT: %w", parseErr)
				}
				weeklyLimit, parseErr := positiveAmount(values[2])
				if parseErr != nil {
					return fmt.Errorf("OWLVIGIL_SMOKE_SPENDING_WEEKLY_LIMIT: %w", parseErr)
				}
				monthlyLimit, parseErr := positiveAmount(values[3])
				if parseErr != nil {
					return fmt.Errorf("OWLVIGIL_SMOKE_SPENDING_MONTHLY_LIMIT: %w", parseErr)
				}
				_, _, err := r.management.UpdateUserSpendingLimit(
					r.ctx,
					r.workspaceID,
					userID,
					&management.UpdateUserSpendingLimitRequest{
						DailyLimit:   &dailyLimit,
						WeeklyLimit:  &weeklyLimit,
						MonthlyLimit: &monthlyLimit,
					},
				)
				return err
			},
		)
	} else {
		limit := limits.Items[0]
		r.write("update user spending limit", "PATCH /v1/workspaces/:workspace_id/governance/financial/spending-limits/users/:user_id", "financial write endpoint not exercised in smoke", func() error {
			_, _, err := r.management.UpdateUserSpendingLimit(r.ctx, r.workspaceID, limit.UserID, &management.UpdateUserSpendingLimitRequest{
				DailyLimit: &limit.DailyLimit, WeeklyLimit: &limit.WeeklyLimit, MonthlyLimit: &limit.MonthlyLimit,
			})
			return err
		})
	}
	r.write("update financial thresholds", "PUT /v1/workspaces/:workspace_id/governance/financial/thresholds", "financial write endpoint not exercised in smoke", func() error {
		thresholds, _, err := r.management.GetFinancialThresholds(r.ctx, r.workspaceID)
		if err != nil {
			return err
		}
		_, _, err = r.management.UpdateFinancialThresholds(r.ctx, r.workspaceID, &management.UpdateThresholdsRequest{
			WarningPercent: &thresholds.WarningPercent, CriticalPercent: &thresholds.CriticalPercent, ExceededAction: &thresholds.ExceededAction,
		})
		return err
	})
}

func (r *runner) runPolicies() {
	workspaceOpt := owlvigil.WithWorkspaceID(r.workspaceID)
	r.call("get gateway policies", "GET /v1/gateway/policies", func() error {
		_, _, err := r.management.GetGatewayPolicies(r.ctx, 0, workspaceOpt)
		return err
	})
	r.call("preview gateway policy effect", "POST /v1/gateway/policies/preview", func() error {
		_, _, err := r.management.PreviewPolicyEffect(r.ctx, &management.PreviewPolicyRequest{WorkspaceID: r.workspaceID, Model: "gpt-4"})
		return err
	})
	var keywordID int64
	keyword := smokeName("SDK Prompt Keyword")
	r.write("add prompt keyword", "POST /v1/gateway/policies/keywords", "prompt keyword write endpoint not exercised in smoke", func() error {
		policies, _, err := r.management.AddPromptKeyword(r.ctx, &management.AddPromptKeywordRequest{
			WorkspaceID: r.workspaceID,
			Keyword:     keyword,
			Description: "temporary SDK smoke prompt restriction",
		})
		if err != nil {
			return err
		}
		for _, policy := range policies.KeywordPolicies {
			if policy.Keyword == keyword {
				keywordID = int64(policy.ID)
				break
			}
		}
		if keywordID == 0 {
			return fmt.Errorf("add prompt keyword response did not contain the created keyword")
		}
		return nil
	})
	if keywordID > 0 {
		r.cleanup("delete prompt keyword", "DELETE /v1/gateway/policies/keywords/:keyword_id", func(ctx context.Context) error {
			_, _, err := r.management.DeletePromptKeyword(ctx, r.workspaceID, keywordID)
			return err
		})
	} else {
		r.skip("delete prompt keyword", "DELETE /v1/gateway/policies/keywords/:keyword_id", "prompt keyword was not created")
	}
	r.configuredWrite(
		"update gateway policy",
		"PATCH /v1/gateway/policies/:policy_id",
		[]string{"OWLVIGIL_SMOKE_POLICY_ID"},
		func(values []string) error {
			policyID, err := strconv.ParseInt(values[0], 10, 64)
			if err != nil || policyID <= 0 {
				return fmt.Errorf("OWLVIGIL_SMOKE_POLICY_ID must be a positive integer")
			}
			_, _, err = r.management.UpdateGatewayPolicy(
				r.ctx,
				policyID,
				&management.UpdateGatewayPolicyRequest{},
				workspaceOpt,
			)
			return err
		},
	)
}

func (r *runner) runWebhooks() {
	workspaceOpt := owlvigil.WithWorkspaceID(r.workspaceID)
	var endpointID int64
	r.call("list webhook endpoints", "GET /v1/webhook-endpoints", func() error {
		_, _, err := r.management.ListWebhookEndpoints(r.ctx, management.ListOptions{Limit: 5}, workspaceOpt)
		return err
	})
	r.configuredWrite(
		"create webhook endpoint",
		"POST /v1/webhook-endpoints",
		[]string{"OWLVIGIL_SMOKE_WEBHOOK_URL"},
		func(values []string) error {
			endpoint, _, err := r.management.CreateWebhookEndpoint(r.ctx, &management.CreateWebhookEndpointRequest{
				WorkspaceID: r.workspaceID,
				URL:         values[0],
				EventTypes:  []string{"gateway.key.updated"},
			}, owlvigil.WithIdempotencyKey(smokeName("sdk-smoke-webhook")))
			if err != nil {
				return err
			}
			endpointID = endpoint.ID
			return nil
		},
	)
	if endpointID > 0 {
		r.call("get webhook endpoint", "GET /v1/webhook-endpoints/:endpoint_id", func() error {
			_, _, err := r.management.GetWebhookEndpoint(r.ctx, endpointID, workspaceOpt)
			return err
		})
		disabled := "disabled"
		r.write("update webhook endpoint", "PATCH /v1/webhook-endpoints/:endpoint_id", "webhook write endpoint not exercised in smoke", func() error {
			_, _, err := r.management.UpdateWebhookEndpoint(r.ctx, endpointID, &management.UpdateWebhookEndpointRequest{
				EventTypes: []string{"gateway.key.updated"},
				Status:     &disabled,
			}, workspaceOpt)
			return err
		})
		r.write("enable webhook endpoint", "POST /v1/webhook-endpoints/:endpoint_id/enable", "webhook write endpoint not exercised in smoke", func() error {
			_, err := r.management.EnableWebhookEndpoint(r.ctx, endpointID, workspaceOpt)
			return err
		})
		r.write("disable webhook endpoint", "POST /v1/webhook-endpoints/:endpoint_id/disable", "webhook write endpoint not exercised in smoke", func() error {
			_, err := r.management.DisableWebhookEndpoint(r.ctx, endpointID, workspaceOpt)
			return err
		})
		r.write("rotate webhook secret", "POST /v1/webhook-endpoints/:endpoint_id/rotate-secret", "webhook write endpoint not exercised in smoke", func() error {
			_, _, err := r.management.RotateWebhookSecret(r.ctx, endpointID, workspaceOpt)
			return err
		})
		r.write("test webhook endpoint", "POST /v1/webhook-endpoints/:endpoint_id/test", "webhook write endpoint not exercised in smoke", func() error {
			_, err := r.management.TestWebhookEndpoint(r.ctx, endpointID, workspaceOpt)
			return err
		})
		r.call("list endpoint webhook events", "GET /v1/webhook-endpoints/:endpoint_id/events", func() error {
			_, _, err := r.management.ListEndpointEvents(r.ctx, endpointID, management.ListOptions{Limit: 5}, workspaceOpt)
			return err
		})
	} else {
		r.skip("get webhook endpoint", "GET /v1/webhook-endpoints/:endpoint_id", "webhook endpoint was not created")
		r.skip("update webhook endpoint", "PATCH /v1/webhook-endpoints/:endpoint_id", "webhook endpoint was not created")
		r.skip("enable webhook endpoint", "POST /v1/webhook-endpoints/:endpoint_id/enable", "webhook endpoint was not created")
		r.skip("disable webhook endpoint", "POST /v1/webhook-endpoints/:endpoint_id/disable", "webhook endpoint was not created")
		r.skip("rotate webhook secret", "POST /v1/webhook-endpoints/:endpoint_id/rotate-secret", "webhook endpoint was not created")
		r.skip("test webhook endpoint", "POST /v1/webhook-endpoints/:endpoint_id/test", "webhook endpoint was not created")
		r.skip("list endpoint webhook events", "GET /v1/webhook-endpoints/:endpoint_id/events", "webhook endpoint was not created")
	}
	r.call("list webhook event types", "GET /v1/webhook-event-types", func() error {
		_, _, err := r.management.ListWebhookEventTypes(r.ctx)
		return err
	})
	var eventID string
	var eventIDInt int
	r.call("list webhook events", "GET /v1/webhook-events", func() error {
		events, _, err := r.management.ListWebhookEvents(r.ctx, management.ListOptions{Limit: 5}, workspaceOpt)
		if err != nil {
			return err
		}
		if len(events.Items) > 0 {
			eventID = events.Items[0].ID
			eventIDInt, _ = strconv.Atoi(eventID)
		}
		if eventID == "" {
			eventID = strings.TrimSpace(os.Getenv("OWLVIGIL_SMOKE_WEBHOOK_EVENT_ID"))
			eventIDInt, _ = strconv.Atoi(eventID)
		}
		return nil
	})
	if eventID != "" {
		r.call("get webhook event", "GET /v1/webhook-events/:event_id", func() error {
			_, _, err := r.management.GetWebhookEvent(r.ctx, eventID, workspaceOpt)
			return err
		})
		r.write("retry webhook event", "POST /v1/webhook-events/:event_id/retry", "webhook write endpoint not exercised in smoke", func() error {
			_, err := r.management.RetryWebhookEvent(r.ctx, eventID, workspaceOpt)
			return err
		})
		r.write("redeliver webhook event", "POST /v1/webhook-events/:event_id/redeliver", "webhook write endpoint not exercised in smoke", func() error {
			_, err := r.management.RedeliverWebhookEvent(r.ctx, eventID, workspaceOpt)
			return err
		})
		r.write("bulk redeliver webhook events", "POST /v1/webhook-events/bulk-redeliver", "webhook write endpoint not exercised in smoke", func() error {
			_, err := r.management.BulkRedeliverWebhookEvents(r.ctx, &management.BulkRedeliverRequest{
				WorkspaceID: r.workspaceID,
				EndpointID:  &endpointID,
				EventIDs:    []int{eventIDInt},
				Limit:       10,
			})
			return err
		})
	} else {
		r.skip("get webhook event", "GET /v1/webhook-events/:event_id", "webhook events list returned no events")
		r.skip("retry webhook event", "POST /v1/webhook-events/:event_id/retry", "webhook events list returned no events")
		r.skip("redeliver webhook event", "POST /v1/webhook-events/:event_id/redeliver", "webhook events list returned no events")
		r.skip("bulk redeliver webhook events", "POST /v1/webhook-events/bulk-redeliver", "webhook events list returned no events")
	}
	if endpointID > 0 {
		r.cleanup("delete webhook endpoint", "DELETE /v1/webhook-endpoints/:endpoint_id", func(ctx context.Context) error {
			_, err := r.management.DeleteWebhookEndpoint(ctx, endpointID, workspaceOpt)
			return err
		})
	} else {
		r.skip("delete webhook endpoint", "DELETE /v1/webhook-endpoints/:endpoint_id", "webhook endpoint was not created")
	}
}

func (r *runner) call(name, contract string, fn func() error) {
	r.recordErr(name, contract, fn())
}

func (r *runner) callSkipKnown(name, contract string, known []string, fn func() error) {
	err := fn()
	if err != nil && containsAny(err.Error(), known) {
		r.skip(name, contract, err.Error())
		return
	}
	r.recordErr(name, contract, err)
}

func (r *runner) write(name, contract, reason string, fn func() error) {
	if !r.writes {
		r.skip(name, contract, reason)
		return
	}
	r.call(name, contract, fn)
}

func (r *runner) configuredWrite(name, contract string, environmentNames []string, fn func([]string) error) {
	if !r.writes {
		r.skip(name, contract, "write endpoint not exercised unless OWLVIGIL_SMOKE_WRITES=1")
		return
	}
	r.configuredCall(name, contract, environmentNames, fn)
}

func (r *runner) configuredCall(name, contract string, environmentNames []string, fn func([]string) error) {
	values := make([]string, len(environmentNames))
	var missing []string
	for index, environmentName := range environmentNames {
		values[index] = os.Getenv(environmentName)
		if values[index] == "" {
			missing = append(missing, environmentName)
		}
	}
	if len(missing) > 0 {
		r.skip(name, contract, "set "+strings.Join(missing, ", "))
		return
	}
	r.call(name, contract, func() error { return fn(values) })
}

func (r *runner) writeSkipKnown(name, contract, reason string, known []string, fn func() error) {
	if !r.writes {
		r.skip(name, contract, reason)
		return
	}
	r.callSkipKnown(name, contract, known, fn)
}

func (r *runner) writeControlled(name, contract, reason string, allowedStatuses []int, fn func() error) {
	if !r.writes {
		r.skip(name, contract, reason)
		return
	}
	err := fn()
	if err == nil {
		r.pass(name, contract)
		return
	}
	var apiErr *owlvigil.APIError
	if errors.As(err, &apiErr) {
		for _, status := range allowedStatuses {
			if apiErr.StatusCode == status {
				r.skip(name, contract, err.Error())
				return
			}
		}
	}
	r.recordErr(name, contract, err)
}

func (r *runner) cleanup(name, contract string, fn func(context.Context) error) {
	ctx, cancel := r.cleanupContext()
	defer cancel()
	r.recordErr(name, contract, fn(ctx))
}

func (r *runner) cleanupContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.WithoutCancel(r.ctx), cleanupTimeout)
}

func (r *runner) recordErr(name, contract string, err error) {
	if err != nil {
		r.steps = append(r.steps, step{Name: name, Contract: contract, Status: "FAIL", Error: err.Error()})
		return
	}
	r.pass(name, contract)
}

func (r *runner) pass(name, contract string) {
	r.steps = append(r.steps, step{Name: name, Contract: contract, Status: "PASS"})
}

func (r *runner) skip(name, contract, reason string) {
	r.steps = append(r.steps, step{Name: name, Contract: contract, Status: "SKIP", Error: reason})
}

func (r *runner) failedCount() int {
	count := 0
	for _, step := range r.steps {
		if step.Status == "FAIL" {
			count++
		}
	}
	return count
}

func (r *runner) skippedCount() int {
	count := 0
	for _, step := range r.steps {
		if step.Status == "SKIP" {
			count++
		}
	}
	return count
}

func (r *runner) complete() bool {
	if r.failedCount() > 0 {
		return false
	}
	if !r.requireAll {
		return true
	}
	coverage := r.managementCoverage()
	if len(coverage) != managementOperationCount {
		return false
	}
	for _, passed := range coverage {
		if !passed {
			return false
		}
	}
	return true
}

func (r *runner) managementCoverage() map[string]bool {
	coverage := make(map[string]bool, managementOperationCount)
	for _, step := range r.steps {
		if !isManagementContract(step.Contract) {
			continue
		}
		if _, exists := coverage[step.Contract]; !exists {
			coverage[step.Contract] = false
		}
		if step.Status == "PASS" {
			coverage[step.Contract] = true
		}
	}
	return coverage
}

func isManagementContract(contract string) bool {
	_, path, ok := strings.Cut(contract, " ")
	return ok && strings.HasPrefix(path, "/v1/")
}

func (r *runner) print() {
	passed, skipped, failed := 0, 0, 0
	for _, step := range r.steps {
		switch step.Status {
		case "PASS":
			passed++
			fmt.Printf("PASS %-38s %q\n", step.Name, step.Contract)
		case "SKIP":
			skipped++
			fmt.Printf("SKIP %-38s %q reason=%q\n", step.Name, step.Contract, step.Error)
		case "FAIL":
			failed++
			fmt.Printf("FAIL %-38s %q error=%q\n", step.Name, step.Contract, step.Error)
		}
	}
	managementPassed := 0
	for _, contractPassed := range r.managementCoverage() {
		if contractPassed {
			managementPassed++
		}
	}
	fmt.Printf("workspace_id=%d user_id=%d\n", r.workspaceID, r.userID)
	fmt.Printf(
		"sdk_openapi_smoke management_expected=%d management_passed=%d steps=%d passed=%d skipped=%d failed=%d require_all=%t complete=%t\n",
		managementOperationCount,
		managementPassed,
		len(r.steps),
		passed,
		skipped,
		failed,
		r.requireAll,
		r.complete(),
	)
}

func envDefault(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func positiveAmount(raw string) (float64, error) {
	amount, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
	if err != nil || math.IsNaN(amount) || math.IsInf(amount, 0) || amount <= 0 {
		return 0, fmt.Errorf("amount must be a positive finite number")
	}
	return amount, nil
}

func oauthEnabled(accessToken, clientID, clientSecret string) bool {
	return strings.TrimSpace(accessToken) != "" || (strings.TrimSpace(clientID) != "" && strings.TrimSpace(clientSecret) != "")
}

func writeSmokeEnabled(value string) bool {
	return strings.TrimSpace(value) == "1"
}

func stripePaymentMethodID(value string) string {
	return strings.TrimSpace(value)
}

func maskToken(token string) string {
	token = strings.TrimSpace(token)
	if token == "" {
		return ""
	}
	if len(token) <= 16 {
		return token[:4] + "..."
	}
	return token[:8] + "..." + token[len(token)-6:]
}

func smokeName(prefix string) string {
	return fmt.Sprintf("%s %d", prefix, time.Now().UnixNano())
}

func containsAny(value string, needles []string) bool {
	value = strings.ToLower(value)
	for _, needle := range needles {
		if strings.Contains(value, strings.ToLower(needle)) {
			return true
		}
	}
	return false
}
