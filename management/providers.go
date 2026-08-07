package management

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	owlvigil "github.com/Syrovex/owlvigil_sdk_go"
)

// Provider describes an upstream model provider configured for a workspace.
type Provider struct {
	ID                int64   `json:"id"`
	WorkspaceID       int64   `json:"workspace_id"`
	Name              string  `json:"name"`
	Type              string  `json:"type"`
	ProviderSource    string  `json:"provider_source,omitempty"`
	ProviderPlatform  *string `json:"provider_platform,omitempty"`
	IsPlatformManaged bool    `json:"is_platform_managed,omitempty"`
	APIKey            string  `json:"api_key,omitempty"`
	BaseURL           string  `json:"base_url,omitempty"`
	DefaultModel      string  `json:"default_model,omitempty"`
	APIMode           string  `json:"api_mode,omitempty"`
	Status            string  `json:"status,omitempty"`
	RequestCount      int64   `json:"request_count,omitempty"`
	ErrorCount        int64   `json:"error_count,omitempty"`
	LastUsedAt        string  `json:"last_used_at,omitempty"`
	CreatedAt         string  `json:"created_at,omitempty"`
	UpdatedAt         string  `json:"updated_at,omitempty"`
}

// CreateProviderRequest creates an upstream provider for a workspace.
type CreateProviderRequest struct {
	WorkspaceID  int64  `json:"workspace_id"`
	Name         string `json:"name"`
	Type         string `json:"type"`
	APIKey       string `json:"api_key"`
	BaseURL      string `json:"base_url,omitempty"`
	APIMode      string `json:"api_mode,omitempty"`
	DefaultModel string `json:"default_model,omitempty"`
}

// UpdateProviderRequest updates selected upstream provider properties.
type UpdateProviderRequest struct {
	WorkspaceID  int64   `json:"workspace_id"`
	Name         *string `json:"name,omitempty"`
	Status       *string `json:"status,omitempty"`
	APIKey       *string `json:"api_key,omitempty"`
	BaseURL      *string `json:"base_url,omitempty"`
	APIMode      *string `json:"api_mode,omitempty"`
	DefaultModel *string `json:"default_model,omitempty"`
}

// VerifyProviderConnectionRequest checks a saved provider or temporary credentials.
type VerifyProviderConnectionRequest struct {
	WorkspaceID  int    `json:"workspace_id"`
	ProviderID   *int   `json:"provider_id,omitempty"`
	Type         string `json:"type,omitempty"`
	APIKey       string `json:"api_key,omitempty"`
	BaseURL      string `json:"base_url,omitempty"`
	APIMode      string `json:"api_mode,omitempty"`
	DefaultModel string `json:"default_model,omitempty"`
}

// ProviderConnectionVerification contains a provider connection test result.
type ProviderConnectionVerification struct {
	Success   bool    `json:"success"`
	Message   string  `json:"message"`
	LatencyMS float64 `json:"latency_ms,omitempty"`
}

// ListProviders lists providers configured for a workspace.
func (c *Client) ListProviders(ctx context.Context, workspaceID int64, reqOpts ...owlvigil.RequestOption) (*ListResponse[Provider], *owlvigil.ResponseMeta, error) {
	var out ListResponse[Provider]
	query := url.Values{"workspace_id": {strconv.FormatInt(workspaceID, 10)}}
	meta, err := c.http.Do(ctx, http.MethodGet, "/gateway/providers", query, nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// CreateProvider creates an upstream provider.
func (c *Client) CreateProvider(ctx context.Context, req *CreateProviderRequest, reqOpts ...owlvigil.RequestOption) (*Provider, *owlvigil.ResponseMeta, error) {
	var out Provider
	meta, err := c.http.Do(ctx, http.MethodPost, "/gateway/providers", nil, req, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// VerifyProviderConnection verifies a saved provider or temporary credentials.
func (c *Client) VerifyProviderConnection(ctx context.Context, req *VerifyProviderConnectionRequest, reqOpts ...owlvigil.RequestOption) (*ProviderConnectionVerification, *owlvigil.ResponseMeta, error) {
	var out ProviderConnectionVerification
	meta, err := c.http.Do(ctx, http.MethodPost, "/gateway/providers/verify-connection", nil, req, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// GetProvider retrieves a provider by ID in a workspace.
func (c *Client) GetProvider(ctx context.Context, workspaceID, providerID int64, reqOpts ...owlvigil.RequestOption) (*Provider, *owlvigil.ResponseMeta, error) {
	var out Provider
	query := url.Values{"workspace_id": {strconv.FormatInt(workspaceID, 10)}}
	path := "/gateway/providers/" + strconv.FormatInt(providerID, 10)
	meta, err := c.http.Do(ctx, http.MethodGet, path, query, nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// UpdateProvider updates a provider in a workspace.
func (c *Client) UpdateProvider(ctx context.Context, workspaceID, providerID int64, req *UpdateProviderRequest, reqOpts ...owlvigil.RequestOption) (*Provider, *owlvigil.ResponseMeta, error) {
	var out Provider
	query := url.Values{"workspace_id": {strconv.FormatInt(workspaceID, 10)}}
	path := "/gateway/providers/" + strconv.FormatInt(providerID, 10)
	body := req
	if req != nil && req.WorkspaceID != 0 && req.WorkspaceID != workspaceID {
		return nil, nil, fmt.Errorf(
			"request workspace_id %d does not match workspaceID %d",
			req.WorkspaceID,
			workspaceID,
		)
	}
	if req != nil && req.WorkspaceID == 0 {
		copyRequest := *req
		copyRequest.WorkspaceID = workspaceID
		body = &copyRequest
	}
	meta, err := c.http.Do(ctx, http.MethodPatch, path, query, body, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// DeleteProvider deletes a provider from a workspace.
func (c *Client) DeleteProvider(ctx context.Context, workspaceID, providerID int64, reqOpts ...owlvigil.RequestOption) (*owlvigil.ResponseMeta, error) {
	_, meta, err := c.DeleteProviderWithResult(ctx, workspaceID, providerID, reqOpts...)
	return meta, err
}

// DeleteProviderWithResult deletes a provider and returns confirmation.
func (c *Client) DeleteProviderWithResult(ctx context.Context, workspaceID, providerID int64, reqOpts ...owlvigil.RequestOption) (*DeleteResponse, *owlvigil.ResponseMeta, error) {
	query := url.Values{"workspace_id": {strconv.FormatInt(workspaceID, 10)}}
	path := "/gateway/providers/" + strconv.FormatInt(providerID, 10)
	var out DeleteResponse
	meta, err := c.http.Do(ctx, http.MethodDelete, path, query, nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}
