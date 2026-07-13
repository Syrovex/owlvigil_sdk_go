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

`Revoke` invalidates an access or refresh token through the OAuth endpoint.
Revoke tokens when a user disconnects the integration, then delete the local
token record. Treat revocation as best-effort during logout: clear the local
session even if the network call fails, and retain only a redacted diagnostic.

## Safe callback handling

Generate a cryptographically random `State` value for every authorization
attempt, store it in the user's server-side session, and compare it before
calling `Exchange`. Do not put a client secret in a browser, mobile bundle, or
redirect URL. Use the exact registered redirect URI, and store refresh tokens
in an encrypted server-side store.

Use `Refresh` before an access token expires and `ClientCredentials` only for a
confidential server-to-server client. OAuth failures are `*owlvigil.OAuthError`;
see [Errors](errors.md) for handling and redaction guidance.

```go
err := client.Revoke(ctx, token.RefreshToken)
```

Token endpoint failures return `*owlvigil.OAuthError`.
