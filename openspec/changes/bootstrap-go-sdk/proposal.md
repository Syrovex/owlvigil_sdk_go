## Why

OwlVigil needs a dedicated Go SDK so customers can integrate Gateway model calls and Open API management workflows without depending on the dashboard backend repository. Creating the SDK as its own project keeps the public client stable, lightweight, and versioned independently from server implementation details.

## What Changes

- Bootstrap the `owlvigil-go` module as the official Go SDK project.
- Add a Gateway client for model calls through `https://api.owlvigil.com`, authenticated by Gateway key.
- Add a Management client for Open API calls through `https://open.owlvigil.com/open/v1`, authenticated by OAuth2.0 user access token.
- Add OAuth2.0 helpers for authorization URL generation, token exchange, token refresh, userinfo, and token revocation.
- Add shared HTTP client behavior for base URLs, authentication, JSON encoding/decoding, request IDs, errors, retries, and context propagation.
- Add public developer documentation that tells customers how to install the SDK, authenticate, call Gateway models, use Management APIs, handle errors, and run examples.
- Keep the SDK as an HTTP client only: it does not connect to databases, does not import dashboard/backend code, and does not bypass the OwlVigil API service.
- Add streaming support for Gateway model responses.
- Add list pagination helpers for Management APIs.
- Add secret redaction, idempotency, webhook signature verification, SDK versioning, and default User-Agent behavior.
- Define the project directory layout for a small public Go SDK library.

## Capabilities

### New Capabilities

- `sdk-core`: Shared Go SDK module, client configuration, HTTP transport, errors, responses, retries, and context-aware requests.
- `gateway-client`: Gateway model invocation client for OpenAI-compatible and provider-compatible model endpoints on `api.owlvigil.com`.
- `management-client`: Open API management client for workspace, Gateway key, model, usage, logs, policy, webhook, and documentation metadata endpoints on `open.owlvigil.com`.
- `oauth2-auth`: OAuth2.0 user authentication helpers for Dashboard and Open API management authorization.
- `public-docs`: Customer-facing documentation, quickstarts, examples, reference pages, and troubleshooting guidance for SDK and API usage.
- `streaming`: SSE and stream handling for Gateway model responses.
- `pagination`: Common list option and paginated response handling for Management APIs.
- `sdk-security`: Secret redaction, safe errors, idempotency, versioning, and User-Agent behavior.
- `webhook-signatures`: Helpers for verifying inbound OwlVigil webhook signatures.
- `repository-layout`: Project directory structure, package boundaries, examples, tests, docs, and generated-code placement.

### Modified Capabilities

- None.

## Impact

- New Go module in the `owlvigil-go` project.
- Public SDK API surface for customer applications.
- Public SDK documentation and runnable examples.
- No dependency on the OwlVigil dashboard/backend module or its internal packages.
- No direct database access from the SDK.
- Requires stable Open API and Gateway endpoint contracts from the OwlVigil service.
