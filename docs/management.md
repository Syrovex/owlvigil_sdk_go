# Management API

The Management client calls `https://api.owlvigil.com/open/v1` and requires a
scoped service-account API key. It is separate from the Gateway key used to
invoke models.

```go
client := management.NewClient(
    owlvigil.WithAPIKey(os.Getenv("OWLVIGIL_API_KEY")),
)
```

Most Management methods return a resource and `*owlvigil.ResponseMeta`. Record
`meta.RequestID` when investigating an API error. List methods use
[`ListOptions`](pagination.md); mutations should receive an idempotency key
when the operation could create a second external effect.

## Guides

| Need | Guide | Main operations |
| --- | --- | --- |
| Select a workspace and manage people or roles | [Access control](access-control.md) | Workspaces, teams, members, invitations, roles, permissions |
| Choose model delivery and credentials | [Model routing](model-routing.md) | Models, routes, providers, Gateway keys, policies |
| Bound and review spend | [Financial governance](financial-governance.md) | Budgets, limits, thresholds, spend summary, usage |
| Change personal account settings | [Account](account.md) | Profile, password, notifications, support, invite links |
| Buy, subscribe, and reconcile payments | [Billing](billing.md) | Plans, subscriptions, top-ups, payment methods, invoices, orders |
| Deliver and diagnose events | [Webhooks](webhooks.md) | Endpoints, event attempts, redelivery, signature verification |
| Inspect request execution | [Gateway](gateway.md) | Usage, request logs, traces, payload access |

## API domain coverage

Every Management source domain has an owner guide. Use the source file and
GoDoc as the field-level reference; these pages explain how to safely combine
the operations.

| Source domain | Covered by | Notes |
| --- | --- | --- |
| `management/workspaces.go` | [Access control](access-control.md) | Workspace selection and activity |
| `management/teams.go`, `management/members.go`, `management/invitations.go` | [Access control](access-control.md) | Team and member lifecycle |
| `management/rbac.go` | [Access control](access-control.md) | Roles and per-member permission overrides |
| `management/models.go`, `management/providers.go`, `management/policies.go` | [Model routing](model-routing.md) | Discovery, routing, provider configuration, policy preview |
| `management/gateway_keys.go` | [Model routing](model-routing.md) | One-time secrets, rotation, state changes |
| `management/financial.go` | [Financial governance](financial-governance.md) | Governance and scoped budgets |
| `management/usage.go`, `management/logs.go` | [Financial governance](financial-governance.md) | Spend, quota, request and trace evidence |
| `management/billing.go`, `management/subscription.go`, `management/topup.go` | [Billing](billing.md) | Billing profile, subscriptions, and top-ups |
| `management/payment_methods.go`, `management/orders.go` | [Billing](billing.md) | Payment method setup and order confirmation |
| `management/webhooks.go` | [Webhooks](webhooks.md) | Outbound endpoints and delivery events |
| `management/user.go` | [Account](account.md) | Profile, support, notifications, and invitations |
| `management/documentation.go` | This page | Navigation, endpoint catalog, OpenAPI/Swagger schema, SDK packages |

## Management metadata and schemas

`DocumentationNavigation`, `ListDocumentedEndpoints`, and
`GetDocumentedEndpoint` retrieve the server-provided endpoint catalog.
`OpenAPISchema` and `SwaggerSchema` return the published API schema as
`map[string]any`; persist it only when you have a schema-consumer that needs a
snapshot. `SDKPackages` lists published SDK package metadata.

These are read-only discovery operations. They do not replace the SDK's typed
methods, which provide request construction and error handling.
