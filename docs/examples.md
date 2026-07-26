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
| `examples/financial-control` | Read financial governance for a workspace |
| `examples/billing-subscription` | Inspect plans and the active subscription |
| `examples/team-management` | List teams and workspace members |
| `examples/openapi-smoke` | Exercise Management operations in an isolated workspace |
| `examples/oauth2-client-credentials` | Obtain an OAuth client-credentials token |
| `examples/oauth2-callback-server` | Receive an OAuth authorization-code callback |
| `examples/oauth2-smoke` | Run a narrow OAuth smoke workflow |
| `examples/multi-environment` | Select production, staging, or local endpoints |

Compile examples:

```bash
go test ./examples/...
```

Examples load the nearest `.env` file without overwriting an exported shell
variable. `OWLVIGIL_GATEWAY_KEY` is for Gateway examples;
`OWLVIGIL_API_KEY` is for Management examples. OAuth examples require the
variables documented in `.env.example`. The `openapi-smoke` example performs
writes and cleanup when `OWLVIGIL_SMOKE_WRITES=1`; set
`OWLVIGIL_SMOKE_REQUIRE_ALL=1` to require a `PASS` for every unique Management
operation.
Password, invitation, secondary-member, policy, Provider, and Stripe
prerequisites are documented in `examples/openapi-smoke/README.md`. Run it only in an isolated
workspace, never against a shared production workspace. Without that flag it
checks read-only paths and records mutating operations as explicit skips; a
skip is not reported as a pass. The catalog test locks the example to all 141
published Management operations, while unit contract tests execute every SDK
method against strict local HTTP fixtures.
