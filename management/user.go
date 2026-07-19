package management

import (
	"context"
	"net/http"

	owlvigil "github.com/Syrovex/owlvigil_sdk_go"
)

// UserProfile describes user profile information.
type UserProfile struct {
	UserID             int64  `json:"user_id"`
	Email              string `json:"email"`
	Name               string `json:"name,omitempty"`
	AvatarURL          string `json:"avatar_url,omitempty"`
	DefaultWorkspaceID int64  `json:"default_workspace_id,omitempty"`
	Status             string `json:"status"`
	CreatedAt          string `json:"created_at,omitempty"`
}

// UpdateUserProfileRequest updates user profile.
type UpdateUserProfileRequest struct {
	Name               *string `json:"name,omitempty"`
	AvatarURL          *string `json:"avatar_url,omitempty"`
	DefaultWorkspaceID *int64  `json:"default_workspace_id,omitempty"`
}

// UpdatePasswordRequest updates user password.
type UpdatePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

// SupportRequest describes a customer support request.
type SupportRequest struct {
	Title       string `json:"title"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

// NotificationPreferences describes user notification settings.
type NotificationPreferences struct {
	BudgetAlerts    bool `json:"budget_alerts"`
	BillingAlerts   bool `json:"billing_alerts"`
	ReportEmails    bool `json:"report_emails"`
	MarketingEmails bool `json:"marketing_emails"`
}

// UpdateNotificationPreferencesRequest updates notification preferences.
type UpdateNotificationPreferencesRequest struct {
	BudgetAlerts    *bool `json:"budget_alerts,omitempty"`
	BillingAlerts   *bool `json:"billing_alerts,omitempty"`
	ReportEmails    *bool `json:"report_emails,omitempty"`
	MarketingEmails *bool `json:"marketing_emails,omitempty"`
}

// InviteLink describes a user invite link.
type InviteLink struct {
	InviteCode string       `json:"invite_code"`
	InviteURL  string       `json:"invite_url"`
	Stats      *InviteStats `json:"stats,omitempty"`
}

// InviteStats describes invitation statistics.
type InviteStats struct {
	TotalInvites    int     `json:"total_invites"`
	AcceptedInvites int     `json:"accepted_invites"`
	PendingInvites  int     `json:"pending_invites"`
	ConversionRate  float64 `json:"conversion_rate"`
}

// UserInvitation describes a user invitation record.
type UserInvitation struct {
	ID         int64  `json:"id"`
	Email      string `json:"email"`
	Status     string `json:"status"`
	SentAt     string `json:"sent_at,omitempty"`
	AcceptedAt string `json:"accepted_at,omitempty"`
}

// SendInvitationRequest sends invitation emails.
type SendInvitationRequest struct {
	Emails  []string `json:"emails"`
	Message string   `json:"message,omitempty"`
}

// GetUserProfile retrieves current user profile.
func (c *Client) GetUserProfile(ctx context.Context, reqOpts ...owlvigil.RequestOption) (*UserProfile, *owlvigil.ResponseMeta, error) {
	var out UserProfile
	meta, err := c.http.Do(ctx, http.MethodGet, "/user/profile", nil, nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// UpdateUserProfile updates current user profile.
func (c *Client) UpdateUserProfile(ctx context.Context, req *UpdateUserProfileRequest, reqOpts ...owlvigil.RequestOption) (*UserProfile, *owlvigil.ResponseMeta, error) {
	var out UserProfile
	meta, err := c.http.Do(ctx, http.MethodPut, "/user/profile", nil, req, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// UpdatePassword updates current user password.
func (c *Client) UpdatePassword(ctx context.Context, req *UpdatePasswordRequest, reqOpts ...owlvigil.RequestOption) (*owlvigil.ResponseMeta, error) {
	return c.http.Do(ctx, http.MethodPut, "/user/password", nil, req, nil, reqOpts...)
}

// CreateSupportRequest submits a customer support request.
func (c *Client) CreateSupportRequest(ctx context.Context, req *SupportRequest, reqOpts ...owlvigil.RequestOption) (*owlvigil.ResponseMeta, error) {
	return c.http.Do(ctx, http.MethodPost, "/user/support-requests", nil, req, nil, reqOpts...)
}

// GetNotificationPreferences retrieves notification preferences.
func (c *Client) GetNotificationPreferences(ctx context.Context, reqOpts ...owlvigil.RequestOption) (*NotificationPreferences, *owlvigil.ResponseMeta, error) {
	var out NotificationPreferences
	meta, err := c.http.Do(ctx, http.MethodGet, "/user/notification-preferences", nil, nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// UpdateNotificationPreferences updates notification preferences.
func (c *Client) UpdateNotificationPreferences(ctx context.Context, req *UpdateNotificationPreferencesRequest, reqOpts ...owlvigil.RequestOption) (*NotificationPreferences, *owlvigil.ResponseMeta, error) {
	var out NotificationPreferences
	meta, err := c.http.Do(ctx, http.MethodPut, "/user/notification-preferences", nil, req, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// GetInviteLink retrieves or creates user invite link.
func (c *Client) GetInviteLink(ctx context.Context, reqOpts ...owlvigil.RequestOption) (*InviteLink, *owlvigil.ResponseMeta, error) {
	var out InviteLink
	meta, err := c.http.Do(ctx, http.MethodGet, "/users/me/invite-link", nil, nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// GetInvitationStats retrieves invitation statistics.
func (c *Client) GetInvitationStats(ctx context.Context, reqOpts ...owlvigil.RequestOption) (*InviteStats, *owlvigil.ResponseMeta, error) {
	var out InviteStats
	meta, err := c.http.Do(ctx, http.MethodGet, "/users/me/invitation-stats", nil, nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// ListUserInvitations lists user's sent invitations.
func (c *Client) ListUserInvitations(ctx context.Context, opts ListOptions, reqOpts ...owlvigil.RequestOption) (*ListResponse[UserInvitation], *owlvigil.ResponseMeta, error) {
	var out ListResponse[UserInvitation]
	meta, err := c.http.Do(ctx, http.MethodGet, "/users/me/invitations", opts.values(), nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// SendInvitation sends invitation emails.
func (c *Client) SendInvitation(ctx context.Context, req *SendInvitationRequest, reqOpts ...owlvigil.RequestOption) (*owlvigil.ResponseMeta, error) {
	return c.http.Do(ctx, http.MethodPost, "/users/me/send-invitation", nil, req, nil, reqOpts...)
}
