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
	EventTypes  []string `json:"event_types,omitempty"`
	Status      string   `json:"status"`
	Secret      string   `json:"secret,omitempty"`
	CreatedAt   string   `json:"created_at,omitempty"`
	UpdatedAt   string   `json:"updated_at,omitempty"`

	// Events is the pre-refactor alias for EventTypes.
	Events []string `json:"-"`
}

// UnmarshalJSON synchronizes current and legacy event type fields.
func (e *WebhookEndpoint) UnmarshalJSON(data []byte) error {
	type alias WebhookEndpoint
	var raw struct {
		alias
		Events []string `json:"events,omitempty"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*e = WebhookEndpoint(raw.alias)
	if len(e.EventTypes) == 0 {
		e.EventTypes = raw.Events
	}
	e.Events = append([]string(nil), e.EventTypes...)
	return nil
}

// CreateWebhookEndpointRequest creates a webhook endpoint.
type CreateWebhookEndpointRequest struct {
	WorkspaceID int64    `json:"workspace_id,omitempty"`
	URL         string   `json:"url"`
	EventTypes  []string `json:"event_types,omitempty"`

	// Events is the pre-refactor alias for EventTypes.
	Events []string `json:"-"`
}

// MarshalJSON emits the current webhook endpoint creation contract.
func (r CreateWebhookEndpointRequest) MarshalJSON() ([]byte, error) {
	eventTypes := r.EventTypes
	if eventTypes == nil {
		eventTypes = r.Events
	}
	return json.Marshal(struct {
		WorkspaceID int64    `json:"workspace_id"`
		URL         string   `json:"url"`
		EventTypes  []string `json:"event_types"`
	}{
		WorkspaceID: r.WorkspaceID,
		URL:         r.URL,
		EventTypes:  eventTypes,
	})
}

// UpdateWebhookEndpointRequest updates webhook endpoint.
type UpdateWebhookEndpointRequest struct {
	URL        *string  `json:"url,omitempty"`
	EventTypes []string `json:"-"`
	Status     *string  `json:"status,omitempty"`

	// Events is the pre-refactor alias for EventTypes.
	Events []string `json:"-"`
}

// MarshalJSON emits the current webhook endpoint update contract and
// preserves an explicitly empty event_types list.
func (r UpdateWebhookEndpointRequest) MarshalJSON() ([]byte, error) {
	eventTypes := r.EventTypes
	if eventTypes == nil {
		eventTypes = r.Events
	}
	var eventTypesPointer *[]string
	if eventTypes != nil {
		eventTypesPointer = &eventTypes
	}
	return json.Marshal(struct {
		URL        *string   `json:"url,omitempty"`
		EventTypes *[]string `json:"event_types,omitempty"`
		Status     *string   `json:"status,omitempty"`
	}{
		URL:        r.URL,
		EventTypes: eventTypesPointer,
		Status:     r.Status,
	})
}

// WebhookEvent describes a webhook delivery event.
type WebhookEvent struct {
	ID          string         `json:"id"`
	EndpointID  int64          `json:"endpoint_id"`
	WorkspaceID int64          `json:"workspace_id"`
	EventType   string         `json:"event_type,omitempty"`
	Status      string         `json:"status"`
	Payload     map[string]any `json:"payload,omitempty"`
	Attempts    int            `json:"attempts"`
	LastError   string         `json:"last_error,omitempty"`
	DeliveredAt string         `json:"delivered_at,omitempty"`
	CreatedAt   string         `json:"created_at,omitempty"`
	UpdatedAt   string         `json:"updated_at,omitempty"`

	// Type is the pre-refactor alias for EventType.
	Type string `json:"type,omitempty"`
}

// UnmarshalJSON accepts numeric and string event IDs and populates legacy aliases.
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
	WorkspaceID int64   `json:"workspace_id"`
	EndpointID  *int64  `json:"endpoint_id,omitempty"`
	EventType   *string `json:"event_type,omitempty"`
	EventIDs    []int   `json:"event_ids,omitempty"`
	StartTime   *string `json:"-"`
	EndTime     *string `json:"-"`
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
	_, meta, err := c.DeleteWebhookEndpointWithResult(ctx, id, reqOpts...)
	return meta, err
}

// DeleteWebhookEndpointWithResult deletes an endpoint and returns confirmation.
func (c *Client) DeleteWebhookEndpointWithResult(ctx context.Context, id int64, reqOpts ...owlvigil.RequestOption) (*DeleteResponse, *owlvigil.ResponseMeta, error) {
	var out DeleteResponse
	meta, err := c.http.Do(ctx, http.MethodDelete, "/webhook-endpoints/"+strconv.FormatInt(id, 10), nil, nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// EnableWebhookEndpoint enables webhook endpoint.
func (c *Client) EnableWebhookEndpoint(ctx context.Context, id int64, reqOpts ...owlvigil.RequestOption) (*owlvigil.ResponseMeta, error) {
	_, meta, err := c.EnableWebhookEndpointWithResult(ctx, id, reqOpts...)
	return meta, err
}

// EnableWebhookEndpointWithResult enables an endpoint and returns its state.
func (c *Client) EnableWebhookEndpointWithResult(ctx context.Context, id int64, reqOpts ...owlvigil.RequestOption) (*WebhookEndpoint, *owlvigil.ResponseMeta, error) {
	var out WebhookEndpoint
	meta, err := c.http.Do(ctx, http.MethodPost, "/webhook-endpoints/"+strconv.FormatInt(id, 10)+"/enable", nil, nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// DisableWebhookEndpoint disables webhook endpoint.
func (c *Client) DisableWebhookEndpoint(ctx context.Context, id int64, reqOpts ...owlvigil.RequestOption) (*owlvigil.ResponseMeta, error) {
	_, meta, err := c.DisableWebhookEndpointWithResult(ctx, id, reqOpts...)
	return meta, err
}

// DisableWebhookEndpointWithResult disables an endpoint and returns its state.
func (c *Client) DisableWebhookEndpointWithResult(ctx context.Context, id int64, reqOpts ...owlvigil.RequestOption) (*WebhookEndpoint, *owlvigil.ResponseMeta, error) {
	var out WebhookEndpoint
	meta, err := c.http.Do(ctx, http.MethodPost, "/webhook-endpoints/"+strconv.FormatInt(id, 10)+"/disable", nil, nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
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
	_, meta, err := c.TestWebhookEndpointWithResult(ctx, id, reqOpts...)
	return meta, err
}

// TestWebhookEndpointWithResult sends a test event and returns its delivery state.
func (c *Client) TestWebhookEndpointWithResult(ctx context.Context, id int64, reqOpts ...owlvigil.RequestOption) (*WebhookEvent, *owlvigil.ResponseMeta, error) {
	var out WebhookEvent
	meta, err := c.http.Do(ctx, http.MethodPost, "/webhook-endpoints/"+strconv.FormatInt(id, 10)+"/test", nil, nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
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
	_, meta, err := c.RetryWebhookEventWithResult(ctx, eventID, reqOpts...)
	return meta, err
}

// RetryWebhookEventWithResult retries an event and returns its updated state.
func (c *Client) RetryWebhookEventWithResult(ctx context.Context, eventID string, reqOpts ...owlvigil.RequestOption) (*WebhookEvent, *owlvigil.ResponseMeta, error) {
	var out WebhookEvent
	meta, err := c.http.Do(ctx, http.MethodPost, "/webhook-events/"+url.PathEscape(eventID)+"/retry", nil, nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// RedeliverWebhookEvent redeliver a webhook event.
func (c *Client) RedeliverWebhookEvent(ctx context.Context, eventID string, reqOpts ...owlvigil.RequestOption) (*owlvigil.ResponseMeta, error) {
	_, meta, err := c.RedeliverWebhookEventWithResult(ctx, eventID, reqOpts...)
	return meta, err
}

// RedeliverWebhookEventWithResult redelivers an event and returns its state.
func (c *Client) RedeliverWebhookEventWithResult(ctx context.Context, eventID string, reqOpts ...owlvigil.RequestOption) (*WebhookEvent, *owlvigil.ResponseMeta, error) {
	var out WebhookEvent
	meta, err := c.http.Do(ctx, http.MethodPost, "/webhook-events/"+url.PathEscape(eventID)+"/redeliver", nil, nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// BulkRedeliverWebhookEvents bulk redeliver webhook events.
func (c *Client) BulkRedeliverWebhookEvents(ctx context.Context, req *BulkRedeliverRequest, reqOpts ...owlvigil.RequestOption) (*owlvigil.ResponseMeta, error) {
	_, meta, err := c.BulkRedeliverWebhookEventsWithResult(ctx, req, reqOpts...)
	return meta, err
}

// BulkRedeliverWebhookEventsWithResult redelivers matching events and returns
// their updated states.
func (c *Client) BulkRedeliverWebhookEventsWithResult(ctx context.Context, req *BulkRedeliverRequest, reqOpts ...owlvigil.RequestOption) (*ListResponse[WebhookEvent], *owlvigil.ResponseMeta, error) {
	var out ListResponse[WebhookEvent]
	meta, err := c.http.Do(ctx, http.MethodPost, "/webhook-events/bulk-redeliver", nil, req, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}
