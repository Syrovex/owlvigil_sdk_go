# Contributing to OwlVigil Go

Thank you for helping improve the OwlVigil Go SDK.

## Development setup

Install the Go version declared in `go.mod`, clone the repository, and run:

<!-- evidence: go.mod, scripts/check-docs.sh -->
```bash
go test -race -count=1 ./...
go vet ./...
golangci-lint run ./...
sh scripts/check-docs.sh
```

Do not add credentials to tests, fixtures, examples, logs, or commits. Copy
`.env.example` to `.env` for local examples; `.env` must remain untracked.

## Changes

- Keep each pull request focused and explain the affected API contract.
- Add focused `httptest` coverage for HTTP method, path, query, body, response,
  and API-error behavior.
- Preserve backward compatibility unless the change is explicitly documented
  as breaking.
- Update public documentation and `CHANGELOG.md` when behavior changes.
- Keep live tests opt-in and read-only by default.

## Documentation examples

Every fenced Go example must be immediately preceded by an evidence marker:

```markdown
<!-- evidence: gateway/client_test.go, examples/gateway-chat/main.go -->
```

Use test or runnable-example paths that prove the API contract shown by the
block. `sh scripts/check-docs.sh` checks every Go block for valid syntax,
current exported types and fields, public client methods, explicit error
handling, use of non-compatibility fields, and existing evidence files. Full
programs should live under `examples/` and be compiled by
`go test ./examples/...`; keep the corresponding documentation block in sync
with a focused test when it is duplicated for copy-and-paste use.

Before submitting, ensure the working tree contains no generated artifacts,
editor settings, local investigation notes, or secrets.

## Releases

Update `Version` and `CHANGELOG.md` together, run every validation command, and
create a matching signed semantic-version tag only from a clean, reviewed
commit.
