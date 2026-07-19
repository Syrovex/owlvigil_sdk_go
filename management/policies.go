package management

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	owlvigil "github.com/Syrovex/owlvigil_sdk_go"
)

// GatewayPolicy describes policies applied to a gateway key.
type GatewayPolicy struct {
	WorkspaceID      int64 `json:"workspace_id,omitempty"`
	KeyID            int64 `json:"key_id,omitempty"`
	ModelPolicies    any   `json:"model_policies,omitempty"`
	ProviderPolicies any   `json:"provider_policies,omitempty"`
	BudgetPolicies   any   `json:"budget_policies,omitempty"`
	LogPolicies      any   `json:"log_policies,omitempty"`
}

// PreviewPolicyRequest previews policy effect on a request.
type PreviewPolicyRequest struct {
	WorkspaceID int64             `json:"workspace_id,omitempty"`
	Model       string            `json:"model"`
	Provider    string            `json:"provider,omitempty"`
	KeyID       int64             `json:"key_id,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// PreviewPolicyResponse describes policy preview result.
type PreviewPolicyResponse struct {
	Allowed    bool     `json:"allowed"`
	Reason     string   `json:"reason,omitempty"`
	ModifiedBy []string `json:"modified_by,omitempty"`
}

// UpdateGatewayPolicyRequest updates selected gateway policy fields.
type UpdateGatewayPolicyRequest struct {
	Action     *string `json:"action,omitempty"`
	RedirectTo *string `json:"redirect_to,omitempty"`
	Priority   *int    `json:"priority,omitempty"`
	Enabled    *bool   `json:"enabled,omitempty"`
}

// GetGatewayPolicies retrieves policies for a gateway key.
func (c *Client) GetGatewayPolicies(ctx context.Context, keyID int64, reqOpts ...owlvigil.RequestOption) (*GatewayPolicy, *owlvigil.ResponseMeta, error) {
	var out GatewayPolicy
	q := url.Values{}
	q.Set("key_id", strconv.FormatInt(keyID, 10))
	meta, err := c.http.Do(ctx, http.MethodGet, "/gateway/policies", q, nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// PreviewPolicyEffect previews policy effect on a request.
func (c *Client) PreviewPolicyEffect(ctx context.Context, req *PreviewPolicyRequest, reqOpts ...owlvigil.RequestOption) (*PreviewPolicyResponse, *owlvigil.ResponseMeta, error) {
	var out PreviewPolicyResponse
	meta, err := c.http.Do(ctx, http.MethodPost, "/gateway/policies/preview", nil, req, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// UpdateGatewayPolicy updates a gateway policy by ID.
func (c *Client) UpdateGatewayPolicy(ctx context.Context, policyID int64, req *UpdateGatewayPolicyRequest, reqOpts ...owlvigil.RequestOption) (*GatewayPolicy, *owlvigil.ResponseMeta, error) {
	var out GatewayPolicy
	path := "/gateway/policies/" + strconv.FormatInt(policyID, 10)
	meta, err := c.http.Do(ctx, http.MethodPatch, path, nil, req, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}
