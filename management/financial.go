package management

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	owlvigil "github.com/owlvigil/owlvigil-go"
)

// FinancialGovernance describes complete financial control configuration.
type FinancialGovernance struct {
	BudgetCaps     *BudgetCaps     `json:"budget_caps,omitempty"`
	SpendingLimits []SpendingLimit `json:"spending_limits,omitempty"`
	Thresholds     *Thresholds     `json:"thresholds,omitempty"`
}

// BudgetCaps describes budget caps at various scopes.
type BudgetCaps struct {
	Workspace   *BudgetCap            `json:"workspace,omitempty"`
	Teams       map[string]*BudgetCap `json:"teams,omitempty"`
	Members     map[string]*BudgetCap `json:"members,omitempty"`
	GatewayKeys map[string]*BudgetCap `json:"gateway_keys,omitempty"`
}

// BudgetCap describes a budget cap for a scope.
type BudgetCap struct {
	ScopeType string  `json:"scope_type"`
	ScopeID   string  `json:"scope_id"`
	Limit     float64 `json:"limit"`
	Used      float64 `json:"used"`
	Currency  string  `json:"currency"`
}

// SpendingLimit describes spending limits for a user.
type SpendingLimit struct {
	UserID       int64   `json:"user_id"`
	DailyLimit   float64 `json:"daily_limit,omitempty"`
	WeeklyLimit  float64 `json:"weekly_limit,omitempty"`
	MonthlyLimit float64 `json:"monthly_limit,omitempty"`
	Currency     string  `json:"currency"`
}

// Thresholds describes financial alert thresholds.
type Thresholds struct {
	WarningPercent  int    `json:"warning_percent"`
	CriticalPercent int    `json:"critical_percent"`
	ExceededAction  string `json:"exceeded_action"`
}

// UpdateFinancialGovernanceRequest updates financial governance config.
type UpdateFinancialGovernanceRequest struct {
	BudgetCaps     *BudgetCaps     `json:"budget_caps,omitempty"`
	SpendingLimits []SpendingLimit `json:"spending_limits,omitempty"`
	Thresholds     *Thresholds     `json:"thresholds,omitempty"`
}

// UpdateBudgetCapsRequest updates budget caps.
type UpdateBudgetCapsRequest struct {
	Workspace   *BudgetCap            `json:"workspace,omitempty"`
	Teams       map[string]*BudgetCap `json:"teams,omitempty"`
	Members     map[string]*BudgetCap `json:"members,omitempty"`
	GatewayKeys map[string]*BudgetCap `json:"gateway_keys,omitempty"`
}

// UpdateScopeBudgetCapRequest updates a single scope's budget cap.
type UpdateScopeBudgetCapRequest struct {
	Limit float64 `json:"limit"`
}

// UpdateSpendingLimitsRequest updates spending limits.
type UpdateSpendingLimitsRequest struct {
	Limits []SpendingLimit `json:"limits"`
}

// UpdateUserSpendingLimitRequest updates a user's spending limit.
type UpdateUserSpendingLimitRequest struct {
	DailyLimit   *float64 `json:"daily_limit,omitempty"`
	WeeklyLimit  *float64 `json:"weekly_limit,omitempty"`
	MonthlyLimit *float64 `json:"monthly_limit,omitempty"`
}

// UpdateThresholdsRequest updates financial thresholds.
type UpdateThresholdsRequest struct {
	WarningPercent  *int    `json:"warning_percent,omitempty"`
	CriticalPercent *int    `json:"critical_percent,omitempty"`
	ExceededAction  *string `json:"exceeded_action,omitempty"`
}

// PreviewFinancialChangesRequest previews financial config changes.
type PreviewFinancialChangesRequest struct {
	BudgetCaps     *BudgetCaps     `json:"budget_caps,omitempty"`
	SpendingLimits []SpendingLimit `json:"spending_limits,omitempty"`
}

// PreviewFinancialChangesResponse describes impact of changes.
type PreviewFinancialChangesResponse struct {
	Valid           bool     `json:"valid"`
	Violations      []string `json:"violations,omitempty"`
	AffectedTeams   []int64  `json:"affected_teams,omitempty"`
	AffectedMembers []int64  `json:"affected_members,omitempty"`
	AffectedKeys    []int64  `json:"affected_keys,omitempty"`
}

// SpendSummary describes spending summary by scope.
type SpendSummary struct {
	Workspace   *ScopeSummary            `json:"workspace,omitempty"`
	Teams       map[string]*ScopeSummary `json:"teams,omitempty"`
	Members     map[string]*ScopeSummary `json:"members,omitempty"`
	GatewayKeys map[string]*ScopeSummary `json:"gateway_keys,omitempty"`
}

// ScopeSummary describes spending for a scope.
type ScopeSummary struct {
	ScopeType string  `json:"scope_type"`
	ScopeID   string  `json:"scope_id"`
	Spent     float64 `json:"spent"`
	Limit     float64 `json:"limit,omitempty"`
	Currency  string  `json:"currency"`
}

