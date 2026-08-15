# API 参考与示例索引

## API 参考

本页是公开能力的导航，不重复维护所有字段签名。完整类型、参数和返回值以 [pkg.go.dev](https://pkg.go.dev/github.com/Syrovex/owlvigil_sdk_go) 和当前源码为准。

### `owlvigil` 共享包

客户端选项：`WithBaseURL`、`WithEnvironment`、`WithEnvironmentFromEnv`、`WithHTTPClient`、`WithTimeout`、`WithUserAgent`、`WithAPIKey`、`WithAPIKeyProvider`、`WithRetry`、`WithoutRetry`。

单次请求选项：`WithIdempotencyKey`（当前仅用于两种受支持的创建接口）、`WithHeader`、`WithQueryParam`、`WithWorkspaceID`。

主要类型：`Config`、`Environment`、`TokenProvider`、`ResponseMeta`、`Envelope[T]`、`APIError`。

### `gateway`

| 能力 | 方法 |
| --- | --- |
| 客户端 | `NewClient`, `BaseURL` |
| 模型 | `ListModels`, `GetModel` |
| Chat | `CreateChatCompletion`, `CreateChatCompletionStream` |
| Responses | `CreateResponse`, `CreateResponseStream` |
| Embeddings | `CreateEmbeddings` |
| Anthropic 兼容 | `CreateAnthropicMessage` |
| Stream | `Next`, `Current`, `Err`, `Close` |

### `webhook`

`VerifySignature` 用于生产验签；`SignPayload` 主要用于测试和本地夹具。错误值包括 `ErrMissingSignature`、`ErrInvalidSignature`、`ErrMissingSecret` 和 `ErrStaleTimestamp`。

### `management`

使用顺序和通用规则见 [Management 概览](05-management.md#management)。以下列表用于按名称查找方法。

#### 工作区和访问控制

- Workspaces：`ListWorkspaces`, `CreateWorkspace`, `GetWorkspace`, `GetWorkspaceOverview`, `UpdateWorkspace`, `DeleteWorkspace`, `ListWorkspaceActivity`, `ListWorkspaceActivityWithFilters`。
- Teams：`ListTeams`, `CreateTeam`, `GetTeam`, `UpdateTeam`, `DeleteTeam`, `DeleteTeamWithResult`。
- Members：`ListMembers`, `ListMembersWithFilters`, `ListRoleOptions`, `CreateMember`, `GetMember`, `UpdateMember`, `DeleteMember`, `DeleteMemberWithResult`。
- Invitations：`ListInvitations`, `CreateInvitation`, `ResendInvitation`, `ResendInvitationWithResult`, `RevokeInvitation`, `RevokeInvitationWithResult`。
- RBAC：`ListRoles`, `CreateRole`, `GetRole`, `UpdateRole`, `DeleteRole`, `DeleteRoleWithResult`, `ListPermissions`, `GetMemberPermissions`, `UpdateMemberPermissions`, `ResetMemberPermissions`。

#### Gateway Key 和模型路由

- Gateway Keys：`ListGatewayKeys`, `CreateGatewayKey`, `GetGatewayKey`, `UpdateGatewayKey`, `RotateGatewayKey`, `EnableGatewayKey`, `EnableGatewayKeyWithResult`, `DisableGatewayKey`, `DisableGatewayKeyWithResult`, `DeleteGatewayKey`, `DeleteGatewayKeyWithResult`。
- Models/Routes：`ListModels`, `GetModel`, `ListRoutes`, `ListRoutesWithFilters`, `GetRoute`, `GetRouteWithFilters`, `PreviewRoute`。
- Providers：`ListProviders`, `CreateProvider`, `VerifyProviderConnection`, `GetProvider`, `UpdateProvider`, `DeleteProvider`, `DeleteProviderWithResult`。
- Policies：`GetGatewayPolicies`, `PreviewPolicyEffect`, `AddPromptKeyword`, `DeletePromptKeyword`, `UpdateGatewayPolicy`。

#### 财务、使用量和日志

- Financial：`GetFinancialGovernance`, `UpdateFinancialGovernance`, `GetBudgetCaps`, `UpdateBudgetCaps`, `UpdateScopeBudgetCap`, `GetSpendingLimits`, `GetSpendingLimitsWithFilters`, `UpdateSpendingLimits`, `UpdateUserSpendingLimit`, `GetFinancialThresholds`, `UpdateFinancialThresholds`, `PreviewFinancialChanges`, `GetSpendSummary`。
- Usage：`ListUsage`, `GetUsageSummary`, `GetQuota`, `GetQuotaSummary`, `GetQuotaUsage`, `GetBalance`, `GetBalanceForWorkspace`, `ListInvoices`, `ListInvoicesForWorkspace`。
- Logs：`ListAuditLogs`, `GetAuditLog`, `GetLoggingSettings`, `UpdateLoggingSettings`, `ListPayloadLogs`, `ListRequestLogs`, `GetRequestLog`, `ListTraces`, `GetTrace`, `GetPayloadAccess`, `GetPayloadLog`。

#### 账单和支付

- Billing：`GetBillingOverview`, `GetBillingOverviewForWorkspace`, `GetBillingDetails`, `UpdateBillingDetails`, `GetInvoice`, `GetInvoiceForWorkspace`。
- Subscription：`ListPlans`, `GetPlan`, `GetSubscription`, `CreateSubscriptionCheckout`, `CreateSubscriptionInApp`, `ConfirmSubscriptionInApp`, `UpgradeSubscription`, `DowngradeSubscription`, `CancelSubscription`, `CancelSubscriptionWithRequest`, `ReactivateSubscription`, `GetSubscriptionCheckoutSession`, `SyncLatestSubscriptionCheckout`。
- Topup：`ListTopupPlans`, `CreateTopupCheckout`, `CreateTopupInApp`, `ConfirmTopupInApp`, `ListTopups`, `ListTopupsWithFilters`, `GetTopup`。
- Payment methods：`ListPaymentMethods`, `ListPaymentMethodsForWorkspace`, `CreatePaymentMethodSetupIntent`, `CreatePaymentMethodSetupIntentForWorkspace`, `SavePaymentMethod`, `SetDefaultPaymentMethod`, `DeletePaymentMethod`, `DeletePaymentMethodWithResult`。
- Orders：`ListOrders`, `ListOrdersWithFilters`, `GetOrder`, `ConfirmStripeSession`。

#### Webhook 和账户

- Webhooks：`ListWebhookEndpoints`, `CreateWebhookEndpoint`, `GetWebhookEndpoint`, `UpdateWebhookEndpoint`, `DeleteWebhookEndpoint`, `DeleteWebhookEndpointWithResult`, `EnableWebhookEndpoint`, `EnableWebhookEndpointWithResult`, `DisableWebhookEndpoint`, `DisableWebhookEndpointWithResult`, `RotateWebhookSecret`, `TestWebhookEndpoint`, `TestWebhookEndpointWithResult`, `ListWebhookEventTypes`, `ListWebhookEvents`, `GetWebhookEvent`, `ListEndpointEvents`, `RetryWebhookEvent`, `RetryWebhookEventWithResult`, `RedeliverWebhookEvent`, `RedeliverWebhookEventWithResult`, `BulkRedeliverWebhookEvents`, `BulkRedeliverWebhookEventsWithResult`。
- User：`GetUserProfile`, `UpdateUserProfile`, `UpdatePassword`, `UpdatePasswordWithResult`, `CreateSupportRequest`, `CreateSupportRequestWithResult`, `GetNotificationPreferences`, `UpdateNotificationPreferences`, `GetInviteLink`, `GetInvitationStats`, `ListUserInvitations`, `SendInvitation`, `SendInvitationWithResult`。

### 如何阅读方法签名

大多数 Gateway/Management 方法返回 `(resource, *owlvigil.ResponseMeta, error)`。必须先检查 error；失败时 resource 通常为 nil，但 metadata 可能仍包含诊断信息。列表方法通常返回 `*management.ListResponse[T]`。方法末尾的 `...owlvigil.RequestOption` 可传工作区 query 等单次请求配置；只有 Gateway Key 创建和 Webhook endpoint 创建支持幂等键。

## 运行示例

仓库的 `examples/` 包含可编译程序。先复制安全配置模板：

<!-- evidence: go.mod, examples/compile_test.go -->
```bash
cp .env.example .env
```

只填写当前示例需要的变量。已导出的 shell 环境变量优先于 `.env`。

| 示例 | 用途 | 是否写入服务端 |
| --- | --- | --- |
| `quickstart` | 完整的首次模型列表调用 | 否 |
| `gateway-models` | 列出模型 | 否 |
| `gateway-chat` | Chat 调用 | 会产生模型用量 |
| `gateway-stream` | 流式 Chat | 会产生模型用量 |
| `management-usage` | 查询 usage | 否 |
| `management-key` | 创建 Gateway Key | 是 |
| `webhook-verify` | 验证签名 | 否 |
| `financial-control` | 读取财务治理 | 否 |
| `billing-subscription` | 读取计划和订阅 | 否 |
| `team-management` | 团队与成员工作流 | 包含写操作 |
| `multi-environment` | 切换环境 | 取决于子操作 |
| `openapi-smoke` | 覆盖 Management 操作 | 开启写模式时大量写入 |

运行只读示例：

<!-- evidence: go.mod, examples/compile_test.go -->
```bash
go run ./examples/gateway-models/main.go
go run ./examples/management-usage/main.go
```

先验证所有示例可以编译：

<!-- evidence: go.mod, examples/compile_test.go -->
```bash
go test ./examples/...
```

`openapi-smoke` 的写模式会创建、更新和删除资源，只能在隔离工作区运行。不要在共享生产工作区设置 `OWLVIGIL_SMOKE_WRITES=1`。详细前置条件见 `examples/openapi-smoke/README.md`。
