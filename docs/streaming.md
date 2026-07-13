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

`CreateResponseStream` provides the same stream contract for the Responses API.
Call `Current` only after `Next` returns true; when it returns false, call
`Err` to distinguish a clean end-of-stream from a transport or decoding error.
Do not start a second goroutine that reads the same stream: one consumer should
own `Next`, `Current`, `Err`, and `Close`.

Streaming requests can produce partial output. Persist or display an event only
after applying the ordering and retry policy appropriate for your product; a
new stream request is not a continuation of a failed stream unless the upstream
API explicitly provides a resume token.
