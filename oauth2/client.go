package oauth2

import (
	"net/http"

	owlvigil "github.com/owlvigil/owlvigil-go"
)

// Client builds OAuth2.0 authorization URLs and calls OwlVigil OAuth endpoints.
type Client struct {
	cfg owlvigil.Config
}

// NewClient creates an OAuth2.0 helper client.
func NewClient(opts ...owlvigil.Option) *Client {
	cfg := owlvigil.DefaultConfig(owlvigil.DefaultOAuthBaseURL)
	for _, opt := range opts {
		opt(&cfg)
	}
	return &Client{cfg: cfg}
}

// BaseURL returns the configured OAuth2.0 issuer URL.
func (c *Client) BaseURL() string {
	return c.cfg.BaseURL
}

func (c *Client) httpClient() *http.Client {
	if c.cfg.HTTPClient != nil {
		return c.cfg.HTTPClient
	}
	return http.DefaultClient
}
