# Errors

Non-2xx API responses return `*owlvigil.APIError`.

```go
var apiErr *owlvigil.APIError
if errors.As(err, &apiErr) {
    fmt.Println(apiErr.StatusCode)
    fmt.Println(apiErr.Code)
    fmt.Println(apiErr.Message)
    fmt.Println(apiErr.RequestID)
}
```

OAuth2.0 token endpoint failures return `*owlvigil.OAuthError`.

```go
var oauthErr *owlvigil.OAuthError
if errors.As(err, &oauthErr) {
    fmt.Println(oauthErr.StatusCode)
    fmt.Println(oauthErr.ErrorCode)
    fmt.Println(oauthErr.ErrorDescription)
}
```

The SDK redacts configured Gateway keys, OAuth tokens, refresh tokens, client secrets, and webhook secrets from structured SDK errors when those values are available to the SDK.

Requests accept `context.Context`; cancellation and deadlines are propagated to HTTP requests. Retry is bounded and can be disabled:

```go
client := gateway.NewClient(owlvigil.WithoutRetry())
```

Set idempotency keys on supported mutating requests:

```go
key, _, err := client.CreateGatewayKey(ctx, req, owlvigil.WithIdempotencyKey("create-key-001"))
```
