# OAuth2.0

OAuth2.0 helpers use `https://open.owlvigil.com` by default.

For staging tests, set `OWLVIGIL_ENV=staging` and pass
`owlvigil.WithEnvironmentFromEnv()` when creating the OAuth2.0 client:

```go
client := oauth2.NewClient(owlvigil.WithEnvironmentFromEnv())
```

## Authorization URL

```go
client := oauth2.NewClient(owlvigil.WithEnvironmentFromEnv())

url, err := client.AuthorizationURL(oauth2.AuthCodeOptions{
    ClientID:    "client_123",
    RedirectURI: "https://app.example.com/callback",
    Scopes:      []string{"workspace:read", "gateway:write"},
    State:       "state_123",
})
```

PKCE parameters are supported:

```go
url, err := client.AuthorizationURL(oauth2.AuthCodeOptions{
    ClientID:            "client_123",
    RedirectURI:         "https://app.example.com/callback",
    Scopes:              []string{"workspace:read"},
    State:               "state_123",
    CodeChallenge:       codeChallenge,
    CodeChallengeMethod: "S256",
})
```

## Token Exchange

```go
token, err := client.Exchange(ctx, oauth2.TokenExchangeRequest{
    ClientID:     "client_123",
    ClientSecret: os.Getenv("OWLVIGIL_CLIENT_SECRET"),
    Code:         r.URL.Query().Get("code"),
    RedirectURI:  "https://app.example.com/callback",
    CodeVerifier: codeVerifier,
})
```

## Refresh

```go
token, err := client.Refresh(ctx, oauth2.RefreshTokenRequest{
    ClientID:     "client_123",
    ClientSecret: os.Getenv("OWLVIGIL_CLIENT_SECRET"),
    RefreshToken: refreshToken,
})
```

## Client Credentials

Use client credentials for server-to-server integrations that do not act on an
interactive user session.

```go
token, err := client.ClientCredentials(ctx, oauth2.ClientCredentialsRequest{
    ClientID:     "client_123",
    ClientSecret: os.Getenv("OWLVIGIL_CLIENT_SECRET"),
    Scopes:       []string{"workspace:read", "gateway:read"},
})
```

Client credentials tokens do not include a refresh token. Request a new access
token when the previous one expires.

## Userinfo

```go
user, err := client.UserInfo(ctx, token.AccessToken)
```

## Revoke

```go
err := client.Revoke(ctx, token.RefreshToken)
```

Token endpoint failures return `*owlvigil.OAuthError`.
