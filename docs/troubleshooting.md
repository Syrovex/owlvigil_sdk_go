# Troubleshooting

## Wrong Domain

Gateway model calls use `https://gateway.owlvigil.com`; OpenAPI Management uses `https://api.owlvigil.com/v1`. OAuth2.0 flows use `https://open.owlvigil.com`.

## 401 Unauthorized

Check that Gateway calls use a Gateway key, and Management calls use an enabled service-account API key.

## 403 Forbidden

Check workspace permissions and the service-account API key scopes.

## 429 Rate Limited

Retry after the service-provided delay when available. Avoid retrying non-idempotent mutations unless you set an idempotency key.

## Request IDs

When reporting an issue, include `APIError.RequestID`, SDK `owlvigil.Version`, endpoint, and a redacted request summary.

## Context Cancellation

If a request returns `context canceled` or `context deadline exceeded`, check the caller's context deadline and the configured HTTP client timeout.

## Duplicate writes after a timeout

For create, update, checkout, invitation, and webhook-delivery operations,
reuse the same `owlvigil.WithIdempotencyKey` value when retrying an ambiguous
failure. If the original call may have reached the service, use the returned or
logged request ID and a read operation before issuing a new write.

## Provider or webhook secrets leaked

Rotate the exposed Gateway, provider, OAuth, or webhook secret immediately;
then update consumers and disable the previous credential. Remove the secret
from logs, tickets, CI output, and local shell history where feasible. Do not
paste it into a support request.

## Documentation and contract mismatch

Run `sh scripts/check-docs.sh` and `go test ./...` from the repository root.
The first checks documentation navigation and Management domain coverage; the
second validates SDK HTTP contracts and compiles all examples without contacting
the live service.
