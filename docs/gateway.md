# Gateway

Gateway uses `https://api.owlvigil.com` by default. It must not use `open.owlvigil.com` for model invocation.

## Chat Completions

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
fmt.Println(meta.RequestID, resp.Choices[0].Message.Content)
```

## Responses

```go
resp, _, err := client.CreateResponse(ctx, &gateway.ResponseRequest{
    Model: "gpt-4o-mini",
    Input: "Summarize this request.",
})
```

## Models

```go
models, _, err := client.ListModels(ctx)
model, _, err := client.GetModel(ctx, "gpt-4o-mini")
```

## Embeddings

```go
embeddings, _, err := client.CreateEmbeddings(ctx, &gateway.EmbeddingsRequest{
    Model: "text-embedding-3-small",
    Input: []string{"first document", "second document"},
})
```

## Anthropic-Compatible Messages

```go
message, _, err := client.CreateAnthropicMessage(ctx, &gateway.AnthropicMessageRequest{
    Model: "claude-3-5-sonnet",
    Messages: []gateway.AnthropicMessage{
        {Role: "user", Content: "Hello"},
    },
})
```
