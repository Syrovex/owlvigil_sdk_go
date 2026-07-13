# Authentication

OwlVigil has two public calling surfaces:

| Surface | Default URL | Auth |
| --- | --- | --- |
| Gateway model calls | `https://api.owlvigil.com` | `Authorization: Bearer ov_sk_xxx` |
| Management Open API | `https://api.owlvigil.com/open/v1` | `Authorization: Bearer <OWLVIGIL_API_KEY>` |

Gateway keys are for model invocation. Management operations such as listing workspaces, creating Gateway keys, reading usage, and managing webhooks require a scoped service-account API key. OAuth2.0 remains available only for the separate OAuth endpoints.

Static tokens:

```go
gatewayClient := gateway.NewClient(owlvigil.WithAPIKey(os.Getenv("OWLVIGIL_GATEWAY_KEY")))
managementClient := management.NewClient(owlvigil.WithAPIKey(os.Getenv("OWLVIGIL_API_KEY")))
```

Dynamic token providers:

```go
managementClient := management.NewClient(owlvigil.WithAPIKeyProvider(func(ctx context.Context) (string, error) {
    return tokenStore.CurrentAPIKey(ctx)
}))
```

The SDK sets `User-Agent: owlvigil-go/<version>` by default.

## Select the right credential

Use a Gateway key only with `gateway.NewClient`; use a scoped service-account
API key only with `management.NewClient`; use OAuth access tokens only with the
OAuth2 client and the APIs for which they were issued. A `403` normally means a
valid credential lacks the required workspace or scope, while `401` usually
means the credential is absent, disabled, expired, or sent to the wrong API
surface.

For short-lived or rotated values, prefer `WithAPIKeyProvider` or
`WithAccessTokenProvider`. The provider is called for each request, so it can
read a refreshed value from an in-memory cache or secret manager. Keep provider
errors free of token text because they may be surfaced to callers.

## Rotation and storage

Store only variable names and safe placeholders in `.env.example`; `.env` is
local and ignored by Git. Gateway-key creation and rotation can expose a
one-time secret: write it directly to a secret store, update consumers, then
disable the old key. Never send API keys, OAuth client secrets, refresh tokens,
or webhook signing secrets to logs, analytics, issue trackers, or browser
clients.
