## ADDED Requirements

### Requirement: Common list options
The SDK SHALL provide common list options for Management API list endpoints.

#### Scenario: Caller sets cursor and limit
- **WHEN** a caller passes cursor and limit options to a list method
- **THEN** the SDK MUST encode those values in the Management API request.

### Requirement: Paginated list response
The SDK SHALL expose typed items and pagination metadata for list responses.

#### Scenario: API returns next cursor
- **WHEN** a list endpoint returns a next cursor
- **THEN** the SDK MUST expose that cursor to the caller.

### Requirement: Endpoint-specific filters
The SDK SHALL allow list methods to include endpoint-specific filters without losing common pagination behavior.

#### Scenario: Caller filters request logs by key
- **WHEN** a caller lists request logs with a Gateway key filter and common pagination options
- **THEN** the SDK MUST include both the filter and pagination parameters in the request.

### Requirement: Pagination docs
The SDK public docs SHALL explain how to iterate through paginated Management API results.

#### Scenario: Customer reads pagination guide
- **WHEN** a customer reads the pagination documentation
- **THEN** they MUST see an example loop that follows the next cursor until no more pages remain.
