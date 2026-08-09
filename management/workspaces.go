package management

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"

	owlvigil "github.com/Syrovex/owlvigil_sdk_go"
)

// Workspace describes an OwlVigil workspace visible to the user.
type Workspace struct {
	ID          int64  `json:"id"`
	WorkspaceID int64  `json:"workspace_id,omitempty"`
	Name        string `json:"name"`
	Slug        string `json:"slug,omitempty"`
	Description string `json:"description,omitempty"`
	Type        string `json:"type,omitempty"`
	Status      string `json:"status,omitempty"`
	Region      string `json:"region,omitempty"`
	Role        string `json:"role,omitempty"`
	Plan        string `json:"plan,omitempty"`
	CreatedAt   string `json:"created_at,omitempty"`
	UpdatedAt   string `json:"updated_at,omitempty"`
}

// CreateWorkspaceRequest creates a workspace.
type CreateWorkspaceRequest struct {
	Name        string  `json:"name"`
	Slug        string  `json:"slug,omitempty"`
	Description *string `json:"description,omitempty"`
	Type        string  `json:"type,omitempty"`
	Region      string  `json:"region,omitempty"`
}

// UpdateWorkspaceRequest updates workspace fields.
type UpdateWorkspaceRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

// WorkspaceOverviewOptions filters workspace overview data.
type WorkspaceOverviewOptions struct {
	Range string
}

func (o WorkspaceOverviewOptions) values() url.Values {
	q := url.Values{}
	if o.Range != "" {
		q.Set("range", o.Range)
	}
	return q
}

// WorkspaceActivityOptions filters and paginates workspace activity.
type WorkspaceActivityOptions struct {
	Limit  int
	Offset int
	Tab    string
	Search string
}

func (o WorkspaceActivityOptions) values() url.Values {
	q := url.Values{}
	if o.Limit > 0 {
		q.Set("limit", intString(o.Limit))
	}
	if o.Offset > 0 {
		q.Set("offset", intString(o.Offset))
	}
	addFilter(q, "tab", o.Tab)
	addFilter(q, "search", o.Search)
	return q
}

// WorkspaceOverview contains the aggregate data displayed by Workspace Overview.
type WorkspaceOverview struct {
	Range          string                          `json:"range"`
	Summary        WorkspaceOverviewSummary        `json:"summary"`
	CostAnalysis   WorkspaceMetricAnalysis         `json:"cost_analysis"`
	UsageAnalysis  WorkspaceMetricAnalysis         `json:"usage_analysis"`
	Breakdown      WorkspaceBreakdown              `json:"breakdown"`
	RecentActivity WorkspaceOverviewRecentActivity `json:"recent_activity"`
}

// WorkspaceOverviewSummary contains the overview headline metrics.
type WorkspaceOverviewSummary struct {
	Cost         WorkspaceCostSummary     `json:"cost"`
	TokenUsage   WorkspaceTokenSummary    `json:"token_usage"`
	Resources    WorkspaceResourceSummary `json:"resources"`
	SystemStatus WorkspaceSystemStatus    `json:"system_status"`
}

// WorkspaceCostSummary contains current and weekly cost and budget values.
type WorkspaceCostSummary struct {
	Current       float64 `json:"current"`
	Budget        float64 `json:"budget"`
	Percent       int     `json:"percent"`
	WeeklyCurrent float64 `json:"weekly_current"`
	WeeklyBudget  float64 `json:"weekly_budget"`
}

// WorkspaceTokenSummary contains current and weekly token usage.
type WorkspaceTokenSummary struct {
	Current       int64  `json:"current"`
	Label         string `json:"label"`
	WeeklyCurrent int64  `json:"weekly_current"`
	WeeklyLabel   string `json:"weekly_label"`
}

// WorkspaceResourceSummary contains workspace resource counts.
type WorkspaceResourceSummary struct {
	ActiveKeys int `json:"active_keys"`
	Members    int `json:"members"`
	Providers  int `json:"providers"`
	Teams      int `json:"teams"`
}

// WorkspaceSystemStatus contains workspace health and latency values.
type WorkspaceSystemStatus struct {
	HealthPercent int     `json:"health_pct"`
	SuccessRate   float64 `json:"success_rate"`
	AvgLatencyMS  int64   `json:"avg_latency_ms"`
	Tone          string  `json:"tone"`
}

