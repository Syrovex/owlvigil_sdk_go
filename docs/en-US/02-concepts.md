# Core Concepts

## SDK and API responsibilities

The SDK is a Go client for OwlVigil HTTP APIs. It builds authenticated requests, applies the configured timeout and limited retry policy, and decodes responses. OwlVigil services evaluate permissions, select configured model routes, and return billing and resource state.

Do not infer server decisions from SDK types or examples. Treat model availability, route state, billing results, resource status, and event types returned by the API as authoritative.

## Public clients

Use `gateway.Client` for model calls and authenticate it with a Gateway Key. Use `management.Client` for workspace and control-plane operations and authenticate it with a Management API Key. The credentials are not interchangeable.

The `webhook` package does not create an API client. It verifies signatures on events delivered to your HTTP service.

## Client options and request options

An `owlvigil.Option` configures a client for its full lifetime. Examples include the API key, base URL, HTTP client, timeout, and retry policy.

An `owlvigil.RequestOption` affects one request. It is used for values such as a required workspace query parameter or an idempotency key. Do not place a request-scoped value in global client configuration.

## Context and request lifetime

Every network method accepts `context.Context`. Use it to cancel work when the caller disconnects and to apply a deadline for one business operation. Do not store and reuse an already canceled context.

For non-streaming calls, the earlier of the context deadline and HTTP client timeout ends the request. Streaming calls must be ended through context cancellation or `Stream.Close`.

## Resources, metadata, and errors

Most Gateway and Management methods return a resource, `*owlvigil.ResponseMeta`, and an error. Check the error before using the resource. Successful calls may include a Request ID in the metadata; API failures expose diagnostic fields through `*owlvigil.APIError`.

Resource IDs and Request IDs are generally safe for diagnosis. Management API Keys, Gateway Keys, upstream-provider credentials, and Webhook signing secrets are sensitive and must not be written to ordinary logs.

## Workspaces and pagination

Most Management resources belong to a workspace. Depending on the method, the workspace ID may be part of the path, request body, or a request option. Follow the method signature instead of manually constructing URLs.

List cursors are opaque. Pass the returned cursor back unchanged and stop when the response reports no more pages or does not return a next cursor.

## Retries and idempotency

Read operations are usually suitable for bounded retries. The current service accepts an idempotency key only when creating a Gateway Key or a Webhook endpoint. Reuse the same key and request body when retrying either operation.

Other writes—including rotation, payment, invitation, and redelivery operations—must not be retried automatically after an uncertain timeout. Read the affected resource or use the Request ID to determine the result first.

Continue with [Authentication and Client Configuration](03-authentication-configuration.md), then choose the Gateway or Management guide for your task.
