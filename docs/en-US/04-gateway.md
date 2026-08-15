# Model Calls and Streaming

## Gateway

Gateway uses `https://gateway.owlvigil.com` by default. Do not send model-invocation requests to the Management API domain.

### Chat Completions

<!-- evidence: gateway/client_test.go, gateway/stream_test.go -->
```go
client := gateway.NewClient(owlvigil.WithAPIKey(os.Getenv("OWLVIGIL_GATEWAY_KEY")))

resp, meta, err := client.CreateChatCompletion(ctx, &gateway.ChatCompletionRequest{
	Model: "gpt-4o-mini",
	Messages: []gateway.Message{
		{Role: "user", Content: "Write one sentence about observability."},
	},
})
if err != nil {
	return err
}
if len(resp.Choices) == 0 || resp.Choices[0].Message == nil {
	return fmt.Errorf("empty chat response: request_id=%s", meta.RequestID)
}
fmt.Println(meta.RequestID, resp.Choices[0].Message.Content)

```

### Responses

<!-- evidence: gateway/client_test.go, gateway/stream_test.go -->
```go
resp, _, err := client.CreateResponse(ctx, &gateway.ResponseRequest{
	Model: "gpt-4o-mini",
	Input: "Summarize this request.",
})
if err != nil {
	return err
}
_ = resp

```

### Models

<!-- evidence: gateway/client_test.go, gateway/stream_test.go -->
```go
models, _, err := client.ListModels(ctx)
if err != nil {
	return err
}
model, _, err := client.GetModel(ctx, "gpt-4o-mini")
if err != nil {
	return err
}
_ = models
_ = model

```

### Embeddings

<!-- evidence: gateway/client_test.go, gateway/stream_test.go -->
```go
embeddings, _, err := client.CreateEmbeddings(ctx, &gateway.EmbeddingsRequest{
	Model: "text-embedding-3-small",
	Input: []string{"first document", "second document"},
})
if err != nil {
	return err
}
_ = embeddings

```

### Anthropic-Compatible Messages

<!-- evidence: gateway/client_test.go, gateway/stream_test.go -->
```go
message, _, err := client.CreateAnthropicMessage(ctx, &gateway.AnthropicMessageRequest{
	Model: "claude-3-5-sonnet",
	Messages: []gateway.AnthropicMessage{
		{Role: "user", Content: "Hello"},
	},
})
if err != nil {
	return err
}
_ = message

```

### Request metadata and streams

Every successful call returns `*owlvigil.ResponseMeta`; record its `RequestID`
when diagnosing unexpected model behavior. Chat and Responses APIs also expose
streaming variants. See [Streaming](04-gateway.md#streaming) for event ownership and
cleanup rules.

Gateway requests must use `OWLVIGIL_GATEWAY_KEY`. They do not accept the
Management service-account key. Before presenting models to an end user, use
`ListModels` and `GetModel` to refresh metadata. For workspace routing,
providers, and Gateway-key lifecycle, use the Management client described in
[Model routing](07-management-operations.md#model-routing-and-provider-management).

## Streaming

Gateway streaming methods return a stream with `Next`, `Current`, `Err`, and `Close`.

<!-- evidence: gateway/client_test.go, gateway/stream_test.go -->
```go
stream, err := client.CreateChatCompletionStream(ctx, &gateway.ChatCompletionRequest{
	Model: "gpt-4o-mini",
	Messages: []gateway.Message{
		{Role: "user", Content: "Count to three."},
	},
})
if err != nil {
	return err
}
defer stream.Close()

for stream.Next() {
	event := stream.Current()
	fmt.Println(event.Event, string(event.Data))
}
if err := stream.Err(); err != nil {
	return err
}

```

Cancel the context to stop waiting for more events. Always close the stream to release the response body.

`CreateResponseStream` provides the same stream contract for the Responses API.
Call `Current` only after `Next` returns true; when it returns false, call
`Err` to distinguish a clean end-of-stream from a transport or decoding error.
Do not start a second goroutine that reads the same stream: one consumer should
own `Next`, `Current`, `Err`, and `Close`.

Streaming requests can produce partial output. Persist or display an event only
after applying the ordering and retry policy appropriate for your product; a
new stream request is not a continuation of a failed stream unless the upstream
API explicitly provides a resume token.
