package gateway

import (
	"context"
	"net/http"
	"net/url"

	owlvigil "github.com/Syrovex/owlvigil_sdk_go"
)

// Model describes a Gateway model.
type Model struct {
	ID      string `json:"id"`
	Object  string `json:"object,omitempty"`
	OwnedBy string `json:"owned_by,omitempty"`
}

// ModelsResponse is returned by ListModels.
type ModelsResponse struct {
	Object string  `json:"object"`
	Data   []Model `json:"data"`
}

// ListModels calls GET /v1/models.
func (c *Client) ListModels(ctx context.Context, opts ...owlvigil.RequestOption) (*ModelsResponse, *owlvigil.ResponseMeta, error) {
	var out ModelsResponse
	meta, err := c.http.Do(ctx, http.MethodGet, "/v1/models", nil, nil, &out, opts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// GetModel calls GET /v1/models/{model}.
func (c *Client) GetModel(ctx context.Context, model string, opts ...owlvigil.RequestOption) (*Model, *owlvigil.ResponseMeta, error) {
	var out Model
	meta, err := c.http.Do(ctx, http.MethodGet, "/v1/models/"+url.PathEscape(model), nil, nil, &out, opts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}
