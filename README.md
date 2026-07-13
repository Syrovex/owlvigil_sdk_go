# owlvigil-go

Go SDK for OwlVigil Gateway model calls and Open API management workflows.

Gateway model calls use `https://api.owlvigil.com` with a Gateway key. Management API calls use `https://api.owlvigil.com/open/v1` with a scoped service-account API key.

## Install

```bash
go get github.com/owlvigil/owlvigil-go
```

## Run examples

Copy the complete example configuration, then set only the credentials needed
by the example you want to run. The template contains Gateway, Management, and
OAuth variable names without secrets.

```bash
cp .env.example .env
# Edit .env, then run an example.
go run ./examples/gateway-models/main.go
```

You can also use shell variables instead of `.env`; exported values always take
precedence over values in `.env`:

```bash
export OWLVIGIL_GATEWAY_KEY='your-gateway-key'
go run ./examples/gateway-models/main.go
```

## Gateway Quickstart

The runnable Gateway examples automatically load the nearest `.env` file without overriding environment variables that are already exported. Set `OWLVIGIL_GATEWAY_KEY` in the repository-root `.env`, then run `go run ./examples/gateway-models/main.go` directly.

```go
client := gateway.NewClient(owlvigil.WithAPIKey(os.Getenv("OWLVIGIL_GATEWAY_KEY")))

resp, meta, err := client.CreateChatCompletion(ctx, &gateway.ChatCompletionRequest{
    Model: "gpt-4o-mini",
    Messages: []gateway.Message{
        {Role: "user", Content: "Say hello from OwlVigil."},
    },
})
if err != nil {
    return err
}
fmt.Println(meta.RequestID, resp.ID)
```

## Management Quickstart

```go
client := management.NewClient(owlvigil.WithAPIKey(os.Getenv("OWLVIGIL_API_KEY")))

keys, _, err := client.ListGatewayKeys(ctx, management.ListOptions{Limit: 20}, "")
if err != nil {
    return err
}
fmt.Println(keys.Items)
```

## OAuth2.0

Use the OAuth2.0 helper for user authorization and Management API access tokens:

```go
auth := oauth2.NewClient(owlvigil.WithEnvironmentFromEnv())
url, err := auth.AuthorizationURL(oauth2.AuthCodeOptions{
    ClientID:    "client_123",
    RedirectURI: "https://app.example.com/callback",
    Scopes:      []string{"workspace:read", "gateway:write"},
    State:       "state_123",
})
```

## Docs

- [Quickstart](docs/quickstart.md)
- [Authentication](docs/authentication.md)
- [Gateway](docs/gateway.md)
- [Management](docs/management.md)
- [OAuth2.0](docs/oauth2.md)
- [Streaming](docs/streaming.md)
- [Webhooks](docs/webhooks.md)
- [Errors](docs/errors.md)
- [Pagination](docs/pagination.md)
- [Examples](docs/examples.md)
- [Troubleshooting](docs/troubleshooting.md)

## Verify

```bash
go test ./...
```

This is the default offline suite: SDK HTTP contracts and example behavior run
against local `httptest` servers. To run the opt-in, read-only live smoke with
credentials from `.env`, use:

```bash
sh scripts/test-live-readonly.sh
```

The live smoke validates only the configured, read-only Gateway and Management
surfaces. Do not run `examples/openapi-smoke` against a shared workspace: it
intentionally creates, updates, and deletes test resources.

The SDK is HTTP-only. It does not import OwlVigil dashboard/backend code and does not connect directly to databases, Redis, queues, or server-side infrastructure.
