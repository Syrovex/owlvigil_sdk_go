# Webhook

Webhook 有两个独立环节：Management 客户端管理 OwlVigil 向哪里投递事件；`webhook` 包验证你的服务收到的事件。两者不要混为一套认证。

## 验证入站签名

必须使用未经 JSON 解码、格式化或重新编码的原始请求体：

<!-- evidence: webhook/verify_test.go, webhook/example_test.go -->
```go
func webhookHandler(w http.ResponseWriter, r *http.Request) {
	const maxWebhookBody = 1 << 20
	r.Body = http.MaxBytesReader(w, r.Body, maxWebhookBody)
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	timestamp := r.Header.Get("OW-Webhook-Timestamp")
	signature := r.Header.Get("OW-Webhook-Signature")
	signatureHeader := "t=" + timestamp + "," + signature
	err = webhook.VerifySignature(
		payload,
		signatureHeader,
		os.Getenv("OWLVIGIL_WEBHOOK_SECRET"),
		webhook.VerifyOptions{},
	)
	if err != nil {
		http.Error(w, "invalid signature", http.StatusUnauthorized)
		return
	}

	// 签名通过后再解码和处理 payload。
	w.WriteHeader(http.StatusNoContent)
}
```

OwlVigil 分别发送 `OW-Webhook-Timestamp` 和 `OW-Webhook-Signature`。SDK 验证器接收规范化的组合值 `t=<timestamp>,v1=<hex-hmac>`。不要用事件 ID 代替时间戳。默认允许时间戳与本机时间相差 5 分钟，并使用 HMAC 常量时间比较。服务器必须保持时钟同步。

可识别的错误包括 `ErrMissingSignature`、`ErrInvalidSignature`、`ErrMissingSecret` 和 `ErrStaleTimestamp`。对外只返回通用错误；内部监控可以按错误类型计数，但不能记录 Webhook 签名密钥或完整的事件正文。

## 防止重复处理

Webhook 是至少一次投递场景，测试、重试和 redelivery 都可能产生重复事件。消费者应：

1. 验证签名。
2. 读取 `OW-Webhook-Event-Id`。
3. 在数据库中以唯一约束原子记录事件 ID。
4. 已处理事件直接返回成功。
5. 业务处理成功后提交状态。

不要仅靠内存 map 去重；进程重启和多实例部署会丢失状态。

## 管理端点

`management.Client` 提供 endpoint 的创建、查询、更新、删除、启用、禁用、secret 轮换和测试。需要 `workspace_id` 查询参数的方法应传 `WithWorkspaceID`。

| 任务 | 方法 |
| --- | --- |
| 查询和维护接收地址 | `ListWebhookEndpoints`, `CreateWebhookEndpoint`, `GetWebhookEndpoint`, `UpdateWebhookEndpoint`, `DeleteWebhookEndpoint`, `DeleteWebhookEndpointWithResult` |
| 启用和禁用 | `EnableWebhookEndpoint`, `EnableWebhookEndpointWithResult`, `DisableWebhookEndpoint`, `DisableWebhookEndpointWithResult` |
| 轮换签名密钥和发送测试事件 | `RotateWebhookSecret`, `TestWebhookEndpoint`, `TestWebhookEndpointWithResult` |
| 查询服务端支持的事件类型 | `ListWebhookEventTypes` |

<!-- evidence: webhook/verify_test.go, webhook/example_test.go -->
```go
endpoint, meta, err := client.CreateWebhookEndpoint(
	ctx,
	&management.CreateWebhookEndpointRequest{
		WorkspaceID: workspaceID,
		URL:         receiverURL,
		EventTypes:  []string{"event.name"},
	},
	owlvigil.WithIdempotencyKey("webhook-endpoint-primary-v1"),
)
if err != nil {
	return err
}
_ = endpoint
_ = meta

```

具体字段以当前 Go 类型为准；事件名称应来自 `ListWebhookEventTypes`，不要照抄占位符。

## 事件历史与重投

使用 `ListWebhookEvents`、`GetWebhookEvent` 和 `ListEndpointEvents` 查看投递历史。`RetryWebhookEvent`、`RedeliverWebhookEvent`、批量 redelivery 和 `TestWebhookEndpoint` 会产生真实网络投递，应只对能够去重的接收端执行。

单条事件操作提供 `RetryWebhookEvent`、`RetryWebhookEventWithResult`、`RedeliverWebhookEvent` 和 `RedeliverWebhookEventWithResult`；批量操作使用 `BulkRedeliverWebhookEvents` 或 `BulkRedeliverWebhookEventsWithResult`。需要服务端结果对象时使用 `WithResult` 版本。

轮换签名密钥后，以轮换响应和服务端当前契约为准更新接收端；本 SDK 不保证新旧密钥存在重叠有效期。
