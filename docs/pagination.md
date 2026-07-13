# Pagination

Management list methods accept `management.ListOptions` and return `management.ListResponse[T]`.

```go
var cursor string
for {
    page, _, err := client.ListGatewayKeys(ctx, management.ListOptions{
        Cursor: cursor,
        Limit:  100,
    }, "")
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
}
```

Endpoint-specific filters can be combined with common pagination. For example, `ListRequestLogs` accepts a Gateway key filter in addition to `ListOptions`.

Set `Limit` only when you want to override the service default; the SDK omits
zero or negative limits. Treat `NextCursor` as opaque: persist and send it back
unchanged, and never attempt to derive offsets from it. A list response may be
decoded from either a bare array or an object containing `items` and
`page_info`, so always use `Items` and `PageInfo` rather than decoding response
JSON yourself.

For long exports, retain the cursor only after successfully handling a page.
If the process is restarted, resume from that stored cursor and deduplicate by
resource ID where the downstream destination requires exactly-once processing.
