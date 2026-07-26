package management

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"

	owlvigil "github.com/Syrovex/owlvigil_sdk_go"
)

// MemberListOptions filters and paginates workspace members.
type MemberListOptions struct {
	Limit  int
	Offset int
	Search string
	Role   string
	Status string
	Team   string
}

func (o MemberListOptions) values() url.Values {
	q := url.Values{}
	if o.Limit > 0 {
		q.Set("limit", intString(o.Limit))
	}
	if o.Offset > 0 {
		q.Set("offset", intString(o.Offset))
	}
	addFilter(q, "search", o.Search)
	addFilter(q, "role", o.Role)
	addFilter(q, "status", o.Status)
	addFilter(q, "team", o.Team)
	return q
}

// Member describes a workspace member.
type Member struct {
	ID                  int64      `json:"id"`
	UserID              int64      `json:"user_id"`
	WorkspaceID         int64      `json:"workspace_id"`
	Email               string     `json:"email"`
	Name                string     `json:"name,omitempty"`
	Role                string     `json:"role,omitempty"`
	Status              string     `json:"status"`
	Team                string     `json:"team,omitempty"`
	TeamID              *int64     `json:"team_id,omitempty"`
	InvitedBy           string     `json:"invited_by,omitempty"`
	JoinedAt            string     `json:"joined_at,omitempty"`
	User                *UserBrief `json:"user,omitempty"`
	MemberType          string     `json:"member_type,omitempty"`
	InvitationID        *int64     `json:"invitation_id,omitempty"`
	InvitationToken     string     `json:"invitation_token,omitempty"`
	InvitationExpiresAt *string    `json:"invitation_expires_at,omitempty"`

	// Deprecated compatibility fields from the pre-refactor API.
	RoleIDs      []int64 `json:"role_ids,omitempty"`
	TeamIDs      []int64 `json:"team_ids,omitempty"`
	LastActiveAt string  `json:"last_active_at,omitempty"`
}

