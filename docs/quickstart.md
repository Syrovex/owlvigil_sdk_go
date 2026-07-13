# Quickstart

Install:

```bash
go get github.com/owlvigil/owlvigil-go
```

Gateway calls go to `https://api.owlvigil.com` and use Gateway keys:

```go
client := gateway.NewClient(owlvigil.WithAPIKey(os.Getenv("OWLVIGIL_GATEWAY_KEY")))
```

Management calls go to `https://api.owlvigil.com/open/v1` and use scoped service-account API keys:

```go
client := management.NewClient(owlvigil.WithAPIKey(os.Getenv("OWLVIGIL_API_KEY")))
```

For tests or private deployments, override domains:

```go
gatewayClient := gateway.NewClient(owlvigil.WithBaseURL("https://api.private.example.com"))
managementClient := management.NewClient(owlvigil.WithBaseURL("https://open.private.example.com/open/v1"))
oauthClient := oauth2.NewClient(owlvigil.WithBaseURL("https://open.private.example.com"))
```

Run local verification:

```bash
go test ./...
```
