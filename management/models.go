package management

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"

	owlvigil "github.com/Syrovex/owlvigil_sdk_go"
)

// DeleteResponse confirms that a resource was deleted.
type DeleteResponse struct {
	Deleted bool `json:"deleted"`
}

// Model describes a gateway model.
type Model struct {
	ID         string          `json:"id"`
	ModelID    string          `json:"model_id,omitempty"`
	Developer  string          `json:"developer,omitempty"`
	Type       string          `json:"type,omitempty"`
	Name       string          `json:"name"`
	Icon       string          `json:"icon,omitempty"`
	Group      string          `json:"group,omitempty"`
	Status     string          `json:"status"`
	RouteCount int             `json:"route_count,omitempty"`
	CreatedAt  string          `json:"created_at,omitempty"`
	UpdatedAt  string          `json:"updated_at,omitempty"`
	ModelCard  json.RawMessage `json:"model_card,omitempty"`
	Routes     []Route         `json:"routes,omitempty"`

	// Legacy fields retained for source compatibility.
	Provider      string             `json:"provider,omitempty"`
	Capabilities  []string           `json:"capabilities,omitempty"`
	ContextWindow int                `json:"context_window,omitempty"`
	Pricing       map[string]float64 `json:"pricing,omitempty"`
}

func (m *Model) UnmarshalJSON(data []byte) error {
	type alias Model
	var raw struct {
		alias
		ID json.RawMessage `json:"id"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*m = Model(raw.alias)
	m.ID = stringFromJSON(raw.ID)
	if m.ModelID != "" {
		m.ID = m.ModelID
	}
	if m.Provider == "" {
		m.Provider = m.Developer
	}
	return nil
}

// Route describes a model routing rule.
type Route struct {
	ID               string          `json:"id"`
	RouteID          string          `json:"route_id,omitempty"`
	WorkspaceID      int64           `json:"workspace_id,omitempty"`
	Model            string          `json:"model,omitempty"`
	ActualModel      string          `json:"actual_model,omitempty"`
	MatchSource      string          `json:"match_source,omitempty"`
	ChannelID        int             `json:"channel_id,omitempty"`
	ChannelName      string          `json:"channel_name,omitempty"`
	ChannelType      string          `json:"channel_type,omitempty"`
	ChannelStatus    string          `json:"channel_status,omitempty"`
	ProviderSource   string          `json:"provider_source,omitempty"`
	ProviderPlatform *string         `json:"provider_platform,omitempty"`
	Price            json.RawMessage `json:"price,omitempty"`
	PriceReferenceID string          `json:"price_reference_id,omitempty"`

	// Legacy fields retained for source compatibility.
	ModelID         string   `json:"model_id,omitempty"`
	Providers       []string `json:"providers,omitempty"`
	Priority        int      `json:"priority,omitempty"`
	FallbackEnabled bool     `json:"fallback_enabled,omitempty"`
}

func (r *Route) UnmarshalJSON(data []byte) error {
	type alias Route
	var raw struct {
		alias
		ID json.RawMessage `json:"id"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*r = Route(raw.alias)
	r.ID = stringFromJSON(raw.ID)
	if r.RouteID != "" {
		r.ID = r.RouteID
	}
	if r.ModelID == "" {
		r.ModelID = r.Model
	}
	return nil
}

// PreviewRouteRequest previews model routing.
type PreviewRouteRequest struct {
	WorkspaceID int64          `json:"workspace_id"`
	Model       string         `json:"model"`
	KeyID       int64          `json:"key_id,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

// PreviewRouteResponse describes routing preview result.
type PreviewRouteResponse struct {
	WorkspaceID     int64          `json:"workspace_id,omitempty"`
	Model           string         `json:"model,omitempty"`
	CandidateCount  int            `json:"candidate_count,omitempty"`
	Candidates      []Route        `json:"candidates,omitempty"`
	PreviewMetadata map[string]any `json:"preview_metadata,omitempty"`

	// Legacy fields retained for source compatibility.
	Provider  string   `json:"provider,omitempty"`
	Channel   string   `json:"channel,omitempty"`
	Fallbacks []string `json:"fallbacks,omitempty"`
}

// RouteListOptions filters and paginates gateway route candidates.
type RouteListOptions struct {
	Cursor string
	Limit  int
	KeyID  int64
	Model  string
}

func (o RouteListOptions) values() url.Values {
	q := ListOptions{Cursor: o.Cursor, Limit: o.Limit}.values()
	if o.KeyID > 0 {
		q.Set("key_id", strconv.FormatInt(o.KeyID, 10))
	}
	addFilter(q, "model", o.Model)
	return q
}

// RouteDetailOptions narrows a gateway route lookup.
type RouteDetailOptions struct {
	KeyID int64
	Model string
}

func (o RouteDetailOptions) values() url.Values {
	q := url.Values{}
	if o.KeyID > 0 {
		q.Set("key_id", strconv.FormatInt(o.KeyID, 10))
	}
	addFilter(q, "model", o.Model)
	return q
}

// ListModels lists available gateway models.
func (c *Client) ListModels(ctx context.Context, opts ListOptions, reqOpts ...owlvigil.RequestOption) (*ListResponse[Model], *owlvigil.ResponseMeta, error) {
	var out ListResponse[Model]
	meta, err := c.http.Do(ctx, http.MethodGet, "/gateway/models", opts.values(), nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// GetModel retrieves model details by ID.
func (c *Client) GetModel(ctx context.Context, modelID string, reqOpts ...owlvigil.RequestOption) (*Model, *owlvigil.ResponseMeta, error) {
	var out Model
	meta, err := c.http.Do(ctx, http.MethodGet, "/gateway/models/"+url.PathEscape(modelID), nil, nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// ListRoutes lists model routing rules.
func (c *Client) ListRoutes(ctx context.Context, opts ListOptions, reqOpts ...owlvigil.RequestOption) (*ListResponse[Route], *owlvigil.ResponseMeta, error) {
	return c.ListRoutesWithFilters(
		ctx,
		RouteListOptions{Cursor: opts.Cursor, Limit: opts.Limit},
		reqOpts...,
	)
}

// ListRoutesWithFilters lists model routing candidates using every filter
// published by the current Open API.
func (c *Client) ListRoutesWithFilters(ctx context.Context, opts RouteListOptions, reqOpts ...owlvigil.RequestOption) (*ListResponse[Route], *owlvigil.ResponseMeta, error) {
	var out ListResponse[Route]
	meta, err := c.http.Do(ctx, http.MethodGet, "/gateway/routes", opts.values(), nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// GetRoute retrieves a gateway route by ID.
func (c *Client) GetRoute(ctx context.Context, routeID string, reqOpts ...owlvigil.RequestOption) (*Route, *owlvigil.ResponseMeta, error) {
	return c.GetRouteWithFilters(ctx, routeID, RouteDetailOptions{}, reqOpts...)
}

// GetRouteWithFilters retrieves a gateway route using the optional key and
// model selectors published by the current Open API.
func (c *Client) GetRouteWithFilters(ctx context.Context, routeID string, opts RouteDetailOptions, reqOpts ...owlvigil.RequestOption) (*Route, *owlvigil.ResponseMeta, error) {
	var out Route
	meta, err := c.http.Do(ctx, http.MethodGet, "/gateway/routes/"+url.PathEscape(routeID), opts.values(), nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// PreviewRoute previews model routing for a request.
func (c *Client) PreviewRoute(ctx context.Context, req *PreviewRouteRequest, reqOpts ...owlvigil.RequestOption) (*PreviewRouteResponse, *owlvigil.ResponseMeta, error) {
	var out PreviewRouteResponse
	meta, err := c.http.Do(ctx, http.MethodPost, "/gateway/routes/preview", nil, req, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}
