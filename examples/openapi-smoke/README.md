# Open API Management smoke

This program is the end-to-end SDK use-case catalog for all 141 published
Management operations. Its test locks the sorted route set to the Dashboard and
Open API catalog fingerprint so a missing or extra operation fails CI.

When the Open API repository is available next to this SDK, compare the live
source catalogs directly:

```bash
sh scripts/check-openapi-alignment.sh ../owlvigil_openapi
```

Run read-only checks with a fresh, unshared Management API key:

```bash
go run ./examples/openapi-smoke
```

Enable reversible resource writes only for a disposable account/workspace:

```bash
OWLVIGIL_SMOKE_WRITES=1 go run ./examples/openapi-smoke
```

The runner prints `PASS`, `FAIL`, or `SKIP` for every use case. `SKIP` means the
operation was not live-tested; it is never counted as a pass. To require every
one of the 141 unique Management contracts to have at least one `PASS`, use:

```bash
OWLVIGIL_SMOKE_WRITES=1 \
OWLVIGIL_SMOKE_REQUIRE_ALL=1 \
go run ./examples/openapi-smoke
```

OAuth helper steps are reported separately and do not affect the 141-operation
Management completeness result. Any `FAIL`, including an OAuth failure when
OAuth credentials are configured, still fails the run.

Billing operations that depend on Stripe require
`OWLVIGIL_STRIPE_TEST_PAYMENT_METHOD_ID`. Top-up and subscription write
scenarios additionally use:

- `OWLVIGIL_SMOKE_TOPUP_AMOUNT`, `OWLVIGIL_SMOKE_SUCCESS_URL`,
  `OWLVIGIL_SMOKE_CANCEL_URL`, and `OWLVIGIL_SMOKE_RETURN_URL`.
- `OWLVIGIL_SMOKE_PLAN_ID`, `OWLVIGIL_SMOKE_PLAN_INTERVAL`,
  `OWLVIGIL_SMOKE_TOPUP_ID`, `OWLVIGIL_SMOKE_INVOICE_ID`, and
  `OWLVIGIL_SMOKE_ORDER_ID` as fallbacks when billing lists are empty.
- `OWLVIGIL_SMOKE_STRIPE_ORDER_ID` and
  `OWLVIGIL_SMOKE_STRIPE_SESSION_ID` for an already completed disposable
  Checkout session.
- `OWLVIGIL_SMOKE_TOPUP_PAYMENT_INTENT_ID` and
  `OWLVIGIL_SMOKE_TOPUP_CLIENT_SECRET` for disposable in-app confirmation.
- `OWLVIGIL_SMOKE_UPGRADE_PLAN_ID`, `OWLVIGIL_SMOKE_UPGRADE_INTERVAL`,
  `OWLVIGIL_SMOKE_DOWNGRADE_PLAN_ID`, and
  `OWLVIGIL_SMOKE_DOWNGRADE_INTERVAL` for explicit subscription transitions.

These values can create test charges and change subscription state. Use Stripe
test mode and an account intended to be destroyed after the run.
Temporary provider tests can use `OWLVIGIL_SMOKE_PROVIDER_TYPE`,
`OWLVIGIL_SMOKE_PROVIDER_API_KEY`, and
`OWLVIGIL_SMOKE_PROVIDER_BASE_URL`.

Account and workspace lifecycle tests additionally use:

- `OWLVIGIL_SMOKE_CURRENT_PASSWORD` and `OWLVIGIL_SMOKE_NEW_PASSWORD`; the
  runner changes the password and immediately restores the original.
- `OWLVIGIL_SMOKE_INVITE_EMAIL` for the current-user referral invitation.
- `OWLVIGIL_SMOKE_MEMBER_EMAIL` for a second registered disposable user that
  can be added, updated, permission-tested, and removed.
- `OWLVIGIL_SMOKE_WORKSPACE_INVITE_EMAIL` for the create/resend/revoke
  workspace-invitation lifecycle.
- `OWLVIGIL_SMOKE_POLICY_ID` for an existing disposable policy that may receive
  a no-op update.
- `OWLVIGIL_SMOKE_AUDIT_LOG_ID`, `OWLVIGIL_SMOKE_MODEL_ID`,
  `OWLVIGIL_SMOKE_ROUTE_ID`, `OWLVIGIL_SMOKE_REQUEST_ID`,
  `OWLVIGIL_SMOKE_TRACE_ID`, and `OWLVIGIL_SMOKE_PAYLOAD_ID` as detail-call
  fallbacks when a new workspace has no list data.
- `OWLVIGIL_SMOKE_WEBHOOK_URL` for a disposable HTTPS receiver and
  `OWLVIGIL_SMOKE_WEBHOOK_EVENT_ID` when testing the endpoint does not create
  a delivery event synchronously.
- `OWLVIGIL_SMOKE_BUDGET_SCOPE_TYPE`, `OWLVIGIL_SMOKE_BUDGET_SCOPE_ID`,
  `OWLVIGIL_SMOKE_BUDGET_ENABLED`, and
  `OWLVIGIL_SMOKE_BUDGET_MONTHLY_AMOUNT` when no scoped cap exists.
- `OWLVIGIL_SMOKE_SPENDING_USER_ID`, `OWLVIGIL_SMOKE_SPENDING_DAILY_LIMIT`,
  `OWLVIGIL_SMOKE_SPENDING_WEEKLY_LIMIT`, and
  `OWLVIGIL_SMOKE_SPENDING_MONTHLY_LIMIT` when no user limit exists.

The full write run submits one support request and sends invitation email, so
use only isolated accounts and inboxes intended for automated testing.

Never use an API key pasted into chat or committed to a file. Put a newly
created disposable key in the process environment or a local ignored `.env`.
