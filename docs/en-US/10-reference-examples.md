# API Reference and Example Index

## API reference

### Shared `owlvigil` package

Client options include `WithBaseURL`, `WithEnvironment`, `WithEnvironmentFromEnv`, `WithHTTPClient`, `WithTimeout`, `WithUserAgent`, `WithAPIKey`, `WithAPIKeyProvider`, `WithRetry`, and `WithoutRetry`.

Request options include `WithIdempotencyKey`, `WithWorkspaceID`, `WithAccessToken`, `WithAccessTokenProvider`, and `WithQueryParam`. Use only the options documented by the target operation.

Inspect `*owlvigil.APIError` for structured API failures and `*owlvigil.ResponseMeta` for successful response metadata.

### `gateway`

| Capability | Methods |
| --- | --- |
| Models | `ListModels`, `GetModel` |
| Chat Completions | `CreateChatCompletion`, `CreateChatCompletionStream` |
| Responses | `CreateResponse`, `CreateResponseStream` |
| Embeddings | `CreateEmbeddings` |
| Anthropic-compatible Messages | `CreateAnthropicMessage` |
| Stream lifecycle | `Next`, `Current`, `Err`, `Close` |

### `webhook`

Use `VerifySignature` in an event receiver. `SignPayload` is intended primarily for tests and local fixtures. Verification errors include `ErrMissingSignature`, `ErrInvalidSignature`, `ErrMissingSecret`, and `ErrStaleTimestamp`.

### `management`

#### Workspaces and access control

- Workspaces: `ListWorkspaces`, `CreateWorkspace`, `GetWorkspace`, `GetWorkspaceOverview`, `UpdateWorkspace`, `DeleteWorkspace`, `ListWorkspaceActivity`, `ListWorkspaceActivityWithFilters`.
- Teams: `ListTeams`, `CreateTeam`, `GetTeam`, `UpdateTeam`, `DeleteTeam`, `DeleteTeamWithResult`.
- Members: `ListMembers`, `ListMembersWithFilters`, `ListRoleOptions`, `CreateMember`, `GetMember`, `UpdateMember`, `DeleteMember`, `DeleteMemberWithResult`.
- Invitations: `ListInvitations`, `CreateInvitation`, `ResendInvitation`, `ResendInvitationWithResult`, `RevokeInvitation`, `RevokeInvitationWithResult`.
- RBAC: `ListRoles`, `CreateRole`, `GetRole`, `UpdateRole`, `DeleteRole`, `DeleteRoleWithResult`, `ListPermissions`, `GetMemberPermissions`, `UpdateMemberPermissions`, `ResetMemberPermissions`.

#### Gateway Keys and model routing

- Gateway Keys: `ListGatewayKeys`, `CreateGatewayKey`, `GetGatewayKey`, `UpdateGatewayKey`, `RotateGatewayKey`, `EnableGatewayKey`, `EnableGatewayKeyWithResult`, `DisableGatewayKey`, `DisableGatewayKeyWithResult`, `DeleteGatewayKey`, `DeleteGatewayKeyWithResult`.
- Models and routes: `ListModels`, `GetModel`, `ListRoutes`, `ListRoutesWithFilters`, `GetRoute`, `GetRouteWithFilters`, `PreviewRoute`.
- Providers: `ListProviders`, `CreateProvider`, `VerifyProviderConnection`, `GetProvider`, `UpdateProvider`, `DeleteProvider`, `DeleteProviderWithResult`.
- Policies: `GetGatewayPolicies`, `PreviewPolicyEffect`, `AddPromptKeyword`, `DeletePromptKeyword`, `UpdateGatewayPolicy`.

#### Financial controls, usage, and logs

- Financial controls: `GetFinancialGovernance`, `UpdateFinancialGovernance`, `GetBudgetCaps`, `UpdateBudgetCaps`, `UpdateScopeBudgetCap`, `GetSpendingLimits`, `GetSpendingLimitsWithFilters`, `UpdateSpendingLimits`, `UpdateUserSpendingLimit`, `GetFinancialThresholds`, `UpdateFinancialThresholds`, `PreviewFinancialChanges`, `GetSpendSummary`.
- Usage: `ListUsage`, `GetUsageSummary`, `GetQuota`, `GetQuotaSummary`, `GetQuotaUsage`, `GetBalance`, `GetBalanceForWorkspace`, `ListInvoices`, `ListInvoicesForWorkspace`.
- Logs: `ListAuditLogs`, `GetAuditLog`, `GetLoggingSettings`, `UpdateLoggingSettings`, `ListPayloadLogs`, `ListRequestLogs`, `GetRequestLog`, `ListTraces`, `GetTrace`, `GetPayloadAccess`, `GetPayloadLog`.