// WorkspaceMetricAnalysis contains total and dimensioned time series.
type WorkspaceMetricAnalysis struct {
	Total      []map[string]any       `json:"total"`
	ByTeam     []map[string]any       `json:"by_team"`
	ByMember   []map[string]any       `json:"by_member"`
	ByProvider []map[string]any       `json:"by_provider"`
	Series     WorkspaceSeriesCatalog `json:"series"`
}

// WorkspaceSeriesCatalog lists the dimensions represented in overview series.
type WorkspaceSeriesCatalog struct {
	Teams     []string `json:"teams"`
	Members   []string `json:"members"`
	Providers []string `json:"providers"`
}

// WorkspaceBreakdown contains usage and cost grouped by workspace dimensions.
type WorkspaceBreakdown struct {
	ByTeam     []WorkspaceBreakdownRow `json:"by_team"`
	ByMember   []WorkspaceBreakdownRow `json:"by_member"`
	ByProvider []WorkspaceBreakdownRow `json:"by_provider"`
}

// WorkspaceBreakdownRow contains totals for one workspace dimension.
type WorkspaceBreakdownRow struct {
	ID           int64   `json:"id,omitempty"`
	Name         string  `json:"name"`
	TokenUsed    int64   `json:"token_used"`
	TokenLabel   string  `json:"token_label"`
	Cost         float64 `json:"cost"`
	MemberCount  int     `json:"member_count,omitempty"`
	KeyCount     int     `json:"key_count,omitempty"`
	RequestCount int64   `json:"request_count,omitempty"`
}

// WorkspaceOverviewRecentActivity contains request and audit records.
type WorkspaceOverviewRecentActivity struct {
	RequestLogs        []WorkspaceOverviewRequestLog `json:"request_logs"`
	AuditLogsAvailable bool                          `json:"audit_logs_available"`
	AuditLogs          []WorkspaceOverviewAuditLog   `json:"audit_logs"`
}

// WorkspaceOverviewRequestLog is the request-log subset embedded in an overview.
type WorkspaceOverviewRequestLog struct {
	ID           int64   `json:"id"`
	CreatedAt    string  `json:"created_at"`
	Member       string  `json:"member,omitempty"`
	TraceID      string  `json:"trace_id,omitempty"`
	Team         string  `json:"team,omitempty"`
	KeyName      string  `json:"key_name"`
	Provider     string  `json:"provider"`
	Model        string  `json:"model"`
	Status       string  `json:"status"`
	StatusCode   int     `json:"status_code"`
	LatencyMS    *int64  `json:"latency_ms,omitempty"`
	InputTokens  int64   `json:"input_tokens"`
	OutputTokens int64   `json:"output_tokens"`
	TotalTokens  int64   `json:"total_tokens"`
	TotalCost    float64 `json:"total_cost"`
}

// WorkspaceOverviewAuditLog is the audit-log subset embedded in an overview.
type WorkspaceOverviewAuditLog struct {
	ID           int64  `json:"id"`
	CreatedAt    string `json:"created_at"`
	Action       string `json:"action"`
	ResourceType string `json:"resource_type"`
	ResourceName string `json:"resource_name"`
	UserName     string `json:"user_name"`
	Status       string `json:"status"`
	Before       string `json:"before,omitempty"`
	After        string `json:"after,omitempty"`
	IPAddress    string `json:"ip_address,omitempty"`
}

// ActivityRecord describes a workspace activity log entry.
type ActivityRecord struct {
	ID          string `json:"id"`
	WorkspaceID int64  `json:"workspace_id"`
	ActorID     int64  `json:"actor_id"`
	ActorEmail  string `json:"actor_email,omitempty"`
	Action      string `json:"action"`
	Resource    string `json:"resource"`
	ResourceID  string `json:"resource_id,omitempty"`
	Details     any    `json:"details,omitempty"`
	Timestamp   string `json:"timestamp"`
	Tab         string `json:"tab,omitempty"`
	Who         string `json:"who,omitempty"`
	What        string `json:"what,omitempty"`
	Meta        string `json:"meta,omitempty"`
	IP          string `json:"ip,omitempty"`
	CreatedAt   string `json:"created_at,omitempty"`
}

