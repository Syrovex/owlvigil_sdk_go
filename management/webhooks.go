package management

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"

	owlvigil "github.com/Syrovex/owlvigil_sdk_go"
)

// WebhookEndpoint describes a configured webhook endpoint.
type WebhookEndpoint struct {
	ID          int64    `json:"id"`
	WorkspaceID int64    `json:"workspace_id,omitempty"`
	URL         string   `json:"url"`
	Events      []string `json:"events,omitempty"`
	EventTypes  []string `json:"event_types,omitempty"`
	Status      string   `json:"status"`
	Secret      string   `json:"secret,omitempty"`
	CreatedAt   string   `json:"created_at,omitempty"`
}

// CreateWebhookEndpointRequest creates a webhook endpoint.
type CreateWebhookEndpointRequest struct {
	WorkspaceID int64    `json:"workspace_id,omitempty"`
	URL         string   `json:"url"`
	Events      []string `json:"events,omitempty"`
	EventTypes  []string `json:"event_types,omitempty"`
}

// UpdateWebhookEndpointRequest updates webhook endpoint.
type UpdateWebhookEndpointRequest struct {
	URL        *string  `json:"url,omitempty"`
	Events     []string `json:"events,omitempty"`
	EventTypes []string `json:"event_types,omitempty"`
	Status     *string  `json:"status,omitempty"`
}

// WebhookEvent describes a webhook delivery event.
type WebhookEvent struct {
	ID          string `json:"id"`
	EndpointID  int64  `json:"endpoint_id"`
	Type        string `json:"type"`
	EventType   string `json:"event_type,omitempty"`
	Status      string `json:"status"`
	Payload     any    `json:"payload,omitempty"`
	CreatedAt   string `json:"created_at,omitempty"`
	DeliveredAt string `json:"delivered_at,omitempty"`
}

