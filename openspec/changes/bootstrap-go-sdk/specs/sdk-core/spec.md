## ADDED Requirements

### Requirement: Standalone Go module
The SDK SHALL be a standalone Go module that does not import OwlVigil dashboard or backend implementation packages.

#### Scenario: Customer installs SDK without backend dependencies
- **WHEN** a customer runs `go get` for the SDK module
- **THEN** the SDK MUST resolve without requiring the OwlVigil dashboard/backend module.

### Requirement: HTTP-only service boundary
The SDK SHALL communicate with OwlVigil only through documented HTTP APIs and MUST NOT directly access OwlVigil databases, Redis, queues, or server-side infrastructure.

#### Scenario: SDK dependency review
- **WHEN** maintainers review SDK dependencies
- **THEN** the SDK MUST NOT include database drivers, Redis clients, queue clients, or imports from OwlVigil server implementation packages.

#### Scenario: Customer performs management operation
- **WHEN** a customer uses the SDK to manage Gateway keys, usage, logs, or webhooks
- **THEN** the SDK MUST send HTTP requests to the configured Management API instead of reading or writing service storage directly.

### Requirement: Configurable HTTP client
The SDK SHALL allow callers to configure base URL, HTTP client, user agent, timeout behavior, and authentication through client options.

#### Scenario: Caller overrides HTTP client
- **WHEN** a caller constructs a client with a custom `*http.Client`
- **THEN** all SDK requests MUST use the provided HTTP client.

### Requirement: Context-aware requests
The SDK SHALL accept `context.Context` on network operations and MUST propagate cancellation and deadlines to HTTP requests.

#### Scenario: Context is canceled before request completes
- **WHEN** the provided context is canceled during an SDK request
- **THEN** the SDK MUST stop waiting for the response and return the context error.

### Requirement: Structured API errors
The SDK SHALL convert non-2xx API responses into structured errors containing status code, request ID, service error code, message, and response body excerpt.

#### Scenario: API returns validation error
- **WHEN** the API returns a non-2xx JSON error response
- **THEN** the SDK MUST return an error exposing the HTTP status, service code, message, and request ID.

### Requirement: JSON response envelope support
The SDK SHALL decode OwlVigil response envelopes and expose the typed `data` payload to callers.

#### Scenario: API returns successful envelope
- **WHEN** the API returns a JSON response with `request_id`, `code`, `message`, and `data`
- **THEN** the SDK MUST decode `data` into the caller-requested response type and preserve the request ID.

### Requirement: Retry policy
The SDK SHALL provide a bounded retry policy for transient failures and MUST allow callers to disable retries.

#### Scenario: Transient gateway failure
- **WHEN** a retryable request receives a transient 502, 503, 504, or network timeout
- **THEN** the SDK MUST retry within the configured maximum attempts and return the final result.

### Requirement: Examples compile
The SDK SHALL include examples that compile under `go test`.

#### Scenario: Running all tests
- **WHEN** maintainers run `go test ./...`
- **THEN** all examples and SDK tests MUST compile and pass.
