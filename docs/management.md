# Management

Management APIs use `https://api.owlvigil.com/open/v1` by default and require a scoped service-account API key.

```go
client := management.NewClient(owlvigil.WithAPIKey(os.Getenv("OWLVIGIL_API_KEY")))
```

## Workspaces

```go
workspaces, _, err := client.ListWorkspaces(ctx, management.ListOptions{Limit: 20})
workspace, _, err := client.GetWorkspace(ctx, 123)
```

## Gateway Keys

```go
key, _, err := client.CreateGatewayKey(
    ctx,
    &management.CreateGatewayKeyRequest{
        Name: "production",
        Scopes: []string{"gateway:invoke"},
    },
    owlvigil.WithIdempotencyKey("create-production-key-001"),
)
```

The `Secret` field is only present when the API returns a one-time visible key secret.

```go
keys, _, err := client.ListGatewayKeys(ctx, management.ListOptions{Limit: 50}, "enabled")
rotated, _, err := client.RotateGatewayKey(ctx, key.ID)
_, err = client.DisableGatewayKey(ctx, key.ID)
```

## Usage, Quota, Billing

```go
summary, _, err := client.GetUsageSummary(ctx)
quota, _, err := client.GetQuota(ctx)
balance, _, err := client.GetBalance(ctx)
invoices, _, err := client.ListInvoices(ctx, management.ListOptions{Limit: 20})
```

## Logs And Traces

```go
logs, _, err := client.ListRequestLogs(ctx, management.ListOptions{Limit: 20}, "")
log, _, err := client.GetRequestLog(ctx, "req_123")
trace, _, err := client.GetTrace(ctx, "trace_123")
```

## Webhooks

```go
endpoint, _, err := client.CreateWebhookEndpoint(ctx, &management.CreateWebhookEndpointRequest{
    URL: "https://app.example.com/owlvigil/webhooks",
    Events: []string{"request.completed"},
})
_, err = client.TestWebhookEndpoint(ctx, endpoint.ID)
events, _, err := client.ListWebhookEvents(ctx, management.ListOptions{Limit: 20})
```
