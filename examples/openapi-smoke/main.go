package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	owlvigil "github.com/owlvigil/owlvigil-go"
	"github.com/owlvigil/owlvigil-go/examples/internal/envfile"
	"github.com/owlvigil/owlvigil-go/management"
	oauth2 "github.com/owlvigil/owlvigil-go/oauth2"
)

const defaultScope = "workspace:read workspace:write gateway:read gateway:write usage:read billing:read billing:write webhook:read webhook:write profile:read profile:write support:write notifications:read notifications:write invites:read invites:write audit_logs:read teams:read teams:write members:read members:write rbac:read rbac:write financial:read financial:write policies:read policies:write"

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
	}
	r.runAll()
	r.print()
	if r.failedCount() > 0 {
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
	r.skip("change user password", "PUT /v1/user/password", "user write endpoint not exercised in smoke")
	r.skip("submit support request", "POST /v1/user/support-requests", "user write endpoint not exercised in smoke")
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
	r.skip("send invitation", "POST /v1/users/me/send-invitation", "user write endpoint not exercised in smoke")

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
	r.call("get workspace", "GET /v1/workspaces/:workspace_id", func() error {
		_, _, err := r.management.GetWorkspace(r.ctx, r.workspaceID)
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
	r.call("get workspace quota summary", "GET /v1/workspaces/:workspace_id/quota-summary", func() error {
		_, _, err := r.management.GetQuotaSummary(r.ctx, r.workspaceID)
		return err
	})
	r.call("get workspace quota usage", "GET /v1/workspaces/:workspace_id/quota-usage", func() error {
		_, _, err := r.management.GetQuotaUsage(r.ctx, r.workspaceID)
		return err
	})
	r.call("list workspace teams", "GET /v1/workspaces/:workspace_id/teams", func() error {
		teams, _, err := r.management.ListTeams(r.ctx, r.workspaceID, management.ListOptions{Limit: 5})
		if err != nil {
			return err
		}
		if len(teams.Items) > 0 {
			_, _, err = r.management.GetTeam(r.ctx, r.workspaceID, teams.Items[0].ID)
			r.recordErr("get workspace team", "GET /v1/workspaces/:workspace_id/teams/:team_id", err)
		} else {
			r.skip("get workspace team", "GET /v1/workspaces/:workspace_id/teams/:team_id", "team list returned no teams")
		}
		return nil
	})
	var createdTeamID int64
	r.write("create workspace team", "POST /v1/workspaces/:workspace_id/teams", "workspace write endpoint not exercised in smoke", func() error {
		team, _, err := r.management.CreateTeam(r.ctx, r.workspaceID, &management.CreateTeamRequest{Name: smokeName("SDK Smoke Team")})
		if err == nil {
			createdTeamID = team.ID
		}
		return err
	})
	if createdTeamID > 0 {
		updatedName := smokeName("SDK Smoke Team Updated")
		r.call("update workspace team", "PATCH /v1/workspaces/:workspace_id/teams/:team_id", func() error {
			_, _, err := r.management.UpdateTeam(r.ctx, r.workspaceID, createdTeamID, &management.UpdateTeamRequest{Name: &updatedName})
			return err
		})
		r.call("delete workspace team", "DELETE /v1/workspaces/:workspace_id/teams/:team_id", func() error {
			_, err := r.management.DeleteTeam(r.ctx, r.workspaceID, createdTeamID)
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
	r.skip("add workspace member", "POST /v1/workspaces/:workspace_id/members", "workspace write endpoint not exercised in smoke")
	if !r.writes {
		r.skip("update workspace member", "PATCH /v1/workspaces/:workspace_id/members/:member_id", "workspace write endpoint not exercised in smoke")
	} else {
		members, _, err := r.management.ListMembers(r.ctx, r.workspaceID, management.ListOptions{Limit: 100})
		if err != nil {
			r.recordErr("update workspace member", "PATCH /v1/workspaces/:workspace_id/members/:member_id", err)
		} else {
			var candidate *management.Member
			for i := range members.Items {
				if members.Items[i].UserID != r.userID {
					candidate = &members.Items[i]
					break
				}
			}
			if candidate == nil {
				r.skip("update workspace member", "PATCH /v1/workspaces/:workspace_id/members/:member_id", "no non-owner member is available for a safe update")
			} else {
				r.call("update workspace member", "PATCH /v1/workspaces/:workspace_id/members/:member_id", func() error {
					_, _, err := r.management.UpdateMember(r.ctx, r.workspaceID, candidate.UserID, &management.UpdateMemberRequest{
						RoleIDs: candidate.RoleIDs,
						TeamIDs: candidate.TeamIDs,
						Status:  &candidate.Status,
					})
					return err
				})
			}
		}
	}
	r.skip("remove workspace member", "DELETE /v1/workspaces/:workspace_id/members/:member_id", "workspace write endpoint not exercised in smoke")

	r.call("list workspace invitations", "GET /v1/workspaces/:workspace_id/invitations", func() error {
		_, _, err := r.management.ListInvitations(r.ctx, r.workspaceID, management.ListOptions{Limit: 5})
		return err
	})
	r.skip("invite workspace member", "POST /v1/workspaces/:workspace_id/invitations", "workspace write endpoint not exercised in smoke")
	r.skip("resend workspace invitation", "POST /v1/workspaces/:workspace_id/invitations/:invitation_id/resend", "workspace write endpoint not exercised in smoke")
	r.skip("revoke workspace invitation", "POST /v1/workspaces/:workspace_id/invitations/:invitation_id/revoke", "workspace write endpoint not exercised in smoke")

	r.callSkipKnown("list workspace roles", "GET /v1/workspaces/:workspace_id/roles", []string{"feature.rbac is not included"}, func() error {
		roles, _, err := r.management.ListRoles(r.ctx, r.workspaceID, management.ListOptions{Limit: 5})
		if err != nil {
			return err
		}
		if len(roles.Items) > 0 {
			_, _, err = r.management.GetRole(r.ctx, r.workspaceID, roles.Items[0].ID)
			r.recordErr("get workspace role", "GET /v1/workspaces/:workspace_id/roles/:role_id", err)
		} else {
			r.skip("get workspace role", "GET /v1/workspaces/:workspace_id/roles/:role_id", "roles list returned no roles")
		}
		return nil
	})
	var createdRoleID int64
	r.write("create workspace role", "POST /v1/workspaces/:workspace_id/roles", "workspace write endpoint not exercised in smoke", func() error {
		role, _, err := r.management.CreateRole(r.ctx, r.workspaceID, &management.CreateRoleRequest{Name: smokeName("SDK Smoke Role")})
		if err == nil {
			createdRoleID = role.ID
		}
		return err
	})
	if createdRoleID > 0 {
		updatedName := smokeName("SDK Smoke Role Updated")
		r.call("update workspace role", "PATCH /v1/workspaces/:workspace_id/roles/:role_id", func() error {
			_, _, err := r.management.UpdateRole(r.ctx, r.workspaceID, createdRoleID, &management.UpdateRoleRequest{Name: &updatedName})
			return err
		})
		r.call("delete workspace role", "DELETE /v1/workspaces/:workspace_id/roles/:role_id", func() error {
			_, err := r.management.DeleteRole(r.ctx, r.workspaceID, createdRoleID)
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
	r.skip("update workspace member permissions", "PUT /v1/workspaces/:workspace_id/members/:member_id/permissions", "workspace write endpoint not exercised in smoke")
	r.skip("reset workspace member permissions", "POST /v1/workspaces/:workspace_id/members/:member_id/permissions/reset", "workspace write endpoint not exercised in smoke")
}

func (r *runner) runGateway() {
	workspaceOpt := owlvigil.WithQueryParam("workspace_id", strconv.FormatInt(r.workspaceID, 10))
	r.call("list gateway keys", "GET /v1/gateway/keys", func() error {
		_, _, err := r.management.ListGatewayKeys(r.ctx, management.ListOptions{Limit: 5}, "", workspaceOpt)
		return err
	})
	var keyID int64
	r.callSkipKnown("create gateway key", "POST /v1/gateway/keys", []string{"quota.gateway_keys limit exceeded"}, func() error {
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
		r.call("update gateway key", "PATCH /v1/gateway/keys/:key_id", func() error {
			_, _, err := r.management.UpdateGatewayKey(r.ctx, keyID, &management.UpdateGatewayKeyRequest{Name: &updatedName}, workspaceOpt)
			return err
		})
		r.call("rotate gateway key", "POST /v1/gateway/keys/:key_id/rotate", func() error {
			_, _, err := r.management.RotateGatewayKey(r.ctx, keyID, workspaceOpt)
			return err
		})
		r.call("disable gateway key", "POST /v1/gateway/keys/:key_id/disable", func() error {
			_, err := r.management.DisableGatewayKey(r.ctx, keyID, workspaceOpt)
			return err
		})
		r.call("enable gateway key", "POST /v1/gateway/keys/:key_id/enable", func() error {
			_, err := r.management.EnableGatewayKey(r.ctx, keyID, workspaceOpt)
			return err
		})
		r.call("delete gateway key", "DELETE /v1/gateway/keys/:key_id", func() error {
			_, err := r.management.DeleteGatewayKey(r.ctx, keyID, workspaceOpt)
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
		return nil
	})
	if modelID != "" {
		r.call("get model", "GET /v1/gateway/models/:model_id", func() error {
			_, _, err := r.management.GetModel(r.ctx, modelID, workspaceOpt)
			return err
		})
		r.call("preview route", "POST /v1/gateway/routes/preview", func() error {
			_, _, err := r.management.PreviewRoute(r.ctx, &management.PreviewRouteRequest{WorkspaceID: r.workspaceID, Model: modelID}, workspaceOpt)
			return err
		})
	} else {
		r.skip("get model", "GET /v1/gateway/models/:model_id", "models list returned no models")
		r.skip("preview route", "POST /v1/gateway/routes/preview", "models list returned no models")
	}
	r.call("list routes", "GET /v1/gateway/routes", func() error {
		_, _, err := r.management.ListRoutes(r.ctx, management.ListOptions{Limit: 5}, workspaceOpt)
		return err
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
			r.skip("get request log", "GET /v1/gateway/request-logs/:request_id", "request logs list returned no logs")
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
			r.skip("get trace", "GET /v1/gateway/traces/:trace_id", "traces list returned no traces")
		}
		return nil
	})
	r.call("payload log access", "GET /v1/gateway/payload-logs/access", func() error {
		_, _, err := r.management.GetPayloadAccess(r.ctx, workspaceOpt)
		return err
	})
	r.skip("get payload log", "GET /v1/gateway/payload-logs/:payload_id", "requires payload logging and an existing payload id")
}

func (r *runner) runBilling() {
	workspaceOpt := owlvigil.WithQueryParam("workspace_id", strconv.FormatInt(r.workspaceID, 10))
	stripePaymentMethodID := stripePaymentMethodID(os.Getenv("OWLVIGIL_STRIPE_TEST_PAYMENT_METHOD_ID"))
	r.call("billing overview", "GET /v1/billing/overview", func() error {
		_, _, err := r.management.GetBillingOverview(r.ctx, workspaceOpt)
		return err
	})
	r.call("billing balance", "GET /v1/billing/balance", func() error {
		_, _, err := r.management.GetBalance(r.ctx, workspaceOpt)
		return err
	})
	var planID, planInterval string
	r.call("billing plans", "GET /v1/billing/plans", func() error {
		plans, _, err := r.management.ListPlans(r.ctx, management.ListOptions{Limit: 5}, workspaceOpt)
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
		return nil
	})
	if planID != "" {
		r.call("billing plan", "GET /v1/billing/plans/:plan_id", func() error {
			_, _, err := r.management.GetPlan(r.ctx, planID, workspaceOpt)
			return err
		})
	} else {
		r.skip("billing plan", "GET /v1/billing/plans/:plan_id", "plans list returned no plans")
	}
	r.call("billing subscription", "GET /v1/billing/subscription", func() error {
		_, _, err := r.management.GetSubscription(r.ctx, workspaceOpt)
		return err
	})
	r.call("billing topup plans", "GET /v1/billing/topup-plans", func() error {
		_, _, err := r.management.ListTopupPlans(r.ctx, management.ListOptions{Limit: 5}, workspaceOpt)
		return err
	})
	r.call("billing topups", "GET /v1/billing/topups", func() error {
		topups, _, err := r.management.ListTopups(r.ctx, management.ListOptions{Limit: 5}, workspaceOpt)
		if err != nil {
			return err
		}
		if len(topups.Items) > 0 && topups.Items[0].ID != "" {
			_, _, err = r.management.GetTopup(r.ctx, topups.Items[0].ID, workspaceOpt)
			r.recordErr("billing topup", "GET /v1/billing/topups/:topup_id", err)
		} else {
			r.skip("billing topup", "GET /v1/billing/topups/:topup_id", "topups list returned no topups")
		}
		return nil
	})
	r.call("billing invoices", "GET /v1/billing/invoices", func() error {
		invoices, _, err := r.management.ListInvoices(r.ctx, management.ListOptions{Limit: 5}, workspaceOpt)
		if err != nil {
			return err
		}
		if len(invoices.Items) > 0 && invoices.Items[0].ID != "" {
			_, _, err = r.management.GetInvoice(r.ctx, invoices.Items[0].ID, workspaceOpt)
			r.recordErr("billing invoice", "GET /v1/billing/invoices/:invoice_id", err)
		} else {
			r.skip("billing invoice", "GET /v1/billing/invoices/:invoice_id", "invoices list returned no invoices")
		}
		return nil
	})
	r.call("billing orders", "GET /v1/billing/orders", func() error {
		orders, _, err := r.management.ListOrders(r.ctx, management.ListOptions{Limit: 5}, "", workspaceOpt)
		if err != nil {
			return err
		}
		if len(orders.Items) > 0 && orders.Items[0].ID != "" {
			_, _, err = r.management.GetOrder(r.ctx, orders.Items[0].ID, workspaceOpt)
			r.recordErr("billing order", "GET /v1/billing/orders/:order_id", err)
		} else {
			r.skip("billing order", "GET /v1/billing/orders/:order_id", "orders list returned no orders")
		}
		return nil
	})
	r.call("billing details", "GET /v1/workspaces/:workspace_id/billing-details", func() error {
		_, _, err := r.management.GetBillingDetails(r.ctx, r.workspaceID)
		return err
	})
	var originalDefaultPaymentMethodID string
	r.call("list payment methods", "GET /v1/billing/payment-methods", func() error {
		methods, _, err := r.management.ListPaymentMethods(r.ctx, management.ListOptions{Limit: 5}, workspaceOpt)
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
	r.skip("update workspace billing details", "PUT /v1/workspaces/:workspace_id/billing-details", "billing write endpoint not exercised in smoke")
	r.skip("billing checkout", "POST /v1/billing/topups/checkout", "payment write endpoint not exercised in smoke")
	r.skip("billing in-app", "POST /v1/billing/topups/in-app", "payment write endpoint not exercised in smoke")
	r.skip("billing confirm stripe session", "POST /v1/billing/orders/:order_id/confirm-stripe-session", "payment write endpoint not exercised in smoke")
	r.skip("create subscription checkout", "POST /v1/billing/subscription/checkout", "subscription write endpoint not exercised in smoke")
	r.write("create payment method setup intent", "POST /v1/billing/payment-methods/setup-intent", "billing write endpoint not exercised in smoke", func() error {
		setupIntent, _, err := r.management.CreatePaymentMethodSetupIntentForWorkspace(r.ctx, r.workspaceID, workspaceOpt)
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
			}, workspaceOpt)
			if err == nil {
				savedPaymentMethodID = method.ID
			}
			return err
		})
		if savedPaymentMethodID == "" {
			r.skip("set Stripe test payment method default", "PUT /v1/billing/payment-methods/:payment_method_id/default", "Stripe test payment method was not saved")
		} else {
			r.call("set Stripe test payment method default", "PUT /v1/billing/payment-methods/:payment_method_id/default", func() error {
				_, _, err := r.management.SetDefaultPaymentMethod(r.ctx, savedPaymentMethodID, workspaceOpt)
				return err
			})
		}
	}

	var paymentIntentID string
	if savedPaymentMethodID == "" || planID == "" || planInterval == "" {
		r.skip("create subscription in-app", "POST /v1/billing/subscription/in-app", "requires a saved Stripe test payment method and a saleable plan")
		r.skip("confirm subscription in-app", "POST /v1/billing/subscription/in-app/confirm", "subscription payment was not created")
	} else {
		r.write("create subscription in-app", "POST /v1/billing/subscription/in-app", "subscription write endpoint not exercised in smoke", func() error {
			payment, _, err := r.management.CreateSubscriptionInApp(r.ctx, &management.CreateSubscriptionInAppRequest{PlanID: planID, Interval: planInterval}, workspaceOpt)
			if err == nil {
				paymentIntentID = payment.PaymentIntentID
				if paymentIntentID == "" {
					return errors.New("subscription payment intent is empty")
				}
			}
			return err
		})
		if paymentIntentID == "" {
			r.skip("confirm subscription in-app", "POST /v1/billing/subscription/in-app/confirm", "subscription payment was not created")
		} else {
			r.call("confirm subscription in-app", "POST /v1/billing/subscription/in-app/confirm", func() error {
				_, _, err := r.management.ConfirmSubscriptionInApp(r.ctx, &management.ConfirmSubscriptionInAppRequest{PaymentIntentID: paymentIntentID}, workspaceOpt)
				return err
			})
		}
	}
	r.skip("upgrade subscription", "POST /v1/billing/subscription/upgrade", "subscription write endpoint not exercised in smoke")
	r.skip("downgrade subscription", "POST /v1/billing/subscription/downgrade", "subscription write endpoint not exercised in smoke")
	if paymentIntentID == "" {
		r.skip("cancel subscription", "POST /v1/billing/subscription/cancel", "subscription payment was not created")
	} else {
		r.call("cancel subscription", "POST /v1/billing/subscription/cancel", func() error {
			_, _, err := r.management.CancelSubscription(r.ctx, workspaceOpt)
			return err
		})
	}
	r.skip("reactivate subscription", "POST /v1/billing/subscription/reactivate", "subscription write endpoint not exercised in smoke")
	r.skip("get subscription checkout session", "GET /v1/billing/subscription/checkout-sessions/:session_id", "requires an existing checkout session")
	r.skip("sync latest subscription checkout", "POST /v1/billing/subscription/checkout-sessions/sync-latest", "subscription write endpoint not exercised in smoke")
	if savedPaymentMethodID == "" {
		r.skip("restore default payment method", "PUT /v1/billing/payment-methods/:payment_method_id/default", "Stripe test payment method was not saved")
		r.skip("delete Stripe test payment method", "DELETE /v1/billing/payment-methods/:payment_method_id", "Stripe test payment method was not saved")
	} else {
		if originalDefaultPaymentMethodID != "" && originalDefaultPaymentMethodID != savedPaymentMethodID {
			r.call("restore default payment method", "PUT /v1/billing/payment-methods/:payment_method_id/default", func() error {
				_, _, err := r.management.SetDefaultPaymentMethod(r.ctx, originalDefaultPaymentMethodID, workspaceOpt)
				return err
			})
		} else {
			r.skip("restore default payment method", "PUT /v1/billing/payment-methods/:payment_method_id/default", "no previous default payment method exists")
		}
		r.call("delete Stripe test payment method", "DELETE /v1/billing/payment-methods/:payment_method_id", func() error {
			_, err := r.management.DeletePaymentMethod(r.ctx, savedPaymentMethodID, workspaceOpt)
			return err
		})
	}
	r.skip("confirm topup in-app", "POST /v1/billing/topups/in-app/confirm", "payment write endpoint not exercised in smoke")
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
		_, _, err := r.management.GetSpendingLimits(r.ctx, r.workspaceID, management.ListOptions{Limit: 5})
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
			BudgetCaps:     governance.BudgetCaps,
			SpendingLimits: governance.SpendingLimits,
			Thresholds:     governance.Thresholds,
		})
		return err
	})
	r.write("update budget caps", "PUT /v1/workspaces/:workspace_id/governance/financial/budget-caps", "financial write endpoint not exercised in smoke", func() error {
		caps, _, err := r.management.GetBudgetCaps(r.ctx, r.workspaceID)
		if err != nil {
			return err
		}
		_, _, err = r.management.UpdateBudgetCaps(r.ctx, r.workspaceID, &management.UpdateBudgetCapsRequest{
			Workspace: caps.Workspace, Teams: caps.Teams, Members: caps.Members, GatewayKeys: caps.GatewayKeys,
		})
		return err
	})
	if !r.writes {
		r.skip("update scope budget cap", "PATCH /v1/workspaces/:workspace_id/governance/financial/budget-caps/:scope_type/:scope_id", "financial write endpoint not exercised in smoke")
	} else if caps, _, err := r.management.GetBudgetCaps(r.ctx, r.workspaceID); err != nil {
		r.recordErr("update scope budget cap", "PATCH /v1/workspaces/:workspace_id/governance/financial/budget-caps/:scope_type/:scope_id", err)
	} else if caps.Workspace == nil || caps.Workspace.ScopeType == "" || caps.Workspace.ScopeID == "" {
		r.skip("update scope budget cap", "PATCH /v1/workspaces/:workspace_id/governance/financial/budget-caps/:scope_type/:scope_id", "workspace budget cap is not configured")
	} else {
		r.call("update scope budget cap", "PATCH /v1/workspaces/:workspace_id/governance/financial/budget-caps/:scope_type/:scope_id", func() error {
			_, _, err := r.management.UpdateScopeBudgetCap(r.ctx, r.workspaceID, caps.Workspace.ScopeType, caps.Workspace.ScopeID, &management.UpdateScopeBudgetCapRequest{Limit: caps.Workspace.Limit})
			return err
		})
	}
	r.write("update spending limits", "PUT /v1/workspaces/:workspace_id/governance/financial/spending-limits", "financial write endpoint not exercised in smoke", func() error {
		limits, _, err := r.management.GetSpendingLimits(r.ctx, r.workspaceID, management.ListOptions{Limit: 100})
		if err != nil {
			return err
		}
		_, _, err = r.management.UpdateSpendingLimits(r.ctx, r.workspaceID, &management.UpdateSpendingLimitsRequest{Limits: limits.Items})
		return err
	})
	if !r.writes {
		r.skip("update user spending limit", "PATCH /v1/workspaces/:workspace_id/governance/financial/spending-limits/users/:user_id", "financial write endpoint not exercised in smoke")
	} else if limits, _, err := r.management.GetSpendingLimits(r.ctx, r.workspaceID, management.ListOptions{Limit: 1}); err != nil {
		r.recordErr("update user spending limit", "PATCH /v1/workspaces/:workspace_id/governance/financial/spending-limits/users/:user_id", err)
	} else if len(limits.Items) == 0 {
		r.skip("update user spending limit", "PATCH /v1/workspaces/:workspace_id/governance/financial/spending-limits/users/:user_id", "no user spending limit is configured")
	} else {
		limit := limits.Items[0]
		r.call("update user spending limit", "PATCH /v1/workspaces/:workspace_id/governance/financial/spending-limits/users/:user_id", func() error {
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
	workspaceOpt := owlvigil.WithQueryParam("workspace_id", strconv.FormatInt(r.workspaceID, 10))
	r.call("get gateway policies", "GET /v1/gateway/policies", func() error {
		_, _, err := r.management.GetGatewayPolicies(r.ctx, 0, workspaceOpt)
		return err
	})
	r.call("preview gateway policy effect", "POST /v1/gateway/policies/preview", func() error {
		_, _, err := r.management.PreviewPolicyEffect(r.ctx, &management.PreviewPolicyRequest{WorkspaceID: r.workspaceID, Model: "gpt-4"}, workspaceOpt)
		return err
	})
	r.skip("update gateway policy", "PATCH /v1/gateway/policies/:policy_id", "policy write endpoint not exercised in smoke")
}

func (r *runner) runWebhooks() {
	workspaceOpt := owlvigil.WithQueryParam("workspace_id", strconv.FormatInt(r.workspaceID, 10))
	var endpointID int64
	r.call("list webhook endpoints", "GET /v1/webhook-endpoints", func() error {
		_, _, err := r.management.ListWebhookEndpoints(r.ctx, management.ListOptions{Limit: 5}, workspaceOpt)
		return err
	})
	r.call("create webhook endpoint", "POST /v1/webhook-endpoints", func() error {
		endpoint, _, err := r.management.CreateWebhookEndpoint(r.ctx, &management.CreateWebhookEndpointRequest{
			WorkspaceID: r.workspaceID,
			URL:         "https://example.com/owlvigil/sdk-openapi-smoke",
			EventTypes:  []string{"webhook.test", "gateway.request.completed"},
		}, owlvigil.WithIdempotencyKey(smokeName("sdk-smoke-webhook")))
		if err != nil {
			return err
		}
		endpointID = endpoint.ID
		return nil
	})
	if endpointID > 0 {
		r.call("get webhook endpoint", "GET /v1/webhook-endpoints/:endpoint_id", func() error {
			_, _, err := r.management.GetWebhookEndpoint(r.ctx, endpointID, workspaceOpt)
			return err
		})
		disabled := "disabled"
		r.call("update webhook endpoint", "PATCH /v1/webhook-endpoints/:endpoint_id", func() error {
			_, _, err := r.management.UpdateWebhookEndpoint(r.ctx, endpointID, &management.UpdateWebhookEndpointRequest{
				EventTypes: []string{"webhook.test"},
				Status:     &disabled,
			}, workspaceOpt)
			return err
		})
		r.call("enable webhook endpoint", "POST /v1/webhook-endpoints/:endpoint_id/enable", func() error {
			_, err := r.management.EnableWebhookEndpoint(r.ctx, endpointID, workspaceOpt)
			return err
		})
		r.call("disable webhook endpoint", "POST /v1/webhook-endpoints/:endpoint_id/disable", func() error {
			_, err := r.management.DisableWebhookEndpoint(r.ctx, endpointID, workspaceOpt)
			return err
		})
		r.call("rotate webhook secret", "POST /v1/webhook-endpoints/:endpoint_id/rotate-secret", func() error {
			_, _, err := r.management.RotateWebhookSecret(r.ctx, endpointID, workspaceOpt)
			return err
		})
		r.call("test webhook endpoint", "POST /v1/webhook-endpoints/:endpoint_id/test", func() error {
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
		return nil
	})
	if eventID != "" {
		r.call("get webhook event", "GET /v1/webhook-events/:event_id", func() error {
			_, _, err := r.management.GetWebhookEvent(r.ctx, eventID, workspaceOpt)
			return err
		})
		r.call("retry webhook event", "POST /v1/webhook-events/:event_id/retry", func() error {
			_, err := r.management.RetryWebhookEvent(r.ctx, eventID, workspaceOpt)
			return err
		})
		r.call("redeliver webhook event", "POST /v1/webhook-events/:event_id/redeliver", func() error {
			_, err := r.management.RedeliverWebhookEvent(r.ctx, eventID, workspaceOpt)
			return err
		})
		r.call("bulk redeliver webhook events", "POST /v1/webhook-events/bulk-redeliver", func() error {
			_, err := r.management.BulkRedeliverWebhookEvents(r.ctx, &management.BulkRedeliverRequest{
				WorkspaceID: r.workspaceID,
				EndpointID:  &endpointID,
				EventIDs:    []int{eventIDInt},
				Limit:       10,
			}, workspaceOpt)
			return err
		})
	} else {
		r.skip("get webhook event", "GET /v1/webhook-events/:event_id", "webhook events list returned no events")
		r.skip("retry webhook event", "POST /v1/webhook-events/:event_id/retry", "webhook events list returned no events")
		r.skip("redeliver webhook event", "POST /v1/webhook-events/:event_id/redeliver", "webhook events list returned no events")
		r.skip("bulk redeliver webhook events", "POST /v1/webhook-events/bulk-redeliver", "webhook events list returned no events")
	}
	if endpointID > 0 {
		r.call("delete webhook endpoint", "DELETE /v1/webhook-endpoints/:endpoint_id", func() error {
			_, err := r.management.DeleteWebhookEndpoint(r.ctx, endpointID, workspaceOpt)
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
	fmt.Printf("workspace_id=%d user_id=%d\n", r.workspaceID, r.userID)
	fmt.Printf("sdk_openapi_smoke expected=%d passed=%d skipped=%d failed=%d passed=%t\n", len(r.steps), passed, skipped, failed, failed == 0)
}

func envDefault(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
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
