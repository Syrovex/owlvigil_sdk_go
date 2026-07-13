package management

import (
	"context"
	"net/http"
	"strconv"

	owlvigil "github.com/owlvigil/owlvigil-go"
)

// Role describes a workspace role.
type Role struct {
	ID          int64    `json:"id"`
	WorkspaceID int64    `json:"workspace_id"`
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	IsSystem    bool     `json:"is_system"`
	Permissions []string `json:"permissions,omitempty"`
	CreatedAt   string   `json:"created_at,omitempty"`
}

// Permission describes a permission.
type Permission struct {
	ID          string `json:"id,omitempty"`
	Key         string `json:"key,omitempty"`
	Name        string `json:"name,omitempty"`
	Label       string `json:"label,omitempty"`
	Description string `json:"description,omitempty"`
	Group       string `json:"group,omitempty"`
	ScopeType   string `json:"scope_type,omitempty"`
	Default     bool   `json:"default,omitempty"`
	Effective   bool   `json:"effective,omitempty"`
}

// PermissionGroup describes a grouped set of permissions.
type PermissionGroup struct {
	ID          string       `json:"id,omitempty"`
	Name        string       `json:"name,omitempty"`
	Label       string       `json:"label,omitempty"`
	Permissions []Permission `json:"permissions,omitempty"`
}

// MemberPermissions describes a member's effective permissions.
type MemberPermissions struct {
	MemberID             int64             `json:"member_id,omitempty"`
	UserID               int64             `json:"user_id,omitempty"`
	WorkspaceID          int64             `json:"workspace_id,omitempty"`
	Role                 string            `json:"role,omitempty"`
	Effective            []string          `json:"effective,omitempty"`
	RolePermissions      []string          `json:"role_permissions,omitempty"`
	OverridePermissions  []string          `json:"override_permissions,omitempty"`
	EffectivePermissions []string          `json:"effective_permissions,omitempty"`
	Groups               []PermissionGroup `json:"groups,omitempty"`
}

// CreateRoleRequest creates a custom role.
type CreateRoleRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
}

// UpdateRoleRequest updates a role.
type UpdateRoleRequest struct {
	Name        *string  `json:"name,omitempty"`
	Description *string  `json:"description,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
}

// UpdateMemberPermissionsRequest updates member permissions.
type UpdateMemberPermissionsRequest struct {
	Permissions []string `json:"permissions"`
}

// ListRoles lists workspace roles.
func (c *Client) ListRoles(ctx context.Context, workspaceID int64, opts ListOptions, reqOpts ...owlvigil.RequestOption) (*ListResponse[Role], *owlvigil.ResponseMeta, error) {
	var out ListResponse[Role]
	path := "/workspaces/" + strconv.FormatInt(workspaceID, 10) + "/roles"
	meta, err := c.http.Do(ctx, http.MethodGet, path, opts.values(), nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// CreateRole creates a custom role.
func (c *Client) CreateRole(ctx context.Context, workspaceID int64, req *CreateRoleRequest, reqOpts ...owlvigil.RequestOption) (*Role, *owlvigil.ResponseMeta, error) {
	var out Role
	path := "/workspaces/" + strconv.FormatInt(workspaceID, 10) + "/roles"
	meta, err := c.http.Do(ctx, http.MethodPost, path, nil, req, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// GetRole retrieves role details by ID.
func (c *Client) GetRole(ctx context.Context, workspaceID, roleID int64, reqOpts ...owlvigil.RequestOption) (*Role, *owlvigil.ResponseMeta, error) {
	var out Role
	path := "/workspaces/" + strconv.FormatInt(workspaceID, 10) + "/roles/" + strconv.FormatInt(roleID, 10)
	meta, err := c.http.Do(ctx, http.MethodGet, path, nil, nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// UpdateRole updates a custom role.
func (c *Client) UpdateRole(ctx context.Context, workspaceID, roleID int64, req *UpdateRoleRequest, reqOpts ...owlvigil.RequestOption) (*Role, *owlvigil.ResponseMeta, error) {
	var out Role
	path := "/workspaces/" + strconv.FormatInt(workspaceID, 10) + "/roles/" + strconv.FormatInt(roleID, 10)
	meta, err := c.http.Do(ctx, http.MethodPatch, path, nil, req, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// DeleteRole deletes a custom role.
func (c *Client) DeleteRole(ctx context.Context, workspaceID, roleID int64, reqOpts ...owlvigil.RequestOption) (*owlvigil.ResponseMeta, error) {
	path := "/workspaces/" + strconv.FormatInt(workspaceID, 10) + "/roles/" + strconv.FormatInt(roleID, 10)
	return c.http.Do(ctx, http.MethodDelete, path, nil, nil, nil, reqOpts...)
}

// ListPermissions lists available permissions.
func (c *Client) ListPermissions(ctx context.Context, workspaceID int64, reqOpts ...owlvigil.RequestOption) (*ListResponse[Permission], *owlvigil.ResponseMeta, error) {
	var out ListResponse[Permission]
	path := "/workspaces/" + strconv.FormatInt(workspaceID, 10) + "/permissions"
	meta, err := c.http.Do(ctx, http.MethodGet, path, nil, nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// GetMemberPermissions retrieves a member's effective permissions by user ID.
func (c *Client) GetMemberPermissions(ctx context.Context, workspaceID, userID int64, reqOpts ...owlvigil.RequestOption) (*MemberPermissions, *owlvigil.ResponseMeta, error) {
	var out MemberPermissions
	path := "/workspaces/" + strconv.FormatInt(workspaceID, 10) + "/members/" + strconv.FormatInt(userID, 10) + "/permissions"
	meta, err := c.http.Do(ctx, http.MethodGet, path, nil, nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// UpdateMemberPermissions sets member-level permission overrides by user ID.
func (c *Client) UpdateMemberPermissions(ctx context.Context, workspaceID, userID int64, req *UpdateMemberPermissionsRequest, reqOpts ...owlvigil.RequestOption) (*MemberPermissions, *owlvigil.ResponseMeta, error) {
	var out MemberPermissions
	path := "/workspaces/" + strconv.FormatInt(workspaceID, 10) + "/members/" + strconv.FormatInt(userID, 10) + "/permissions"
	meta, err := c.http.Do(ctx, http.MethodPut, path, nil, req, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// ResetMemberPermissions clears member permission overrides by user ID.
func (c *Client) ResetMemberPermissions(ctx context.Context, workspaceID, userID int64, reqOpts ...owlvigil.RequestOption) (*MemberPermissions, *owlvigil.ResponseMeta, error) {
	var out MemberPermissions
	path := "/workspaces/" + strconv.FormatInt(workspaceID, 10) + "/members/" + strconv.FormatInt(userID, 10) + "/permissions/reset"
	meta, err := c.http.Do(ctx, http.MethodPost, path, nil, nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}
