# Repository Guidelines

## Project Structure & Module Organization

This repository is the `github.com/owlvigil/owlvigil-go` Go SDK. Public API
packages are organized by surface: `gateway/` for model calls,
`management/` for `/open/v1` operations, `oauth2/` for OAuth flows, and
`webhook/` for signature verification. Shared HTTP, retry, redaction, and SSE
code belongs in `internal/`. Runnable programs live in `examples/`; their
environment-file loader is in `examples/internal/envfile`. Keep API guides in
`docs/` and OpenSpec work in `openspec/`.

## Build, Test, and Development Commands

Use Go 1.25.5 or later, as declared in `go.mod`.

```bash
go test ./...                         # run all unit and contract tests
go test ./management -run TestName    # run one management test
go run ./examples/gateway-models/main.go
```

Copy `.env.example` to `.env` for examples. Exported variables override the
file. Use `OWLVIGIL_GATEWAY_KEY` for gateway examples and
`OWLVIGIL_API_KEY` for management examples; never print or commit either.

## Coding Style & Naming Conventions

Format all modified Go files with `gofmt`. Follow idiomatic Go: tabs for
indentation, `MixedCaps` identifiers, `ID`, `URL`, and `API` initialisms, and
short receiver names such as `c *Client`. Keep public types and methods
documented. Add endpoint code to the appropriate public package; do not expose
implementation details from `internal/`.

## Testing Guidelines

Write table-driven or focused `httptest` contract tests for every new HTTP
method. Verify HTTP method, path, query parameters, request body, response
decoding, and API-error handling. Name tests `Test<TypeOrMethod>_<Behavior>`.
Examples must compile under `go test ./examples/...`; keep live tests opt-in
and read-only by default. Run `go test ./...` before handing off changes.

## Commit & Pull Request Guidelines

Git history is not available in this workspace, so use concise imperative
commits such as `add provider management client`. Keep each commit focused.
Pull requests should explain the API contract affected, list test commands and
results, link the relevant issue or OpenSpec change, and call out any live API
or credential requirements.

## Security & Configuration

Keep `.env` local; update `.env.example` only with variable names and safe
placeholders. Redact keys, access tokens, webhook secrets, and provider
credentials from logs, tests, fixtures, and review comments.
