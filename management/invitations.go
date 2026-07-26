package management

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	owlvigil "github.com/Syrovex/owlvigil_sdk_go"
)

// Invitation describes a workspace invitation.
type Invitation struct {
	ID              int64      `json:"id"`
	WorkspaceID     int64      `json:"workspace_id"`
	ProjectID       int64      `json:"project_id,omitempty"`
	TeamID          *int64     `json:"team_id,omitempty"`
	Email           string     `json:"email"`
	Role            string     `json:"role,omitempty"`
	Status          string     `json:"status"`
	Token           string     `json:"token,omitempty"`
	InvitedByUserID int64      `json:"invited_by_user_id,omitempty"`
	InvitedBy       *UserBrief `json:"invited_by,omitempty"`
	ExpiresAt       string     `json:"expires_at,omitempty"`
	CreatedAt       string     `json:"created_at,omitempty"`
	UpdatedAt       string     `json:"updated_at,omitempty"`

	// Deprecated compatibility fields from the pre-refactor API.
	RoleIDs    []int64 `json:"role_ids,omitempty"`
	TeamIDs    []int64 `json:"team_ids,omitempty"`
	InviteLink string  `json:"invite_link,omitempty"`
}

// CreateInvitationRequest creates an invitation.
type CreateInvitationRequest struct {
	Email          string `json:"-"`
	Role           string `json:"-"`
	TeamID         *int64 `json:"-"`
	ExpiresInHours int    `json:"-"`

	// Deprecated compatibility fields from the pre-refactor API.
	RoleIDs   []int64 `json:"-"`
	TeamIDs   []int64 `json:"-"`
	ExpiresAt string  `json:"-"`
}

// MarshalJSON emits the current workspace invitation contract.
func (r CreateInvitationRequest) MarshalJSON() ([]byte, error) {
	role := r.Role
	if role == "" && len(r.RoleIDs) > 0 {
		role = "member"
	}
	teamID := r.TeamID
	if teamID == nil && len(r.TeamIDs) > 0 {
		teamID = &r.TeamIDs[0]
	}
	return json.Marshal(struct {
		Email          string `json:"email"`
		Role           string `json:"role"`
		TeamID         *int64 `json:"team_id,omitempty"`
		ExpiresInHours int    `json:"expires_in_hours,omitempty"`
	}{
		Email:          r.Email,
		Role:           role,
		TeamID:         teamID,
		ExpiresInHours: r.ExpiresInHours,
	})
}

// ListInvitations lists workspace invitations.
func (c *Client) ListInvitations(ctx context.Context, workspaceID int64, opts ListOptions, reqOpts ...owlvigil.RequestOption) (*ListResponse[Invitation], *owlvigil.ResponseMeta, error) {
	var out ListResponse[Invitation]
	path := "/workspaces/" + strconv.FormatInt(workspaceID, 10) + "/invitations"
	// The current Open API declares no query parameters on this route.
	_ = opts
	meta, err := c.http.Do(ctx, http.MethodGet, path, nil, nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// CreateInvitation creates a new invitation.
func (c *Client) CreateInvitation(ctx context.Context, workspaceID int64, req *CreateInvitationRequest, reqOpts ...owlvigil.RequestOption) (*Invitation, *owlvigil.ResponseMeta, error) {
	var out Invitation
	path := "/workspaces/" + strconv.FormatInt(workspaceID, 10) + "/invitations"
	meta, err := c.http.Do(ctx, http.MethodPost, path, nil, req, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// ResendInvitation resends an invitation email.
func (c *Client) ResendInvitation(ctx context.Context, workspaceID, invitationID int64, reqOpts ...owlvigil.RequestOption) (*owlvigil.ResponseMeta, error) {
	_, meta, err := c.ResendInvitationWithResult(ctx, workspaceID, invitationID, reqOpts...)
	return meta, err
}

// ResendInvitationWithResult resends an invitation and returns its updated state.
func (c *Client) ResendInvitationWithResult(ctx context.Context, workspaceID, invitationID int64, reqOpts ...owlvigil.RequestOption) (*Invitation, *owlvigil.ResponseMeta, error) {
	path := "/workspaces/" + strconv.FormatInt(workspaceID, 10) + "/invitations/" + strconv.FormatInt(invitationID, 10) + "/resend"
	var out Invitation
	meta, err := c.http.Do(ctx, http.MethodPost, path, nil, nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// RevokeInvitation revokes an invitation.
func (c *Client) RevokeInvitation(ctx context.Context, workspaceID, invitationID int64, reqOpts ...owlvigil.RequestOption) (*owlvigil.ResponseMeta, error) {
	_, meta, err := c.RevokeInvitationWithResult(ctx, workspaceID, invitationID, reqOpts...)
	return meta, err
}

// RevokeInvitationWithResult revokes an invitation and returns its updated state.
func (c *Client) RevokeInvitationWithResult(ctx context.Context, workspaceID, invitationID int64, reqOpts ...owlvigil.RequestOption) (*Invitation, *owlvigil.ResponseMeta, error) {
	path := "/workspaces/" + strconv.FormatInt(workspaceID, 10) + "/invitations/" + strconv.FormatInt(invitationID, 10) + "/revoke"
	var out Invitation
	meta, err := c.http.Do(ctx, http.MethodPost, path, nil, nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}
