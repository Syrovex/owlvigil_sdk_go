## ADDED Requirements

### Requirement: Gateway client domain
The Gateway client SHALL use `https://gateway.owlvigil.com` as its default domain and MUST NOT use `open.owlvigil.com` for model invocation.

#### Scenario: Default Gateway base URL
- **WHEN** a caller constructs a Gateway client without overriding the base URL
- **THEN** the client MUST target `https://gateway.owlvigil.com`.

### Requirement: Gateway key authentication
The Gateway client SHALL authenticate model requests with `Authorization: Bearer <gateway_key>`.

#### Scenario: Chat completion request includes Gateway key
- **WHEN** a caller invokes a Gateway model endpoint with a configured Gateway key
- **THEN** the SDK MUST send the key in the `Authorization` header.

### Requirement: OpenAI-compatible chat completions
The Gateway client SHALL expose a method for `POST /v1/chat/completions`.

#### Scenario: Create chat completion
- **WHEN** a caller creates a chat completion request
- **THEN** the SDK MUST POST the request body to `/v1/chat/completions` on `gateway.owlvigil.com`.

### Requirement: OpenAI-compatible responses
The Gateway client SHALL expose a method for `POST /v1/responses`.

#### Scenario: Create response
- **WHEN** a caller creates a response request
- **THEN** the SDK MUST POST the request body to `/v1/responses` on `gateway.owlvigil.com`.

### Requirement: Model listing
The Gateway client SHALL expose methods for listing models and retrieving a model by ID.

#### Scenario: List models
- **WHEN** a caller requests available models
- **THEN** the SDK MUST call `GET /v1/models` and return the decoded model list.

### Requirement: Embeddings
The Gateway client SHALL expose a method for `POST /v1/embeddings`.

#### Scenario: Create embeddings
- **WHEN** a caller creates an embeddings request
- **THEN** the SDK MUST POST the request body to `/v1/embeddings` on `gateway.owlvigil.com`.

### Requirement: Provider-compatible endpoints
The Gateway client SHALL support provider-compatible endpoint groups that use the Gateway domain, including Anthropic-compatible messages.

#### Scenario: Anthropic-compatible message
- **WHEN** a caller invokes an Anthropic-compatible message request
- **THEN** the SDK MUST POST to `/anthropic/v1/messages` on `gateway.owlvigil.com`.
