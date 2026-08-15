# 模型调用与流式响应

## Gateway 模型调用

`gateway.Client` 用于模型推理。它提供 OpenAI 兼容的 Chat Completions、Responses 和 Embeddings，Anthropic 兼容的 Messages，以及模型发现接口。

### 创建客户端

<!-- evidence: gateway/client_test.go, gateway/stream_test.go -->
```go
client := gateway.NewClient(
	owlvigil.WithAPIKey(os.Getenv("OWLVIGIL_GATEWAY_KEY")),
)

```

一个客户端可安全地复用其底层 HTTP Client。应用通常应在启动时创建客户端，而不是每次请求创建。

### 发现可用模型

不要把示例模型名当作所有工作区都可用的固定值。先读取当前 Gateway Key 可访问的模型：

<!-- evidence: gateway/client_test.go, gateway/stream_test.go -->
```go
models, meta, err := client.ListModels(ctx)
if err != nil {
	return err
}
for _, model := range models.Data {
	fmt.Println(model.ID, model.OwnedBy)
}
slog.Info("listed Gateway models", "request_id", meta.RequestID)

```

读取单个模型时，SDK 会对模型 ID 进行 URL path 转义：

<!-- evidence: gateway/client_test.go, gateway/stream_test.go -->
```go
model, _, err := client.GetModel(ctx, modelID)
if err != nil {
	return err
}
_ = model

```

### Chat Completions

<!-- evidence: gateway/client_test.go, gateway/stream_test.go -->
```go
temperature := 0.2
maxTokens := 500

resp, meta, err := client.CreateChatCompletion(ctx, &gateway.ChatCompletionRequest{
	Model: modelID,
	Messages: []gateway.Message{
		{Role: "system", Content: "回答要简洁。"},
		{Role: "user", Content: "解释什么是幂等请求。"},
	},
	Temperature: &temperature,
	MaxTokens:   &maxTokens,
})
if err != nil {
	return err
}
if len(resp.Choices) == 0 || resp.Choices[0].Message == nil {
	return fmt.Errorf("empty response: request_id=%s", meta.RequestID)
}
fmt.Println(resp.Choices[0].Message.Content)

```

`Temperature` 和 `MaxTokens` 使用指针，以区分“未发送”与数值零。

### Responses API

`Input` 是 `any`，可承载上游接口支持的字符串或结构化输入。SDK 不会在本地验证其具体形状。

<!-- evidence: gateway/client_test.go, gateway/stream_test.go -->
```go
resp, meta, err := client.CreateResponse(ctx, &gateway.ResponseRequest{
	Model: modelID,
	Input: "把这段内容概括为三个要点。",
})
if err != nil {
	return err
}
_ = resp
_ = meta

```

处理 `Response.Output` 时应根据服务端返回契约做类型检查，不要直接断言为某个 Go 类型。

### Embeddings

<!-- evidence: gateway/client_test.go, gateway/stream_test.go -->
```go
resp, _, err := client.CreateEmbeddings(ctx, &gateway.EmbeddingsRequest{
	Model: embeddingModelID,
	Input: []string{"第一段文本", "第二段文本"},
})
if err != nil {
	return err
}
for _, item := range resp.Data {
	fmt.Println(item.Index, len(item.Embedding))
}

```

向量维度由模型决定。下游存储必须验证实际长度，不能只依赖示例。

### Anthropic 兼容 Messages

<!-- evidence: gateway/client_test.go, gateway/stream_test.go -->
```go
resp, meta, err := client.CreateAnthropicMessage(ctx, &gateway.AnthropicMessageRequest{
	Model: modelID,
	Messages: []gateway.AnthropicMessage{
		{Role: "user", Content: "你好"},
	},
	MaxTokens: 256,
})
if err != nil {
	return err
}
_ = resp
_ = meta

```

`Content` 使用 `any` 以支持兼容接口的不同内容结构。传入结构化内容时，以 OwlVigil 服务端当前契约为准。

### 流式调用

Chat 和 Responses 分别提供 `CreateChatCompletionStream` 与 `CreateResponseStream`。流式方法返回 SSE 事件，不会自动组装成最终响应。参见[流式响应](04-gateway.md#流式响应)。

### 生产注意事项

- 为每次业务调用设置 context deadline。
- 记录成功响应的 `meta.RequestID` 和错误中的 `APIError.RequestID`。
- 在展示或持久化模型输出前处理空结果、内容类型和业务安全策略。
- 不要记录完整 prompt、响应或认证头，除非已经完成数据分类、授权和脱敏。
- SDK 可自动重试只读 Gateway 请求。模型生成调用不支持本文所述的 Management 幂等键；发生不确定结果时不要自动重放，以免产生重复用量或输出。

## 消费流式响应

Gateway 的 Chat 和 Responses 流式方法返回 `*gateway.Stream`。每个 Stream 只能由一个消费者读取。

### 基本循环

<!-- evidence: gateway/client_test.go, gateway/stream_test.go -->
```go
stream, err := client.CreateChatCompletionStream(ctx, &gateway.ChatCompletionRequest{
	Model: modelID,
	Messages: []gateway.Message{
		{Role: "user", Content: "从一数到三。"},
	},
})
if err != nil {
	return err
}
defer stream.Close()

for stream.Next() {
	event := stream.Current()
	fmt.Printf("event=%s data=%s\n", event.Event, event.Data)
}
if err := stream.Err(); err != nil {
	return err
}

```

调用顺序固定为：`Next` 返回 `true` 后读取 `Current`；循环结束后读取 `Err`；所有路径都调用 `Close`。

### 事件数据

`StreamEvent` 包含：

- `Event`：SSE event 名称。
- `Data`：未经业务解码的 `json.RawMessage`。
- `Raw`：原始 data 字符串。

SDK 只负责拆分 SSE 帧，不会假设所有事件都有相同 JSON 结构。事件名称和 payload 结构属于服务端协议；应用应依据当前 Gateway 协议按 `Event` 分派并将 `Data` 解码到对应类型，不要根据本页自行发明事件名称。

### 取消和清理

- 用户断开连接或业务超时时，取消创建 Stream 时使用的 context。
- `Close` 释放响应体和底层连接；即使消费到正常结束也要调用。
- 不要让多个 goroutine 同时调用 `Next` 或 `Current`。
- 流中断后已经展示的内容属于部分结果。除非协议提供恢复 token，否则重新请求不是原流的续传。

### 背压

`Next` 按消费者速度读取。如果下游写入很慢，应使用有界队列或直接施加背压，不要创建无限增长的内存缓冲。客户端断开时立即取消上游 context，避免继续产生无消费者的 token。
