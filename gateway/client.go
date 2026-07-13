package gateway

import (
	owlvigil "github.com/owlvigil/owlvigil-go"
	"github.com/owlvigil/owlvigil-go/internal/owlvigilhttp"
)

// Client calls OwlVigil Gateway endpoints on api.owlvigil.com by default.
type Client struct {
	http *owlvigilhttp.Client
}

// NewClient creates a Gateway client.
func NewClient(opts ...owlvigil.Option) *Client {
	cfg := owlvigil.DefaultConfig(owlvigil.DefaultGatewayBaseURL)
	for _, opt := range opts {
		opt(&cfg)
	}
	return &Client{http: owlvigilhttp.New(cfg)}
}

// BaseURL returns the configured Gateway base URL.
func (c *Client) BaseURL() string {
	return c.http.BaseURL()
}
