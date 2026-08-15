# owlvigil-go

[![CI](https://github.com/Syrovex/owlvigil_sdk_go/actions/workflows/test.yml/badge.svg)](https://github.com/Syrovex/owlvigil_sdk_go/actions/workflows/test.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/Syrovex/owlvigil_sdk_go.svg)](https://pkg.go.dev/github.com/Syrovex/owlvigil_sdk_go)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

Go SDK for OwlVigil Gateway model calls and Open API management workflows.

Documentation: [English](docs/en-US/README.md) | [简体中文](docs/zh-CN/README.md)

Gateway model calls use `https://gateway.owlvigil.com` with a Gateway key. OpenAPI Management calls use `https://api.owlvigil.com/v1` with a scoped service-account API key.

## Install

<!-- evidence: go.mod, .env.example, scripts/check-docs.sh -->
```bash
go get github.com/Syrovex/owlvigil_sdk_go
```

## Run examples

Copy the complete example configuration, then set only the credentials needed
by the example you want to run. The template contains safe variable names and
placeholders without secrets.

<!-- evidence: go.mod, .env.example, scripts/check-docs.sh -->
```bash
cp .env.example .env
# Edit .env, then run an example.
go run ./examples/gateway-models/main.go
```

You can also use shell variables instead of `.env`; exported values always take
precedence over values in `.env`:

<!-- evidence: go.mod, .env.example, scripts/check-docs.sh -->
```bash
export OWLVIGIL_GATEWAY_KEY='your-gateway-key'
go run ./examples/gateway-models/main.go
```

## Gateway Quickstart

The runnable Gateway examples automatically load the nearest `.env` file without overriding environment variables that are already exported. Set `OWLVIGIL_GATEWAY_KEY` in the repository-root `.env`, then run `go run ./examples/gateway-models/main.go` directly.

<!-- evidence: examples/compile_test.go, management/all_operations_usecase_test.go -->
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

<!-- evidence: examples/compile_test.go, management/all_operations_usecase_test.go -->
```go
client := management.NewClient(owlvigil.WithAPIKey(os.Getenv("OWLVIGIL_API_KEY")))

keys, _, err := client.ListGatewayKeys(
	ctx,
	management.ListOptions{Limit: 20},
	"",
	owlvigil.WithWorkspaceID(workspaceID),
)
if err != nil {
	return err
}
fmt.Println(keys.Items)

```

## Docs

- [English documentation](docs/en-US/README.md)
- [简体中文文档](docs/zh-CN/README.md)

## Verify

<!-- evidence: go.mod, .env.example, scripts/check-docs.sh -->
```bash
sh scripts/check-docs.sh
sh scripts/check-openapi-alignment.sh ../owlvigil_openapi
go test ./...
```

The documentation check validates public-guide navigation, local Markdown links,
and Management API-domain coverage. The Go suite validates SDK HTTP contracts
and example behavior against local `httptest` servers. To run the opt-in,
read-only live smoke with
credentials from `.env`, use:

<!-- evidence: go.mod, .env.example, scripts/check-docs.sh -->
```bash
sh scripts/test-live-readonly.sh
```

The live smoke validates only the configured, read-only Gateway and Management
surfaces. Do not run `examples/openapi-smoke` against a shared workspace: it
intentionally creates, updates, and deletes test resources.

The SDK is HTTP-only. It does not import OwlVigil dashboard/backend code and does not connect directly to databases, Redis, queues, or server-side infrastructure.

## Community and security

- Read [CONTRIBUTING.md](CONTRIBUTING.md) before opening a pull request.
- Report vulnerabilities using the process in [SECURITY.md](SECURITY.md).
- Participation is governed by [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md).
- Releases follow semantic versioning and are documented in [CHANGELOG.md](CHANGELOG.md).
