# Streaming

Gateway streaming methods return a stream with `Next`, `Current`, `Err`, and `Close`.

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
