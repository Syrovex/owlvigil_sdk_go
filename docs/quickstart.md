# Quickstart

## Choose a client

Create a Gateway client for model inference and a Management client for
workspace configuration. They use different credentials and production hosts:

```go
gatewayClient := gateway.NewClient(
    owlvigil.WithAPIKey(os.Getenv("OWLVIGIL_GATEWAY_KEY")),
)
managementClient := management.NewClient(
    owlvigil.WithAPIKey(os.Getenv("OWLVIGIL_API_KEY")),
)
```

Use a context with a deadline for application requests, check every returned
error, and record the response metadata request ID for writes. Continue with
[Gateway](gateway.md) for model calls or [Management](management.md) for
workspace operations.

Install:

```bash
go get github.com/Syrovex/owlvigil_sdk_go
```

Gateway calls go to `https://gateway.owlvigil.com` and use Gateway keys:

```go
client := gateway.NewClient(owlvigil.WithAPIKey(os.Getenv("OWLVIGIL_GATEWAY_KEY")))
```

OpenAPI Management calls go to `https://api.owlvigil.com/v1` and use scoped service-account API keys:

```go
client := management.NewClient(owlvigil.WithAPIKey(os.Getenv("OWLVIGIL_API_KEY")))
```

For tests or private deployments, override domains:

```go
gatewayClient := gateway.NewClient(owlvigil.WithBaseURL("https://gateway.private.example.com"))
managementClient := management.NewClient(owlvigil.WithBaseURL("https://api.private.example.com/v1"))
oauthClient := oauth2.NewClient(owlvigil.WithBaseURL("https://open.private.example.com"))
```

Run local verification:

```bash
go test ./...
```