// UserBrief is a compact member user reference.
type UserBrief struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// RoleOption describes an available role.
type RoleOption struct {
	ID          int64    `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	IsSystem    bool     `json:"is_system"`
	Permissions []string `json:"permissions,omitempty"`
	Value       string   `json:"value,omitempty"`
	Label       string   `json:"label,omitempty"`
	Enabled     bool     `json:"enabled,omitempty"`
}

// CreateMemberRequest invites or adds a member.
type CreateMemberRequest struct {
	Email  string `json:"-"`
	Role   string `json:"-"`
	TeamID *int64 `json:"-"`

	// Deprecated compatibility fields from the pre-refactor API.
	RoleIDs   []int64 `json:"-"`
	TeamIDs   []int64 `json:"-"`
	ExpiresAt string  `json:"-"`
}

// MarshalJSON emits the current member invitation contract.
func (r CreateMemberRequest) MarshalJSON() ([]byte, error) {
	role := r.Role
	if role == "" && len(r.RoleIDs) > 0 {
		role = "member"
	}
	teamID := r.TeamID
	if teamID == nil && len(r.TeamIDs) > 0 {
		teamID = &r.TeamIDs[0]
	}
	return json.Marshal(struct {
		Email  string `json:"email"`
		Role   string `json:"role"`
		TeamID *int64 `json:"team_id,omitempty"`
	}{
		Email:  r.Email,
		Role:   role,
		TeamID: teamID,
	})
}

// UpdateMemberRequest updates member information.
type UpdateMemberRequest struct {
	Role   string  `json:"-"`
	Status *string `json:"-"`
	TeamID *int64  `json:"-"`

	// Deprecated compatibility fields from the pre-refactor API.
	RoleIDs []int64 `json:"-"`
	TeamIDs []int64 `json:"-"`
}

// MarshalJSON emits the current mutable member contract.
func (r UpdateMemberRequest) MarshalJSON() ([]byte, error) {
	role := r.Role
	if role == "" && len(r.RoleIDs) > 0 {
		role = "member"
	}
	teamID := r.TeamID
	if teamID == nil && len(r.TeamIDs) > 0 {
		teamID = &r.TeamIDs[0]
	}
	status := ""
	if r.Status != nil {
		status = *r.Status
	}
	return json.Marshal(struct {
		Role   string `json:"role,omitempty"`
		Status string `json:"status,omitempty"`
		TeamID *int64 `json:"team_id,omitempty"`
	}{
		Role:   role,
		Status: status,
		TeamID: teamID,
	})
}

// ListMembers lists workspace members.
func (c *Client) ListMembers(ctx context.Context, workspaceID int64, opts ListOptions, reqOpts ...owlvigil.RequestOption) (*ListResponse[Member], *owlvigil.ResponseMeta, error) {
	return c.ListMembersWithFilters(
		ctx,
		workspaceID,
		MemberListOptions{Limit: opts.Limit},
		reqOpts...,
	)
}

// ListMembersWithFilters lists workspace members using every filter published
// by the current Open API.
func (c *Client) ListMembersWithFilters(ctx context.Context, workspaceID int64, opts MemberListOptions, reqOpts ...owlvigil.RequestOption) (*ListResponse[Member], *owlvigil.ResponseMeta, error) {
	var out ListResponse[Member]
	path := "/workspaces/" + strconv.FormatInt(workspaceID, 10) + "/members"
	meta, err := c.http.Do(ctx, http.MethodGet, path, opts.values(), nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// ListRoleOptions lists available roles for assignment.
func (c *Client) ListRoleOptions(ctx context.Context, workspaceID int64, reqOpts ...owlvigil.RequestOption) (*ListResponse[RoleOption], *owlvigil.ResponseMeta, error) {
	var out ListResponse[RoleOption]
	path := "/workspaces/" + strconv.FormatInt(workspaceID, 10) + "/members/role-options"
	meta, err := c.http.Do(ctx, http.MethodGet, path, nil, nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// CreateMember invites or adds a member to the workspace.
func (c *Client) CreateMember(ctx context.Context, workspaceID int64, req *CreateMemberRequest, reqOpts ...owlvigil.RequestOption) (*Member, *owlvigil.ResponseMeta, error) {
	var out Member
	path := "/workspaces/" + strconv.FormatInt(workspaceID, 10) + "/members"
	meta, err := c.http.Do(ctx, http.MethodPost, path, nil, req, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// GetMember retrieves member details by user ID.
func (c *Client) GetMember(ctx context.Context, workspaceID, userID int64, reqOpts ...owlvigil.RequestOption) (*Member, *owlvigil.ResponseMeta, error) {
	var out Member
	path := "/workspaces/" + strconv.FormatInt(workspaceID, 10) + "/members/" + strconv.FormatInt(userID, 10)
	meta, err := c.http.Do(ctx, http.MethodGet, path, nil, nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// UpdateMember updates member roles, teams, or status by user ID.
func (c *Client) UpdateMember(ctx context.Context, workspaceID, userID int64, req *UpdateMemberRequest, reqOpts ...owlvigil.RequestOption) (*Member, *owlvigil.ResponseMeta, error) {
	var out Member
	path := "/workspaces/" + strconv.FormatInt(workspaceID, 10) + "/members/" + strconv.FormatInt(userID, 10)
	meta, err := c.http.Do(ctx, http.MethodPatch, path, nil, req, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// DeleteMember removes a member from the workspace by user ID.
func (c *Client) DeleteMember(ctx context.Context, workspaceID, userID int64, reqOpts ...owlvigil.RequestOption) (*owlvigil.ResponseMeta, error) {
	_, meta, err := c.DeleteMemberWithResult(ctx, workspaceID, userID, reqOpts...)
	return meta, err
}

// DeleteMemberWithResult removes a member and returns confirmation.
func (c *Client) DeleteMemberWithResult(ctx context.Context, workspaceID, userID int64, reqOpts ...owlvigil.RequestOption) (*DeleteResponse, *owlvigil.ResponseMeta, error) {
	path := "/workspaces/" + strconv.FormatInt(workspaceID, 10) + "/members/" + strconv.FormatInt(userID, 10)
	var out DeleteResponse
	meta, err := c.http.Do(ctx, http.MethodDelete, path, nil, nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}
