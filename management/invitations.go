package management

import (
	"context"
	"net/http"
	"strconv"

	owlvigil "github.com/Syrovex/owlvigil_sdk_go"
)

// Invitation describes a workspace invitation.
type Invitation struct {
	ID          int64   `json:"id"`
	WorkspaceID int64   `json:"workspace_id"`
	Email       string  `json:"email"`
	Status      string  `json:"status"`
	RoleIDs     []int64 `json:"role_ids,omitempty"`
	TeamIDs     []int64 `json:"team_ids,omitempty"`
	InviteLink  string  `json:"invite_link,omitempty"`
	ExpiresAt   string  `json:"expires_at,omitempty"`
	CreatedAt   string  `json:"created_at,omitempty"`
}

// CreateInvitationRequest creates an invitation.
type CreateInvitationRequest struct {
	Email     string  `json:"email"`
	RoleIDs   []int64 `json:"role_ids,omitempty"`
	TeamIDs   []int64 `json:"team_ids,omitempty"`
	ExpiresAt string  `json:"expires_at,omitempty"`
}

// ListInvitations lists workspace invitations.
func (c *Client) ListInvitations(ctx context.Context, workspaceID int64, opts ListOptions, reqOpts ...owlvigil.RequestOption) (*ListResponse[Invitation], *owlvigil.ResponseMeta, error) {
	var out ListResponse[Invitation]
	path := "/workspaces/" + strconv.FormatInt(workspaceID, 10) + "/invitations"
	meta, err := c.http.Do(ctx, http.MethodGet, path, opts.values(), nil, &out, reqOpts...)
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
	path := "/workspaces/" + strconv.FormatInt(workspaceID, 10) + "/invitations/" + strconv.FormatInt(invitationID, 10) + "/resend"
	return c.http.Do(ctx, http.MethodPost, path, nil, nil, nil, reqOpts...)
}

// RevokeInvitation revokes an invitation.
func (c *Client) RevokeInvitation(ctx context.Context, workspaceID, invitationID int64, reqOpts ...owlvigil.RequestOption) (*owlvigil.ResponseMeta, error) {
	path := "/workspaces/" + strconv.FormatInt(workspaceID, 10) + "/invitations/" + strconv.FormatInt(invitationID, 10) + "/revoke"
	return c.http.Do(ctx, http.MethodPost, path, nil, nil, nil, reqOpts...)
}
