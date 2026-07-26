package management

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"

	owlvigil "github.com/Syrovex/owlvigil_sdk_go"
)

// RequestLog describes a Gateway request log entry.
type RequestLog struct {
	ID             int     `json:"id"`
	RequestID      string  `json:"request_id"`
	TraceID        string  `json:"trace_id"`
	WorkspaceID    int     `json:"workspace_id"`
	Provider       string  `json:"provider"`
	ProviderID     int     `json:"provider_id"`
	ProviderSource string  `json:"provider_source"`
	Model          string  `json:"model"`
	Format         string  `json:"format,omitempty"`
	Status         string  `json:"status"`
	StatusCode     int     `json:"status_code"`
	LatencyMS      *int64  `json:"latency_ms"`
	InputTokens    int64   `json:"input_tokens"`
	OutputTokens   int64   `json:"output_tokens"`
	TotalTokens    int64   `json:"total_tokens"`
	TotalCost      float64 `json:"total_cost"`
	GatewayKeyID   int     `json:"gateway_key_id"`
	KeyName        string  `json:"key_name"`
	UserID         int     `json:"user_id"`
	MemberName     string  `json:"member_name"`
	TeamID         *int64  `json:"team_id"`
	TeamName       string  `json:"team_name"`
	Stream         bool    `json:"stream"`
	CreatedAt      string  `json:"created_at,omitempty"`

	// Legacy aliases retained for source compatibility.
	Cost     float64 `json:"cost,omitempty"`
	Duration int64   `json:"duration,omitempty"`
}

// UnmarshalJSON synchronizes current request-log values with legacy aliases.
func (l *RequestLog) UnmarshalJSON(data []byte) error {
	type alias RequestLog
	var out alias
	if err := json.Unmarshal(data, &out); err != nil {
		return err
	}
	*l = RequestLog(out)
	if l.TotalCost == 0 {
		l.TotalCost = l.Cost
	}
	if l.Cost == 0 {
		l.Cost = l.TotalCost
	}
	if l.LatencyMS == nil && l.Duration != 0 {
		l.LatencyMS = &l.Duration
	}
	if l.Duration == 0 && l.LatencyMS != nil {
		l.Duration = *l.LatencyMS
	}
	return nil
}

// Trace describes Gateway trace details.
type Trace struct {
	ID            int    `json:"id"`
	TraceID       string `json:"trace_id"`
	WorkspaceID   int    `json:"workspace_id"`
	ThreadID      string `json:"-"`
	ThreadIDValue int    `json:"thread_id,omitempty"`
	CreatedAt     string `json:"created_at,omitempty"`
	UpdatedAt     string `json:"updated_at,omitempty"`

	// Events is retained for pre-refactor trace responses.
	Events []any `json:"events,omitempty"`
}