// GetFinancialGovernance retrieves complete financial control configuration.
func (c *Client) GetFinancialGovernance(ctx context.Context, workspaceID int64, reqOpts ...owlvigil.RequestOption) (*FinancialGovernance, *owlvigil.ResponseMeta, error) {
	var out FinancialGovernance
	path := "/workspaces/" + strconv.FormatInt(workspaceID, 10) + "/governance/financial"
	meta, err := c.http.Do(ctx, http.MethodGet, path, nil, nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// UpdateFinancialGovernance updates financial control configuration.
func (c *Client) UpdateFinancialGovernance(ctx context.Context, workspaceID int64, req *UpdateFinancialGovernanceRequest, reqOpts ...owlvigil.RequestOption) (*FinancialGovernance, *owlvigil.ResponseMeta, error) {
	var out FinancialGovernance
	path := "/workspaces/" + strconv.FormatInt(workspaceID, 10) + "/governance/financial"
	meta, err := c.http.Do(ctx, http.MethodPut, path, nil, req, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// GetBudgetCaps retrieves budget caps.
func (c *Client) GetBudgetCaps(ctx context.Context, workspaceID int64, reqOpts ...owlvigil.RequestOption) (*BudgetCaps, *owlvigil.ResponseMeta, error) {
	var out BudgetCaps
	path := "/workspaces/" + strconv.FormatInt(workspaceID, 10) + "/governance/financial/budget-caps"
	meta, err := c.http.Do(ctx, http.MethodGet, path, nil, nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// UpdateBudgetCaps updates budget caps.
func (c *Client) UpdateBudgetCaps(ctx context.Context, workspaceID int64, req *UpdateBudgetCapsRequest, reqOpts ...owlvigil.RequestOption) (*BudgetCaps, *owlvigil.ResponseMeta, error) {
	var out BudgetCaps
	path := "/workspaces/" + strconv.FormatInt(workspaceID, 10) + "/governance/financial/budget-caps"
	meta, err := c.http.Do(ctx, http.MethodPut, path, nil, req, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// UpdateScopeBudgetCap updates a specific scope's budget cap.
func (c *Client) UpdateScopeBudgetCap(ctx context.Context, workspaceID int64, scopeType, scopeID string, req *UpdateScopeBudgetCapRequest, reqOpts ...owlvigil.RequestOption) (*BudgetCap, *owlvigil.ResponseMeta, error) {
	var out BudgetCap
	path := "/workspaces/" + strconv.FormatInt(workspaceID, 10) + "/governance/financial/budget-caps/" + url.PathEscape(scopeType) + "/" + url.PathEscape(scopeID)
	meta, err := c.http.Do(ctx, http.MethodPatch, path, nil, req, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// GetSpendingLimits retrieves spending limits.
func (c *Client) GetSpendingLimits(ctx context.Context, workspaceID int64, opts ListOptions, reqOpts ...owlvigil.RequestOption) (*ListResponse[SpendingLimit], *owlvigil.ResponseMeta, error) {
	var out ListResponse[SpendingLimit]
	path := "/workspaces/" + strconv.FormatInt(workspaceID, 10) + "/governance/financial/spending-limits"
	meta, err := c.http.Do(ctx, http.MethodGet, path, opts.values(), nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// UpdateSpendingLimits updates spending limits in batch.
func (c *Client) UpdateSpendingLimits(ctx context.Context, workspaceID int64, req *UpdateSpendingLimitsRequest, reqOpts ...owlvigil.RequestOption) (*ListResponse[SpendingLimit], *owlvigil.ResponseMeta, error) {
	var out ListResponse[SpendingLimit]
	path := "/workspaces/" + strconv.FormatInt(workspaceID, 10) + "/governance/financial/spending-limits"
	meta, err := c.http.Do(ctx, http.MethodPut, path, nil, req, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// UpdateUserSpendingLimit updates a user's spending limit.
func (c *Client) UpdateUserSpendingLimit(ctx context.Context, workspaceID, userID int64, req *UpdateUserSpendingLimitRequest, reqOpts ...owlvigil.RequestOption) (*SpendingLimit, *owlvigil.ResponseMeta, error) {
	var out SpendingLimit
	path := "/workspaces/" + strconv.FormatInt(workspaceID, 10) + "/governance/financial/spending-limits/users/" + strconv.FormatInt(userID, 10)
	meta, err := c.http.Do(ctx, http.MethodPatch, path, nil, req, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// GetFinancialThresholds retrieves financial thresholds.
func (c *Client) GetFinancialThresholds(ctx context.Context, workspaceID int64, reqOpts ...owlvigil.RequestOption) (*Thresholds, *owlvigil.ResponseMeta, error) {
	var out Thresholds
	path := "/workspaces/" + strconv.FormatInt(workspaceID, 10) + "/governance/financial/thresholds"
	meta, err := c.http.Do(ctx, http.MethodGet, path, nil, nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// UpdateFinancialThresholds updates financial thresholds.
func (c *Client) UpdateFinancialThresholds(ctx context.Context, workspaceID int64, req *UpdateThresholdsRequest, reqOpts ...owlvigil.RequestOption) (*Thresholds, *owlvigil.ResponseMeta, error) {
	var out Thresholds
	path := "/workspaces/" + strconv.FormatInt(workspaceID, 10) + "/governance/financial/thresholds"
	meta, err := c.http.Do(ctx, http.MethodPut, path, nil, req, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// PreviewFinancialChanges previews financial configuration changes.
func (c *Client) PreviewFinancialChanges(ctx context.Context, workspaceID int64, req *PreviewFinancialChangesRequest, reqOpts ...owlvigil.RequestOption) (*PreviewFinancialChangesResponse, *owlvigil.ResponseMeta, error) {
	var out PreviewFinancialChangesResponse
	path := "/workspaces/" + strconv.FormatInt(workspaceID, 10) + "/governance/financial/preview"
	meta, err := c.http.Do(ctx, http.MethodPost, path, nil, req, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// GetSpendSummary retrieves spending summary by scope.
func (c *Client) GetSpendSummary(ctx context.Context, workspaceID int64, reqOpts ...owlvigil.RequestOption) (*SpendSummary, *owlvigil.ResponseMeta, error) {
	var out SpendSummary
	path := "/workspaces/" + strconv.FormatInt(workspaceID, 10) + "/governance/financial/spend-summary"
	meta, err := c.http.Do(ctx, http.MethodGet, path, nil, nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}
