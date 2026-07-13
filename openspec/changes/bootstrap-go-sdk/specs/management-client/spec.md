## ADDED Requirements

### Requirement: Management client domain
The Management client SHALL use `https://open.owlvigil.com/open/v1` as its default base URL.

#### Scenario: Default Management base URL
- **WHEN** a caller constructs a Management client without overriding the base URL
- **THEN** the client MUST target `https://open.owlvigil.com/open/v1`.

### Requirement: OAuth2.0 access token authentication
The Management client SHALL authenticate Open API management requests with `Authorization: Bearer <user_access_token>`.

#### Scenario: Management request includes user access token
- **WHEN** a caller invokes a Management endpoint with a configured access token
- **THEN** the SDK MUST send the token in the `Authorization` header.

### Requirement: Workspace management
The Management client SHALL provide methods for listing accessible workspaces and retrieving workspace details.

#### Scenario: List workspaces
- **WHEN** a caller lists workspaces
- **THEN** the SDK MUST call `GET /workspaces` on the Management base URL.

### Requirement: Gateway key management
The Management client SHALL provide methods for creating, listing, retrieving, updating, rotating, enabling, disabling, and deleting Gateway keys.

#### Scenario: Create Gateway key
- **WHEN** a caller creates a Gateway key
- **THEN** the SDK MUST call `POST /gateway/keys` and return the one-time visible key secret when provided by the API.

### Requirement: Usage and quota access
The Management client SHALL provide methods for retrieving Gateway usage, usage summary, quota, balance, and invoice data.

#### Scenario: Get usage summary
- **WHEN** a caller requests usage summary
- **THEN** the SDK MUST call `GET /gateway/usage/summary` and return typed usage totals.

### Requirement: Logs and traces access
The Management client SHALL provide methods for request logs, trace lookup, and payload log access workflows.

#### Scenario: Get request log detail
- **WHEN** a caller requests a request log by ID
- **THEN** the SDK MUST call `GET /gateway/request-logs/{request_id}`.

### Requirement: Webhook management
The Management client SHALL provide methods for webhook endpoint and webhook event workflows.

#### Scenario: Test webhook endpoint
- **WHEN** a caller tests a webhook endpoint
- **THEN** the SDK MUST call `POST /webhook-endpoints/{endpoint_id}/test`.
