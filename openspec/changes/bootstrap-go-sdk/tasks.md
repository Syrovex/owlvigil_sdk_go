## 1. Module Setup

- [x] 1.1 Initialize `go.mod` with the final SDK module path.
- [x] 1.2 Add repository metadata files including `README.md`, `CHANGELOG.md`, `.gitignore`, and license placeholder if needed.
- [x] 1.3 Create root package files for shared options, errors, version metadata, and common types.
- [x] 1.4 Create public packages for `gateway`, `management`, `oauth2`, and `webhook`.
- [x] 1.5 Create private packages under `internal/` for HTTP plumbing, retry, redaction, response decoding, and SSE scanning.
- [x] 1.6 Create `examples/` directories for Gateway chat, Gateway streaming, Management key operations, OAuth2.0 flow, and webhook verification.
- [x] 1.7 Create `docs/` files for quickstart, authentication, Gateway, Management, OAuth2.0, streaming, webhooks, errors, pagination, and troubleshooting.
- [x] 1.8 Add co-located `_test.go` placeholders or initial tests for root, public packages, and internal packages.
- [x] 1.9 Add baseline CI or local verification command documentation for `go test ./...`.

## 2. SDK Core

- [x] 2.1 Implement shared client configuration with functional options for base URL, HTTP client, user agent, timeout, retry policy, and auth token providers.
- [x] 2.2 Implement shared HTTP request builder with context propagation and JSON request encoding.
- [x] 2.3 Implement response envelope decoding with request ID preservation.
- [x] 2.4 Implement structured API error type for non-2xx responses.
- [x] 2.5 Implement bounded retry behavior for transient network and 5xx failures with an option to disable retries.
- [x] 2.6 Implement secret redaction for API keys, OAuth tokens, refresh tokens, client secrets, and webhook secrets.
- [x] 2.7 Implement default `User-Agent: owlvigil-go/<version>` and exported SDK version metadata.
- [x] 2.8 Implement request option support for idempotency keys on supported mutating requests.
- [x] 2.9 Add tests for custom HTTP client usage, context cancellation, envelope decoding, API errors, retry behavior, redaction, User-Agent, and idempotency headers.

## 3. OAuth2.0 Authentication

- [x] 3.1 Implement OAuth2.0 client configuration with default issuer `https://open.owlvigil.com`.
- [x] 3.2 Implement authorization URL generation with scopes, state, redirect URI, and optional PKCE parameters.
- [x] 3.3 Implement authorization code token exchange against `/oauth/token`.
- [x] 3.4 Implement refresh token flow against `/oauth/token/refresh`.
- [x] 3.5 Implement userinfo lookup against `/oauth/userinfo`.
- [x] 3.6 Implement token revocation against `/oauth/revoke`.
- [x] 3.7 Add tests for OAuth2.0 URL generation, token decoding, refresh, userinfo, revoke, and structured OAuth errors.

## 4. Gateway Client

- [x] 4.1 Implement `GatewayClient` with default domain `https://gateway.owlvigil.com`.
- [x] 4.2 Implement Gateway key authentication with `Authorization: Bearer ov_sk_xxx`.
- [x] 4.3 Implement OpenAI-compatible chat completions for `POST /v1/chat/completions`.
- [x] 4.4 Implement OpenAI-compatible responses for `POST /v1/responses`.
- [x] 4.5 Implement model listing and model detail for `GET /v1/models` and `GET /v1/models/{model}`.
- [x] 4.6 Implement embeddings for `POST /v1/embeddings`.
- [x] 4.7 Implement Anthropic-compatible messages for `POST /anthropic/v1/messages`.
- [x] 4.8 Implement streaming chat completions and streaming responses with SSE decoding.
- [x] 4.9 Implement streaming lifecycle methods for `Next`, `Current`, `Err`, and `Close`.
- [x] 4.10 Add Gateway client tests using `httptest.Server` fixtures for headers, paths, payloads, response decoding, streaming events, stream close, and context cancellation.

## 5. Management Client

- [x] 5.1 Implement `ManagementClient` with default base URL `https://open.owlvigil.com/open/v1`.
- [x] 5.2 Implement OAuth2.0 access token authentication for Management requests.
- [x] 5.3 Implement common list options and paginated response metadata.
- [x] 5.4 Implement workspace list and detail methods.
- [x] 5.5 Implement Gateway key create, list, get, update, rotate, enable, disable, and delete methods.
- [x] 5.6 Implement usage, usage summary, quota, balance, and invoice retrieval methods.
- [x] 5.7 Implement request logs, trace lookup, and payload log access methods.
- [x] 5.8 Implement webhook endpoint and webhook event methods.
- [x] 5.9 Add Management client tests using `httptest.Server` fixtures for headers, paths, payloads, pagination, and response decoding.

## 6. Webhook Signatures

- [x] 6.1 Implement webhook signature verification helper.
- [x] 6.2 Implement timestamp tolerance validation.
- [x] 6.3 Implement constant-time signature comparison.
- [x] 6.4 Add webhook verification tests for valid signatures, invalid signatures, stale timestamps, and malformed headers.

## 7. Public Documentation and Examples

- [x] 7.1 Create `docs/` structure for quickstart, authentication, Gateway, Management, OAuth2.0, streaming, webhooks, errors, pagination, examples, and troubleshooting.
- [x] 7.2 Write README quickstart for installing the SDK and choosing Gateway vs Management clients.
- [x] 7.3 Write Gateway calling guide with copy-pasteable chat completions, streaming, responses, model listing, and embeddings examples.
- [x] 7.4 Write Management API guide with workspace, Gateway key, usage, logs, webhooks, and pagination examples.
- [x] 7.5 Write OAuth2.0 guide covering authorization URL, callback, token exchange, refresh, userinfo, and revoke.
- [x] 7.6 Write webhook signature verification guide.
- [x] 7.7 Document default domains and how to override base URLs for tests or private deployments.
- [x] 7.8 Document error handling, request IDs, retries, context cancellation, idempotency, secret redaction, and common troubleshooting steps.
- [x] 7.9 Add runnable examples for chat completions, streaming chat, model listing, Gateway key creation, usage summary, OAuth2.0 token exchange, and webhook verification.
- [x] 7.10 Ensure documentation-linked examples compile under `go test ./...`.

## 8. Verification and Release Prep

- [x] 8.1 Run `go test ./...` and fix all failures.
- [x] 8.2 Review exported Go identifiers for godoc quality.
- [x] 8.3 Confirm examples compile through tests.
- [x] 8.4 Confirm the SDK does not import OwlVigil dashboard/backend modules.
- [x] 8.5 Confirm the SDK has no direct database, Redis, queue, or server-infrastructure dependencies.
- [x] 8.6 Confirm repository layout matches the accepted SDK directory design.
- [x] 8.7 Prepare first version tag notes and customer-facing release summary.
