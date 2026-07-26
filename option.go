package owlvigil

import (
	"context"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	// DefaultGatewayBaseURL deliberately omits /v1 because Gateway client
	// methods include their versioned paths.
	DefaultGatewayBaseURL    = "https://gateway.owlvigil.com"
	DefaultManagementBaseURL = "https://api.owlvigil.com/v1"
	DefaultOAuthBaseURL      = "https://open.owlvigil.com"
	DefaultHTTPTimeout       = 60 * time.Second
)

// Environment represents the API environment.
type Environment string

const (
	// EnvironmentProduction is the production environment (default).
	EnvironmentProduction Environment = "production"

	// EnvironmentStaging is the staging environment for testing.
	EnvironmentStaging Environment = "staging"

	// EnvironmentLocal is the local development environment.
	EnvironmentLocal Environment = "local"
)

// Config contains shared SDK client configuration.
type Config struct {
	BaseURL             string
	HTTPClient          *http.Client
	UserAgent           string
	APIKey              string
	AccessToken         string
	APIKeyProvider      TokenProvider
	AccessTokenProvider TokenProvider
	RetryMax            int
	RetryWait           time.Duration
}

// Option configures SDK clients.
type Option func(*Config)

// TokenProvider returns an authentication token for a request.
type TokenProvider func(context.Context) (string, error)

// DefaultConfig returns conservative default client configuration.
// The baseURL parameter must be a valid HTTP(S) URL.
func DefaultConfig(baseURL string) Config {
	// Validate baseURL is not empty
	if baseURL == "" {
		baseURL = DefaultGatewayBaseURL
	}
	return Config{
		BaseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: DefaultHTTPTimeout},
		UserAgent:  UserAgent(),
		RetryMax:   2,
		RetryWait:  200 * time.Millisecond,
	}
}

// WithBaseURL overrides the default API base URL.
func WithBaseURL(baseURL string) Option {
	return func(c *Config) {
		c.BaseURL = baseURL
	}
}

// WithEnvironment sets the SDK environment.
//
// Supported environments:
//   - EnvironmentProduction (default): Production environment
//   - EnvironmentStaging: Staging environment for testing
//   - EnvironmentLocal: Local development environment
//
// Example:
//
//	client := management.NewClient(
//	    owlvigil.WithEnvironment(owlvigil.EnvironmentStaging),
//	    owlvigil.WithAccessToken(token),
//	)
//
// Note: WithEnvironment modifies the BaseURL. If you also use WithBaseURL,
// call WithEnvironment first, or WithBaseURL will override the environment setting.
func WithEnvironment(env Environment) Option {
	return func(c *Config) {
		if c.BaseURL == "" {
			return
		}

		switch env {
		case EnvironmentStaging:
			c.BaseURL = convertToStagingURL(c.BaseURL)
		case EnvironmentLocal:
			c.BaseURL = convertToLocalURL(c.BaseURL)
		case EnvironmentProduction:
			// Production is default, no change needed
		default:
			// Unknown environment, treat as production
		}
	}
}

// WithEnvironmentFromEnv reads the environment from the OWLVIGIL_ENV environment variable.
//
// Valid values:
//   - "production" or empty: Production environment (default)
//   - "staging": Staging environment
//   - "local": Local development environment
//
// Example:
//
//	// Set environment variable
//	// export OWLVIGIL_ENV=staging
//
//	client := management.NewClient(
//	    owlvigil.WithEnvironmentFromEnv(),
//	    owlvigil.WithAccessToken(os.Getenv("OWLVIGIL_ACCESS_TOKEN")),
//	)
func WithEnvironmentFromEnv() Option {
	env := os.Getenv("OWLVIGIL_ENV")
	if env == "" {
		return func(c *Config) {} // noop, use default
	}
	return WithEnvironment(Environment(env))
}

// convertToStagingURL converts a production URL to staging URL.
func convertToStagingURL(prodURL string) string {
	// Parse and validate URL structure
	u, err := url.Parse(prodURL)
	if err != nil {
		// Fallback to string replacement if parsing fails
		url := strings.Replace(prodURL, "gateway.owlvigil.com", "staginggateway.owlvigil.com", 1)
		url = strings.Replace(url, "api.owlvigil.com", "stagingapi.owlvigil.com", 1)
		url = strings.Replace(url, "open.owlvigil.com", "stagingapi.owlvigil.com", 1)
		return url
	}

	// Replace host if it matches known production hosts
	switch u.Host {
	case "gateway.owlvigil.com":
		u.Host = "staginggateway.owlvigil.com"
	case "api.owlvigil.com":
		u.Host = "stagingapi.owlvigil.com"
	case "open.owlvigil.com":
		u.Host = "stagingapi.owlvigil.com"
	}
	return u.String()
}

