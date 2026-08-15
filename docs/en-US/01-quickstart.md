# Quickstart

## Choose a client

Create a Gateway client for model inference and a Management client for
workspace configuration. They use different credentials and production hosts:

<!-- evidence: option_test.go, environment_test.go -->
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
[Gateway](04-gateway.md#gateway) for model calls or [Management](05-management.md#management-api) for
workspace operations.

Install:

<!-- evidence: go.mod, examples/compile_test.go -->
```bash
go get github.com/Syrovex/owlvigil_sdk_go
```

Gateway calls go to `https://gateway.owlvigil.com` and use Gateway keys:

<!-- evidence: option_test.go, environment_test.go -->
```go
client := gateway.NewClient(owlvigil.WithAPIKey(os.Getenv("OWLVIGIL_GATEWAY_KEY")))

```

OpenAPI Management calls go to `https://api.owlvigil.com/v1` and use scoped service-account API keys:

<!-- evidence: option_test.go, environment_test.go -->
```go
client := management.NewClient(owlvigil.WithAPIKey(os.Getenv("OWLVIGIL_API_KEY")))

```

For tests or private deployments, override domains:

<!-- evidence: option_test.go, environment_test.go -->
```go
gatewayClient := gateway.NewClient(owlvigil.WithBaseURL("https://gateway.private.example.com"))
managementClient := management.NewClient(owlvigil.WithBaseURL("https://api.private.example.com/v1"))

```

Run local verification:

<!-- evidence: go.mod, examples/compile_test.go -->
```bash
go test ./...
```