#### Billing and payments

- Billing: `GetBillingOverview`, `GetBillingOverviewForWorkspace`, `GetBillingDetails`, `UpdateBillingDetails`, `GetInvoice`, `GetInvoiceForWorkspace`.
- Subscriptions: `ListPlans`, `GetPlan`, `GetSubscription`, `CreateSubscriptionCheckout`, `CreateSubscriptionInApp`, `ConfirmSubscriptionInApp`, `UpgradeSubscription`, `DowngradeSubscription`, `CancelSubscription`, `CancelSubscriptionWithRequest`, `ReactivateSubscription`, `GetSubscriptionCheckoutSession`, `SyncLatestSubscriptionCheckout`.
- Top-ups: `ListTopupPlans`, `CreateTopupCheckout`, `CreateTopupInApp`, `ConfirmTopupInApp`, `ListTopups`, `ListTopupsWithFilters`, `GetTopup`.
- Payment methods: `ListPaymentMethods`, `ListPaymentMethodsForWorkspace`, `CreatePaymentMethodSetupIntent`, `CreatePaymentMethodSetupIntentForWorkspace`, `SavePaymentMethod`, `SetDefaultPaymentMethod`, `DeletePaymentMethod`, `DeletePaymentMethodWithResult`.
- Orders: `ListOrders`, `ListOrdersWithFilters`, `GetOrder`, `ConfirmStripeSession`.

#### Webhooks and account operations

- Webhooks: `ListWebhookEndpoints`, `CreateWebhookEndpoint`, `GetWebhookEndpoint`, `UpdateWebhookEndpoint`, `DeleteWebhookEndpoint`, `DeleteWebhookEndpointWithResult`, `EnableWebhookEndpoint`, `EnableWebhookEndpointWithResult`, `DisableWebhookEndpoint`, `DisableWebhookEndpointWithResult`, `RotateWebhookSecret`, `TestWebhookEndpoint`, `TestWebhookEndpointWithResult`, `ListWebhookEventTypes`, `ListWebhookEvents`, `GetWebhookEvent`, `ListEndpointEvents`, `RetryWebhookEvent`, `RetryWebhookEventWithResult`, `RedeliverWebhookEvent`, `RedeliverWebhookEventWithResult`, `BulkRedeliverWebhookEvents`, `BulkRedeliverWebhookEventsWithResult`.
- Account: `GetUserProfile`, `UpdateUserProfile`, `UpdatePassword`, `UpdatePasswordWithResult`, `CreateSupportRequest`, `CreateSupportRequestWithResult`, `GetNotificationPreferences`, `UpdateNotificationPreferences`, `GetInviteLink`, `GetInvitationStats`, `ListUserInvitations`, `SendInvitation`, `SendInvitationWithResult`.

Most Gateway and Management methods return a resource, `*owlvigil.ResponseMeta`, and an error. Check the error before using the resource. List methods usually return `*management.ListResponse[T]`.

## Runnable examples

Runnable examples live under `examples/`.

| Directory | Workflow |
| --- | --- |
| `examples/quickstart` | Complete first Gateway model-list call |
| `examples/gateway-chat` | Create a Gateway chat completion |
| `examples/gateway-stream` | Read a Gateway streaming chat response |
| `examples/gateway-models` | List Gateway models |
| `examples/management-key` | Create a Gateway key through Management API |
| `examples/management-usage` | Read usage summary |
| `examples/webhook-verify` | Verify a webhook signature |
| `examples/financial-control` | Read financial governance for a workspace |
| `examples/billing-subscription` | Inspect plans and the active subscription |
| `examples/team-management` | List teams and workspace members |
| `examples/openapi-smoke` | Exercise Management operations in an isolated workspace |
| `examples/multi-environment` | Select production, staging, or local endpoints |

Compile examples:

<!-- evidence: go.mod, examples/compile_test.go -->
```bash
go test ./examples/...
```

Examples load the nearest `.env` file without overwriting an exported shell
variable. `OWLVIGIL_GATEWAY_KEY` is for Gateway examples;
`OWLVIGIL_API_KEY` is for Management examples. The `openapi-smoke` example performs
writes and cleanup when `OWLVIGIL_SMOKE_WRITES=1`; set
`OWLVIGIL_SMOKE_REQUIRE_ALL=1` to require a `PASS` for every unique Management
operation.
Password, invitation, secondary-member, policy, Provider, and Stripe
prerequisites are documented in `examples/openapi-smoke/README.md`. Run it only in an isolated
workspace, never against a shared production workspace. Without that flag it
checks read-only paths and records mutating operations as explicit skips; a
skip is not reported as a pass. The catalog test locks the example to all 141
published Management operations, while unit contract tests execute every SDK
method against strict local HTTP fixtures.