// convertToLocalURL converts a production URL to local development URL.
func convertToLocalURL(prodURL string) string {
	// Parse URL to determine service type
	u, err := url.Parse(prodURL)
	if err != nil {
		// Fallback: check string patterns
		if strings.Contains(prodURL, "/v1") {
			return "http://localhost:8081/v1"
		}
		if strings.Contains(prodURL, "gateway.owlvigil.com") {
			return "http://localhost:8080"
		} else if strings.Contains(prodURL, "api.owlvigil.com") {
			return "http://localhost:8081/v1"
		} else if strings.Contains(prodURL, "open.owlvigil.com") {
			return "http://localhost:8081"
		}
		return "http://localhost:8080"
	}

	// Management APIs use the /v1 base path and a distinct local port.
	if strings.HasPrefix(u.Path, "/v1") {
		return "http://localhost:8081/v1"
	}

	// Determine service based on host
	switch u.Host {
	case "gateway.owlvigil.com", "staginggateway.owlvigil.com":
		// Gateway API
		return "http://localhost:8080"
	case "api.owlvigil.com":
		// Management API
		return "http://localhost:8081/v1"
	case "open.owlvigil.com", "stagingapi.owlvigil.com":
		// Management/OAuth API
		return "http://localhost:8081"
	default:
		// Unknown host, default to gateway
		return "http://localhost:8080"
	}
}

// WithHTTPClient overrides the HTTP client used for requests.
func WithHTTPClient(client *http.Client) Option {
	return func(c *Config) {
		if client != nil {
			c.HTTPClient = client
		}
	}
}

// WithTimeout sets the timeout on the SDK HTTP client.
// Note: This creates a shallow copy of the HTTP client. The Transport and other
// pointer fields will be shared with the original client.
func WithTimeout(timeout time.Duration) Option {
	return func(c *Config) {
		if timeout <= 0 {
			return
		}
		client := &http.Client{}
		if c.HTTPClient != nil {
			copyClient := *c.HTTPClient
			client = &copyClient
		}
		client.Timeout = timeout
		c.HTTPClient = client
	}
}

// WithUserAgent overrides the SDK User-Agent header.
func WithUserAgent(userAgent string) Option {
	return func(c *Config) {
		c.UserAgent = userAgent
	}
}

// WithAPIKey configures API key authentication.
func WithAPIKey(apiKey string) Option {
	return func(c *Config) {
		c.APIKey = apiKey
	}
}

// WithAPIKeyProvider configures dynamic API key authentication.
func WithAPIKeyProvider(provider TokenProvider) Option {
	return func(c *Config) {
		c.APIKeyProvider = provider
	}
}

// WithAccessToken configures OAuth2 access-token authentication.
func WithAccessToken(token string) Option {
	return func(c *Config) {
		c.AccessToken = token
	}
}

// WithAccessTokenProvider configures dynamic OAuth2 access-token authentication.
func WithAccessTokenProvider(provider TokenProvider) Option {
	return func(c *Config) {
		c.AccessTokenProvider = provider
	}
}

// WithRetry configures retry attempts and wait duration for transient failures.
func WithRetry(maxAttempts int, wait time.Duration) Option {
	return func(c *Config) {
		c.RetryMax = maxAttempts
		c.RetryWait = wait
	}
}

// WithoutRetry disables SDK retries.
func WithoutRetry() Option {
	return func(c *Config) {
		c.RetryMax = 0
	}
}

// RequestConfig contains per-request configuration.
type RequestConfig struct {
	IdempotencyKey string
	Headers        http.Header
	Query          map[string]string
}

// RequestOption configures a single SDK request.
type RequestOption func(*RequestConfig)

// WithIdempotencyKey sets the Idempotency-Key header for supported mutating requests.
func WithIdempotencyKey(key string) RequestOption {
	return func(c *RequestConfig) {
		c.IdempotencyKey = key
	}
}

// WithHeader sets an additional request header.
func WithHeader(key, value string) RequestOption {
	return func(c *RequestConfig) {
		if c.Headers == nil {
			c.Headers = make(http.Header)
		}
		c.Headers.Set(key, value)
	}
}

// WithQueryParam adds or overrides a query parameter for a single SDK request.
func WithQueryParam(key, value string) RequestOption {
	return func(c *RequestConfig) {
		if c.Query == nil {
			c.Query = make(map[string]string)
		}
		c.Query[key] = value
	}
}

// WithWorkspaceID selects the workspace for Management routes whose Open API
// contract requires a workspace_id query parameter.
func WithWorkspaceID(workspaceID int64) RequestOption {
	return WithQueryParam("workspace_id", strconv.FormatInt(workspaceID, 10))
}
