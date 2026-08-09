package management

import (
	"context"
	"encoding/json"
	"net/http"

	owlvigil "github.com/Syrovex/owlvigil_sdk_go"
)

// MessageResponse contains an action result message.
type MessageResponse struct {
	Message string `json:"message"`
}

// InvitationDeliveryResponse reports how many invitation emails were sent.
type InvitationDeliveryResponse struct {
	Message    string `json:"message"`
	EmailsSent int    `json:"emails_sent"`
}

// UserProfile describes user profile information.
type UserProfile struct {
	ID                       int64              `json:"id"`
	UserID                   int64              `json:"user_id"`
	Username                 string             `json:"username"`
	Email                    string             `json:"email"`
	Name                     string             `json:"name,omitempty"`
	DisplayName              *string            `json:"display_name,omitempty"`
	AvatarURL                string             `json:"avatar_url,omitempty"`
	DefaultWorkspaceID       int64              `json:"default_workspace_id,omitempty"`
	Status                   string             `json:"status"`
	BalanceNotifyEnabled     bool               `json:"balance_notify_enabled"`
	BalanceNotifyThreshold   *float64           `json:"balance_notify_threshold,omitempty"`
	BalanceNotifyExtraEmails []NotifyEmailEntry `json:"balance_notify_extra_emails,omitempty"`
	CreatedAt                string             `json:"created_at,omitempty"`
	UpdatedAt                string             `json:"updated_at,omitempty"`
}

// NotifyEmailEntry is an additional balance-notification recipient.
type NotifyEmailEntry struct {
	Email    string `json:"email"`
	Disabled bool   `json:"disabled"`
	Verified bool   `json:"verified"`
}

// UpdateUserProfileRequest updates user profile.
type UpdateUserProfileRequest struct {
	Username                 *string             `json:"-"`
	AvatarURL                *string             `json:"-"`
	DefaultWorkspaceID       *int64              `json:"-"`
	BalanceNotifyEnabled     *bool               `json:"-"`
	BalanceNotifyThreshold   *float64            `json:"-"`
	BalanceNotifyExtraEmails *[]NotifyEmailEntry `json:"-"`

	// ClearAvatarURL and ClearBalanceNotifyThreshold send an explicit JSON null.
	// A clear flag takes precedence when the corresponding value is also set.
	ClearAvatarURL              bool `json:"-"`
	ClearBalanceNotifyThreshold bool `json:"-"`

	// Name is retained as a compatibility alias for Username.
	//
	// Deprecated: use Username. Name is serialized as username, not display name.
	Name *string `json:"-"`
}

// MarshalJSON emits the refactored Open API profile contract.
func (r UpdateUserProfileRequest) MarshalJSON() ([]byte, error) {
	username := r.Username
	if username == nil {
		username = r.Name
	}
	var avatarURL json.RawMessage
	switch {
	case r.ClearAvatarURL:
		avatarURL = json.RawMessage("null")
	case r.AvatarURL != nil:
		avatarURL, _ = json.Marshal(*r.AvatarURL)
	}
	var balanceNotifyThreshold json.RawMessage
	switch {
	case r.ClearBalanceNotifyThreshold:
		balanceNotifyThreshold = json.RawMessage("null")
	case r.BalanceNotifyThreshold != nil:
		balanceNotifyThreshold, _ = json.Marshal(*r.BalanceNotifyThreshold)
	}
	return json.Marshal(struct {
		Username                 *string             `json:"username,omitempty"`
		AvatarURL                json.RawMessage     `json:"avatar_url,omitempty"`
		DefaultWorkspaceID       *int64              `json:"default_workspace_id,omitempty"`
		BalanceNotifyEnabled     *bool               `json:"balance_notify_enabled,omitempty"`
		BalanceNotifyThreshold   json.RawMessage     `json:"balance_notify_threshold,omitempty"`
		BalanceNotifyExtraEmails *[]NotifyEmailEntry `json:"balance_notify_extra_emails,omitempty"`
	}{
		Username:                 username,
		AvatarURL:                avatarURL,
		DefaultWorkspaceID:       r.DefaultWorkspaceID,
		BalanceNotifyEnabled:     r.BalanceNotifyEnabled,
		BalanceNotifyThreshold:   balanceNotifyThreshold,
		BalanceNotifyExtraEmails: r.BalanceNotifyExtraEmails,
	})
}

