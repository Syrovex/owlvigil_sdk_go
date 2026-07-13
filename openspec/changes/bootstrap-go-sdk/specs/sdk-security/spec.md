## ADDED Requirements

### Requirement: Secret redaction
The SDK SHALL redact Gateway keys, OAuth access tokens, refresh tokens, client secrets, and webhook secrets from error strings and debug output.

#### Scenario: API error includes submitted secret
- **WHEN** an API error payload or request context contains a credential value
- **THEN** the SDK MUST redact the credential before returning formatted error text.

### Requirement: Idempotency keys
The SDK SHALL allow callers to set an idempotency key on supported mutating requests.

#### Scenario: Caller creates Gateway key with idempotency
- **WHEN** a caller creates a Gateway key with an idempotency key option
- **THEN** the SDK MUST send the idempotency key header for that request.

### Requirement: Default User-Agent
The SDK SHALL send a default User-Agent containing the SDK name and version.

#### Scenario: Caller uses default client
- **WHEN** a caller sends any SDK request with the default configuration
- **THEN** the SDK MUST include a `User-Agent` header containing `owlvigil-go/`.

### Requirement: Version metadata
The SDK SHALL expose its version as public metadata for diagnostics.

#### Scenario: Customer reports issue
- **WHEN** a customer needs to report their SDK version
- **THEN** the SDK MUST provide an exported version constant or function.

### Requirement: Private deployment base URLs
The SDK SHALL allow callers to override Gateway, Management, and OAuth base URLs for private deployments.

#### Scenario: Customer uses private deployment
- **WHEN** a caller configures custom base URLs
- **THEN** the SDK MUST send requests to those configured URLs instead of OwlVigil cloud defaults.
