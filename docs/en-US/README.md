# OwlVigil Go SDK Documentation

These guides explain how to integrate the OwlVigil Go SDK. Method signatures and exported types in the current repository remain the final source of truth.

## Recommended path

New users should read chapters 1 through 4 in order. After completing the initial integration, go directly to the chapter for the task at hand.

1. [Quickstart](01-quickstart.md): install the SDK and make the first Gateway request.
2. [Core Concepts](02-concepts.md): understand clients, credentials, workspaces, request lifetimes, and return values.
3. [Authentication and Client Configuration](03-authentication-configuration.md): select the correct key and configure environments, timeouts, retries, and HTTP clients.
4. [Model Calls and Streaming](04-gateway.md): discover models and use regular or streaming model APIs.
5. [Management and Workspaces](05-management.md): learn common Management rules, workspace operations, pagination, and account tasks.
6. [Members, Access Control, and Gateway Keys](06-management-access-and-keys.md): manage people, permissions, and model-call credentials.
7. [Routing, Observability, and Billing](07-management-operations.md): configure model delivery, inspect operations, and manage financial workflows.
8. [Webhooks](08-webhooks.md): verify events and manage delivery endpoints and redelivery.
9. [Errors and Troubleshooting](09-errors-troubleshooting.md): handle failures and diagnose common symptoms.
10. [API Reference and Example Index](10-reference-examples.md): find public capabilities and runnable programs.

## Public packages

| Package | Purpose |
| --- | --- |
| `owlvigil` | Shared configuration, request options, errors, and response metadata |
| `gateway` | Models, Chat Completions, Responses, Embeddings, Anthropic-compatible calls, and streams |
| `management` | Workspaces, access control, Gateway Keys, routing, billing, logs, and Webhook management |
| `webhook` | Signature verification for inbound Webhook deliveries |

## Requirements

- Go 1.25.12 or later.
- Pin the SDK version with Go Modules.
- Review `CHANGELOG.md` before upgrading.
