package management

import (
	"context"
	"net/http"
	"net/url"

	owlvigil "github.com/owlvigil/owlvigil-go"
)

// DocumentationNavigation contains the groups shown in the developer documentation.
type DocumentationNavigation struct {
	Groups []DocumentationGroup `json:"groups"`
}

// DocumentationGroup describes one developer documentation group.
type DocumentationGroup struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

// DocumentedEndpoint describes one endpoint in the developer documentation.
type DocumentedEndpoint struct {
	ID          string `json:"id"`
	Group       string `json:"group"`
	Method      string `json:"method"`
	Path        string `json:"path"`
	Scope       string `json:"scope"`
	Status      string `json:"status"`
	Description string `json:"description"`
}

// DocumentedEndpointListOptions filters documented API endpoints.
type DocumentedEndpointListOptions struct {
	Group  string
	Scope  string
	Status string
}

func (o DocumentedEndpointListOptions) values() url.Values {
	query := url.Values{}
	if o.Group != "" {
		query.Set("group", o.Group)
	}
	if o.Scope != "" {
		query.Set("scope", o.Scope)
	}
	if o.Status != "" {
		query.Set("status", o.Status)
	}
	return query
}

// SDKPackage describes availability of an OwlVigil SDK package.
type SDKPackage struct {
	Language string `json:"language"`
	Package  string `json:"package"`
	Status   string `json:"status"`
}

// DocumentationNavigation retrieves developer documentation navigation groups.
func (c *Client) DocumentationNavigation(ctx context.Context, reqOpts ...owlvigil.RequestOption) (*DocumentationNavigation, *owlvigil.ResponseMeta, error) {
	var out DocumentationNavigation
	meta, err := c.http.Do(ctx, http.MethodGet, "/docs/navigation", nil, nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// ListDocumentedEndpoints lists developer documentation endpoints with optional filters.
func (c *Client) ListDocumentedEndpoints(ctx context.Context, opts DocumentedEndpointListOptions, reqOpts ...owlvigil.RequestOption) (*ListResponse[DocumentedEndpoint], *owlvigil.ResponseMeta, error) {
	var out ListResponse[DocumentedEndpoint]
	meta, err := c.http.Do(ctx, http.MethodGet, "/docs/endpoints", opts.values(), nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// GetDocumentedEndpoint retrieves documentation metadata by endpoint ID.
func (c *Client) GetDocumentedEndpoint(ctx context.Context, endpointID string, reqOpts ...owlvigil.RequestOption) (*DocumentedEndpoint, *owlvigil.ResponseMeta, error) {
	var out DocumentedEndpoint
	meta, err := c.http.Do(ctx, http.MethodGet, "/docs/endpoints/"+url.PathEscape(endpointID), nil, nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// OpenAPISchema retrieves the legacy public OpenAPI schema.
func (c *Client) OpenAPISchema(ctx context.Context, reqOpts ...owlvigil.RequestOption) (map[string]any, *owlvigil.ResponseMeta, error) {
	var out map[string]any
	meta, err := c.http.Do(ctx, http.MethodGet, "/openapi.json", nil, nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return out, meta, nil
}

// SwaggerSchema retrieves the public OpenAPI schema used by developer documentation.
func (c *Client) SwaggerSchema(ctx context.Context, reqOpts ...owlvigil.RequestOption) (map[string]any, *owlvigil.ResponseMeta, error) {
	var out map[string]any
	meta, err := c.http.Do(ctx, http.MethodGet, "/swagger.json", nil, nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return out, meta, nil
}

// SDKPackages lists available OwlVigil SDK packages.
func (c *Client) SDKPackages(ctx context.Context, reqOpts ...owlvigil.RequestOption) (*ListResponse[SDKPackage], *owlvigil.ResponseMeta, error) {
	var out ListResponse[SDKPackage]
	meta, err := c.http.Do(ctx, http.MethodGet, "/sdk/packages", nil, nil, &out, reqOpts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}
