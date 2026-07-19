package management

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	owlvigil "github.com/Syrovex/owlvigil_sdk_go"
)

// Model describes a gateway model.
type Model struct {
	ID            string             `json:"id"`
	ModelID       string             `json:"model_id,omitempty"`
	Name          string             `json:"name"`
	Provider      string             `json:"provider"`
	Developer     string             `json:"developer,omitempty"`
	Capabilities  []string           `json:"capabilities,omitempty"`
	ContextWindow int                `json:"context_window,omitempty"`
	Status        string             `json:"status"`
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
	ID              string   `json:"id"`
	RouteID         string   `json:"route_id,omitempty"`
	ModelID         string   `json:"model_id"`
	Model           string   `json:"model,omitempty"`
	Providers       []string `json:"providers,omitempty"`
	Priority        int      `json:"priority"`
	FallbackEnabled bool     `json:"fallback_enabled"`
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
	WorkspaceID int64          `json:"workspace_id,omitempty"`
	Model       string         `json:"model"`
	KeyID       int64          `json:"key_id,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

// PreviewRouteResponse describes routing preview result.
type PreviewRouteResponse struct {
	Provider  string   `json:"provider"`
	Channel   string   `json:"channel,omitempty"`
	Fallbacks []string `json:"fallbacks,omitempty"`
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
	var out ListResponse[Route]
	meta, err := c.http.Do(ctx, http.MethodGet, "/gateway/routes", opts.values(), nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// GetRoute retrieves a gateway route by ID.
func (c *Client) GetRoute(ctx context.Context, routeID string, reqOpts ...owlvigil.RequestOption) (*Route, *owlvigil.ResponseMeta, error) {
	var out Route
	meta, err := c.http.Do(ctx, http.MethodGet, "/gateway/routes/"+url.PathEscape(routeID), nil, nil, &out, reqOpts...)
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