// UpdatePasswordRequest updates user password.
type UpdatePasswordRequest struct {
	OldPassword string `json:"-"`
	NewPassword string `json:"-"`

	// CurrentPassword is retained as a compatibility alias for OldPassword.
	CurrentPassword string `json:"-"`
}

// MarshalJSON emits old_password as required by the refactored Open API.
func (r UpdatePasswordRequest) MarshalJSON() ([]byte, error) {
	oldPassword := r.OldPassword
	if oldPassword == "" {
		oldPassword = r.CurrentPassword
	}
	return json.Marshal(struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}{
		OldPassword: oldPassword,
		NewPassword: r.NewPassword,
	})
}

// SupportRequest describes a customer support request.
type SupportRequest struct {
	Subject     string `json:"-"`
	IssueType   string `json:"-"`
	Description string `json:"description"`

	// Title and Type are retained as compatibility aliases.
	Title string `json:"-"`
	Type  string `json:"-"`
}

// MarshalJSON emits subject and issue_type as required by the refactored Open API.
func (r SupportRequest) MarshalJSON() ([]byte, error) {
	subject := r.Subject
	if subject == "" {
		subject = r.Title
	}
	issueType := r.IssueType
	if issueType == "" {
		issueType = r.Type
	}
	return json.Marshal(struct {
		Subject     string `json:"subject"`
		IssueType   string `json:"issue_type"`
		Description string `json:"description"`
	}{
		Subject:     subject,
		IssueType:   issueType,
		Description: r.Description,
	})
}

// NotificationPreferences describes user notification settings.
type NotificationPreferences struct {
	Budget    bool `json:"budget"`
	Billing   bool `json:"billing"`
	Reports   bool `json:"reports"`
	Marketing bool `json:"marketing"`

	// Deprecated compatibility aliases.
	BudgetAlerts    bool `json:"-"`
	BillingAlerts   bool `json:"-"`
	ReportEmails    bool `json:"-"`
	MarketingEmails bool `json:"-"`
}

