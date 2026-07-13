package oauth2

import (
	"net/url"
	"path"
	"strings"
)

// AuthCodeOptions configures OAuth2.0 authorization URL generation.
type AuthCodeOptions struct {
	ClientID             string
	RedirectURI          string
	Scopes               []string
	State                string
	CodeChallenge        string
	CodeChallengeMethod  string
	AdditionalParameters url.Values
}

// AuthorizationURL builds an OAuth2.0 authorization URL.
func (c *Client) AuthorizationURL(opts AuthCodeOptions) (string, error) {
	u, err := url.Parse(c.cfg.BaseURL)
	if err != nil {
		return "", err
	}
	u.Path = path.Join(strings.TrimRight(u.Path, "/"), "/oauth/authorize")
	q := u.Query()
	q.Set("response_type", "code")
	q.Set("client_id", opts.ClientID)
	q.Set("redirect_uri", opts.RedirectURI)
	if len(opts.Scopes) > 0 {
		q.Set("scope", strings.Join(opts.Scopes, " "))
	}
	if opts.State != "" {
		q.Set("state", opts.State)
	}
	if opts.CodeChallenge != "" {
		q.Set("code_challenge", opts.CodeChallenge)
		if opts.CodeChallengeMethod != "" {
			q.Set("code_challenge_method", opts.CodeChallengeMethod)
		}
	}
	for key, values := range opts.AdditionalParameters {
		for _, value := range values {
			q.Add(key, value)
		}
	}
	u.RawQuery = q.Encode()
	return u.String(), nil
}
