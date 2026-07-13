## ADDED Requirements

### Requirement: OAuth2.0 domain
OAuth2.0 helper requests SHALL use `https://open.owlvigil.com` as the default authorization server domain.

#### Scenario: Default authorization URL
- **WHEN** a caller builds an authorization URL without overriding the issuer base URL
- **THEN** the generated URL MUST use `https://open.owlvigil.com`.

### Requirement: Authorization URL generation
The SDK SHALL generate OAuth2.0 authorization URLs with client ID, redirect URI, response type, scopes, state, and optional PKCE parameters.

#### Scenario: Build authorization URL
- **WHEN** a caller provides client ID, redirect URI, scopes, and state
- **THEN** the SDK MUST return a URL targeting `/oauth/authorize` with those parameters encoded.

### Requirement: Token exchange
The SDK SHALL exchange authorization codes for access and refresh tokens.

#### Scenario: Exchange authorization code
- **WHEN** a caller exchanges an authorization code
- **THEN** the SDK MUST call `POST /oauth/token` and decode access token, refresh token, expiry, token type, and scopes.

### Requirement: Token refresh
The SDK SHALL refresh access tokens with refresh tokens.

#### Scenario: Refresh access token
- **WHEN** a caller provides a refresh token
- **THEN** the SDK MUST call `POST /oauth/token/refresh` and return a new token response.

### Requirement: Userinfo
The SDK SHALL retrieve OAuth2.0 userinfo with a user access token.

#### Scenario: Get current user info
- **WHEN** a caller requests userinfo with an access token
- **THEN** the SDK MUST call `GET /oauth/userinfo` and return user identity fields.

### Requirement: Token revocation
The SDK SHALL revoke access tokens or refresh tokens.

#### Scenario: Revoke refresh token
- **WHEN** a caller revokes a refresh token
- **THEN** the SDK MUST call `POST /oauth/revoke` with the provided token.

### Requirement: OAuth2.0 errors
The SDK SHALL return structured OAuth2.0 errors for authorization and token endpoint failures.

#### Scenario: Token exchange fails
- **WHEN** the authorization server returns an OAuth2.0 error response
- **THEN** the SDK MUST expose the OAuth error code, description, URI, and HTTP status.
