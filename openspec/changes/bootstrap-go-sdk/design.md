## Context

`owlvigil-go` is a new standalone Go SDK project for customer integrations. The SDK must not import the OwlVigil dashboard/backend module because backend internals are implementation details, may change independently, and include dependencies that are inappropriate for customer applications.

The SDK is an HTTP client library. It must not connect directly to OwlVigil databases, Redis, queues, or other server-side infrastructure. All customer operations go through the public OwlVigil API service.

The public API surface has two distinct domains:

- Gateway model calls use `https://gateway.owlvigil.com` with Gateway key authentication.
- Dashboard and Open API management workflows use `https://open.owlvigil.com` with OAuth2.0 user access tokens.

The SDK should expose simple hand-written clients first, with room to generate lower-level request/response types from OpenAPI once the service contract stabilizes.

## Goals / Non-Goals

**Goals:**

- Provide a standalone Go module for OwlVigil customers.
- Support context-aware HTTP requests, custom transports, base URL overrides, JSON request/response handling, structured API errors, and request ID propagation.
- Support streaming Gateway responses with context cancellation and explicit close semantics.
- Provide a Gateway client that works naturally with OpenAI-compatible usage patterns.
- Provide a Management client for Open API endpoints on `open.owlvigil.com/open/v1`.
- Provide OAuth2.0 helpers for user authorization and Management API access tokens.
- Provide pagination, secret redaction, idempotency, webhook verification, and SDK version identification.
- Include public documentation, examples, and tests that make the first customer integration obvious.

**Non-Goals:**

- Do not import dashboard/backend packages from the OwlVigil server repository.
- Do not connect directly to databases, Redis, queues, or server-side infrastructure.
- Do not implement the OwlVigil service or mock all backend behavior.
- Do not require OAuth2.0 for Gateway model calls.
- Do not support every future endpoint in the first SDK version.
- Do not add code generation as a hard dependency for the initial implementation.
- Do not build a CLI, local daemon, proxy, or embedded server.

## Decisions

1. **Standalone module over backend submodule**

   The SDK will live as its own Go module and repository. This keeps customer dependencies small and lets SDK versions follow public API compatibility instead of backend release cadence.

   Alternative considered: place SDK code inside the dashboard repository. Rejected because it would blur internal and public API boundaries and make release management harder.

2. **HTTP-only SDK boundary**

   The SDK will communicate only through documented HTTP endpoints. Data storage, authorization enforcement, billing, audit logs, rate limits, provider routing, and database access remain inside the OwlVigil service.

   Alternative considered: expose database access or import server packages for reuse. Rejected because it would leak internal schemas, bypass service-side policy enforcement, and make customer applications fragile.

3. **Two top-level clients**

   The SDK will expose `GatewayClient` and `ManagementClient`. `GatewayClient` targets `https://gateway.owlvigil.com`; `ManagementClient` targets `https://open.owlvigil.com/open/v1`.

   Alternative considered: one client with all methods. Rejected because Gateway and Management use different domains, authentication methods, and customer mental models.

4. **Gateway key for Gateway, OAuth2.0 token for Management**

   Gateway calls use `Authorization: Bearer ov_sk_xxx` to preserve OpenAI-compatible SDK expectations. Management calls use OAuth2.0 user access tokens because they operate on user/workspace-controlled resources.

   Alternative considered: use OAuth2.0 for all SDK calls. Rejected because model invocation should be simple, service-to-service friendly, and compatible with existing model SDK conventions.

5. **Hand-written ergonomic client first**

   The first implementation will hand-write core client methods and shared HTTP plumbing. Generated code can be introduced later for broad endpoint coverage without exposing generated internals as the primary customer API.

   Alternative considered: generate the entire SDK immediately from OpenAPI. Rejected because the API contract is still being shaped and generated SDKs are often less ergonomic for first-party customer examples.

6. **Functional options for configuration**

   Clients will use functional options such as `WithAPIKey`, `WithAccessToken`, `WithBaseURL`, `WithHTTPClient`, and `WithUserAgent`.

   Alternative considered: large exported config structs only. Rejected because options compose better as the SDK grows while still allowing a config struct internally.

7. **Streaming as a typed iterator**

   Streaming Gateway methods will return a stream object with `Next()`, `Current()`, `Err()`, and `Close()` style behavior, backed by SSE parsing. Context cancellation must stop the stream promptly, and callers must be able to close streams explicitly.

   Alternative considered: expose raw `io.ReadCloser` only. Rejected because callers would need to reimplement SSE parsing, error handling, and provider-specific event conversion.

8. **Common pagination model**

   Management list methods will accept common list options and return typed list results with pagination metadata. The SDK should support cursor-based pagination first while leaving room for page/page-size APIs.

   Alternative considered: each list method defines unrelated pagination fields. Rejected because customers need consistent list workflows across keys, logs, usage, webhooks, and workspaces.

