## ADDED Requirements

### Requirement: Public documentation structure
The SDK project SHALL include customer-facing documentation that explains installation, authentication, Gateway usage, Management usage, OAuth2.0, error handling, examples, and troubleshooting.

#### Scenario: Customer opens docs directory
- **WHEN** a customer opens the SDK documentation
- **THEN** the docs MUST provide clear entry points for quickstart, authentication, Gateway calls, Management API calls, OAuth2.0, errors, and examples.

### Requirement: Installation quickstart
The public docs SHALL show how to install the Go SDK and initialize Gateway and Management clients.

#### Scenario: Customer installs SDK
- **WHEN** a customer follows the installation quickstart
- **THEN** they MUST see the Go module path, `go get` command, import path, and minimal client initialization examples.

### Requirement: Domain and authentication guidance
The public docs SHALL clearly distinguish Gateway and Management domains and authentication methods.

#### Scenario: Customer chooses a base URL
- **WHEN** a customer reads authentication documentation
- **THEN** they MUST understand that Gateway calls use `https://api.owlvigil.com` with Gateway keys and Management calls use `https://open.owlvigil.com/open/v1` with OAuth2.0 user access tokens.

### Requirement: Gateway calling guide
The public docs SHALL include copy-pasteable examples for Gateway model calls.

#### Scenario: Customer calls chat completions
- **WHEN** a customer follows the Gateway guide
- **THEN** they MUST be able to create a chat completion request with a Gateway key and the default Gateway client.

### Requirement: Management API guide
The public docs SHALL include examples for common Management API workflows.

#### Scenario: Customer manages Gateway keys
- **WHEN** a customer follows the Management guide
- **THEN** they MUST be able to list workspaces and create, list, or rotate Gateway keys using an OAuth2.0 access token.

### Requirement: OAuth2.0 guide
The public docs SHALL explain the OAuth2.0 authorization code flow for Management API access.

#### Scenario: Customer implements login
- **WHEN** a customer follows the OAuth2.0 guide
- **THEN** they MUST understand authorization URL generation, callback handling, token exchange, refresh, userinfo, and revocation.

### Requirement: Error and troubleshooting guide
The public docs SHALL explain SDK error types, request IDs, retries, common HTTP status codes, and troubleshooting steps.

#### Scenario: API request fails
- **WHEN** a customer receives an SDK error
- **THEN** the docs MUST show how to inspect status code, service code, message, request ID, and retry guidance.

### Requirement: Runnable examples
The public docs SHALL link to runnable examples that compile under `go test ./...`.

#### Scenario: Maintainer verifies docs examples
- **WHEN** maintainers run `go test ./...`
- **THEN** documentation-linked examples MUST compile successfully.
