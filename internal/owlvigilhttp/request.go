package owlvigilhttp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	owlvigil "github.com/owlvigil/owlvigil-go"
)

// Client is the shared HTTP client used by public SDK packages.
type Client struct {
	cfg owlvigil.Config
}

// New creates a shared HTTP client.
func New(cfg owlvigil.Config) *Client {
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = http.DefaultClient
	}
	if cfg.UserAgent == "" {
		cfg.UserAgent = owlvigil.UserAgent()
	}
	return &Client{cfg: cfg}
}

// BaseURL returns the configured base URL.
func (c *Client) BaseURL() string {
	return c.cfg.BaseURL
}

// Config returns the underlying client configuration.
func (c *Client) Config() owlvigil.Config {
	return c.cfg
}

// Do sends a JSON request and decodes either an OwlVigil envelope or a raw JSON response.
func (c *Client) Do(ctx context.Context, method, endpoint string, query url.Values, body any, out any, opts ...owlvigil.RequestOption) (*owlvigil.ResponseMeta, error) {
	bodyBytes, err := encodeBody(body)
	if err != nil {
		return nil, err
	}

	reqCfg := owlvigil.RequestConfig{}
	for _, opt := range opts {
		opt(&reqCfg)
	}

	attempts := c.cfg.RetryMax + 1
	if attempts < 1 {
		attempts = 1
	}

	var lastErr error
	for attempt := 0; attempt < attempts; attempt++ {
		if attempt > 0 && c.cfg.RetryWait > 0 {
			timer := time.NewTimer(c.cfg.RetryWait)
			select {
			case <-ctx.Done():
				timer.Stop()
				return nil, ctx.Err()
			case <-timer.C:
			}
		}

		req, err := c.newRequest(ctx, method, endpoint, mergeQuery(query, reqCfg.Query), bodyBytes, reqCfg)
		if err != nil {
			return nil, err
		}

		resp, err := c.cfg.HTTPClient.Do(req)
		if err != nil {
			lastErr = err
			if !isRetryableError(err) || attempt == attempts-1 {
				return nil, err
			}
			continue
		}

		meta, err := DecodeResponse(resp, out, c.cfg.APIKey, c.cfg.AccessToken)
		if err != nil {
			lastErr = err
			var apiErr *owlvigil.APIError
			if errors.As(err, &apiErr) && isRetryableStatus(apiErr.StatusCode) && attempt < attempts-1 {
				continue
			}
			return meta, err
		}
		return meta, nil
	}

	if lastErr != nil {
		return nil, lastErr
	}
	return nil, fmt.Errorf("owlvigil: request failed")
}

// NewStreamRequest creates an authenticated request for streaming endpoints.
func (c *Client) NewStreamRequest(ctx context.Context, method, endpoint string, body any, opts ...owlvigil.RequestOption) (*http.Request, error) {
	bodyBytes, err := encodeBody(body)
	if err != nil {
		return nil, err
	}
	reqCfg := owlvigil.RequestConfig{}
	for _, opt := range opts {
		opt(&reqCfg)
	}
	return c.newRequest(ctx, method, endpoint, mergeQuery(nil, reqCfg.Query), bodyBytes, reqCfg)
}

// HTTPClient returns the configured HTTP client.
func (c *Client) HTTPClient() *http.Client {
	return c.cfg.HTTPClient
}

// StreamHTTPClient returns a client suitable for long-lived streaming responses.
func (c *Client) StreamHTTPClient() *http.Client {
	if c.cfg.HTTPClient == nil {
		return &http.Client{}
	}
	if c.cfg.HTTPClient.Timeout == 0 {
		return c.cfg.HTTPClient
	}
	copyClient := *c.cfg.HTTPClient
	copyClient.Timeout = 0
	return &copyClient
}

func (c *Client) newRequest(ctx context.Context, method, endpoint string, query url.Values, body []byte, reqCfg owlvigil.RequestConfig) (*http.Request, error) {
	u, err := joinURL(c.cfg.BaseURL, endpoint)
	if err != nil {
		return nil, err
	}
	if len(query) > 0 {
		u.RawQuery = query.Encode()
	}

	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, u.String(), reader)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.cfg.UserAgent)
	apiKey := c.cfg.APIKey
	if c.cfg.APIKeyProvider != nil {
		token, err := c.cfg.APIKeyProvider(ctx)
		if err != nil {
			return nil, fmt.Errorf("owlvigil: api key provider: %w", err)
		}
		apiKey = token
	}
	accessToken := c.cfg.AccessToken
	if c.cfg.AccessTokenProvider != nil {
		token, err := c.cfg.AccessTokenProvider(ctx)
		if err != nil {
			return nil, fmt.Errorf("owlvigil: access token provider: %w", err)
		}
		accessToken = token
	}
	// Prefer accessToken over apiKey when both are provided (Management API uses OAuth2)
	if accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	} else if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}
	if reqCfg.IdempotencyKey != "" {
		req.Header.Set("Idempotency-Key", reqCfg.IdempotencyKey)
	}
	for key, values := range reqCfg.Headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}
	return req, nil
}

func encodeBody(body any) ([]byte, error) {
	if body == nil {
		return nil, nil
	}
	switch v := body.(type) {
	case []byte:
		return v, nil
	case json.RawMessage:
		return v, nil
	default:
		return json.Marshal(body)
	}
}

func joinURL(base, endpoint string) (*url.URL, error) {
	u, err := url.Parse(base)
	if err != nil {
		return nil, err
	}
	if endpoint == "" {
		return u, nil
	}
	if strings.HasPrefix(endpoint, "http://") || strings.HasPrefix(endpoint, "https://") {
		return url.Parse(endpoint)
	}
	basePath := strings.TrimRight(u.Path, "/")
	endpointPath := strings.TrimLeft(endpoint, "/")
	u.Path = path.Join(basePath, endpointPath)
	if strings.HasPrefix(endpoint, "/") && basePath == "" {
		u.Path = "/" + endpointPath
	}
	return u, nil
}

func mergeQuery(base url.Values, extra map[string]string) url.Values {
	if len(extra) == 0 {
		return base
	}
	out := url.Values{}
	for key, values := range base {
		for _, value := range values {
			out.Add(key, value)
		}
	}
	for key, value := range extra {
		out.Set(key, value)
	}
	return out
}

func isRetryableError(err error) bool {
	// Retry on timeout errors
	var netErr interface{ Timeout() bool }
	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}
	// Retry on temporary network errors (DNS, connection reset, etc.)
	var tempErr interface{ Temporary() bool }
	if errors.As(err, &tempErr) && tempErr.Temporary() {
		return true
	}
	return false
}

func isRetryableStatus(status int) bool {
	return status == http.StatusBadGateway || status == http.StatusServiceUnavailable || status == http.StatusGatewayTimeout
}
