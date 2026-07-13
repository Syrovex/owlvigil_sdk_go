# Examples

Runnable examples live under `examples/`.

| Directory | Workflow |
| --- | --- |
| `examples/gateway-chat` | Create a Gateway chat completion |
| `examples/gateway-stream` | Read a Gateway streaming chat response |
| `examples/gateway-models` | List Gateway models |
| `examples/management-key` | Create a Gateway key through Management API |
| `examples/management-usage` | Read usage summary |
| `examples/oauth2-flow` | Build authorization URL and exchange a code |
| `examples/webhook-verify` | Verify a webhook signature |

Compile examples:

```bash
go test ./examples/...
```