func (e *WebhookEvent) UnmarshalJSON(data []byte) error {
	type alias WebhookEvent
	var raw struct {
		alias
		ID json.RawMessage `json:"id"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*e = WebhookEvent(raw.alias)
	e.ID = stringFromJSON(raw.ID)
	if e.Type == "" {
		e.Type = e.EventType
	}
	if e.EventType == "" {
		e.EventType = e.Type
	}
	return nil
}

// WebhookEventType describes an event type.
type WebhookEventType struct {
	Type        string `json:"type"`
	Group       string `json:"group"`
	Description string `json:"description,omitempty"`
}

// BulkRedeliverRequest bulk redeliver webhook events.
type BulkRedeliverRequest struct {
	WorkspaceID int64   `json:"workspace_id,omitempty"`
	EndpointID  *int64  `json:"endpoint_id,omitempty"`
	EventType   *string `json:"event_type,omitempty"`
	EventIDs    []int   `json:"event_ids,omitempty"`
	StartTime   *string `json:"start_time,omitempty"`
	EndTime     *string `json:"end_time,omitempty"`
	Limit       int     `json:"limit,omitempty"`
}

// ListWebhookEndpoints lists webhook endpoints.
func (c *Client) ListWebhookEndpoints(ctx context.Context, opts ListOptions, reqOpts ...owlvigil.RequestOption) (*ListResponse[WebhookEndpoint], *owlvigil.ResponseMeta, error) {
	var out ListResponse[WebhookEndpoint]
	meta, err := c.http.Do(ctx, http.MethodGet, "/webhook-endpoints", opts.values(), nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// CreateWebhookEndpoint creates a webhook endpoint.
func (c *Client) CreateWebhookEndpoint(ctx context.Context, req *CreateWebhookEndpointRequest, reqOpts ...owlvigil.RequestOption) (*WebhookEndpoint, *owlvigil.ResponseMeta, error) {
	var out WebhookEndpoint
	meta, err := c.http.Do(ctx, http.MethodPost, "/webhook-endpoints", nil, req, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// GetWebhookEndpoint retrieves webhook endpoint details.
func (c *Client) GetWebhookEndpoint(ctx context.Context, id int64, reqOpts ...owlvigil.RequestOption) (*WebhookEndpoint, *owlvigil.ResponseMeta, error) {
	var out WebhookEndpoint
	meta, err := c.http.Do(ctx, http.MethodGet, "/webhook-endpoints/"+strconv.FormatInt(id, 10), nil, nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// UpdateWebhookEndpoint updates webhook endpoint.
func (c *Client) UpdateWebhookEndpoint(ctx context.Context, id int64, req *UpdateWebhookEndpointRequest, reqOpts ...owlvigil.RequestOption) (*WebhookEndpoint, *owlvigil.ResponseMeta, error) {
	var out WebhookEndpoint
	meta, err := c.http.Do(ctx, http.MethodPatch, "/webhook-endpoints/"+strconv.FormatInt(id, 10), nil, req, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// DeleteWebhookEndpoint deletes webhook endpoint.
func (c *Client) DeleteWebhookEndpoint(ctx context.Context, id int64, reqOpts ...owlvigil.RequestOption) (*owlvigil.ResponseMeta, error) {
	return c.http.Do(ctx, http.MethodDelete, "/webhook-endpoints/"+strconv.FormatInt(id, 10), nil, nil, nil, reqOpts...)
}

// EnableWebhookEndpoint enables webhook endpoint.
func (c *Client) EnableWebhookEndpoint(ctx context.Context, id int64, reqOpts ...owlvigil.RequestOption) (*owlvigil.ResponseMeta, error) {
	return c.http.Do(ctx, http.MethodPost, "/webhook-endpoints/"+strconv.FormatInt(id, 10)+"/enable", nil, nil, nil, reqOpts...)
}

// DisableWebhookEndpoint disables webhook endpoint.
func (c *Client) DisableWebhookEndpoint(ctx context.Context, id int64, reqOpts ...owlvigil.RequestOption) (*owlvigil.ResponseMeta, error) {
	return c.http.Do(ctx, http.MethodPost, "/webhook-endpoints/"+strconv.FormatInt(id, 10)+"/disable", nil, nil, nil, reqOpts...)
}

// RotateWebhookSecret rotates webhook signing secret.
func (c *Client) RotateWebhookSecret(ctx context.Context, id int64, reqOpts ...owlvigil.RequestOption) (*WebhookEndpoint, *owlvigil.ResponseMeta, error) {
	var out WebhookEndpoint
	meta, err := c.http.Do(ctx, http.MethodPost, "/webhook-endpoints/"+strconv.FormatInt(id, 10)+"/rotate-secret", nil, nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// TestWebhookEndpoint sends a test event to webhook endpoint.
func (c *Client) TestWebhookEndpoint(ctx context.Context, id int64, reqOpts ...owlvigil.RequestOption) (*owlvigil.ResponseMeta, error) {
	return c.http.Do(ctx, http.MethodPost, "/webhook-endpoints/"+strconv.FormatInt(id, 10)+"/test", nil, nil, nil, reqOpts...)
}

// ListWebhookEventTypes lists available webhook event types.
func (c *Client) ListWebhookEventTypes(ctx context.Context, reqOpts ...owlvigil.RequestOption) (*ListResponse[WebhookEventType], *owlvigil.ResponseMeta, error) {
	var out ListResponse[WebhookEventType]
	meta, err := c.http.Do(ctx, http.MethodGet, "/webhook-event-types", nil, nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// ListWebhookEvents lists webhook delivery events.
func (c *Client) ListWebhookEvents(ctx context.Context, opts ListOptions, reqOpts ...owlvigil.RequestOption) (*ListResponse[WebhookEvent], *owlvigil.ResponseMeta, error) {
	var out ListResponse[WebhookEvent]
	meta, err := c.http.Do(ctx, http.MethodGet, "/webhook-events", opts.values(), nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// GetWebhookEvent retrieves webhook event details.
func (c *Client) GetWebhookEvent(ctx context.Context, eventID string, reqOpts ...owlvigil.RequestOption) (*WebhookEvent, *owlvigil.ResponseMeta, error) {
	var out WebhookEvent
	meta, err := c.http.Do(ctx, http.MethodGet, "/webhook-events/"+url.PathEscape(eventID), nil, nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// ListEndpointEvents lists events for a specific endpoint.
func (c *Client) ListEndpointEvents(ctx context.Context, endpointID int64, opts ListOptions, reqOpts ...owlvigil.RequestOption) (*ListResponse[WebhookEvent], *owlvigil.ResponseMeta, error) {
	var out ListResponse[WebhookEvent]
	path := "/webhook-endpoints/" + strconv.FormatInt(endpointID, 10) + "/events"
	meta, err := c.http.Do(ctx, http.MethodGet, path, opts.values(), nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// RetryWebhookEvent retries a failed webhook event.
func (c *Client) RetryWebhookEvent(ctx context.Context, eventID string, reqOpts ...owlvigil.RequestOption) (*owlvigil.ResponseMeta, error) {
	return c.http.Do(ctx, http.MethodPost, "/webhook-events/"+url.PathEscape(eventID)+"/retry", nil, nil, nil, reqOpts...)
}

// RedeliverWebhookEvent redeliver a webhook event.
func (c *Client) RedeliverWebhookEvent(ctx context.Context, eventID string, reqOpts ...owlvigil.RequestOption) (*owlvigil.ResponseMeta, error) {
	return c.http.Do(ctx, http.MethodPost, "/webhook-events/"+url.PathEscape(eventID)+"/redeliver", nil, nil, nil, reqOpts...)
}

// BulkRedeliverWebhookEvents bulk redeliver webhook events.
func (c *Client) BulkRedeliverWebhookEvents(ctx context.Context, req *BulkRedeliverRequest, reqOpts ...owlvigil.RequestOption) (*owlvigil.ResponseMeta, error) {
	return c.http.Do(ctx, http.MethodPost, "/webhook-events/bulk-redeliver", nil, req, nil, reqOpts...)
}
