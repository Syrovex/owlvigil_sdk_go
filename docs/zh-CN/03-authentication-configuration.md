# 认证与客户端配置

## 认证与凭证

本文介绍模型调用和 Management 操作使用的两类 API Key。选择客户端时先选择凭证，避免把“能管理工作区”和“能调用模型”当成同一种权限。

| 使用场景 | 客户端 | 推荐环境变量 | SDK 配置 |
| --- | --- | --- | --- |
| 调用模型 | `gateway.Client` | `OWLVIGIL_GATEWAY_KEY` | `WithAPIKey` |
| 管理工作区和资源 | `management.Client` | `OWLVIGIL_API_KEY` | `WithAPIKey` |

### 静态 API Key

<!-- evidence: option_test.go, environment_test.go -->
```go
gatewayClient := gateway.NewClient(
	owlvigil.WithAPIKey(os.Getenv("OWLVIGIL_GATEWAY_KEY")),
)

managementClient := management.NewClient(
	owlvigil.WithAPIKey(os.Getenv("OWLVIGIL_API_KEY")),
)

```

应用启动时应检查环境变量非空。SDK 会发送 Bearer 认证头，但不会替你判断某个字符串是不是正确类型的密钥。

### 存储与轮换

- 生产环境通过环境变量或专用的密钥管理服务提供凭证，不要把凭证直接写入源代码或编译后的程序。
- `.env` 文件只用于本地开发，并且不能提交到 Git 仓库。
- 创建或轮换 Gateway Key 时，接口返回的完整密钥可能只显示一次。收到后应立即保存到安全的密钥管理服务中，不要输出到日志。
- 轮换 Gateway Key 时，先让调用模型的应用改用新 Key；确认调用成功后，再禁用旧 Key。
- 日志、监控数据、调用链记录、Issue 和支持请求中都不能包含 Management API Key、Gateway Key、上游模型凭证或 Webhook 签名密钥。

### 认证错误

- `401 Unauthorized` 通常表示凭证缺失、过期、禁用或发给了错误的客户端。
- `403 Forbidden` 通常表示凭证有效，但缺少所需的权限范围（scope）、角色、工作区访问权或资源权限。

捕获 `*owlvigil.APIError` 并记录 `RequestID`，不要记录请求头。详见[错误处理](09-errors-troubleshooting.md#错误处理)。

如果还不熟悉 Client、Option 和 RequestOption，先阅读[核心概念](02-concepts.md)。

## 客户端配置

所有客户端都通过 `owlvigil.Option` 配置。选项按传入顺序执行；后面的选项可以覆盖前面的结果。

### 默认值

| 配置 | 默认值 |
| --- | --- |
| Gateway Base URL | `https://gateway.owlvigil.com` |
| Management Base URL | `https://api.owlvigil.com/v1` |
| HTTP 超时 | 60 秒 |
| 重试次数 | 2 |
| 首次重试等待 | 200 毫秒 |
| User-Agent | `owlvigil-go/<version>` |

需要确认客户端当前使用的基础地址时，可调用 `BaseURL`；不要依赖未导出的配置字段。

### 超时和上下文

<!-- evidence: option_test.go, environment_test.go -->
```go
client := gateway.NewClient(
	owlvigil.WithAPIKey(key),
	owlvigil.WithTimeout(20*time.Second),
)

ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
defer cancel()

```

对于非流式请求，调用会在 context deadline 或 HTTP Client timeout 中更早到期的一侧停止。流式调用会复制 HTTP Client 并清除其总超时，避免长连接被固定时长截断，因此必须通过 context deadline 或主动取消来限制生命周期。

### 自定义 HTTP Client

<!-- evidence: option_test.go, environment_test.go -->
```go
transport := http.DefaultTransport.(*http.Transport).Clone()
transport.MaxIdleConns = 100
transport.MaxIdleConnsPerHost = 20

httpClient := &http.Client{
	Transport: transport,
	Timeout:   30 * time.Second,
}

client := management.NewClient(
	owlvigil.WithHTTPClient(httpClient),
	owlvigil.WithAPIKey(apiKey),
)

```

复用客户端和 Transport，避免每次请求重新创建连接池。`WithTimeout` 会浅拷贝已有 HTTP Client，因此 Transport 等指针字段仍然共享。

### 环境和自定义地址

<!-- evidence: option_test.go, environment_test.go -->
```go
client := management.NewClient(
	owlvigil.WithEnvironment(owlvigil.EnvironmentStaging),
	owlvigil.WithAPIKey(apiKey),
)

```

也可设置 `OWLVIGIL_ENV=production|staging|local` 后使用 `WithEnvironmentFromEnv()`。未知环境会在发起请求时返回配置错误。

`WithEnvironment` 会修改当前 Base URL。需要显式地址时先传环境，再传 `WithBaseURL`：

<!-- evidence: option_test.go, environment_test.go -->
```go
client := management.NewClient(
	owlvigil.WithEnvironment(owlvigil.EnvironmentStaging),
	owlvigil.WithBaseURL(testServer.URL+"/v1"),
)

```

生产应用通常不应允许普通终端用户控制 Base URL，否则可能造成服务端请求伪造或凭证泄露。

### 重试

<!-- evidence: option_test.go, environment_test.go -->
```go
client := gateway.NewClient(
	owlvigil.WithRetry(3, 500*time.Millisecond),
)

```

`WithRetry(3, 500*time.Millisecond)` 表示初次请求失败后最多再重试 3 次，重试之间固定等待 500 毫秒。SDK 只自动重试网络 timeout、实现了 `Temporary() == true` 的网络错误，以及 HTTP `502`、`503`、`504`；它不会解析 `Retry-After`，也不会自动重试 `429`。传入 `WithIdempotencyKey` 会让 SDK transport 允许写请求重试，但当前服务只在 Gateway Key 创建和 Webhook endpoint 创建接口支持该 header，因此不得把它用于其他写操作：

<!-- evidence: option_test.go, environment_test.go -->
```go
resource, meta, err := client.CreateGatewayKey(
	ctx,
	req,
	owlvigil.WithIdempotencyKey("gateway-key-create-production-v1"),
)
if err != nil {
	return err
}
_ = resource
_ = meta

```

如果应用有统一的指数退避、抖动或 `Retry-After` 处理层，使用 `WithoutRetry()` 避免重试叠加。

### 单次请求选项

| 选项 | 用途 |
| --- | --- |
| `WithIdempotencyKey` | 仅为 Gateway Key 创建或 Webhook endpoint 创建设置 `Idempotency-Key` |
| `WithHeader` | 设置附加请求头；不要覆盖认证头或写入不可信值 |
| `WithQueryParam` | 添加或覆盖查询参数 |
| `WithWorkspaceID` | 设置需要 `workspace_id` 查询参数的 Management 请求 |
