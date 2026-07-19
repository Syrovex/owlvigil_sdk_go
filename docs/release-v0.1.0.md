# v0.1.0 Draft Release Notes

Initial customer-facing Go SDK for OwlVigil.

## Highlights

- Gateway client for chat completions, responses, models, embeddings, Anthropic-compatible messages, and SSE streaming on `https://gateway.owlvigil.com`.
- OpenAPI Management client for workspaces, Gateway keys, usage, quota, balance, invoices, logs, traces, webhook endpoints, and webhook events on `https://api.owlvigil.com/v1`.
- OAuth2.0 helpers for authorization URLs, token exchange, refresh, userinfo, and revocation on `https://open.owlvigil.com`.
- Shared HTTP behavior for custom clients, timeouts, retries, request IDs, API errors, redaction, idempotency keys, context cancellation, and `User-Agent: owlvigil-go/<version>`.
- Webhook signature verification with timestamp tolerance and constant-time comparison.
- Runnable examples and customer documentation.

## Verification

```bash
go test ./...
go vet ./...
```

## Tag

Use `v0.1.0` for the first public preview tag after the corresponding OwlVigil API contracts are confirmed.
