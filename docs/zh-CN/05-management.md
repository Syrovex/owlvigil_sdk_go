# Management 与工作区

## Management 概览

`management.Client` 用于配置和观察 OwlVigil 工作区。它属于控制面，不负责模型推理；调用模型请使用 `gateway.Client`。

### 创建客户端

<!-- evidence: management/all_operations_usecase_test.go, management/refactored_openapi_contract_test.go -->
```go
apiKey := os.Getenv("OWLVIGIL_API_KEY")
if apiKey == "" {
	return errors.New("OWLVIGIL_API_KEY is required")
}

client := management.NewClient(
	owlvigil.WithAPIKey(apiKey),
	owlvigil.WithTimeout(30*time.Second),
)

```

客户端会使用 SDK 定义的默认地址。除测试、私有部署或明确的环境切换外，不需要手工拼接 endpoint。

### 按任务选择文档

| 任务 | 文档 | 主要资源 |
| --- | --- | --- |
| 建立资源边界 | [工作区](05-management.md#工作区) | workspace、overview、activity |
| 管理人员与权限 | [成员与访问控制](06-management-access-and-keys.md#成员与访问控制) | team、member、invitation、role、permission |
| 发放模型调用凭证 | [Gateway Key 管理](06-management-access-and-keys.md#gateway-key-管理) | Gateway Key 生命周期 |
| 选择模型和上游 | [模型路由与 Provider](07-management-operations.md#management模型路由与-provider) | model、route、provider、policy |
| 观察运行情况 | [用量、日志与审计](07-management-operations.md#management用量日志与审计) | usage、quota、request log、trace、audit log |
| 控制和核对支出 | [财务与账单](07-management-operations.md#management财务与账单) | budget、subscription、topup、payment、invoice、order |
| 配置事件投递 | [Webhook](08-webhooks.md) | endpoint、event、redelivery |

### 通用调用规则

#### 明确工作区

很多 Management 资源属于某个工作区。部分方法把工作区 ID 放在 URL 路径或请求体中，另一些方法需要传入以下单次请求选项：

<!-- evidence: management/all_operations_usecase_test.go, management/refactored_openapi_contract_test.go -->
```go
owlvigil.WithWorkspaceID(workspaceID)

```

严格按照当前方法签名和 GoDoc 传递，不要同时在多个位置重复添加。

#### 处理返回值

<!-- evidence: management/all_operations_usecase_test.go, management/refactored_openapi_contract_test.go -->
```go
workspaces, meta, err := client.ListWorkspaces(ctx, management.ListOptions{})
if err != nil {
	return err
}
slog.Info("listed OwlVigil workspaces", "request_id", meta.RequestID)
_ = workspaces

```

先检查 error，再使用资源结果。关键写操作保存 Request ID，用于排查不确定结果。

#### 安全执行写操作

创建、轮换、结账、删除和重投可能产生外部副作用。当前只有 Gateway Key 创建和 Webhook endpoint 创建接受幂等键；这两种操作重试时必须复用原键和相同请求体。其他写操作不得自动重试，应通过读接口或 Request ID 核对结果。

<!-- evidence: management/all_operations_usecase_test.go, management/refactored_openapi_contract_test.go -->
```go
owlvigil.WithIdempotencyKey("gateway-key-create-primary-v1")

```

#### 选择兼容方法

部分删除、启停和重投方法同时提供普通版本与 `WithResult` 版本。只关心成功与否时使用普通版本；需要读取服务端确认对象时使用 `WithResult`。

#### 遍历列表

列表方法通常接受 `ListOptions`，返回 `ListResponse[T]`。完整循环见[分页](05-management.md#分页)。全部公开方法见 [API 参考](10-reference-examples.md#api-参考)。

## Management：工作区

工作区是大多数 Management 资源的归属边界。接入 Management API 时，首先要确定当前业务租户对应的工作区 ID，而不是直接创建 Key。

### 列出并选择工作区

<!-- evidence: management/all_operations_usecase_test.go, management/refactored_openapi_contract_test.go -->
```go
page, meta, err := client.ListWorkspaces(ctx, management.ListOptions{Limit: 50})
if err != nil {
	return err
}
for _, workspace := range page.Items {
	fmt.Println(workspace.ID, workspace.Name)
}
slog.Info("listed OwlVigil workspaces", "request_id", meta.RequestID)

```

不要默认选择列表第一项。生产应用应使用配置、用户选择或经过验证的租户映射。

### 生命周期

- `ListWorkspaces`：分页列出当前凭证可访问的工作区。
- `CreateWorkspace`：创建工作区。
- `GetWorkspace`：读取单个工作区。
- `GetWorkspaceOverview`：读取概览数据。
- `UpdateWorkspace`：更新可变字段。
- `DeleteWorkspace`：删除工作区。

创建和删除属于高影响写操作，当前都不支持幂等键自动重试。发生不确定结果时先通过列表或详情核对。删除前确认目标 ID、环境和依赖资源，并按服务端约束处理非空工作区。

### 活动记录

`ListWorkspaceActivity` 提供通用列表；`ListWorkspaceActivityWithFilters` 使用类型化筛选条件。Cursor 分页规则见[分页](05-management.md#分页)。

活动记录可用于审计，但不能替代应用自己的业务事务日志。输出活动详情前检查其中是否包含用户或资源敏感信息。

### 下一步

确定工作区 ID 后，可以继续配置[成员与访问控制](06-management-access-and-keys.md#成员与访问控制)、[Gateway Key](06-management-access-and-keys.md#gateway-key-管理)或[模型路由](07-management-operations.md#management模型路由与-provider)。

## Management 分页

常见列表方法接受 `management.ListOptions`，并返回 `management.ListResponse[T]`。

<!-- evidence: management/all_operations_usecase_test.go, management/refactored_openapi_contract_test.go -->
```go
var cursor string
for {
	page, meta, err := client.ListGatewayKeys(
		ctx,
		management.ListOptions{Cursor: cursor, Limit: 100},
		"",
		owlvigil.WithWorkspaceID(workspaceID),
	)
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
	_ = meta
}

```

### 规则

- `Limit <= 0` 时 SDK 不发送 limit，由服务端决定默认页大小。
- Cursor 是不透明字符串，只能原样传回，不能解析为 offset。
- 同时检查 `HasMore` 和非空 `NextCursor`，避免异常响应造成死循环。
- 成功处理完一页后再持久化 cursor；需要 exactly-once 效果时，还要按资源 ID 去重。
- `ListResponse` 可兼容裸数组和包含 `items`、`page_info` 的对象响应；应用始终读取 `Items` 和 `PageInfo`。

部分资源提供专用筛选类型和 `WithFilters` 方法，例如 routes、members、workspace activity、orders、topups 和 spending limits。优先使用这些类型，不要通过 `WithQueryParam` 猜测未发布的查询参数。

## 账户、通知与支持

这些方法操作当前 Management API 用户的资料或面向当前用户的功能：

| 任务 | 方法 |
| --- | --- |
| 读取和更新个人资料 | `GetUserProfile`, `UpdateUserProfile` |
| 修改密码 | `UpdatePassword`, `UpdatePasswordWithResult` |
| 提交支持请求 | `CreateSupportRequest`, `CreateSupportRequestWithResult` |
| 读取和更新通知偏好 | `GetNotificationPreferences`, `UpdateNotificationPreferences` |
| 查看邀请入口和统计 | `GetInviteLink`, `GetInvitationStats`, `ListUserInvitations` |
| 发送邀请 | `SendInvitation`, `SendInvitationWithResult` |

密码、个人资料和通知偏好属于敏感账户操作。修改后需要服务端确认对象时使用 `WithResult` 版本；否则使用普通版本并检查返回的 `error` 和 Request ID。发送邀请可能向外部联系人发送邮件，不能自动重试。
