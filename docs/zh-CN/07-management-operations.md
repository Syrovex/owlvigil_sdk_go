# Management：路由、可观测性与账单

## Management：模型路由与 Provider

模型路由决定 Gateway 请求可以选择哪些模型，以及请求怎样匹配上游 Provider。修改路由前先读取和预览，避免直接影响生产流量。

### 四个概念

- Model：Management 模型目录中的模型定义。
- Route：模型请求到可用上游的路由配置。
- Provider：上游模型服务及其连接配置。
- Gateway policy：请求进入 Gateway 时应用的策略。

这些是 Management 控制面资源；真正调用模型仍使用 `gateway.Client`。

### 推荐变更顺序

1. `ListModels` / `GetModel` 确认目标模型。
2. `ListProviders` / `GetProvider` 查看可用 Provider。
3. 创建或更新 Provider 后调用 `VerifyProviderConnection`。
4. `ListRoutesWithFilters` 查找相关路由。
5. 使用 `PreviewRoute` 验证预期匹配结果。
6. 修改 policy 前使用 `PreviewPolicyEffect` 查看影响。
7. 写入后重新读取，并从 Gateway 做一次受控验证。

完整方法分组如下：

| 资源 | 方法 |
| --- | --- |
| 模型 | `ListModels`, `GetModel` |
| 路由 | `ListRoutes`, `ListRoutesWithFilters`, `GetRoute`, `GetRouteWithFilters`, `PreviewRoute` |
| Provider | `ListProviders`, `CreateProvider`, `VerifyProviderConnection`, `GetProvider`, `UpdateProvider`, `DeleteProvider`, `DeleteProviderWithResult` |
| Gateway policy | `GetGatewayPolicies`, `PreviewPolicyEffect`, `AddPromptKeyword`, `DeletePromptKeyword`, `UpdateGatewayPolicy` |

### Provider secret

创建或更新 Provider 的请求可能包含第三方 credential。请求对象、HTTP body 和返回值都不能直接进入普通日志。Provider 删除前确认没有生产 route 依赖它。

`DeleteProviderWithResult` 返回服务端确认对象；只需要成功/失败时可用兼容删除方法。

### Gateway policy

Policy 方法包括读取、预览、更新以及 prompt keyword 的添加和删除。关键词可能包含业务敏感内容，日志和审计展示时需要脱敏。

Policy 变更可能影响整个工作区的请求。推荐先 preview，再在非生产环境或小流量 Key 上验证，最后逐步推广。

### 动态数据

模型 ID、route 状态、Provider 类型和 policy 字段以 API 返回为准。示例值不能作为固定枚举写入应用；需要展示选择项时从读取接口刷新。

## Management：用量、日志与审计

Management 提供三层观察数据：聚合用量回答“用了多少”，request log/trace 回答“一次调用发生了什么”，audit log 回答“谁修改了什么”。

### 用量、配额和余额

- `ListUsage`：分页读取用量记录。
- `GetUsageSummary`：读取聚合用量。
- `GetQuota` / `GetQuotaSummary` / `GetQuotaUsage`：读取配额状态。
- `GetBalance` / `GetBalanceForWorkspace`：读取余额。
- `ListInvoices` / `ListInvoicesForWorkspace`：读取发票列表。

金额和 token 汇总由服务端计算。不要通过抽样 request log 自行推导账单结果。

### Request log 与 Trace

`ListRequestLogs` / `GetRequestLog` 用于定位单次请求；`ListTraces` / `GetTrace` 用于查看关联执行。优先从应用保存的 Request ID 开始排查，而不是大范围导出 payload。

### Audit log

`ListAuditLogs` 支持类型化筛选，`GetAuditLog` 读取单条审计详情。Audit log 适合追踪配置变更，但应用仍应保留自己的业务授权和审批记录。

### Logging settings 与 Payload

`GetLoggingSettings` / `UpdateLoggingSettings` 控制日志采集设置。修改正文、header 或 stream chunk 采集前，先评估个人信息、商业数据和密钥泄露风险。

