package oauth2

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	owlvigil "github.com/owlvigil/owlvigil-go"
	"github.com/owlvigil/owlvigil-go/internal/owlvigilhttp"
)

// TokenResponse is returned by OAuth2.0 token endpoints.
type TokenResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int       `json:"expires_in,omitempty"`
	ExpiresAt    time.Time `json:"-"`
	Scope        string    `json:"scope,omitempty"`
}

// TokenExchangeRequest exchanges an authorization code for tokens.
type TokenExchangeRequest struct {
	ClientID     string
	ClientSecret string
	Code         string
	RedirectURI  string
	CodeVerifier string
}

// RefreshTokenRequest refreshes an access token.
type RefreshTokenRequest struct {
	ClientID     string
	ClientSecret string
	RefreshToken string
}

// ClientCredentialsRequest exchanges client credentials for a machine access token.
type ClientCredentialsRequest struct {
	ClientID     string
	ClientSecret string
	Scopes       []string
}

// Exchange exchanges an authorization code for OAuth2.0 tokens.
func (c *Client) Exchange(ctx context.Context, req TokenExchangeRequest) (*TokenResponse, error) {
	values := url.Values{}
	values.Set("grant_type", "authorization_code")
	values.Set("client_id", req.ClientID)
	values.Set("code", req.Code)
	values.Set("redirect_uri", req.RedirectURI)
	if req.ClientSecret != "" {
		values.Set("client_secret", req.ClientSecret)
	}
	if req.CodeVerifier != "" {
		values.Set("code_verifier", req.CodeVerifier)
	}
	return c.postForm(ctx, "/oauth/token", values)
}

// Refresh refreshes an access token using a refresh token.
func (c *Client) Refresh(ctx context.Context, req RefreshTokenRequest) (*TokenResponse, error) {
	values := url.Values{}
	values.Set("grant_type", "refresh_token")
	values.Set("client_id", req.ClientID)
	values.Set("refresh_token", req.RefreshToken)
	if req.ClientSecret != "" {
		values.Set("client_secret", req.ClientSecret)
	}
	return c.postForm(ctx, "/oauth/token/refresh", values)
}

// ClientCredentials exchanges client credentials for a machine access token.
func (c *Client) ClientCredentials(ctx context.Context, req ClientCredentialsRequest) (*TokenResponse, error) {
	values := url.Values{}
	values.Set("grant_type", "client_credentials")
	values.Set("client_id", req.ClientID)
	values.Set("client_secret", req.ClientSecret)
	if len(req.Scopes) > 0 {
		values.Set("scope", strings.Join(req.Scopes, " "))
	}
	return c.postForm(ctx, "/oauth/token", values)
}

// Revoke revokes an OAuth2.0 access token or refresh token.
func (c *Client) Revoke(ctx context.Context, token string) error {
	values := url.Values{}
	values.Set("token", token)
	_, err := c.postForm(ctx, "/oauth/revoke", values)
	return err
}

func (c *Client) postForm(ctx context.Context, endpoint string, values url.Values) (*TokenResponse, error) {
	u, err := c.endpoint(endpoint)
	if err != nil {
		return nil, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, u, strings.NewReader(values.Encode()))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("User-Agent", c.cfg.UserAgent)
	resp, err := c.httpClient().Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	secrets := []string{
		values.Get("client_secret"),
		values.Get("refresh_token"),
		values.Get("code"),
		values.Get("code_verifier"),
		values.Get("token"),
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var oauthErr struct {
			Error            string `json:"error"`
			ErrorDescription string `json:"error_description"`
			ErrorURI         string `json:"error_uri"`
		}
		_ = json.Unmarshal(body, &oauthErr)
		return nil, &owlvigil.OAuthError{
			StatusCode:       resp.StatusCode,
			ErrorCode:        owlvigilhttp.Redact(oauthErr.Error, secrets...),
			ErrorDescription: owlvigilhttp.Redact(oauthErr.ErrorDescription, secrets...),
			ErrorURI:         owlvigilhttp.Redact(oauthErr.ErrorURI, secrets...),
		}
	}
	if len(strings.TrimSpace(string(body))) == 0 {
		return nil, nil
	}
	var out TokenResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, err
	}
	// Calculate ExpiresAt based on ExpiresIn
	// Note: This uses local system time. Ensure system clock is accurate.
	// Clock skew may cause tokens to expire earlier or later than expected.
	if out.ExpiresIn > 0 {
		out.ExpiresAt = time.Now().Add(time.Duration(out.ExpiresIn) * time.Second)
	}
	return &out, nil
}

func (c *Client) endpoint(endpoint string) (string, error) {
	u, err := url.Parse(c.cfg.BaseURL)
	if err != nil {
		return "", err
	}
	u.Path = path.Join(strings.TrimRight(u.Path, "/"), endpoint)
	return u.String(), nil
}
