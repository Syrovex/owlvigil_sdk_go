# Management and Workspaces

## Management API

The OpenAPI Management client calls `https://api.owlvigil.com/v1` and requires a
scoped service-account API key. It is separate from the Gateway key used to
invoke models.

<!-- evidence: management/all_operations_usecase_test.go, management/refactored_openapi_contract_test.go -->
```go
client := management.NewClient(
	owlvigil.WithAPIKey(os.Getenv("OWLVIGIL_API_KEY")),
)

```

Most Management methods return a resource and `*owlvigil.ResponseMeta`. Record
`meta.RequestID` when investigating an API error. List methods use
[`ListOptions`](05-management.md#pagination). Only Gateway-key creation and Webhook-endpoint
creation currently accept an idempotency key. Do not attach one to other writes
or assume they are safe to retry.

### Guides

| Need | Guide | Main operations |
| --- | --- | --- |
| Select a workspace and manage people or roles | [Access control](06-management-access-and-keys.md#access-control) | Workspaces, teams, members, invitations, roles, permissions |
| Choose model delivery and credentials | [Model routing](07-management-operations.md#model-routing-and-provider-management) | Models, routes, providers, Gateway keys, policies |
| Bound and review spend | [Financial governance](07-management-operations.md#financial-governance) | Budgets, limits, thresholds, spend summary, usage |
| Change personal account settings | [Account](05-management.md#account-settings-and-support) | Profile, password, notifications, support, invite links |
| Buy, subscribe, and reconcile payments | [Billing](07-management-operations.md#billing-and-subscriptions) | Plans, subscriptions, top-ups, payment methods, invoices, orders |
| Deliver and diagnose events | [Webhooks](08-webhooks.md) | Endpoints, event attempts, redelivery, signature verification |
| Inspect request execution | [Gateway](04-gateway.md#gateway) | Usage, request logs, traces, payload access |

### API domain coverage

Every Management source domain has an owner guide. Use the source file and
GoDoc as the field-level reference; these pages explain how to safely combine
the operations.

| Source domain | Covered by | Notes |
| --- | --- | --- |
| `management/workspaces.go` | [Access control](06-management-access-and-keys.md#access-control) | Workspace selection and activity |
| `management/teams.go`, `management/members.go`, `management/invitations.go` | [Access control](06-management-access-and-keys.md#access-control) | Team and member lifecycle |
| `management/rbac.go` | [Access control](06-management-access-and-keys.md#access-control) | Roles and per-member permission overrides |
| `management/models.go`, `management/providers.go`, `management/policies.go` | [Model routing](07-management-operations.md#model-routing-and-provider-management) | Discovery, routing, provider configuration, policy preview |
| `management/gateway_keys.go` | [Model routing](07-management-operations.md#model-routing-and-provider-management) | One-time secrets, rotation, state changes |
| `management/financial.go` | [Financial governance](07-management-operations.md#financial-governance) | Governance and scoped budgets |
| `management/usage.go`, `management/logs.go` | [Financial governance](07-management-operations.md#financial-governance) | Spend, quota, request and trace evidence |
| `management/billing.go`, `management/subscription.go`, `management/topup.go` | [Billing](07-management-operations.md#billing-and-subscriptions) | Billing profile, subscriptions, and top-ups |
| `management/payment_methods.go`, `management/orders.go` | [Billing](07-management-operations.md#billing-and-subscriptions) | Payment method setup and order confirmation |
| `management/webhooks.go` | [Webhooks](08-webhooks.md) | Outbound endpoints and delivery events |
| `management/user.go` | [Account](05-management.md#account-settings-and-support) | Profile, support, notifications, and invitations |

## Pagination

Management list methods accept `management.ListOptions` and return `management.ListResponse[T]`.

<!-- evidence: management/all_operations_usecase_test.go, management/refactored_openapi_contract_test.go -->
```go
var cursor string
for {
	page, _, err := client.ListGatewayKeys(ctx, management.ListOptions{
		Cursor: cursor,
		Limit:  100,
	}, "", owlvigil.WithWorkspaceID(workspaceID))
	if err != nil {
		return err
	}

	for _, key := range page.Items {
		fmt.Println(key.ID, key.Name)
	}

	if !page.PageInfo.HasMore || page.PageInfo.NextCursor == "" {
		break
	}
	cursor = page.PageInfo.NextCursor
}

```

Endpoint-specific typed options expose the current Open API filters:
`ListRoutesWithFilters`, `ListMembersWithFilters`,
`ListWorkspaceActivityWithFilters`, `ListOrdersWithFilters`, and
`ListTopupsWithFilters`. Do not send undeclared filters: the Open API rejects
unknown query parameters.

Set `Limit` only when you want to override the service default; the SDK omits
zero or negative limits. Treat `NextCursor` as opaque: persist and send it back
unchanged, and never attempt to derive offsets from it. A list response may be
decoded from either a bare array or an object containing `items` and
`page_info`, so always use `Items` and `PageInfo` rather than decoding response
JSON yourself.

For long exports, retain the cursor only after successfully handling a page.
If the process is restarted, resume from that stored cursor and deduplicate by
resource ID where the downstream destination requires exactly-once processing.

## Account Settings and Support

These Management methods operate on the authenticated user rather than a
Gateway key. Use `OWLVIGIL_API_KEY` and do not expose profile or invitation
data to another user without authorization.

### Profile and preferences

`GetUserProfile` reads the current profile; `UpdateUserProfile` updates only
the selected fields of `UpdateUserProfileRequest`. The current contract uses
`Username`; `Name` remains a compatibility alias. Set `ClearAvatarURL` or
`ClearBalanceNotifyThreshold` to send the explicit JSON `null` required to
clear those nullable values.

<!-- evidence: management/all_operations_usecase_test.go, management/refactored_openapi_contract_test.go -->
```go
name := "Ada Lovelace"
profile, _, err := client.UpdateUserProfile(ctx,
	&management.UpdateUserProfileRequest{Username: &name},
)
if err != nil {
	return err
}
_ = profile

```

`GetNotificationPreferences` and `UpdateNotificationPreferences` manage budget,
billing, report, and marketing notification choices. The update is a PUT
replacement: read the current values first and submit all four choices. Any
unset request field becomes `false`.

### Password and support

`UpdatePassword` changes credentials using `UpdatePasswordRequest`. Handle the
password only in a protected form submission; do not log the request or return
it in an API response. `CreateSupportRequest` sends a typed support request
with a subject, issue type, and description. Include the SDK response request ID and a
sanitized reproduction in support tickets when possible.

Use `UpdatePasswordWithResult`, `CreateSupportRequestWithResult`, or
`SendInvitationWithResult` when the Open API action message or invitation
delivery count is needed. The original methods remain available for callers
that only need response metadata and an error.

### Invitations

`GetInviteLink` and `GetInvitationStats` retrieve the authenticated user's
invitation information. `ListUserInvitations` returns the complete published
list and accepts no pagination query; `SendInvitation` creates one. Invite
links can grant access, so do not place
them in public issues, analytics events, or client-side logs.

For workspace-admin invitations, member assignment, and role management, see
[Access control](06-management-access-and-keys.md#access-control).

## Complete workspace lifecycle

Workspace operations are `ListWorkspaces`, `CreateWorkspace`, `GetWorkspace`, `GetWorkspaceOverview`, `UpdateWorkspace`, `DeleteWorkspace`, `ListWorkspaceActivity`, and `ListWorkspaceActivityWithFilters`.
