# Errors and Troubleshooting

## Errors

Non-2xx API responses return `*owlvigil.APIError`.

<!-- evidence: internal/owlvigilhttp/request_test.go, internal/owlvigilhttp/response_test.go -->
```go
var apiErr *owlvigil.APIError
if errors.As(err, &apiErr) {
	fmt.Println(apiErr.StatusCode)
	fmt.Println(apiErr.Code)
	fmt.Println(apiErr.Message)
	fmt.Println(apiErr.RequestID)
}

```

The SDK redacts configured Gateway keys, Management API keys, provider credentials, and webhook secrets from structured SDK errors when those values are available to the SDK.

Requests accept `context.Context`; cancellation and deadlines are propagated to HTTP requests. Retry is bounded and can be disabled:

<!-- evidence: internal/owlvigilhttp/request_test.go, internal/owlvigilhttp/response_test.go -->
```go
client := gateway.NewClient(owlvigil.WithoutRetry())

```

Set idempotency keys on supported mutating requests:

<!-- evidence: internal/owlvigilhttp/request_test.go, internal/owlvigilhttp/response_test.go -->
```go
key, _, err := client.CreateGatewayKey(ctx, req, owlvigil.WithIdempotencyKey("create-key-001"))
if err != nil {
	return err
}
_ = key

```

### Retry and recovery

Read-only requests retry conservatively by default (up to two retries after
the initial attempt, with a fixed short wait). The SDK retries network timeout
and temporary errors plus HTTP `502`, `503`, and `504`; it does not parse
`Retry-After` or automatically retry `429`. The SDK transport enables retries
for a mutation when an idempotency key is present, but the current service
contract accepts that key only for `CreateGatewayKey` and
`CreateWebhookEndpoint`. Do not attach it to other writes. Use `WithRetry` to tune that behavior or `WithoutRetry` when
the surrounding application owns retries. Do not blindly retry a mutation
after an ambiguous timeout. For either supported create route, reuse the same
key with an identical body. For every other write, read the resource or inspect
the request ID before deciding what to do next.

`ResponseMeta` is returned on successful SDK calls and includes `RequestID`,
`Code`, and `Message`. Preserve `RequestID` alongside application logs for both
successful writes and `APIError` failures. It lets support correlate the exact
server request without receiving credential or payload data.

For `400` and `422`, correct the request rather than retrying. For `401` and
`403`, verify credential type, scope, and workspace. Handle `429` in the caller.
The SDK automatically retries only eligible requests that fail with `502`,
`503`, or `504`.

## Troubleshooting

### Wrong Domain

Gateway model calls use `https://gateway.owlvigil.com`; OpenAPI Management uses `https://api.owlvigil.com/v1`.

### 401 Unauthorized

Check that Gateway calls use a Gateway key, and Management calls use an enabled service-account API key.

### 403 Forbidden

Check workspace permissions and the service-account API key scopes.

### 429 Rate Limited

The SDK does not automatically retry `429` or parse `Retry-After`. Apply the
service-provided delay in the caller. Do not retry a mutation merely by adding
an idempotency key: only Gateway-key creation and Webhook-endpoint creation
document support for that header.

### Request IDs

When reporting an issue, include `APIError.RequestID`, SDK `owlvigil.Version`, endpoint, and a redacted request summary.

### Context Cancellation

If a request returns `context canceled` or `context deadline exceeded`, check the caller's context deadline and the configured HTTP client timeout.

### Duplicate writes after a timeout

If Gateway-key creation or Webhook-endpoint creation fails ambiguously, reuse
the same `owlvigil.WithIdempotencyKey` value with the identical body. For every
other create, update, checkout, invitation, or webhook-delivery operation, do
not automatically retry. Use the returned or logged request ID and an
appropriate read operation to determine whether the first write succeeded.

### Provider or webhook secrets leaked

Rotate the exposed Gateway, provider, Management, or webhook secret immediately;
then update consumers and disable the previous credential. Remove the secret
from logs, tickets, CI output, and local shell history where feasible. Do not
paste it into a support request.

### Documentation and contract mismatch

Run `sh scripts/check-docs.sh` and `go test ./...` from the repository root.
The first checks documentation navigation and Management domain coverage; the
second validates SDK HTTP contracts and compiles all examples without contacting
the live service.
