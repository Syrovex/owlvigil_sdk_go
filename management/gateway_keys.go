package management

import (
	"context"
	"net/http"
	"strconv"

	owlvigil "github.com/Syrovex/owlvigil_sdk_go"
)

// GatewayKey describes a Gateway key managed through the Management API.
type GatewayKey struct {
	ID          int64    `json:"id"`
	WorkspaceID int64    `json:"workspace_id,omitempty"`
	Name        string   `json:"name"`
	Status      string   `json:"status"`
	Scopes      []string `json:"scopes,omitempty"`
	Secret      string   `json:"secret,omitempty"`
	CreatedAt   string   `json:"created_at,omitempty"`
}

// CreateGatewayKeyRequest creates a Gateway key.
type CreateGatewayKeyRequest struct {
	WorkspaceID    int64    `json:"workspace_id,omitempty"`
	Name           string   `json:"name"`
	ProviderSource string   `json:"provider_source,omitempty"`
	Scopes         []string `json:"scopes,omitempty"`
}

// UpdateGatewayKeyRequest updates mutable Gateway key fields.
type UpdateGatewayKeyRequest struct {
	Name   *string  `json:"name,omitempty"`
	Scopes []string `json:"scopes,omitempty"`
}

// ListGatewayKeys lists Gateway keys, optionally filtered by status.
func (c *Client) ListGatewayKeys(ctx context.Context, opts ListOptions, status string, reqOpts ...owlvigil.RequestOption) (*ListResponse[GatewayKey], *owlvigil.ResponseMeta, error) {
	q := addFilter(opts.values(), "status", status)
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
	return c.http.Do(ctx, http.MethodPost, "/gateway/keys/"+strconv.FormatInt(id, 10)+"/enable", nil, nil, nil, reqOpts...)
}

// DisableGatewayKey disables a Gateway key.
func (c *Client) DisableGatewayKey(ctx context.Context, id int64, reqOpts ...owlvigil.RequestOption) (*owlvigil.ResponseMeta, error) {
	return c.http.Do(ctx, http.MethodPost, "/gateway/keys/"+strconv.FormatInt(id, 10)+"/disable", nil, nil, nil, reqOpts...)
}

// DeleteGatewayKey deletes a Gateway key.
func (c *Client) DeleteGatewayKey(ctx context.Context, id int64, reqOpts ...owlvigil.RequestOption) (*owlvigil.ResponseMeta, error) {
	return c.http.Do(ctx, http.MethodDelete, "/gateway/keys/"+strconv.FormatInt(id, 10), nil, nil, nil, reqOpts...)
}
