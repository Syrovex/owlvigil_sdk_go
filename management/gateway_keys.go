package management

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	owlvigil "github.com/Syrovex/owlvigil_sdk_go"
)

// GatewayKey describes a Gateway key managed through the Management API.
type GatewayKey struct {
	ID               int64   `json:"id"`
	WorkspaceID      int64   `json:"workspace_id,omitempty"`
	Name             string  `json:"name"`
	Key              string  `json:"key,omitempty"`
	MaskedKey        string  `json:"masked_key,omitempty"`
	Status           string  `json:"status"`
	ProviderSource   string  `json:"provider_source,omitempty"`
	ProviderPlatform *string `json:"provider_platform,omitempty"`
	TeamID           *int64  `json:"team_id,omitempty"`
	AssigneeUserID   *int    `json:"assignee_user_id,omitempty"`
	RateLimitRPM     *int    `json:"rate_limit_rpm,omitempty"`
	UsageMultiplier  float64 `json:"usage_multiplier"`
	CreatedAt        string  `json:"created_at,omitempty"`
	UpdatedAt        string  `json:"updated_at,omitempty"`

	// Deprecated compatibility aliases.
	Scopes []string `json:"scopes,omitempty"`
	Secret string   `json:"secret,omitempty"`
}

// UnmarshalJSON keeps the one-time key available through both Key and Secret.
func (k *GatewayKey) UnmarshalJSON(data []byte) error {
	type gatewayKeyAlias GatewayKey
	var out gatewayKeyAlias
	if err := json.Unmarshal(data, &out); err != nil {
		return err
	}
	*k = GatewayKey(out)
	if k.Key == "" {
		k.Key = k.Secret
	}
	if k.Secret == "" {
		k.Secret = k.Key
	}
	return nil
}

// CreateGatewayKeyRequest creates a Gateway key.
type CreateGatewayKeyRequest struct {
	WorkspaceID      int64    `json:"workspace_id"`
	Name             string   `json:"name"`
	ProviderSource   string   `json:"provider_source,omitempty"`
	ProviderPlatform *string  `json:"provider_platform,omitempty"`
	TeamID           *int64   `json:"team_id,omitempty"`
	AssigneeUserID   *int     `json:"assignee_user_id,omitempty"`
	RateLimitRPM     *int     `json:"rate_limit_rpm,omitempty"`
	BudgetLimit      *float64 `json:"budget_limit,omitempty"`
	UsageMultiplier  *float64 `json:"usage_multiplier,omitempty"`

	// Scopes is retained for source compatibility and is not part of the current Open API.
	Scopes []string `json:"-"`
}

// UpdateGatewayKeyRequest updates mutable Gateway key fields.
type UpdateGatewayKeyRequest struct {
	Name                *string  `json:"name,omitempty"`
	ProviderSource      *string  `json:"provider_source,omitempty"`
	ProviderPlatform    *string  `json:"provider_platform,omitempty"`
	TeamID              *int64   `json:"team_id,omitempty"`
	ClearTeamID         bool     `json:"clear_team_id,omitempty"`
	AssigneeUserID      *int     `json:"assignee_user_id,omitempty"`
	ClearAssigneeUserID bool     `json:"clear_assignee_user_id,omitempty"`
	RateLimitRPM        *int     `json:"rate_limit_rpm,omitempty"`
	BudgetLimit         *float64 `json:"budget_limit,omitempty"`
	UsageMultiplier     *float64 `json:"usage_multiplier,omitempty"`

	// Scopes is retained for source compatibility and is not part of the current Open API.
	Scopes []string `json:"-"`
}

