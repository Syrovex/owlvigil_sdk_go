package management

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"

	owlvigil "github.com/Syrovex/owlvigil_sdk_go"
)

// GatewayPolicy describes policies applied to a gateway key.
type GatewayPolicy struct {
	WorkspaceID     int64             `json:"workspace_id,omitempty"`
	KeyID           int64             `json:"key_id,omitempty"`
	ModelPolicies   []ModelPolicy     `json:"model_policies"`
	KeywordPolicies []KeywordPolicy   `json:"keyword_policies"`
	BudgetPolicies  []BudgetPolicy    `json:"budget_policies"`
	LogPolicies     []LogPolicy       `json:"log_policies"`
	RateLimits      []PolicyRateLimit `json:"rate_limits"`
	Warnings        []string          `json:"warnings,omitempty"`

	// Fields returned by PATCH /gateway/policies/{policy_id}.
	ID         int      `json:"id,omitempty"`
	Provider   string   `json:"provider,omitempty"`
	ModelID    string   `json:"model_id,omitempty"`
	Action     string   `json:"action,omitempty"`
	RedirectTo *string  `json:"redirect_to,omitempty"`
	Priority   int      `json:"priority,omitempty"`
	Scope      string   `json:"scope,omitempty"`
	AppliesTo  []string `json:"applies_to,omitempty"`
	Reason     string   `json:"reason,omitempty"`

	// ProviderPolicies is retained for pre-refactor responses.
	ProviderPolicies any `json:"provider_policies,omitempty"`
}

// UnmarshalJSON decodes the current typed policy arrays. Pre-refactor
// deployments returned model_policies, budget_policies, or log_policies as
// objects; those legacy shapes are tolerated so unrelated response fields
// remain usable during a rolling deployment.
func (p *GatewayPolicy) UnmarshalJSON(data []byte) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}

	modelPolicies := fields["model_policies"]
	budgetPolicies := fields["budget_policies"]
	logPolicies := fields["log_policies"]
	delete(fields, "model_policies")
	delete(fields, "budget_policies")
	delete(fields, "log_policies")

	base, err := json.Marshal(fields)
	if err != nil {
		return err
	}
	type alias GatewayPolicy
	var decoded alias
	if err := json.Unmarshal(base, &decoded); err != nil {
		return err
	}
	*p = GatewayPolicy(decoded)

	if err := decodePolicyList(modelPolicies, &p.ModelPolicies); err != nil {
		return err
	}
	if err := decodePolicyList(budgetPolicies, &p.BudgetPolicies); err != nil {
		return err
	}
	return decodePolicyList(logPolicies, &p.LogPolicies)
}

func decodePolicyList[T any](raw json.RawMessage, target *[]T) error {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
		return nil
	}
	if raw[0] != '[' {
		// Legacy object form has no lossless mapping to the current item list.
		return nil
	}
	return json.Unmarshal(raw, target)
}

// ModelPolicy describes model allow, block, or redirect behavior.
type ModelPolicy struct {
	ID         int      `json:"id"`
	Provider   string   `json:"provider"`
	ModelID    string   `json:"model_id"`
	Action     string   `json:"action"`
	RedirectTo *string  `json:"redirect_to,omitempty"`
	Priority   int      `json:"priority"`
	Scope      string   `json:"scope"`
	AppliesTo  []string `json:"applies_to,omitempty"`
	Reason     string   `json:"reason,omitempty"`
}

// KeywordPolicy describes one prompt keyword policy.
type KeywordPolicy struct {
	ID          int    `json:"id"`
	Keyword     string `json:"keyword"`
	Action      string `json:"action"`
	Enabled     bool   `json:"enabled"`
	Scope       string `json:"scope"`
	Description string `json:"description,omitempty"`
}

// BudgetPolicy describes one budget enforcement rule.
type BudgetPolicy struct {
	ID              int     `json:"id"`
	Scope           string  `json:"scope"`
	ScopeID         int64   `json:"scope_id"`
	MonthlyLimit    float64 `json:"monthly_limit"`
	CurrentSpend    float64 `json:"current_spend"`
	EnforcementMode string  `json:"enforcement_mode"`
}

