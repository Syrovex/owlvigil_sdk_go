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
