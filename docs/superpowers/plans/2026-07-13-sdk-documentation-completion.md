# SDK Documentation Completion Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Document every public SDK product domain with navigable, safe, API-accurate guides and prevent coverage regressions.

**Architecture:** Retain README as the navigation root and use `docs/management.md` as the Management API index. Keep existing topic guides, add focused guides for uncovered domains, and enforce the information architecture with a shell validation script.

**Tech Stack:** Markdown, POSIX shell, Go SDK source, `go test`.

## Global Constraints

- Use only methods and types present in this repository.
- Never include credentials or non-placeholder secrets in documentation.
- Mark mutating operations and recommend idempotency where request options support it.
- Keep examples compatible with Go 1.25.5 and package APIs.

---

### Task 1: Build documentation navigation and Management coverage map

**Files:** `README.md`, `docs/management.md`

- [ ] Add every top-level guide to the README documentation list.
- [ ] Replace the Management overview with grouped links and an API-domain mapping for all Management source files.
- [ ] Check all Markdown links target existing files.
- [ ] Commit navigation changes with `docs: index management API domains`.

### Task 2: Document access, routing, governance, and account operations

**Files:** `docs/access-control.md`, `docs/model-routing.md`, `docs/financial-governance.md`, `docs/account.md`, `docs/teams.md`

- [ ] Write access-control guidance for workspaces, teams, members, invitations, roles, and permission overrides.
- [ ] Write model-routing guidance for model discovery, route preview, providers, and gateway policies.
- [ ] Write financial-governance guidance for budgets, limits, thresholds, previews, and spend reports.
- [ ] Write account guidance for profile, notifications, invitations, support, and password changes.
- [ ] Link the existing teams guide to the authoritative access-control page.
- [ ] Verify every named SDK method exists and commit with `docs: cover management operations`.

### Task 3: Complete operational and SDK behavior guides

**Files:** `docs/authentication.md`, `docs/errors.md`, `docs/pagination.md`, `docs/streaming.md`, `docs/webhooks.md`, `docs/environments.md`, `docs/gateway.md`, `docs/oauth2.md`, `docs/troubleshooting.md`, `docs/examples.md`, `docs/billing.md`

- [ ] Add credential selection, secret rotation, environment precedence, retry, idempotency, response metadata, pagination, and stream cleanup behavior.
- [ ] Distinguish webhook endpoint management from inbound signature verification.
- [ ] Add endpoint-specific operational guidance for Gateway, OAuth2, billing, troubleshooting, and examples.
- [ ] Verify no unsafe literal credential is present and commit with `docs: complete SDK operations guides`.

### Task 4: Add documentation contract checks and verify the repository

**Files:** `scripts/check-docs.sh`, `AGENTS.md`, `README.md`

- [ ] Write a shell checker for required guides, README links, and Management domain mappings.
- [ ] Document the checker in contributor and user verification commands.
- [ ] Run `sh scripts/check-docs.sh && go test ./...` and commit with `test: validate SDK documentation coverage`.

## Plan Self-Review

The four tasks cover every design requirement: navigation, all twenty
Management domains, operational behavior, safety guidance, and an automated
regression check. All named methods will be verified against repository source;
no task relies on undocumented server behavior.
