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

## Retry and recovery

Clients retry conservatively by default (two retry attempts with a short wait).
Use `WithRetry` to tune that behavior or `WithoutRetry` when the surrounding
application owns retries. Do not blindly retry a mutation after an ambiguous
timeout: reuse the same idempotency key, then read the resource or inspect the
request ID to determine whether the service accepted it.

`ResponseMeta` is returned on successful SDK calls and includes `RequestID`,
`Code`, and `Message`. Preserve `RequestID` alongside application logs for both
successful writes and `APIError` failures. It lets support correlate the exact
server request without receiving credential or payload data.

For `400` and `422`, correct the request rather than retrying. For `401` and
`403`, verify credential type, scope, and workspace. For `429` or transient
`5xx`, back off and retry only when the operation is read-only or idempotent.
