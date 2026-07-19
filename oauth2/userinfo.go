package oauth2

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	owlvigil "github.com/Syrovex/owlvigil_sdk_go"
)

// UserInfo describes the current OAuth2.0 user.
type UserInfo struct {
	Subject string `json:"sub"`
	Email   string `json:"email,omitempty"`
	Name    string `json:"name,omitempty"`
	Picture string `json:"picture,omitempty"`
}

// UserInfo retrieves identity information for an access token.
func (c *Client) UserInfo(ctx context.Context, accessToken string) (*UserInfo, error) {
	u, err := c.endpoint("/oauth/userinfo")
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.cfg.UserAgent)
	resp, err := c.httpClient().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &owlvigil.OAuthError{StatusCode: resp.StatusCode, ErrorCode: http.StatusText(resp.StatusCode)}
	}
	var out UserInfo
	return &out, json.Unmarshal(body, &out)
}
