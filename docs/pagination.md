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
