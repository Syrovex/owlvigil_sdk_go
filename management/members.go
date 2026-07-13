package management

import (
	"context"
	"net/http"
	"strconv"

	owlvigil "github.com/owlvigil/owlvigil-go"
)

// Member describes a workspace member.
type Member struct {
	ID           int64   `json:"id"`
	UserID       int64   `json:"user_id"`
	WorkspaceID  int64   `json:"workspace_id"`
	Email        string  `json:"email"`
	Name         string  `json:"name,omitempty"`
	Status       string  `json:"status"`
	RoleIDs      []int64 `json:"role_ids,omitempty"`
	TeamIDs      []int64 `json:"team_ids,omitempty"`
	JoinedAt     string  `json:"joined_at,omitempty"`
	LastActiveAt string  `json:"last_active_at,omitempty"`
}

// RoleOption describes an available role.
type RoleOption struct {
	ID          int64    `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	IsSystem    bool     `json:"is_system"`
	Permissions []string `json:"permissions,omitempty"`
}

// CreateMemberRequest invites or adds a member.
type CreateMemberRequest struct {
	Email     string  `json:"email"`
	RoleIDs   []int64 `json:"role_ids,omitempty"`
	TeamIDs   []int64 `json:"team_ids,omitempty"`
	ExpiresAt string  `json:"expires_at,omitempty"`
}

// UpdateMemberRequest updates member information.
type UpdateMemberRequest struct {
	RoleIDs []int64 `json:"role_ids,omitempty"`
	TeamIDs []int64 `json:"team_ids,omitempty"`
	Status  *string `json:"status,omitempty"`
}

// ListMembers lists workspace members.
func (c *Client) ListMembers(ctx context.Context, workspaceID int64, opts ListOptions, reqOpts ...owlvigil.RequestOption) (*ListResponse[Member], *owlvigil.ResponseMeta, error) {
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
	path := "/workspaces/" + strconv.FormatInt(workspaceID, 10) + "/members/" + strconv.FormatInt(userID, 10)
	return c.http.Do(ctx, http.MethodDelete, path, nil, nil, nil, reqOpts...)
}
