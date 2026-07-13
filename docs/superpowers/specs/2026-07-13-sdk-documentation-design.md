# SDK Documentation Completion Design

## Goal

Make every public OwlVigil Go SDK capability discoverable from the README and
documented by its product domain, with safe runnable patterns and an automated
coverage check.

## Information Architecture

`README.md` remains the entry point. `docs/management.md` becomes the Open API
index and links to focused guides rather than duplicating every endpoint.

New Management guides are grouped by the decisions a caller makes:

1. `access-control.md`: workspaces, teams, members, invitations, roles, and
   permissions.
2. `model-routing.md`: catalog models, routes, providers, and gateway policy.
3. `financial-governance.md`: budgets, spend limits, thresholds, governance,
   and spend summaries.
4. `account.md`: profile, password, support, notification preferences, and
   invite links.

Existing `billing.md`, `teams.md`, `environments.md`, and the Gateway,
OAuth2, Webhook, Streaming, Pagination, Errors, Authentication, and
Troubleshooting guides are expanded where they lack operational constraints.

## Documentation Contract

Each public product domain has a page linked from `README.md` and
`docs/management.md`. Each page states the required credential, contains at
least one API-accurate example, distinguishes read-only from mutating calls,
and records pagination, idempotency, or secret-handling rules when relevant.

The documentation does not reproduce all request/response fields. GoDoc is
the type-level reference; the guides explain choosing and safely composing
operations. Every Management source domain maps to one guide, documented in
`docs/management.md`.

## Validation

`scripts/check-docs.sh` verifies that required documentation files exist, all
Management source domains are represented by the index, and README references
every top-level guide. `go test ./...` continues to compile examples and check
HTTP contracts. The checker is suitable for CI once the workflow token can be
granted permission to publish `.github/workflows/test.yml`.

## Non-goals

This change does not invent server behavior, expose a generated endpoint
reference, or make live API calls. It documents only the SDK methods and types
present in this repository.
