package management

import (
	owlvigil "github.com/Syrovex/owlvigil_sdk_go"
	"github.com/Syrovex/owlvigil_sdk_go/internal/owlvigilhttp"
)

// Client calls OwlVigil OpenAPI Management endpoints on api.owlvigil.com by default.
type Client struct {
	http *owlvigilhttp.Client
}

// NewClient creates a Management client.
func NewClient(opts ...owlvigil.Option) *Client {
	cfg := owlvigil.DefaultConfig(owlvigil.DefaultManagementBaseURL)
	for _, opt := range opts {
		opt(&cfg)
	}
	return &Client{http: owlvigilhttp.New(cfg)}
}

// BaseURL returns the configured Management base URL.
func (c *Client) BaseURL() string {
	return c.http.BaseURL()
}