// UnmarshalJSON accepts the current contract and the pre-refactor field names.
func (p *NotificationPreferences) UnmarshalJSON(data []byte) error {
	var raw struct {
		Budget          *bool `json:"budget"`
		Billing         *bool `json:"billing"`
		Reports         *bool `json:"reports"`
		Marketing       *bool `json:"marketing"`
		BudgetAlerts    *bool `json:"budget_alerts"`
		BillingAlerts   *bool `json:"billing_alerts"`
		ReportEmails    *bool `json:"report_emails"`
		MarketingEmails *bool `json:"marketing_emails"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	p.Budget = firstBool(raw.Budget, raw.BudgetAlerts)
	p.Billing = firstBool(raw.Billing, raw.BillingAlerts)
	p.Reports = firstBool(raw.Reports, raw.ReportEmails)
	p.Marketing = firstBool(raw.Marketing, raw.MarketingEmails)
	p.BudgetAlerts = p.Budget
	p.BillingAlerts = p.Billing
	p.ReportEmails = p.Reports
	p.MarketingEmails = p.Marketing
	return nil
}

// UpdateNotificationPreferencesRequest updates notification preferences.
type UpdateNotificationPreferencesRequest struct {
	Budget    *bool `json:"-"`
	Billing   *bool `json:"-"`
	Reports   *bool `json:"-"`
	Marketing *bool `json:"-"`

	// Deprecated compatibility aliases.
	BudgetAlerts    *bool `json:"-"`
	BillingAlerts   *bool `json:"-"`
	ReportEmails    *bool `json:"-"`
	MarketingEmails *bool `json:"-"`
}

// MarshalJSON emits the refactored notification preference field names.
func (r UpdateNotificationPreferencesRequest) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Budget    bool `json:"budget"`
		Billing   bool `json:"billing"`
		Reports   bool `json:"reports"`
		Marketing bool `json:"marketing"`
	}{
		Budget:    firstBool(r.Budget, r.BudgetAlerts),
		Billing:   firstBool(r.Billing, r.BillingAlerts),
		Reports:   firstBool(r.Reports, r.ReportEmails),
		Marketing: firstBool(r.Marketing, r.MarketingEmails),
	})
}

// InviteLink describes a user invite link.
type InviteLink struct {
	InviteCode           string  `json:"invite_code"`
	InviteLink           string  `json:"invite_link"`
	TotalInvitations     int     `json:"total_invitations"`
	ConvertedInvitations int     `json:"converted_invitations"`
	PendingInvitations   int     `json:"pending_invitations"`
	ConversionRate       float64 `json:"conversion_rate"`

	// Legacy aliases retained for source compatibility.
	InviteURL string       `json:"invite_url,omitempty"`
	Stats     *InviteStats `json:"stats,omitempty"`
}

// UnmarshalJSON synchronizes the current flat invitation summary with the
// pre-refactor URL and nested stats aliases.
func (l *InviteLink) UnmarshalJSON(data []byte) error {
	type alias InviteLink
	var out alias
	if err := json.Unmarshal(data, &out); err != nil {
		return err
	}
	*l = InviteLink(out)
	if l.InviteLink == "" {
		l.InviteLink = l.InviteURL
	}
	if l.InviteURL == "" {
		l.InviteURL = l.InviteLink
	}
	if l.Stats == nil {
		l.Stats = &InviteStats{
			TotalInvitations:     l.TotalInvitations,
			ConvertedInvitations: l.ConvertedInvitations,
			PendingInvitations:   l.PendingInvitations,
			ConversionRate:       l.ConversionRate,
		}
		l.Stats.syncLegacyAliases()
	}
	return nil
}

// InviteStats describes invitation statistics.
type InviteStats struct {
	InviteCode           string  `json:"invite_code,omitempty"`
	InviteLink           string  `json:"invite_link,omitempty"`
	TotalInvitations     int     `json:"total_invitations"`
	ConvertedInvitations int     `json:"converted_invitations"`
	PendingInvitations   int     `json:"pending_invitations"`
	ConversionRate       float64 `json:"conversion_rate"`

	// Legacy aliases retained for source compatibility.
	TotalInvites    int `json:"total_invites,omitempty"`
	AcceptedInvites int `json:"accepted_invites,omitempty"`
	PendingInvites  int `json:"pending_invites,omitempty"`
}

// UnmarshalJSON accepts both current and legacy invitation statistics.
func (s *InviteStats) UnmarshalJSON(data []byte) error {
	type alias InviteStats
	var out alias
	if err := json.Unmarshal(data, &out); err != nil {
		return err
	}
	*s = InviteStats(out)
	if s.TotalInvitations == 0 {
		s.TotalInvitations = s.TotalInvites
	}
	if s.ConvertedInvitations == 0 {
		s.ConvertedInvitations = s.AcceptedInvites
	}
	if s.PendingInvitations == 0 {
		s.PendingInvitations = s.PendingInvites
	}
	s.syncLegacyAliases()
	return nil
}

func (s *InviteStats) syncLegacyAliases() {
	s.TotalInvites = s.TotalInvitations
	s.AcceptedInvites = s.ConvertedInvitations
	s.PendingInvites = s.PendingInvitations
}

// UserInvitation describes a user invitation record.
type UserInvitation struct {
	ID            int64  `json:"id"`
	InviteCode    string `json:"invite_code"`
	InviterUserID int64  `json:"inviter_user_id"`
	InvitedUserID *int   `json:"invited_user_id"`
	InvitedEmail  string `json:"invited_email"`
	Status        string `json:"status"`
	CreatedAt     string `json:"created_at"`
	ConvertedAt   string `json:"converted_at,omitempty"`

	// Legacy aliases retained for source compatibility.
	Email      string `json:"email,omitempty"`
	SentAt     string `json:"sent_at,omitempty"`
	AcceptedAt string `json:"accepted_at,omitempty"`
}

// UnmarshalJSON synchronizes current and legacy invitation fields.
func (i *UserInvitation) UnmarshalJSON(data []byte) error {
	type alias UserInvitation
	var out alias
	if err := json.Unmarshal(data, &out); err != nil {
		return err
	}
	*i = UserInvitation(out)
	if i.InvitedEmail == "" {
		i.InvitedEmail = i.Email
	}
	if i.Email == "" {
		i.Email = i.InvitedEmail
	}
	if i.CreatedAt == "" {
		i.CreatedAt = i.SentAt
	}
	if i.SentAt == "" {
		i.SentAt = i.CreatedAt
	}
	if i.ConvertedAt == "" {
		i.ConvertedAt = i.AcceptedAt
	}
	if i.AcceptedAt == "" {
		i.AcceptedAt = i.ConvertedAt
	}
	return nil
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
	_, meta, err := c.UpdatePasswordWithResult(ctx, req, reqOpts...)
	return meta, err
}

// UpdatePasswordWithResult updates the password and returns the published
// action response.
func (c *Client) UpdatePasswordWithResult(ctx context.Context, req *UpdatePasswordRequest, reqOpts ...owlvigil.RequestOption) (*MessageResponse, *owlvigil.ResponseMeta, error) {
	var out MessageResponse
	meta, err := c.http.Do(ctx, http.MethodPut, "/user/password", nil, req, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// CreateSupportRequest submits a customer support request.
func (c *Client) CreateSupportRequest(ctx context.Context, req *SupportRequest, reqOpts ...owlvigil.RequestOption) (*owlvigil.ResponseMeta, error) {
	_, meta, err := c.CreateSupportRequestWithResult(ctx, req, reqOpts...)
	return meta, err
}

// CreateSupportRequestWithResult submits a support request and returns the
// published action response.
func (c *Client) CreateSupportRequestWithResult(ctx context.Context, req *SupportRequest, reqOpts ...owlvigil.RequestOption) (*MessageResponse, *owlvigil.ResponseMeta, error) {
	var out MessageResponse
	meta, err := c.http.Do(ctx, http.MethodPost, "/user/support-requests", nil, req, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
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
	// The current Open API returns an array and declares no query parameters.
	_ = opts
	meta, err := c.http.Do(ctx, http.MethodGet, "/users/me/invitations", nil, nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// SendInvitation sends invitation emails.
func (c *Client) SendInvitation(ctx context.Context, req *SendInvitationRequest, reqOpts ...owlvigil.RequestOption) (*owlvigil.ResponseMeta, error) {
	_, meta, err := c.SendInvitationWithResult(ctx, req, reqOpts...)
	return meta, err
}

// SendInvitationWithResult sends invitation emails and returns the delivery
// count published by the Open API.
func (c *Client) SendInvitationWithResult(ctx context.Context, req *SendInvitationRequest, reqOpts ...owlvigil.RequestOption) (*InvitationDeliveryResponse, *owlvigil.ResponseMeta, error) {
	var out InvitationDeliveryResponse
	meta, err := c.http.Do(ctx, http.MethodPost, "/users/me/send-invitation", nil, req, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

func firstBool(primary, compatibility *bool) bool {
	if primary != nil {
		return *primary
	}
	return compatibility != nil && *compatibility
}
