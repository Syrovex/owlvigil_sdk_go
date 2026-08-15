# 核心概念

在写第一段业务代码前，先理解这些概念。它们解释了 SDK 为什么分成多个客户端，以及一次调用从配置到结果是怎样完成的。

## SDK 与 API 的关系

SDK 是 OwlVigil HTTP API 的 Go 客户端。它负责构造请求、发送认证信息、处理有限重试并解析响应。权限校验、模型路由选择和账单结果由 OwlVigil 服务端根据工作区配置执行，并通过 API 返回。

因此：

- Go 类型和方法签名定义客户端如何调用 API。
- 服务端返回当前凭证是否有权执行操作。
- 模型路由、账单结果、资源状态和事件类型以 API 的实际返回为准，客户端不应自行推断或从示例硬编码。

## 两个主要客户端

<!-- evidence: gateway/client.go, management/client.go -->
```text
你的 Go 应用
├── gateway.Client ─── 模型推理：Chat、Responses、Embeddings、Models
└── management.Client ─ 工作区管理：成员、密钥、路由、账单、日志、Webhook
```

`gateway.Client` 处理模型调用，使用 Gateway Key。`management.Client` 处理控制面操作，使用 Management API Key。两个客户端的凭证不能互换。

## Client、Option 与 RequestOption

Client 是可复用的服务对象。应用通常在启动时创建一次：

<!-- evidence: option_test.go, environment_test.go -->
```go
client := gateway.NewClient(
	owlvigil.WithAPIKey(gatewayKey),
	owlvigil.WithTimeout(30*time.Second),
)

```

`owlvigil.Option` 配置整个客户端，例如凭证、Base URL、HTTP Client、超时和重试。

`owlvigil.RequestOption` 只影响一次请求，例如幂等键或 `workspace_id`：

<!-- evidence: option_test.go, environment_test.go -->
```go
result, meta, err := client.CreateGatewayKey(
	ctx,
	req,
	owlvigil.WithIdempotencyKey(operationID),
)
if err != nil {
	return err
}
_ = result
_ = meta

```

## Context 是请求生命周期

每个网络方法都接收 `context.Context`。它负责取消、deadline 和请求级生命周期：

<!-- evidence: option_test.go, environment_test.go -->
```go
ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
defer cancel()

```

不要把已经取消的 context 保存到全局重复使用。HTTP Client timeout 是客户端安全上限，context deadline 是一次业务操作的上限；先到期的一方会停止请求。

## Resource、Metadata 与 Error

多数 Gateway 和 Management 方法返回三个值：

<!-- evidence: option_test.go, environment_test.go -->
```go
models, meta, err := client.ListModels(ctx)
if err != nil {
	return err
}
_ = models
_ = meta

```

- `models`：解码后的业务结果。
- `meta`：请求 ID、服务端 code 和 message 等响应元数据。
- `err`：API、网络、解码或 context 错误。

必须先判断 `err`，再读取 `models`。成功写操作应记录 `meta.RequestID`；API 失败时从 `*owlvigil.APIError` 读取 Request ID。

## 工作区是 Management 的资源边界

成员、Gateway Key、Provider、预算和 Webhook 等资源通常属于工作区。不同方法可能通过 path、请求体或 `workspace_id` query 表达工作区，必须按照方法签名传递，不能自行拼接 endpoint。

列表返回多个工作区时，不要默认选择第一项。生产应用应通过已验证的租户映射、配置或用户选择确定工作区。

## 分页与 Cursor

Management 列表通常返回 `ListResponse[T]`。其中 `Items` 是当前页资源，`PageInfo.NextCursor` 是下一页游标。Cursor 是不透明值，只能原样传回，不能计算或解析为页码。

## 幂等性

幂等键表示“这是同一次业务写入”。当前服务只在 Gateway Key 创建和 Webhook endpoint 创建接口接受该 header。对这两种操作，网络超时后使用同一个键和相同请求体重试，服务端才有机会识别重复请求；重新生成键等同于发起新的业务操作。

只读请求通常可以安全重试。轮换、支付、邀请和 Webhook 重投等其他写操作不支持通过幂等键自动重试，必须先通过读接口或 Request ID 核对结果。

## Secret 与普通 ID

资源 ID 和 Request ID 通常可以用于排查问题。Management API Key、Gateway Key、上游模型服务凭证和 Webhook 签名密钥属于敏感信息；其中部分完整密钥可能只显示一次，应直接保存到密钥管理服务，不能输出到普通应用日志。

下一步阅读[认证与凭证](03-authentication-configuration.md#认证与凭证)和[客户端配置](03-authentication-configuration.md#客户端配置)。
