package management

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"

	owlvigil "github.com/Syrovex/owlvigil_sdk_go"
)

// FinancialGovernance describes complete financial control configuration.
type FinancialGovernance struct {
	WorkspaceID          int             `json:"workspace_id"`
	WorkspaceCap         BudgetCap       `json:"workspace_cap"`
	TeamCaps             []BudgetCap     `json:"team_caps"`
	MemberCaps           []BudgetCap     `json:"member_caps"`
	GatewayKeyCaps       []BudgetCap     `json:"gateway_key_caps"`
	MemberLimits         []SpendingLimit `json:"member_limits"`
	Thresholds           *Thresholds     `json:"thresholds"`
	ExceededAction       string          `json:"exceeded_action"`
	Currency             string          `json:"currency"`
	NotificationChannels []string        `json:"notification_channels"`

	// Legacy aliases retained for source compatibility.
	BudgetCaps     *BudgetCaps     `json:"-"`
	SpendingLimits []SpendingLimit `json:"-"`
}

// UnmarshalJSON decodes the current governance response and synchronizes the
// pre-refactor aggregate aliases.
func (g *FinancialGovernance) UnmarshalJSON(data []byte) error {
	type alias FinancialGovernance
	var out alias
	if err := json.Unmarshal(data, &out); err != nil {
		return err
	}
	*g = FinancialGovernance(out)
	g.BudgetCaps = budgetCapsFromGovernance(g)
	g.SpendingLimits = append([]SpendingLimit(nil), g.MemberLimits...)
	return nil
}

// BudgetCaps describes budget caps at various scopes.
type BudgetCaps struct {
	WorkspaceID          int             `json:"workspace_id"`
	WorkspaceCap         BudgetCap       `json:"workspace_cap"`
	TeamCaps             []BudgetCap     `json:"team_caps"`
	MemberCaps           []BudgetCap     `json:"member_caps"`
	GatewayKeyCaps       []BudgetCap     `json:"gateway_key_caps"`
	MemberLimits         []SpendingLimit `json:"member_limits"`
	Thresholds           Thresholds      `json:"thresholds"`
	ExceededAction       string          `json:"exceeded_action"`
	Currency             string          `json:"currency"`
	NotificationChannels []string        `json:"notification_channels"`

	// Legacy aliases retained for source compatibility.
	Workspace   *BudgetCap            `json:"-"`
	Teams       map[string]*BudgetCap `json:"-"`
	Members     map[string]*BudgetCap `json:"-"`
	GatewayKeys map[string]*BudgetCap `json:"-"`
}

// UnmarshalJSON accepts the current Governance response returned by the
// budget-caps endpoints and the older nested-map representation.
func (b *BudgetCaps) UnmarshalJSON(data []byte) error {
	type alias BudgetCaps
	var current alias
	if err := json.Unmarshal(data, &current); err != nil {
		return err
	}
	*b = BudgetCaps(current)

	var legacy struct {
		Workspace   *BudgetCap            `json:"workspace,omitempty"`
		Teams       map[string]*BudgetCap `json:"teams,omitempty"`
		Members     map[string]*BudgetCap `json:"members,omitempty"`
		GatewayKeys map[string]*BudgetCap `json:"gateway_keys,omitempty"`
	}
	if err := json.Unmarshal(data, &legacy); err != nil {
		return err
	}
	if b.WorkspaceCap.ScopeType == "" && legacy.Workspace != nil {
		b.WorkspaceCap = *legacy.Workspace
	}
	if len(b.TeamCaps) == 0 {
		b.TeamCaps = budgetCapMapValues(legacy.Teams)
	}
	if len(b.MemberCaps) == 0 {
		b.MemberCaps = budgetCapMapValues(legacy.Members)
	}
	if len(b.GatewayKeyCaps) == 0 {
		b.GatewayKeyCaps = budgetCapMapValues(legacy.GatewayKeys)
	}
	b.syncLegacyAliases()
	return nil
}