// LogPolicy describes request and response capture behavior.
type LogPolicy struct {
	ID              int    `json:"id"`
	Scope           string `json:"scope"`
	CaptureRequest  bool   `json:"capture_request"`
	CaptureResponse bool   `json:"capture_response"`
	RetentionDays   int    `json:"retention_days"`
}

// PolicyRateLimit describes one gateway rate-limit policy.
type PolicyRateLimit struct {
	ID              int    `json:"id"`
	Scope           string `json:"scope"`
	RequestsPerMin  int    `json:"requests_per_min,omitempty"`
	TokensPerMin    int    `json:"tokens_per_min,omitempty"`
	EnforcementMode string `json:"enforcement_mode"`
}

// PreviewPolicyRequest previews policy effect on a request.
type PreviewPolicyRequest struct {
	WorkspaceID int64          `json:"workspace_id"`
	KeyID       int64          `json:"key_id,omitempty"`
	Model       string         `json:"model"`
	Request     map[string]any `json:"request,omitempty"`

	// Provider and Metadata are legacy aliases omitted from the current wire
	// contract because the Open API rejects unknown fields.
	Provider string            `json:"-"`
	Metadata map[string]string `json:"-"`
}

// PreviewPolicyResponse describes policy preview result.
type PreviewPolicyResponse struct {
	Allowed        bool     `json:"allowed"`
	BlockedBy      []string `json:"blocked_by"`
	ModifiedBy     []string `json:"modified_by"`
	RedirectedTo   *string  `json:"redirected_to,omitempty"`
	BudgetCheck    string   `json:"budget_check"`
	RateLimitCheck string   `json:"rate_limit_check"`
	Warnings       []string `json:"warnings"`

	// Reason is retained for pre-refactor responses.
	Reason string `json:"reason,omitempty"`
}

// UpdateGatewayPolicyRequest updates selected gateway policy fields.
type UpdateGatewayPolicyRequest struct {
	Action     *string `json:"action,omitempty"`
	RedirectTo *string `json:"redirect_to,omitempty"`
	Priority   *int    `json:"priority,omitempty"`
	Enabled    *bool   `json:"enabled,omitempty"`
}

// AddPromptKeywordRequest creates a case-insensitive workspace prompt restriction.
type AddPromptKeywordRequest struct {
	WorkspaceID int64  `json:"workspace_id"`
	Keyword     string `json:"keyword"`
	Description string `json:"description,omitempty"`
	Enabled     *bool  `json:"enabled,omitempty"`
}

// GetGatewayPolicies retrieves policies for a gateway key.
func (c *Client) GetGatewayPolicies(ctx context.Context, keyID int64, reqOpts ...owlvigil.RequestOption) (*GatewayPolicy, *owlvigil.ResponseMeta, error) {
	var out GatewayPolicy
	q := url.Values{}
	if keyID > 0 {
		q.Set("key_id", strconv.FormatInt(keyID, 10))
	}
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

// AddPromptKeyword creates a workspace prompt keyword policy.
func (c *Client) AddPromptKeyword(ctx context.Context, req *AddPromptKeywordRequest, reqOpts ...owlvigil.RequestOption) (*GatewayPolicy, *owlvigil.ResponseMeta, error) {
	var out GatewayPolicy
	meta, err := c.http.Do(ctx, http.MethodPost, "/gateway/policies/keywords", nil, req, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// DeletePromptKeyword deletes a workspace prompt keyword policy.
func (c *Client) DeletePromptKeyword(ctx context.Context, workspaceID, keywordID int64, reqOpts ...owlvigil.RequestOption) (*GatewayPolicy, *owlvigil.ResponseMeta, error) {
	var out GatewayPolicy
	q := url.Values{"workspace_id": {strconv.FormatInt(workspaceID, 10)}}
	path := "/gateway/policies/keywords/" + strconv.FormatInt(keywordID, 10)
	meta, err := c.http.Do(ctx, http.MethodDelete, path, q, nil, &out, reqOpts...)
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