9. **Security helpers in the SDK core**

   The SDK will redact API keys, OAuth tokens, webhook secrets, and client secrets from error strings and debug output. Idempotency keys will be request options, not global client configuration.

   Alternative considered: leave secret hygiene to customer code. Rejected because SDK errors are often logged directly and should be safe by default.

10. **Webhook verification helper**

   The SDK will provide a small helper to verify OwlVigil webhook signatures for customer webhook receivers. This helper belongs in the SDK because signature verification must be implemented consistently across customer services.

   Alternative considered: only document signature verification. Rejected because hand-rolled verification is error-prone and can create security gaps.

11. **Public docs as a first-class deliverable**

   The SDK project will include customer-facing docs under `docs/` plus README quickstarts and runnable examples. The docs should explain both direct API calling and SDK usage so customers can debug integrations without reading SDK source.

   Alternative considered: only provide README examples. Rejected because customers need a durable reference for domains, auth modes, request flows, error handling, and troubleshooting.

12. **Small library-oriented project layout**

   The project will use a flat public root package plus a few focused subpackages for optional domains. Internal HTTP plumbing will live under `internal/` to keep it private. Runnable examples and customer docs will be first-class directories.

   Proposed layout:

   ```text
   owlvigil-go/
     go.mod
     README.md
     CHANGELOG.md
     LICENSE
     client.go
     client_test.go
     option.go
     error.go
     error_test.go
     version.go
     gateway/
       client.go
       client_test.go
       chat.go
       chat_test.go
       responses.go
       models.go
       embeddings.go
       anthropic.go
       stream.go
       stream_test.go
     management/
       client.go
       client_test.go
       workspaces.go
       gateway_keys.go
       gateway_keys_test.go
       usage.go
       logs.go
       webhooks.go
       pagination.go
       pagination_test.go
     oauth2/
       client.go
       authorize.go
       authorize_test.go
       token.go
       token_test.go
       userinfo.go
     webhook/
       verify.go
       verify_test.go
     internal/
       owlvigilhttp/
         request.go
         request_test.go
         response.go
         response_test.go
         retry.go
         retry_test.go
         redact.go
         redact_test.go
       sse/
         scanner.go
         scanner_test.go
     examples/
       gateway-chat/
         main.go
       gateway-stream/
         main.go
       management-key/
         main.go
       oauth2-flow/
         main.go
       webhook-verify/
         main.go
       examples_test.go
     docs/
       quickstart.md
       authentication.md
       gateway.md
       management.md
       oauth2.md
       streaming.md
       webhooks.md
       errors.md
       pagination.md
       troubleshooting.md
   ```

   Alternative considered: mirror the service architecture with `cmd/`, `internal/server`, and layered service packages. Rejected because this is a public SDK library, not a backend service.

## Risks / Trade-offs

- Public API changes after early customers adopt the SDK -> Keep the initial surface small, document beta status if needed, and use semantic versioning.
- Drift between SDK types and service responses -> Add contract tests with HTTP fixtures and later consider OpenAPI-generated low-level types.
- Confusion between `gateway.owlvigil.com` and `open.owlvigil.com` -> Keep separate clients, defaults, examples, and error messages.
- Accidental coupling to backend internals -> Add tests or dependency checks confirming the SDK imports no dashboard/backend modules and no database drivers.
- Streaming APIs can leak resources if not closed -> Document `defer stream.Close()` and enforce close behavior in tests.
- Secrets can leak through logs -> Redact credentials before formatting SDK errors and include tests for redaction.
- Pagination differences across endpoints can surprise callers -> Normalize common list options and expose endpoint-specific fields only when needed.
- OAuth2.0 details vary by customer integration type -> Keep OAuth helpers focused on standard authorization code, refresh, userinfo, and revoke flows.
- Retry behavior can duplicate non-idempotent requests -> Default retries only to safe transient failures and document how to disable or tune retries.
- Docs can drift from SDK behavior -> Compile examples in tests and keep docs examples copied from runnable snippets where possible.

## Migration Plan

1. Create the Go module and shared SDK foundations.
2. Add Gateway client methods and examples first because model invocation is the primary customer integration path.
3. Add OAuth2.0 helpers and Management client methods for key and usage management.
4. Add public docs, README, examples, tests, and CI.
5. Tag the first SDK version after examples compile and tests pass.

Rollback is simple before first public release: revise the SDK API and OpenSpec artifacts. After release, incompatible changes must go through a new major version.

## Open Questions

- What final Go module path should be used, for example `github.com/Syrovex/owlvigil_sdk_go`?
- Which Gateway endpoints are required in the first public release beyond chat completions, responses, models, and embeddings?
- Which Management endpoints should be included before generated OpenAPI support exists?
- Should OAuth2.0 PKCE helpers be included in the first release or in a follow-up change?
