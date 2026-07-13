package management

import (
	"context"
	"net/http"
	"net/url"

	owlvigil "github.com/owlvigil/owlvigil-go"
)

// RequestLog describes a Gateway request log entry.
type RequestLog struct {
	RequestID string  `json:"request_id"`
	Model     string  `json:"model"`
	Status    string  `json:"status"`
	Cost      float64 `json:"cost,omitempty"`
	Provider  string  `json:"provider,omitempty"`
	Duration  int64   `json:"duration,omitempty"`
	CreatedAt string  `json:"created_at,omitempty"`
}

// Trace describes Gateway trace details.
type Trace struct {
	TraceID   string `json:"trace_id"`
	ThreadID  string `json:"thread_id,omitempty"`
	Events    []any  `json:"events,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
}

// PayloadAccess describes payload log access permission.
type PayloadAccess struct {
	Allowed bool   `json:"allowed"`
	Reason  string `json:"reason,omitempty"`
}

// PayloadLog describes a payload log entry.
type PayloadLog struct {
	PayloadID string `json:"payload_id"`
	RequestID string `json:"request_id"`
	Request   any    `json:"request,omitempty"`
	Response  any    `json:"response,omitempty"`
}

// ListRequestLogs lists Gateway request logs.
func (c *Client) ListRequestLogs(ctx context.Context, opts ListOptions, gatewayKeyID string, reqOpts ...owlvigil.RequestOption) (*ListResponse[RequestLog], *owlvigil.ResponseMeta, error) {
	q := addFilter(opts.values(), "gateway_key_id", gatewayKeyID)
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