func budgetCapsFromGovernance(g *FinancialGovernance) *BudgetCaps {
	b := &BudgetCaps{
		WorkspaceID:          g.WorkspaceID,
		WorkspaceCap:         g.WorkspaceCap,
		TeamCaps:             append([]BudgetCap(nil), g.TeamCaps...),
		MemberCaps:           append([]BudgetCap(nil), g.MemberCaps...),
		GatewayKeyCaps:       append([]BudgetCap(nil), g.GatewayKeyCaps...),
		MemberLimits:         append([]SpendingLimit(nil), g.MemberLimits...),
		Thresholds:           Thresholds{},
		ExceededAction:       g.ExceededAction,
		Currency:             g.Currency,
		NotificationChannels: append([]string(nil), g.NotificationChannels...),
	}
	if g.Thresholds != nil {
		b.Thresholds = *g.Thresholds
	}
	b.syncLegacyAliases()
	return b
}

func (b *BudgetCaps) syncLegacyAliases() {
	b.Workspace = &b.WorkspaceCap
	b.Teams = budgetCapSliceMap(b.TeamCaps)
	b.Members = budgetCapSliceMap(b.MemberCaps)
	b.GatewayKeys = budgetCapSliceMap(b.GatewayKeyCaps)
}

func budgetCapMapValues(values map[string]*BudgetCap) []BudgetCap {
	out := make([]BudgetCap, 0, len(values))
	for _, value := range values {
		if value != nil {
			out = append(out, *value)
		}
	}
	return out
}

func budgetCapSliceMap(values []BudgetCap) map[string]*BudgetCap {
	out := make(map[string]*BudgetCap, len(values))
	for i := range values {
		value := values[i]
		key := strconv.Itoa(i)
		if value.ScopeID != nil {
			key = strconv.FormatInt(*value.ScopeID, 10)
		}
		out[key] = &value
	}
	return out
}

// BudgetCap describes a budget cap for a scope.
type BudgetCap struct {
	ScopeType     string  `json:"scope_type"`
	ScopeID       *int64  `json:"scope_id,omitempty"`
	TeamID        *int64  `json:"team_id,omitempty"`
	UserID        *int    `json:"user_id,omitempty"`
	APIKeyID      *int    `json:"api_key_id,omitempty"`
	Name          string  `json:"name,omitempty"`
	Email         string  `json:"email,omitempty"`
	Enabled       bool    `json:"enabled"`
	MonthlyAmount float64 `json:"monthly_amount"`
	CurrentSpend  float64 `json:"current_spend,omitempty"`

	// Legacy aliases retained for source compatibility.
	Limit    float64 `json:"-"`
	Used     float64 `json:"-"`
	Currency string  `json:"-"`
}

// UnmarshalJSON synchronizes current and legacy cap values.
func (b *BudgetCap) UnmarshalJSON(data []byte) error {
	type alias BudgetCap
	var out alias
	if err := json.Unmarshal(data, &out); err != nil {
		return err
	}
	*b = BudgetCap(out)
	var legacy struct {
		Limit    float64 `json:"limit"`
		Used     float64 `json:"used"`
		Currency string  `json:"currency"`
	}
	if err := json.Unmarshal(data, &legacy); err != nil {
		return err
	}
	if b.MonthlyAmount == 0 {
		b.MonthlyAmount = legacy.Limit
	}
	if b.CurrentSpend == 0 {
		b.CurrentSpend = legacy.Used
	}
	b.Limit = b.MonthlyAmount
	b.Used = b.CurrentSpend
	b.Currency = legacy.Currency
	return nil
}

