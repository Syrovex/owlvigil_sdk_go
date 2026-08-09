# Contributing to OwlVigil Go

Thank you for helping improve the OwlVigil Go SDK.

## Development setup

Install the Go version declared in `go.mod`, clone the repository, and run:

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

Before submitting, ensure the working tree contains no generated artifacts,
editor settings, local investigation notes, or secrets.

## Releases

Update `Version` and `CHANGELOG.md` together, run every validation command, and
create a matching signed semantic-version tag only from a clean, reviewed
commit.
