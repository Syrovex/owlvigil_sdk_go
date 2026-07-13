## ADDED Requirements

### Requirement: Library-oriented root package
The SDK SHALL use a public root package for shared client options, errors, version metadata, and common types.

#### Scenario: Customer imports root package
- **WHEN** a customer imports the SDK root package
- **THEN** they MUST have access to shared options and common error types without importing internal packages.

### Requirement: Focused public subpackages
The SDK SHALL use focused public subpackages for Gateway, Management, OAuth2.0, and webhook helpers.

#### Scenario: Customer only needs Gateway calls
- **WHEN** a customer imports the Gateway package
- **THEN** they MUST NOT need to import Management or OAuth2.0 packages.

### Requirement: Private internal plumbing
Shared HTTP transport, retry, redaction, response decoding, and SSE scanning internals SHALL live under `internal/`.

#### Scenario: Customer attempts internal import
- **WHEN** external customer code tries to import SDK internal plumbing
- **THEN** Go module rules MUST prevent that import.

### Requirement: Runnable examples directory
The SDK SHALL include runnable examples under `examples/` for key customer workflows.

#### Scenario: Maintainer verifies examples
- **WHEN** maintainers run `go test ./...`
- **THEN** example code MUST compile.

### Requirement: Co-located tests
The SDK SHALL include co-located `_test.go` files for public packages and private internal packages that contain request, response, retry, redaction, streaming, OAuth2.0, webhook, and pagination logic.

#### Scenario: Maintainer inspects package tests
- **WHEN** maintainers review the repository layout
- **THEN** packages with SDK behavior MUST have adjacent `_test.go` files covering that behavior.

### Requirement: Example compile tests
The SDK SHALL include tests that compile runnable examples and protect documentation snippets from drifting.

#### Scenario: Maintainer runs examples tests
- **WHEN** maintainers run `go test ./...`
- **THEN** example code and documentation-linked examples MUST compile.

### Requirement: Customer docs directory
The SDK SHALL include a `docs/` directory with task-oriented customer documentation.

#### Scenario: Customer looks for streaming docs
- **WHEN** a customer opens `docs/streaming.md`
- **THEN** they MUST find Gateway streaming usage and cleanup guidance.

### Requirement: No service-style layout
The SDK SHALL NOT use backend service directories such as `cmd/server`, `internal/server`, or database migration directories for SDK implementation.

#### Scenario: Repository layout review
- **WHEN** maintainers review the SDK project structure
- **THEN** service-only directories MUST NOT be required for the SDK to build or test.