// MarshalJSON emits the current Open API cap fields.
func (b BudgetCap) MarshalJSON() ([]byte, error) {
	monthlyAmount := b.MonthlyAmount
	if monthlyAmount == 0 {
		monthlyAmount = b.Limit
	}
	type wireCap struct {
		ScopeType     string  `json:"scope_type"`
		ScopeID       *int64  `json:"scope_id,omitempty"`
		TeamID        *int64  `json:"team_id,omitempty"`
		UserID        *int    `json:"user_id,omitempty"`
		APIKeyID      *int    `json:"api_key_id,omitempty"`
		Name          string  `json:"name,omitempty"`
		Email         string  `json:"email,omitempty"`
		Enabled       bool    `json:"enabled"`
		MonthlyAmount float64 `json:"monthly_amount"`
		CurrentSpend  float64 `json:"current_spend,omitempty"`
	}
	return json.Marshal(wireCap{
		ScopeType:     b.ScopeType,
		ScopeID:       b.ScopeID,
		TeamID:        b.TeamID,
		UserID:        b.UserID,
		APIKeyID:      b.APIKeyID,
		Name:          b.Name,
		Email:         b.Email,
		Enabled:       b.Enabled,
		MonthlyAmount: monthlyAmount,
		CurrentSpend:  b.CurrentSpend,
	})
}

// SpendingLimit describes spending limits for a user.
type SpendingLimit struct {
	UserID       int64   `json:"user_id"`
	Email        string  `json:"email,omitempty"`
	Name         string  `json:"name,omitempty"`
	DailyLimit   float64 `json:"daily_limit,omitempty"`
	WeeklyLimit  float64 `json:"weekly_limit,omitempty"`
	MonthlyLimit float64 `json:"monthly_limit,omitempty"`
	DailySpend   float64 `json:"daily_spend,omitempty"`
	WeeklySpend  float64 `json:"weekly_spend,omitempty"`
	MonthlySpend float64 `json:"monthly_spend,omitempty"`

	// Currency is retained for the old SDK shape and omitted from requests.
	Currency string `json:"-"`
}

// Thresholds describes financial alert thresholds.
type Thresholds struct {
	WarningPercent  float64 `json:"warning_percent"`
	CriticalPercent float64 `json:"critical_percent"`
	ExceededAction  string  `json:"exceeded_action"`
}

// UpdateFinancialGovernanceRequest updates financial governance config.
type UpdateFinancialGovernanceRequest struct {
	WorkspaceCap   *BudgetCap      `json:"workspace_cap,omitempty"`
	TeamCaps       []BudgetCap     `json:"team_caps,omitempty"`
	MemberCaps     []BudgetCap     `json:"member_caps,omitempty"`
	GatewayKeyCaps []BudgetCap     `json:"gateway_key_caps,omitempty"`
	MemberLimits   []SpendingLimit `json:"member_limits,omitempty"`
	Thresholds     *Thresholds     `json:"thresholds,omitempty"`
	ExceededAction *string         `json:"exceeded_action,omitempty"`

	// Legacy aliases retained for source compatibility.
	BudgetCaps     *BudgetCaps     `json:"-"`
	SpendingLimits []SpendingLimit `json:"-"`
}

// MarshalJSON emits the current full-governance update contract.
func (r UpdateFinancialGovernanceRequest) MarshalJSON() ([]byte, error) {
	workspaceCap, teamCaps, memberCaps, gatewayKeyCaps := requestCaps(
		r.WorkspaceCap,
		r.TeamCaps,
		r.MemberCaps,
		r.GatewayKeyCaps,
		r.BudgetCaps,
	)
	memberLimits := r.MemberLimits
	if len(memberLimits) == 0 {
		memberLimits = r.SpendingLimits
	}
	type wireRequest struct {
		WorkspaceCap   *BudgetCap      `json:"workspace_cap,omitempty"`
		TeamCaps       []BudgetCap     `json:"team_caps,omitempty"`
		MemberCaps     []BudgetCap     `json:"member_caps,omitempty"`
		GatewayKeyCaps []BudgetCap     `json:"gateway_key_caps,omitempty"`
		MemberLimits   []SpendingLimit `json:"member_limits,omitempty"`
		Thresholds     *Thresholds     `json:"thresholds,omitempty"`
		ExceededAction *string         `json:"exceeded_action,omitempty"`
	}
	return json.Marshal(wireRequest{
		WorkspaceCap:   workspaceCap,
		TeamCaps:       teamCaps,
		MemberCaps:     memberCaps,
		GatewayKeyCaps: gatewayKeyCaps,
		MemberLimits:   memberLimits,
		Thresholds:     r.Thresholds,
		ExceededAction: r.ExceededAction,
	})
}

