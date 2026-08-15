# Management Routing, Observability, and Billing

## Model Routing and Providers

This guide covers the Management configuration that determines which models are
available and how Gateway traffic is routed. It requires `OWLVIGIL_API_KEY`.
Gateway invocation itself still uses `OWLVIGIL_GATEWAY_KEY`; see
[Gateway](04-gateway.md#gateway).

### Discover models and routes

Use `ListModels` and `GetModel` to present available models. `ListRoutes` and
`GetRoute` inspect configured routing rules. All four are read-only and list
methods use `management.ListOptions`. All four require
`owlvigil.WithWorkspaceID(workspaceID)`. Use `ListRoutesWithFilters` or
`GetRouteWithFilters` when selecting a key or model explicitly.

Before changing a route, use `PreviewRoute` with a
`management.PreviewRouteRequest`. A preview evaluates a proposed route without
making it active, which lets an operator inspect the selected provider and
model before traffic changes.

### Providers

Providers are workspace-scoped upstream credentials. Use `ListProviders`,
`GetProvider`, `CreateProvider`, `UpdateProvider`, and `DeleteProvider`.

<!-- evidence: management/all_operations_usecase_test.go, management/refactored_openapi_contract_test.go -->
```go
provider, _, err := client.CreateProvider(ctx, &management.CreateProviderRequest{
	WorkspaceID:  workspaceID,
	Name:         "primary-openai",
	Type:         "openai",
	APIKey:       os.Getenv("UPSTREAM_PROVIDER_API_KEY"),
	DefaultModel: "gpt-4o-mini",
})
if err != nil {
	return err
}
_ = provider

```

Provider `APIKey` is a secret. Read it only from a secret store or process
environment, never log the request, and never put it in source control. Create
and update calls are mutations and do not document idempotency-key support;
do not automatically retry them after an ambiguous failure.
Deleting a provider can make routes unusable, so preview and migrate routes
first.

### Gateway keys and policies

`ListGatewayKeys`, `GetGatewayKey`, `CreateGatewayKey`, `UpdateGatewayKey`,
`RotateGatewayKey`, `EnableGatewayKey`, `DisableGatewayKey`, and
`DeleteGatewayKey` manage credentials used by callers of the Gateway API.
Except for create, these routes require
`owlvigil.WithWorkspaceID(workspaceID)`; list also requires it.
The key `Secret` is returned only when the service makes it visible. Store it
once, then discard it from application memory and logs.
The `WithResult` action variants return the updated key or delete
confirmation while the original signatures retain compatibility.

`GetGatewayPolicies` retrieves policy configuration for a key.
`PreviewPolicyEffect` evaluates a proposed policy effect, and
`UpdateGatewayPolicy` changes selected fields. Preview before every update and
keep the returned request ID with the change record. Reads and updates require
`owlvigil.WithWorkspaceID(workspaceID)`; previews carry `workspace_id` in the
request body.

### Safe rollout sequence

1. Discover a model and current route with `ListModels` and `ListRoutes`.
2. Create or update the provider with secret-store data.
3. Call `PreviewRoute` or `PreviewPolicyEffect`.
4. Apply the approved write once and retain its request ID.
5. Verify the result through a read-only `GetProvider`, `GetRoute`, or
   `GetGatewayPolicies` call.
6. Rotate or disable old Gateway keys only after dependent callers use the
   replacement.

## Financial Governance

Financial controls are workspace-scoped Management writes. Require
`OWLVIGIL_API_KEY`, surface a confirmation to a human operator, and retain the
returned request ID. Use [Billing](07-management-operations.md#billing-and-subscriptions) for payment and subscription
workflows.

### Inspect before changing

Read the current state with `GetFinancialGovernance`, `GetBudgetCaps`,
`GetSpendingLimits`, `GetFinancialThresholds`, and `GetSpendSummary`. The
spending-limits call is not paginated; use `GetSpendingLimitsWithFilters` for
the published `team_id` and `user_id` filters. `GetQuotaSummary`,
`GetQuotaUsage`, `GetUsageSummary`, `ListUsage`, and
`GetBalanceForWorkspace` provide complementary quota, usage, and balance
evidence. Gateway usage methods require `owlvigil.WithWorkspaceID(workspaceID)`.

<!-- evidence: management/all_operations_usecase_test.go, management/refactored_openapi_contract_test.go -->
```go
governance, meta, err := client.GetFinancialGovernance(ctx, workspaceID)
if err != nil {
	return err
}
slog.Info("loaded financial governance", "request_id", meta.RequestID)
_ = governance

```

### Set controls

`UpdateFinancialGovernance` can update budget caps, spending limits, and
thresholds together. Use the more focused `UpdateBudgetCaps`,
`UpdateSpendingLimits`, or `UpdateFinancialThresholds` when changing one
control family. `UpdateScopeBudgetCap` changes an individual workspace, team,
member, or Gateway-key scope; `UpdateUserSpendingLimit` changes one user's
limit.

<!-- evidence: management/all_operations_usecase_test.go, management/refactored_openapi_contract_test.go -->
```go
warning, critical := 80, 95
thresholds, _, err := client.UpdateFinancialThresholds(ctx, workspaceID,
	&management.UpdateThresholdsRequest{
		WarningPercent:  &warning,
		CriticalPercent: &critical,
	},
)
if err != nil {
	return err
}
_ = thresholds

```

Only populate pointer fields you intend to change. Before a broad update, call
`PreviewFinancialChanges` with a `PreviewFinancialChangesRequest`; it lets you
review the prospective result before persisting it. Avoid issuing automatic
budget changes based solely on a transient usage sample.

### Monitoring and evidence

Use `GetSpendSummary` for a workspace-level summary and `ListUsage` for
cursor-paginated records. `ListRequestLogs`, `GetRequestLog`, `ListTraces`, and
`GetTrace` connect spend to individual requests. Payload access and payload-log
methods may expose sensitive request content: request them only for an approved
incident and follow your data-retention policy.

See [Pagination](05-management.md#pagination) for list traversal and [Errors](09-errors-troubleshooting.md#errors) for
handling quota, authorization, and rate-limit failures.

## Billing and Subscriptions

Billing combines read-only catalog and status operations with payment mutations that can create external financial effects. Use [Financial governance](07-management-operations.md#financial-governance) for budgets and spend limits; use this page for plans, subscriptions, top-ups, payment methods, orders, invoices, and billing details.

Field-level signatures are published in the [management package documentation](https://pkg.go.dev/github.com/Syrovex/owlvigil_sdk_go/management). Do not copy fields from an older SDK version: the current subscription contract uses names such as `SubscriptionStatus`, `SubscriptionTier`, and `SubscriptionCancelAtPeriodEnd`; legacy aliases remain only for source compatibility.

### Read-only starting points

- Plans: `ListPlans`, `GetPlan`.
- Current subscription: `GetSubscription`.
- Billing: `GetBillingOverview`, `GetBillingOverviewForWorkspace`, `GetBillingDetails`.
- Top-ups and orders: `ListTopupPlans`, `ListTopups`, `GetTopup`, `ListOrders`, `GetOrder`.
- Payment methods: `ListPaymentMethods`, `ListPaymentMethodsForWorkspace`.
- Invoices: `ListInvoices`, `ListInvoicesForWorkspace`, `GetInvoice`, `GetInvoiceForWorkspace`.

Use the workspace-specific variant when the method signature requires a workspace ID. Treat IDs and status values as opaque service data rather than deriving them from display names.

### Mutating workflows

Subscription mutations include checkout and in-app creation, confirmation, upgrade, downgrade, cancellation, reactivation, and checkout synchronization. Top-up mutations include checkout and in-app creation and confirmation. Payment-method mutations include setup-intent creation, save, set-default, and delete.

For every payment mutation:

1. Prevent duplicate submission in the application and send the request once.
2. Retain its request ID.
3. If the response is ambiguous, do not automatically retry: these payment routes do not document idempotency-key support.
4. Read the order, checkout session, top-up, or subscription to determine final state.
5. Treat redirects and client-side payment confirmation as intermediate signals, not authoritative completion.

Do not calculate invoice totals or final payment state from local floating-point arithmetic. Use the values returned by the service and payment provider.

### Sensitive values

Billing details, payment confirmation secrets, invoice URLs, and payment-provider identifiers may be sensitive. Pass them only to the trusted component that needs them. Never place them in ordinary logs, analytics, issues, or support tickets.

### Verified examples

- [`examples/billing-subscription`](../../examples/billing-subscription/) reads plans and the active subscription.
- [`examples/financial-control`](../../examples/financial-control/) reads financial governance without changing it.
- The isolated [`examples/openapi-smoke`](../../examples/openapi-smoke/) workflow exercises mutating contracts only when writes are explicitly enabled.

All example programs are compiled by `go test ./examples/...`. Request paths and JSON bodies are covered by `management/refactored_openapi_contract_test.go` and `management/all_operations_usecase_test.go`.

### Complete operational method map

- Provider maintenance: `CreateProvider`, `GetProvider`, `UpdateProvider`, `VerifyProviderConnection`, `DeleteProvider`, `DeleteProviderWithResult`.
- Policy keywords: `AddPromptKeyword`, `DeletePromptKeyword`.
- Quota, balance, and operational evidence: `GetQuota`, `GetBalance`, `ListAuditLogs`, `GetAuditLog`, `GetLoggingSettings`, `UpdateLoggingSettings`, `ListPayloadLogs`, `GetPayloadAccess`, `GetPayloadLog`.
- Subscriptions: `CreateSubscriptionCheckout`, `CreateSubscriptionInApp`, `ConfirmSubscriptionInApp`, `UpgradeSubscription`, `DowngradeSubscription`, `CancelSubscription`, `CancelSubscriptionWithRequest`, `ReactivateSubscription`, `GetSubscriptionCheckoutSession`, `SyncLatestSubscriptionCheckout`.
- Top-ups: `CreateTopupCheckout`, `CreateTopupInApp`, `ConfirmTopupInApp`.
- Payment methods: `CreatePaymentMethodSetupIntent`, `CreatePaymentMethodSetupIntentForWorkspace`, `SavePaymentMethod`, `SetDefaultPaymentMethod`, `DeletePaymentMethod`, `DeletePaymentMethodWithResult`.
- Billing and order writes: `UpdateBillingDetails`, `ConfirmStripeSession`.

Payment and configuration writes can have external effects. Do not automatically retry them after an uncertain timeout; read the affected resource or use the Request ID to establish the outcome first.