// UnmarshalJSON accepts the current numeric thread_id and older string values.
func (t *Trace) UnmarshalJSON(data []byte) error {
	type alias Trace
	var raw struct {
		alias
		ThreadID json.RawMessage `json:"thread_id"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*t = Trace(raw.alias)
	t.ThreadID = stringFromJSON(raw.ThreadID)
	if value := int64PointerFromJSON(raw.ThreadID); value != nil {
		t.ThreadIDValue = int(*value)
	}
	return nil
}

// PayloadAccess describes payload log access permission.
type PayloadAccess struct {
	Enabled                 bool   `json:"enabled"`
	CaptureRequestBody      bool   `json:"capture_request_body"`
	CaptureResponseBody     bool   `json:"capture_response_body"`
	CaptureStreamChunks     bool   `json:"capture_stream_chunks"`
	Message                 string `json:"message,omitempty"`
	RequestDisabledMessage  string `json:"request_disabled_message,omitempty"`
	ResponseDisabledMessage string `json:"response_disabled_message,omitempty"`

	// Legacy aliases retained for source compatibility.
	Allowed bool   `json:"allowed,omitempty"`
	Reason  string `json:"reason,omitempty"`
}

// UnmarshalJSON synchronizes current and legacy payload-access fields.
func (a *PayloadAccess) UnmarshalJSON(data []byte) error {
	type alias PayloadAccess
	var out alias
	if err := json.Unmarshal(data, &out); err != nil {
		return err
	}
	*a = PayloadAccess(out)
	if !a.Enabled {
		a.Enabled = a.Allowed
	}
	if !a.Allowed {
		a.Allowed = a.Enabled
	}
	if a.Message == "" {
		a.Message = a.Reason
	}
	if a.Reason == "" {
		a.Reason = a.Message
	}
	return nil
}

// PayloadLog describes a payload log entry.
type PayloadLog struct {
	ID                  int                `json:"id"`
	RequestID           string             `json:"request_id"`
	WorkspaceID         int                `json:"workspace_id"`
	APIKeyID            *int               `json:"api_key_id,omitempty"`
	ChannelID           *int               `json:"channel_id,omitempty"`
	TraceID             *int               `json:"trace_id,omitempty"`
	TraceKey            string             `json:"trace_key,omitempty"`
	Model               string             `json:"model"`
	Format              string             `json:"format"`
	Status              string             `json:"status"`
	Stream              bool               `json:"stream"`
	ClientIP            string             `json:"client_ip,omitempty"`
	ExternalID          string             `json:"external_id,omitempty"`
	DurationMS          *int64             `json:"duration_ms,omitempty"`
	FirstTokenMS        *int64             `json:"first_token_ms,omitempty"`
	ReasoningDurationMS *int64             `json:"reasoning_duration_ms,omitempty"`
	RequestBody         any                `json:"request_body,omitempty"`
	ResponseBody        any                `json:"response_body,omitempty"`
	ResponseChunks      any                `json:"response_chunks,omitempty"`
	Executions          []PayloadExecution `json:"executions,omitempty"`
	CreatedAt           string             `json:"created_at"`
	UpdatedAt           string             `json:"updated_at"`

	// Legacy aliases retained for source compatibility.
	PayloadID string `json:"-"`
	Request   any    `json:"-"`
	Response  any    `json:"-"`
}

// PayloadExecution describes one upstream execution in a payload log.
type PayloadExecution struct {
	ID             int    `json:"id"`
	ChannelID      *int   `json:"channel_id,omitempty"`
	ExternalID     string `json:"external_id,omitempty"`
	Model          string `json:"model"`
	Format         string `json:"format"`
	Status         string `json:"status"`
	StatusCode     *int   `json:"status_code,omitempty"`
	Stream         bool   `json:"stream"`
	ErrorMessage   string `json:"error_message,omitempty"`
	RequestBody    any    `json:"request_body,omitempty"`
	ResponseBody   any    `json:"response_body,omitempty"`
	ResponseChunks any    `json:"response_chunks,omitempty"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
}

// UnmarshalJSON decodes the current detailed payload response and fills legacy
// request/response aliases.
func (p *PayloadLog) UnmarshalJSON(data []byte) error {
	type alias PayloadLog
	var raw struct {
		alias
		LegacyPayloadID string `json:"payload_id,omitempty"`
		LegacyRequest   any    `json:"request,omitempty"`
		LegacyResponse  any    `json:"response,omitempty"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*p = PayloadLog(raw.alias)
	p.PayloadID = raw.LegacyPayloadID
	if p.PayloadID == "" && p.ID != 0 {
		p.PayloadID = strconv.Itoa(p.ID)
	}
	p.Request = p.RequestBody
	if p.Request == nil {
		p.Request = raw.LegacyRequest
		p.RequestBody = raw.LegacyRequest
	}
	p.Response = p.ResponseBody
	if p.Response == nil {
		p.Response = raw.LegacyResponse
		p.ResponseBody = raw.LegacyResponse
	}
	return nil
}

// AuditLog describes a workspace audit record.
type AuditLog struct {
	ID                 int64          `json:"id"`
	WorkspaceID        int64          `json:"workspace_id"`
	UserID             int64          `json:"user_id"`
	Action             string         `json:"action"`
	ResourceType       string         `json:"resource_type"`
	ResourceID         string         `json:"resource_id"`
	Status             string         `json:"status"`
	ActorLabel         string         `json:"actor_label"`
	TargetLabel        string         `json:"target_label"`
	BeforeAfterSummary string         `json:"before_after_summary"`
	UserEmail          string         `json:"user_email,omitempty"`
	UserName           string         `json:"user_name,omitempty"`
	ResourceName       string         `json:"resource_name,omitempty"`
	IPAddress          string         `json:"ip_address,omitempty"`
	UserAgent          string         `json:"user_agent,omitempty"`
	BeforeState        map[string]any `json:"before_state,omitempty"`
	AfterState         map[string]any `json:"after_state,omitempty"`
	Metadata           map[string]any `json:"metadata,omitempty"`
	ErrorMessage       string         `json:"error_message,omitempty"`
	CreatedAt          string         `json:"created_at"`
}

// AuditLogListOptions filters and paginates workspace audit logs.
type AuditLogListOptions struct {
	Cursor       string
	Limit        int
	UserID       string
	Action       string
	ResourceType string
	ResourceID   string
	Status       string
	DateFrom     string
	DateTo       string
	Search       string
}

func (o AuditLogListOptions) values() url.Values {
	q := ListOptions{Cursor: o.Cursor, Limit: o.Limit}.values()
	addFilter(q, "user_id", o.UserID)
	addFilter(q, "action", o.Action)
	addFilter(q, "resource_type", o.ResourceType)
	addFilter(q, "resource_id", o.ResourceID)
	addFilter(q, "status", o.Status)
	addFilter(q, "date_from", o.DateFrom)
	addFilter(q, "date_to", o.DateTo)
	addFilter(q, "search", o.Search)
	return q
}

// LoggingSettings contains workspace request and payload logging configuration.
type LoggingSettings struct {
	Enabled              bool          `json:"enabled"`
	RetentionDays        int           `json:"retention_days"`
	CaptureHeaders       bool          `json:"capture_headers"`
	CaptureRequestBody   bool          `json:"capture_request_body"`
	CaptureResponseBody  bool          `json:"capture_response_body"`
	CaptureStreamChunks  bool          `json:"capture_stream_chunks"`
	MaskSensitiveHeaders bool          `json:"mask_sensitive_headers"`
	MaxBodySize          int           `json:"max_body_size"`
	RedactHeaders        []string      `json:"redact_headers"`
	LogDataQuota         *LogDataQuota `json:"log_data_quota,omitempty"`
}

// LogDataQuota contains log-storage quota status.
type LogDataQuota struct {
	UsedBytes          int64   `json:"used_bytes"`
	LimitBytes         int64   `json:"limit_bytes"`
	UsagePercent       float64 `json:"usage_percent"`
	Warning            bool    `json:"warning"`
	OverLimit          bool    `json:"over_limit"`
	CleanupRecommended bool    `json:"cleanup_recommended"`
	LastCalculatedAt   string  `json:"last_calculated_at"`
	LastError          string  `json:"last_error,omitempty"`
	DeletedBytes       int64   `json:"deleted_bytes,omitempty"`
	DeletedObjects     int     `json:"deleted_objects,omitempty"`
}

// UpdateLoggingSettingsRequest updates workspace logging configuration.
type UpdateLoggingSettingsRequest struct {
	Enabled              *bool     `json:"enabled,omitempty"`
	RetentionDays        *int      `json:"retention_days,omitempty"`
	CaptureHeaders       *bool     `json:"capture_headers,omitempty"`
	CaptureRequestBody   *bool     `json:"capture_request_body,omitempty"`
	CaptureResponseBody  *bool     `json:"capture_response_body,omitempty"`
	CaptureStreamChunks  *bool     `json:"capture_stream_chunks,omitempty"`
	MaskSensitiveHeaders *bool     `json:"mask_sensitive_headers,omitempty"`
	MaxBodySize          *int      `json:"max_body_size,omitempty"`
	RedactHeaders        *[]string `json:"redact_headers,omitempty"`
}

// PayloadLogSummary is the body-free payload-log list item.
type PayloadLogSummary struct {
	ID          int    `json:"id"`
	RequestID   string `json:"request_id"`
	WorkspaceID int    `json:"workspace_id"`
	TraceID     *int   `json:"trace_id,omitempty"`
	TraceKey    string `json:"trace_key,omitempty"`
	Model       string `json:"model"`
	Format      string `json:"format"`
	Status      string `json:"status"`
	Stream      bool   `json:"stream"`
	CreatedAt   string `json:"created_at"`
}

// ListAuditLogs lists audit records for a workspace.
func (c *Client) ListAuditLogs(ctx context.Context, workspaceID int64, opts AuditLogListOptions, reqOpts ...owlvigil.RequestOption) (*ListResponse[AuditLog], *owlvigil.ResponseMeta, error) {
	var out ListResponse[AuditLog]
	path := "/workspaces/" + strconv.FormatInt(workspaceID, 10) + "/audit-logs"
	meta, err := c.http.Do(ctx, http.MethodGet, path, opts.values(), nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// GetAuditLog retrieves an audit record for a workspace.
func (c *Client) GetAuditLog(ctx context.Context, workspaceID, auditLogID int64, reqOpts ...owlvigil.RequestOption) (*AuditLog, *owlvigil.ResponseMeta, error) {
	var out AuditLog
	path := "/workspaces/" + strconv.FormatInt(workspaceID, 10) + "/audit-logs/" + strconv.FormatInt(auditLogID, 10)
	meta, err := c.http.Do(ctx, http.MethodGet, path, nil, nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// GetLoggingSettings retrieves workspace logging configuration.
func (c *Client) GetLoggingSettings(ctx context.Context, workspaceID int64, reqOpts ...owlvigil.RequestOption) (*LoggingSettings, *owlvigil.ResponseMeta, error) {
	var out LoggingSettings
	path := "/workspaces/" + strconv.FormatInt(workspaceID, 10) + "/logging-settings"
	meta, err := c.http.Do(ctx, http.MethodGet, path, nil, nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// UpdateLoggingSettings updates workspace logging configuration.
func (c *Client) UpdateLoggingSettings(ctx context.Context, workspaceID int64, req *UpdateLoggingSettingsRequest, reqOpts ...owlvigil.RequestOption) (*LoggingSettings, *owlvigil.ResponseMeta, error) {
	var out LoggingSettings
	path := "/workspaces/" + strconv.FormatInt(workspaceID, 10) + "/logging-settings"
	meta, err := c.http.Do(ctx, http.MethodPut, path, nil, req, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// ListPayloadLogs lists body-free payload logs for a workspace.
func (c *Client) ListPayloadLogs(ctx context.Context, workspaceID int64, opts ListOptions, reqOpts ...owlvigil.RequestOption) (*ListResponse[PayloadLogSummary], *owlvigil.ResponseMeta, error) {
	q := opts.values()
	q.Set("workspace_id", strconv.FormatInt(workspaceID, 10))
	var out ListResponse[PayloadLogSummary]
	meta, err := c.http.Do(ctx, http.MethodGet, "/gateway/payload-logs", q, nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// ListRequestLogs lists Gateway request logs.
func (c *Client) ListRequestLogs(ctx context.Context, opts ListOptions, gatewayKeyID string, reqOpts ...owlvigil.RequestOption) (*ListResponse[RequestLog], *owlvigil.ResponseMeta, error) {
	q := opts.values()
	addFilter(q, "gateway_key_id", gatewayKeyID)
	var out ListResponse[RequestLog]
	meta, err := c.http.Do(ctx, http.MethodGet, "/gateway/request-logs", q, nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// GetRequestLog retrieves a request log by request ID.
func (c *Client) GetRequestLog(ctx context.Context, requestID string, reqOpts ...owlvigil.RequestOption) (*RequestLog, *owlvigil.ResponseMeta, error) {
	var out RequestLog
	meta, err := c.http.Do(ctx, http.MethodGet, "/gateway/request-logs/"+url.PathEscape(requestID), nil, nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// ListTraces lists Gateway traces.
func (c *Client) ListTraces(ctx context.Context, opts ListOptions, reqOpts ...owlvigil.RequestOption) (*ListResponse[Trace], *owlvigil.ResponseMeta, error) {
	var out ListResponse[Trace]
	meta, err := c.http.Do(ctx, http.MethodGet, "/gateway/traces", opts.values(), nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// GetTrace retrieves trace details by trace ID.
func (c *Client) GetTrace(ctx context.Context, traceID string, reqOpts ...owlvigil.RequestOption) (*Trace, *owlvigil.ResponseMeta, error) {
	var out Trace
	meta, err := c.http.Do(ctx, http.MethodGet, "/gateway/traces/"+url.PathEscape(traceID), nil, nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// GetPayloadAccess checks if current key has payload log access.
func (c *Client) GetPayloadAccess(ctx context.Context, reqOpts ...owlvigil.RequestOption) (*PayloadAccess, *owlvigil.ResponseMeta, error) {
	var out PayloadAccess
	meta, err := c.http.Do(ctx, http.MethodGet, "/gateway/payload-logs/access", nil, nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// GetPayloadLog retrieves payload log by ID.
func (c *Client) GetPayloadLog(ctx context.Context, payloadID string, reqOpts ...owlvigil.RequestOption) (*PayloadLog, *owlvigil.ResponseMeta, error) {
	var out PayloadLog
	meta, err := c.http.Do(ctx, http.MethodGet, "/gateway/payload-logs/"+url.PathEscape(payloadID), nil, nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}