// UpdateBudgetCapsRequest updates budget caps.
type UpdateBudgetCapsRequest struct {
	WorkspaceCap   *BudgetCap  `json:"workspace_cap,omitempty"`
	TeamCaps       []BudgetCap `json:"team_caps,omitempty"`
	MemberCaps     []BudgetCap `json:"member_caps,omitempty"`
	GatewayKeyCaps []BudgetCap `json:"gateway_key_caps,omitempty"`

	// Legacy aliases retained for source compatibility.
	Workspace   *BudgetCap            `json:"-"`
	Teams       map[string]*BudgetCap `json:"-"`
	Members     map[string]*BudgetCap `json:"-"`
	GatewayKeys map[string]*BudgetCap `json:"-"`
}

// MarshalJSON emits the current budget-cap replacement contract.
func (r UpdateBudgetCapsRequest) MarshalJSON() ([]byte, error) {
	workspaceCap := r.WorkspaceCap
	if workspaceCap == nil {
		workspaceCap = r.Workspace
	}
	teamCaps := r.TeamCaps
	if len(teamCaps) == 0 {
		teamCaps = budgetCapMapValues(r.Teams)
	}
	memberCaps := r.MemberCaps
	if len(memberCaps) == 0 {
		memberCaps = budgetCapMapValues(r.Members)
	}
	gatewayKeyCaps := r.GatewayKeyCaps
	if len(gatewayKeyCaps) == 0 {
		gatewayKeyCaps = budgetCapMapValues(r.GatewayKeys)
	}
	return json.Marshal(struct {
		WorkspaceCap   *BudgetCap  `json:"workspace_cap,omitempty"`
		TeamCaps       []BudgetCap `json:"team_caps,omitempty"`
		MemberCaps     []BudgetCap `json:"member_caps,omitempty"`
		GatewayKeyCaps []BudgetCap `json:"gateway_key_caps,omitempty"`
	}{
		WorkspaceCap:   workspaceCap,
		TeamCaps:       teamCaps,
		MemberCaps:     memberCaps,
		GatewayKeyCaps: gatewayKeyCaps,
	})
}

// UpdateScopeBudgetCapRequest updates a single scope's budget cap.
type UpdateScopeBudgetCapRequest struct {
	Enabled       *bool    `json:"enabled,omitempty"`
	MonthlyAmount *float64 `json:"monthly_amount,omitempty"`

	// Limit is the pre-refactor alias for MonthlyAmount.
	Limit float64 `json:"-"`
}

// MarshalJSON emits the current scoped-cap update contract.
func (r UpdateScopeBudgetCapRequest) MarshalJSON() ([]byte, error) {
	monthlyAmount := r.MonthlyAmount
	if monthlyAmount == nil && r.Limit != 0 {
		monthlyAmount = &r.Limit
	}
	return json.Marshal(struct {
		Enabled       *bool    `json:"enabled,omitempty"`
		MonthlyAmount *float64 `json:"monthly_amount,omitempty"`
	}{
		Enabled:       r.Enabled,
		MonthlyAmount: monthlyAmount,
	})
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
	WarningPercent  *float64 `json:"warning_percent,omitempty"`
	CriticalPercent *float64 `json:"critical_percent,omitempty"`
	ExceededAction  *string  `json:"exceeded_action,omitempty"`
}

// PreviewFinancialChangesRequest previews financial config changes.
type PreviewFinancialChangesRequest struct {
	WorkspaceCap   *BudgetCap  `json:"workspace_cap,omitempty"`
	TeamCaps       []BudgetCap `json:"team_caps,omitempty"`
	MemberCaps     []BudgetCap `json:"member_caps,omitempty"`
	GatewayKeyCaps []BudgetCap `json:"gateway_key_caps,omitempty"`

	// Legacy aliases retained for source compatibility.
	BudgetCaps     *BudgetCaps     `json:"-"`
	SpendingLimits []SpendingLimit `json:"-"`
}

