## ADDED Requirements

### Requirement: Webhook signature verification
The SDK SHALL provide a helper for verifying inbound OwlVigil webhook signatures.

#### Scenario: Valid webhook signature
- **WHEN** a caller passes a webhook payload, timestamp, signature header, and webhook secret
- **THEN** the SDK MUST confirm the signature is valid.

### Requirement: Webhook timestamp tolerance
The SDK SHALL reject webhook signatures whose timestamp is outside the configured tolerance.

#### Scenario: Stale webhook timestamp
- **WHEN** a webhook timestamp is older than the configured tolerance
- **THEN** the SDK MUST reject the signature.

### Requirement: Constant-time comparison
The SDK SHALL compare webhook signatures using constant-time comparison.

#### Scenario: Signature mismatch
- **WHEN** a webhook signature does not match the expected signature
- **THEN** the SDK MUST reject it without leaking partial match information.

### Requirement: Webhook docs
The public docs SHALL show how to verify webhook signatures before processing webhook events.

#### Scenario: Customer implements webhook receiver
- **WHEN** a customer follows the webhook documentation
- **THEN** they MUST see a runnable example that verifies the signature before decoding the event.
