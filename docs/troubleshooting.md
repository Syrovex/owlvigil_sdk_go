# Troubleshooting

## Wrong Domain

Gateway model calls and Management APIs use `https://api.owlvigil.com` (Management paths start with `/open/v1`). OAuth2.0 flows use `https://open.owlvigil.com`.

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