// MarshalJSON emits the current financial-preview contract.
func (r PreviewFinancialChangesRequest) MarshalJSON() ([]byte, error) {
	workspaceCap, teamCaps, memberCaps, gatewayKeyCaps := requestCaps(
		r.WorkspaceCap,
		r.TeamCaps,
		r.MemberCaps,
		r.GatewayKeyCaps,
		r.BudgetCaps,
	)
	return json.Marshal(struct {
		WorkspaceCap   *BudgetCap  `json:"workspace_cap,omitempty"`
		TeamCaps       []BudgetCap `json:"team_caps,omitempty"`
		MemberCaps     []BudgetCap `json:"member_caps,omitempty"`
		GatewayKeyCaps []BudgetCap `json:"gateway_key_caps,omitempty"`
	}{
		WorkspaceCap:   workspaceCap,
		TeamCaps:       teamCaps,
		MemberCaps:     memberCaps,
		GatewayKeyCaps: gatewayKeyCaps,
	})
}

func requestCaps(
	workspaceCap *BudgetCap,
	teamCaps []BudgetCap,
	memberCaps []BudgetCap,
	gatewayKeyCaps []BudgetCap,
	legacy *BudgetCaps,
) (*BudgetCap, []BudgetCap, []BudgetCap, []BudgetCap) {
	if legacy == nil {
		return workspaceCap, teamCaps, memberCaps, gatewayKeyCaps
	}
	if workspaceCap == nil {
		if legacy.WorkspaceCap.ScopeType != "" {
			workspaceCap = &legacy.WorkspaceCap
		} else {
			workspaceCap = legacy.Workspace
		}
	}
	if len(teamCaps) == 0 {
		teamCaps = legacy.TeamCaps
		if len(teamCaps) == 0 {
			teamCaps = budgetCapMapValues(legacy.Teams)
		}
	}
	if len(memberCaps) == 0 {
		memberCaps = legacy.MemberCaps
		if len(memberCaps) == 0 {
			memberCaps = budgetCapMapValues(legacy.Members)
		}
	}
	if len(gatewayKeyCaps) == 0 {
		gatewayKeyCaps = legacy.GatewayKeyCaps
		if len(gatewayKeyCaps) == 0 {
			gatewayKeyCaps = budgetCapMapValues(legacy.GatewayKeys)
		}
	}
	return workspaceCap, teamCaps, memberCaps, gatewayKeyCaps
}

// PreviewFinancialChangesResponse describes impact of changes.
type PreviewFinancialChangesResponse struct {
	Valid               bool     `json:"valid"`
	Violations          []string `json:"violations,omitempty"`
	AffectedTeams       []int64  `json:"affected_teams,omitempty"`
	AffectedMembers     []int    `json:"affected_members,omitempty"`
	AffectedGatewayKeys []int    `json:"affected_gateway_keys,omitempty"`

	// AffectedKeys is the pre-refactor alias.
	AffectedKeys []int64 `json:"-"`
}

// SpendSummary describes spending summary by scope.
type SpendSummary struct {
	WorkspaceID      int           `json:"workspace_id"`
	CurrentPeriod    string        `json:"current_period"`
	WorkspaceSpend   SpendDetail   `json:"workspace_spend"`
	TeamSpends       []SpendDetail `json:"team_spends,omitempty"`
	MemberSpends     []SpendDetail `json:"member_spends,omitempty"`
	GatewayKeySpends []SpendDetail `json:"gateway_key_spends,omitempty"`

	// Legacy aliases retained for source compatibility.
	Workspace   *ScopeSummary            `json:"-"`
	Teams       map[string]*ScopeSummary `json:"-"`
	Members     map[string]*ScopeSummary `json:"-"`
	GatewayKeys map[string]*ScopeSummary `json:"-"`
}

