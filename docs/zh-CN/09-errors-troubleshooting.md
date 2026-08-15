# 错误处理与故障排查

## 错误处理

当前文档范围内的错误分为 API 错误、网络/解码错误和 context 错误。调用者必须先检查 `err`，再使用响应对象。

### APIError

Gateway 和 Management 的非 2xx 响应返回 `*owlvigil.APIError`：

<!-- evidence: internal/owlvigilhttp/request_test.go, internal/owlvigilhttp/response_test.go -->
```go
var apiErr *owlvigil.APIError
if errors.As(err, &apiErr) {
	slog.Warn("OwlVigil API request failed",
		"status", apiErr.StatusCode,
		"code", apiErr.Code,
		"request_id", apiErr.RequestID,
	)
}

```

`Body` 可能包含经过截断和脱敏的服务端正文。默认不要把它写入普通日志；其中仍可能含有用户业务数据。

### Context 和网络错误

<!-- evidence: internal/owlvigilhttp/request_test.go, internal/owlvigilhttp/response_test.go -->
```go
switch {
case errors.Is(err, context.Canceled):
	// 调用方主动取消。
case errors.Is(err, context.DeadlineExceeded):
	// 超过业务 deadline。
default:
	// 使用 errors.As 检查 APIError，再处理网络或解码错误。
}

```

网络超时不能证明服务端没有执行写操作。只有 Gateway Key 创建和 Webhook endpoint 创建支持幂等键；这两种操作重试时复用原键和相同请求体。其他写操作必须先通过读接口或 Request ID 核对结果，不能自动重试。

### 按状态码处理

| 状态 | 建议 |
| --- | --- |
| `400` / `422` | 修正参数，不自动重试 |
| `401` | 检查凭证类型、值、过期和目标客户端 |
| `403` | 检查权限范围（scope）、角色、工作区和资源权限 |
| `404` | 检查资源 ID、工作区和是否已删除 |
| `409` | 读取当前资源状态，再决定是否重复写入 |
| `429` | 调用方解析服务端等待提示并退避；SDK 不会自动重试 429 |
| `502` / `503` / `504` | SDK 可对只读请求，以及两个明确支持幂等键的创建请求，进行固定间隔、有上限的重试 |
| 其他 `5xx` | 根据业务语义由调用方决定，不属于 SDK 默认自动重试状态 |

### ResponseMeta

成功的 Gateway 和 Management 方法通常返回 `*owlvigil.ResponseMeta`，包含 `RequestID`、`Code` 和 `Message`。对关键写操作记录 Request ID，但不要把业务 payload 或凭证一起写入日志。

## 故障排查

### 先收集安全信息

记录 SDK 版本、客户端类型、方法名、发生时间、HTTP 状态和 Request ID。不要收集 Management API Key、Gateway Key、访问令牌、Webhook 签名密钥、认证请求头或完整的敏感请求正文。

### 401 Unauthorized

检查环境变量是否为空、凭证是否过期或禁用，以及是否把 Gateway Key 用在 Management 客户端。确认应用实际读取到预期的部署 secret，而不是本地 `.env` 中的旧值。

### 403 Forbidden

凭证通常有效，但缺少所需的权限范围（scope）、角色、工作区访问权或资源级权限。核对工作区 ID，并检查基于角色的访问控制（RBAC）和成员权限设置。

### 404 Not Found

确认资源 ID 属于当前工作区、资源尚未被删除，并且没有把完整的接口地址手工拼接到已经包含 `/v1` 的 Management 基础地址后面。

### 429 或 5xx

对只读请求使用有上限的退避重试。只有 Gateway Key 创建和 Webhook endpoint 创建支持幂等键；其他写请求不能自动重试。若应用外层已有重试，SDK 使用 `WithoutRetry()` 避免乘法放大。

### context deadline exceeded

比较业务 context deadline 和 HTTP Client timeout。若总是固定时长失败，检查代理、DNS、TLS 和下游连接池。不要简单无限增大 timeout；先确认延迟发生在哪一层。

### 流提前结束

循环结束后检查 `stream.Err()`，确认创建 Stream 的 context 没被提前取消，并确保只有一个 goroutine 消费。部分输出不能自动视为完整响应。

### Webhook 签名失败

确认验签使用原始字节、正确 header、当前 secret 和同步的系统时间。不要在 JSON decode 后重新 marshal 再验签。轮换期间确认接收端部署顺序。

### 写操作结果不确定

网络超时不代表服务端没有成功。两个支持幂等键的创建操作应复用原键和相同请求体；其他写操作先读取资源、订单、订阅或 Webhook 事件状态确认，不要自动重试，并向支持提供 Request ID。

### 文档与代码不一致

从仓库根目录运行：

<!-- evidence: scripts/check-docs.sh -->
```bash
sh scripts/check-docs.sh
go test ./...
```

方法签名以当前版本的 pkg.go.dev 和源码为准。报告问题时附上 `owlvigil.Version`。
