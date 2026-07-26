## ADDED Requirements

### Requirement: Gateway streaming support
The SDK SHALL support streaming Gateway responses for model endpoints that return server-sent events.

#### Scenario: Stream chat completions
- **WHEN** a caller creates a streaming chat completion request
- **THEN** the SDK MUST connect to `POST /v1/chat/completions` on `gateway.owlvigil.com` and expose decoded stream events.

### Requirement: Stream lifecycle
The SDK SHALL provide explicit stream close behavior and MUST respect context cancellation.

#### Scenario: Caller cancels stream context
- **WHEN** the caller cancels the context used for a streaming request
- **THEN** the SDK MUST stop reading the stream and return the context error.

#### Scenario: Caller closes stream
- **WHEN** the caller calls `Close()` on a stream
- **THEN** the SDK MUST close the underlying response body.

### Requirement: Stream event decoding
The SDK SHALL decode SSE event data into typed Gateway stream events and expose raw event data for unsupported event types.

#### Scenario: Unknown stream event
- **WHEN** the service returns an unsupported SSE event type
- **THEN** the SDK MUST preserve the event name and raw payload for caller inspection.

### Requirement: Stream error handling
The SDK SHALL surface stream protocol errors, HTTP errors, and final stream errors through the stream error API.

#### Scenario: Streaming endpoint returns non-2xx
- **WHEN** a streaming request receives a non-2xx response
- **THEN** the SDK MUST return a structured API error before exposing a stream.