// SpendDetail describes spending for a scope.
type SpendDetail struct {
	ScopeType string  `json:"scope_type"`
	ScopeID   int64   `json:"scope_id,omitempty"`
	Name      string  `json:"name,omitempty"`
	Spend     float64 `json:"spend"`
	Limit     float64 `json:"limit,omitempty"`
	Percent   float64 `json:"percent,omitempty"`
	Status    string  `json:"status"`
}

// ScopeSummary is the pre-refactor spend detail alias.
type ScopeSummary struct {
	ScopeType string  `json:"scope_type"`
	ScopeID   string  `json:"scope_id"`
	Spent     float64 `json:"spent"`
	Limit     float64 `json:"limit,omitempty"`
	Currency  string  `json:"currency"`
}

// UnmarshalJSON decodes the current spend summary and builds legacy aliases.
func (s *SpendSummary) UnmarshalJSON(data []byte) error {
	type alias SpendSummary
	var out alias
	if err := json.Unmarshal(data, &out); err != nil {
		return err
	}
	*s = SpendSummary(out)
	var legacy struct {
		Workspace   *ScopeSummary            `json:"workspace,omitempty"`
		Teams       map[string]*ScopeSummary `json:"teams,omitempty"`
		Members     map[string]*ScopeSummary `json:"members,omitempty"`
		GatewayKeys map[string]*ScopeSummary `json:"gateway_keys,omitempty"`
	}
	if err := json.Unmarshal(data, &legacy); err != nil {
		return err
	}
	if s.WorkspaceSpend.ScopeType == "" && legacy.Workspace != nil {
		s.Workspace = legacy.Workspace
		s.Teams = legacy.Teams
		s.Members = legacy.Members
		s.GatewayKeys = legacy.GatewayKeys
		s.WorkspaceSpend = SpendDetail{
			ScopeType: legacy.Workspace.ScopeType,
			Spend:     legacy.Workspace.Spent,
			Limit:     legacy.Workspace.Limit,
		}
		return nil
	}
	workspace := scopeSummaryFromSpendDetail(s.WorkspaceSpend)
	s.Workspace = &workspace
	s.Teams = scopeSummaryMap(s.TeamSpends)
	s.Members = scopeSummaryMap(s.MemberSpends)
	s.GatewayKeys = scopeSummaryMap(s.GatewayKeySpends)
	return nil
}

func scopeSummaryFromSpendDetail(value SpendDetail) ScopeSummary {
	return ScopeSummary{
		ScopeType: value.ScopeType,
		ScopeID:   strconv.FormatInt(value.ScopeID, 10),
		Spent:     value.Spend,
		Limit:     value.Limit,
	}
}

func scopeSummaryMap(values []SpendDetail) map[string]*ScopeSummary {
	out := make(map[string]*ScopeSummary, len(values))
	for _, value := range values {
		summary := scopeSummaryFromSpendDetail(value)
		out[summary.ScopeID] = &summary
	}
	return out
}

// SpendingLimitOptions contains the filters accepted by the Open API.
type SpendingLimitOptions struct {
	TeamID int64
	UserID int64
}

func (o SpendingLimitOptions) values() url.Values {
	query := url.Values{}
	if o.TeamID > 0 {
		query.Set("team_id", strconv.FormatInt(o.TeamID, 10))
	}
	if o.UserID > 0 {
		query.Set("user_id", strconv.FormatInt(o.UserID, 10))
	}
	return query
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
	// The refactored Open API no longer accepts cursor or limit on this route.
	_ = opts
	meta, err := c.http.Do(ctx, http.MethodGet, path, nil, nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// GetSpendingLimitsWithFilters retrieves spending limits filtered by team
// and/or user using the query parameters declared by the Open API.
func (c *Client) GetSpendingLimitsWithFilters(ctx context.Context, workspaceID int64, opts SpendingLimitOptions, reqOpts ...owlvigil.RequestOption) (*ListResponse[SpendingLimit], *owlvigil.ResponseMeta, error) {
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