读取 payload 前调用 `GetPayloadAccess`。`ListPayloadLogs` 和 `GetPayloadLog` 返回的数据可能包含用户 prompt、响应或上游执行内容，应视为敏感生产数据。

### 安全排障流程

1. 从 Request ID、时间和操作名开始。
2. 先看汇总和状态，再按需读取 trace。
3. 只有必要且有权限时读取 payload。
4. 对复制到支持渠道的内容脱敏。
5. 排障完成后遵循保留策略清理临时导出。

分页导出规则见[分页](05-management.md#分页)，日志字段的完整定义见 [API 参考](10-reference-examples.md#api-参考)。

## Management：财务与账单

本章区分两类能力：财务治理限制“允许花多少”，账单与支付处理“怎样购买、充值和核对结果”。预览和读取不会自动生效，写操作也不能只凭前端跳转判断成功。

### 财务治理

| 目标 | 方法组 |
| --- | --- |
| 读取/更新治理配置 | `GetFinancialGovernance`, `UpdateFinancialGovernance` |
| 管理预算上限 | `GetBudgetCaps`, `UpdateBudgetCaps`, `UpdateScopeBudgetCap` |
| 管理支出限制 | `GetSpendingLimits`, `GetSpendingLimitsWithFilters`, `UpdateSpendingLimits`, `UpdateUserSpendingLimit` |
| 管理阈值 | `GetFinancialThresholds`, `UpdateFinancialThresholds` |
| 预览和汇总 | `PreviewFinancialChanges`, `GetSpendSummary` |

推荐先调用 `PreviewFinancialChanges`，把预览展示给有权限的操作者确认，再执行更新并重新读取最终状态。

### 账单资料和发票

`GetBillingOverview`、`GetBillingOverviewForWorkspace`、`GetBillingDetails` 和 `GetInvoice`、`GetInvoiceForWorkspace` 用于读取账单信息；`UpdateBillingDetails` 是写操作。发票列表使用 `ListInvoices` 或 `ListInvoicesForWorkspace`。

### 订阅

订阅流程包括 plans、当前 subscription、checkout、in-app 确认、升级、降级、取消、恢复和 checkout 同步。支付跳转完成不代表本地业务状态已经确认，应通过 subscription 或 checkout session 读取接口核对。

对应方法为 `ListPlans`、`GetPlan`、`GetSubscription`、`CreateSubscriptionCheckout`、`CreateSubscriptionInApp`、`ConfirmSubscriptionInApp`、`UpgradeSubscription`、`DowngradeSubscription`、`CancelSubscription`、`CancelSubscriptionWithRequest`、`ReactivateSubscription`、`GetSubscriptionCheckoutSession` 和 `SyncLatestSubscriptionCheckout`。

### 充值、支付方式和订单

- Topup：`ListTopupPlans`、`CreateTopupCheckout`、`CreateTopupInApp`、`ConfirmTopupInApp`、`ListTopups`、`ListTopupsWithFilters`、`GetTopup`。
- Payment method：`ListPaymentMethods`、`ListPaymentMethodsForWorkspace`、`CreatePaymentMethodSetupIntent`、`CreatePaymentMethodSetupIntentForWorkspace`、`SavePaymentMethod`、`SetDefaultPaymentMethod`、`DeletePaymentMethod`、`DeletePaymentMethodWithResult`。
- Order：`ListOrders`、`ListOrdersWithFilters`、`GetOrder`、`ConfirmStripeSession`。

创建 checkout、确认支付、变更订阅和删除支付方式都可能产生外部副作用，且当前不支持幂等键自动重试。应用必须防止用户双击；超时后先读取订单或订阅状态，不得直接重放写请求。

### 金额和状态

金额、币种、优惠、税费和允许的状态由服务端及支付提供商决定。不要在客户端用浮点数重新计算应付总额，也不要把 preview 或 pending 状态展示为已完成。

支付数据和支付确认密钥属于敏感信息。只传给需要它的受信任组件，不写入日志、指标、Issue 或支持工单。
