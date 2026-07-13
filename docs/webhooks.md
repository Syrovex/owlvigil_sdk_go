# Webhooks

Verify OwlVigil webhook signatures before decoding or processing the payload.

```go
payload, err := io.ReadAll(r.Body)
if err != nil {
    http.Error(w, "bad request", http.StatusBadRequest)
    return
}

err = webhook.VerifySignature(
    payload,
    r.Header.Get("OwlVigil-Signature"),
    os.Getenv("OWLVIGIL_WEBHOOK_SECRET"),
    webhook.VerifyOptions{},
)
if err != nil {
    http.Error(w, "invalid signature", http.StatusUnauthorized)
    return
}
```

The verifier checks timestamp tolerance and uses constant-time signature comparison.

## Receive safely

Read the unmodified request body before JSON decoding it; signature verification
must use the exact bytes sent by OwlVigil. The default tolerance is five
minutes. Supply `VerifyOptions{Tolerance: ..., Now: ...}` only when your clock
or replay policy requires a different window. Reject missing, malformed, stale,
or invalid signatures and return a non-success HTTP response so the sender can
record the delivery failure.

## Manage outbound endpoints

Inbound verification above is separate from Management API endpoint lifecycle.
Use `CreateWebhookEndpoint`, `GetWebhookEndpoint`, `UpdateWebhookEndpoint`,
`DeleteWebhookEndpoint`, `EnableWebhookEndpoint`, `DisableWebhookEndpoint`,
and `RotateWebhookSecret` to configure where OwlVigil sends events. Endpoint
writes are mutations; use an idempotency key and store a rotated secret only in
a secret manager.

`ListWebhookEventTypes` lists selectable events. `ListWebhookEvents`,
`GetWebhookEvent`, and `ListEndpointEvents` inspect delivery history.
`TestWebhookEndpoint`, `RetryWebhookEvent`, `RedeliverWebhookEvent`, and
`BulkRedeliverWebhookEvents` cause new deliveries, so run them only against a
receiver that can safely process duplicates. Persist an event identifier or use
an idempotent consumer to deduplicate retries.