// ListGatewayKeys lists Gateway keys, optionally filtered by status.
func (c *Client) ListGatewayKeys(ctx context.Context, opts ListOptions, status string, reqOpts ...owlvigil.RequestOption) (*ListResponse[GatewayKey], *owlvigil.ResponseMeta, error) {
	q := opts.values()
	// The refactored Open API no longer declares a status filter.
	_ = status
	var out ListResponse[GatewayKey]
	meta, err := c.http.Do(ctx, http.MethodGet, "/gateway/keys", q, nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// CreateGatewayKey creates a Gateway key.
func (c *Client) CreateGatewayKey(ctx context.Context, req *CreateGatewayKeyRequest, reqOpts ...owlvigil.RequestOption) (*GatewayKey, *owlvigil.ResponseMeta, error) {
	var out GatewayKey
	meta, err := c.http.Do(ctx, http.MethodPost, "/gateway/keys", nil, req, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// GetGatewayKey retrieves a Gateway key by ID.
func (c *Client) GetGatewayKey(ctx context.Context, id int64, reqOpts ...owlvigil.RequestOption) (*GatewayKey, *owlvigil.ResponseMeta, error) {
	var out GatewayKey
	meta, err := c.http.Do(ctx, http.MethodGet, "/gateway/keys/"+strconv.FormatInt(id, 10), nil, nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// UpdateGatewayKey updates a Gateway key.
func (c *Client) UpdateGatewayKey(ctx context.Context, id int64, req *UpdateGatewayKeyRequest, reqOpts ...owlvigil.RequestOption) (*GatewayKey, *owlvigil.ResponseMeta, error) {
	var out GatewayKey
	meta, err := c.http.Do(ctx, http.MethodPatch, "/gateway/keys/"+strconv.FormatInt(id, 10), nil, req, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// RotateGatewayKey rotates a Gateway key and returns the new one-time secret when provided by the API.
func (c *Client) RotateGatewayKey(ctx context.Context, id int64, reqOpts ...owlvigil.RequestOption) (*GatewayKey, *owlvigil.ResponseMeta, error) {
	var out GatewayKey
	meta, err := c.http.Do(ctx, http.MethodPost, "/gateway/keys/"+strconv.FormatInt(id, 10)+"/rotate", nil, nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// EnableGatewayKey enables a Gateway key.
func (c *Client) EnableGatewayKey(ctx context.Context, id int64, reqOpts ...owlvigil.RequestOption) (*owlvigil.ResponseMeta, error) {
	_, meta, err := c.EnableGatewayKeyWithResult(ctx, id, reqOpts...)
	return meta, err
}

// EnableGatewayKeyWithResult enables a Gateway key and returns its updated state.
func (c *Client) EnableGatewayKeyWithResult(ctx context.Context, id int64, reqOpts ...owlvigil.RequestOption) (*GatewayKey, *owlvigil.ResponseMeta, error) {
	var out GatewayKey
	meta, err := c.http.Do(ctx, http.MethodPost, "/gateway/keys/"+strconv.FormatInt(id, 10)+"/enable", nil, nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// DisableGatewayKey disables a Gateway key.
func (c *Client) DisableGatewayKey(ctx context.Context, id int64, reqOpts ...owlvigil.RequestOption) (*owlvigil.ResponseMeta, error) {
	_, meta, err := c.DisableGatewayKeyWithResult(ctx, id, reqOpts...)
	return meta, err
}

// DisableGatewayKeyWithResult disables a Gateway key and returns its updated state.
func (c *Client) DisableGatewayKeyWithResult(ctx context.Context, id int64, reqOpts ...owlvigil.RequestOption) (*GatewayKey, *owlvigil.ResponseMeta, error) {
	var out GatewayKey
	meta, err := c.http.Do(ctx, http.MethodPost, "/gateway/keys/"+strconv.FormatInt(id, 10)+"/disable", nil, nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// DeleteGatewayKey deletes a Gateway key.
func (c *Client) DeleteGatewayKey(ctx context.Context, id int64, reqOpts ...owlvigil.RequestOption) (*owlvigil.ResponseMeta, error) {
	_, meta, err := c.DeleteGatewayKeyWithResult(ctx, id, reqOpts...)
	return meta, err
}

// DeleteGatewayKeyWithResult deletes a Gateway key and returns confirmation.
func (c *Client) DeleteGatewayKeyWithResult(ctx context.Context, id int64, reqOpts ...owlvigil.RequestOption) (*DeleteResponse, *owlvigil.ResponseMeta, error) {
	var out DeleteResponse
	meta, err := c.http.Do(ctx, http.MethodDelete, "/gateway/keys/"+strconv.FormatInt(id, 10), nil, nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}