// UnmarshalJSON accepts numeric and string activity IDs and populates legacy aliases.
func (a *ActivityRecord) UnmarshalJSON(data []byte) error {
	type alias ActivityRecord
	var raw struct {
		alias
		ID json.RawMessage `json:"id"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*a = ActivityRecord(raw.alias)
	a.ID = stringFromJSON(raw.ID)
	if a.Action == "" {
		a.Action = a.What
	}
	if a.Timestamp == "" {
		a.Timestamp = a.CreatedAt
	}
	return nil
}

// ListWorkspaces lists workspaces visible to the authenticated user.
func (c *Client) ListWorkspaces(ctx context.Context, opts ListOptions, reqOpts ...owlvigil.RequestOption) (*ListResponse[Workspace], *owlvigil.ResponseMeta, error) {
	var out ListResponse[Workspace]
	meta, err := c.http.Do(ctx, http.MethodGet, "/workspaces", opts.values(), nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// CreateWorkspace creates a workspace.
func (c *Client) CreateWorkspace(ctx context.Context, req *CreateWorkspaceRequest, reqOpts ...owlvigil.RequestOption) (*Workspace, *owlvigil.ResponseMeta, error) {
	var out Workspace
	meta, err := c.http.Do(ctx, http.MethodPost, "/workspaces", nil, req, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// GetWorkspace retrieves a workspace by ID.
func (c *Client) GetWorkspace(ctx context.Context, id int64, reqOpts ...owlvigil.RequestOption) (*Workspace, *owlvigil.ResponseMeta, error) {
	var out Workspace
	meta, err := c.http.Do(ctx, http.MethodGet, "/workspaces/"+strconv.FormatInt(id, 10), nil, nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// GetWorkspaceOverview retrieves aggregate metrics and recent activity for a workspace.
func (c *Client) GetWorkspaceOverview(ctx context.Context, workspaceID int64, opts WorkspaceOverviewOptions, reqOpts ...owlvigil.RequestOption) (*WorkspaceOverview, *owlvigil.ResponseMeta, error) {
	var out WorkspaceOverview
	path := "/workspaces/" + strconv.FormatInt(workspaceID, 10) + "/overview"
	meta, err := c.http.Do(ctx, http.MethodGet, path, opts.values(), nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// UpdateWorkspace updates workspace information.
func (c *Client) UpdateWorkspace(ctx context.Context, id int64, req *UpdateWorkspaceRequest, reqOpts ...owlvigil.RequestOption) (*Workspace, *owlvigil.ResponseMeta, error) {
	var out Workspace
	meta, err := c.http.Do(ctx, http.MethodPatch, "/workspaces/"+strconv.FormatInt(id, 10), nil, req, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// DeleteWorkspace deletes a workspace.
func (c *Client) DeleteWorkspace(ctx context.Context, workspaceID int64, reqOpts ...owlvigil.RequestOption) (*owlvigil.ResponseMeta, error) {
	path := "/workspaces/" + strconv.FormatInt(workspaceID, 10)
	return c.http.Do(ctx, http.MethodDelete, path, nil, nil, nil, reqOpts...)
}

// ListWorkspaceActivity lists workspace activity logs.
func (c *Client) ListWorkspaceActivity(ctx context.Context, workspaceID int64, opts ListOptions, reqOpts ...owlvigil.RequestOption) (*ListResponse[ActivityRecord], *owlvigil.ResponseMeta, error) {
	return c.ListWorkspaceActivityWithFilters(
		ctx,
		workspaceID,
		WorkspaceActivityOptions{Limit: opts.Limit},
		reqOpts...,
	)
}

// ListWorkspaceActivityWithFilters lists workspace activity using every
// filter published by the current Open API.
func (c *Client) ListWorkspaceActivityWithFilters(ctx context.Context, workspaceID int64, opts WorkspaceActivityOptions, reqOpts ...owlvigil.RequestOption) (*ListResponse[ActivityRecord], *owlvigil.ResponseMeta, error) {
	var out ListResponse[ActivityRecord]
	path := "/workspaces/" + strconv.FormatInt(workspaceID, 10) + "/activity"
	meta, err := c.http.Do(ctx, http.MethodGet, path, opts.values(), nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

func intString(i int) string {
	return strconv.Itoa(i)
}

func addFilter(q url.Values, key, value string) url.Values {
	if value != "" {
		q.Set(key, value)
	}
	return q
}
