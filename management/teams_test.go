package management_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	owlvigil "github.com/owlvigil/owlvigil-go"
	"github.com/owlvigil/owlvigil-go/management"
)

func TestMembersEndpoints(t *testing.T) {
	t.Parallel()

	expected := map[string]string{
		"GET /workspaces/1/members":              `{"items":[{"id":1,"user_id":10,"email":"user@example.com","name":"John Doe","status":"active"}],"page_info":{}}`,
		"GET /workspaces/1/members/role-options": `{"items":[{"id":1,"name":"Admin","is_system":true}],"page_info":{}}`,
		"POST /workspaces/1/members":             `{"id":2,"user_id":20,"email":"new@example.com","status":"pending"}`,
		"GET /workspaces/1/members/1":            `{"id":1,"user_id":10,"email":"user@example.com","status":"active"}`,
		"PATCH /workspaces/1/members/1":          `{"id":1,"user_id":10,"email":"user@example.com","status":"active"}`,
		"DELETE /workspaces/1/members/1":         `{}`,
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

	// Test ListMembers
	members, _, err := client.ListMembers(ctx, 1, management.ListOptions{})
	if err != nil {
		t.Fatalf("ListMembers failed: %v", err)
	}
	if len(members.Items) != 1 || members.Items[0].Email != "user@example.com" {
		t.Fatalf("members = %+v", members)
	}

	// Test ListRoleOptions
	roles, _, err := client.ListRoleOptions(ctx, 1)
	if err != nil {
		t.Fatalf("ListRoleOptions failed: %v", err)
	}
	if len(roles.Items) != 1 || roles.Items[0].Name != "Admin" {
		t.Fatalf("roles = %+v", roles)
	}

	// Test CreateMember
	member, _, err := client.CreateMember(ctx, 1, &management.CreateMemberRequest{
		Email:   "new@example.com",
		RoleIDs: []int64{1},
	})
	if err != nil {
		t.Fatalf("CreateMember failed: %v", err)
	}
	if member.Email != "new@example.com" {
		t.Fatalf("member = %+v", member)
	}

	// Test GetMember
	_, _, err = client.GetMember(ctx, 1, 1)
	if err != nil {
		t.Fatalf("GetMember failed: %v", err)
	}

	// Test UpdateMember
	_, _, err = client.UpdateMember(ctx, 1, 1, &management.UpdateMemberRequest{
		RoleIDs: []int64{1, 2},
	})
	if err != nil {
		t.Fatalf("UpdateMember failed: %v", err)
	}

	// Test DeleteMember
	_, err = client.DeleteMember(ctx, 1, 1)
	if err != nil {
		t.Fatalf("DeleteMember failed: %v", err)
	}
}

func TestTeamsEndpoints(t *testing.T) {
	t.Parallel()

	expected := map[string]string{
		"GET /workspaces/1/teams":      `{"items":[{"id":1,"name":"Engineering","status":"active","member_count":5}],"page_info":{}}`,
		"POST /workspaces/1/teams":     `{"id":2,"name":"Sales","status":"active","member_count":0}`,
		"GET /workspaces/1/teams/1":    `{"id":1,"name":"Engineering","status":"active","member_count":5}`,
		"PATCH /workspaces/1/teams/1":  `{"id":1,"name":"Engineering Team","status":"active","member_count":5}`,
		"DELETE /workspaces/1/teams/1": `{}`,
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

	// Test ListTeams
	teams, _, err := client.ListTeams(ctx, 1, management.ListOptions{})
	if err != nil {
		t.Fatalf("ListTeams failed: %v", err)
	}
	if len(teams.Items) != 1 || teams.Items[0].Name != "Engineering" {
		t.Fatalf("teams = %+v", teams)
	}

	// Test CreateTeam
	team, _, err := client.CreateTeam(ctx, 1, &management.CreateTeamRequest{
		Name:        "Sales",
		Description: "Sales team",
	})
	if err != nil {
		t.Fatalf("CreateTeam failed: %v", err)
	}
	if team.Name != "Sales" {
		t.Fatalf("team = %+v", team)
	}

	// Test GetTeam
	_, _, err = client.GetTeam(ctx, 1, 1)
	if err != nil {
		t.Fatalf("GetTeam failed: %v", err)
	}

	// Test UpdateTeam
	name := "Engineering Team"
	_, _, err = client.UpdateTeam(ctx, 1, 1, &management.UpdateTeamRequest{
		Name: &name,
	})
	if err != nil {
		t.Fatalf("UpdateTeam failed: %v", err)
	}

	// Test DeleteTeam
	_, err = client.DeleteTeam(ctx, 1, 1)
	if err != nil {
		t.Fatalf("DeleteTeam failed: %v", err)
	}
}

func TestInvitationsEndpoints(t *testing.T) {
	t.Parallel()

	expected := map[string]string{
		"GET /workspaces/1/invitations":           `{"items":[{"id":1,"email":"invite@example.com","status":"pending"}],"page_info":{}}`,
		"POST /workspaces/1/invitations":          `{"id":2,"email":"new@example.com","status":"pending","invite_link":"https://app.example.com/invite/abc"}`,
		"POST /workspaces/1/invitations/1/resend": `{}`,
		"POST /workspaces/1/invitations/1/revoke": `{}`,
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

	// Test ListInvitations
	invitations, _, err := client.ListInvitations(ctx, 1, management.ListOptions{})
	if err != nil {
		t.Fatalf("ListInvitations failed: %v", err)
	}
	if len(invitations.Items) != 1 || invitations.Items[0].Email != "invite@example.com" {
		t.Fatalf("invitations = %+v", invitations)
	}

	// Test CreateInvitation
	invitation, _, err := client.CreateInvitation(ctx, 1, &management.CreateInvitationRequest{
		Email:   "new@example.com",
		RoleIDs: []int64{1},
	})
	if err != nil {
		t.Fatalf("CreateInvitation failed: %v", err)
	}
	if invitation.Email != "new@example.com" {
		t.Fatalf("invitation = %+v", invitation)
	}

	// Test ResendInvitation
	_, err = client.ResendInvitation(ctx, 1, 1)
	if err != nil {
		t.Fatalf("ResendInvitation failed: %v", err)
	}

	// Test RevokeInvitation
	_, err = client.RevokeInvitation(ctx, 1, 1)
	if err != nil {
		t.Fatalf("RevokeInvitation failed: %v", err)
	}
}
