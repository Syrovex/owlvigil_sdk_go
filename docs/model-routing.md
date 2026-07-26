# Model Routing and Providers

This guide covers the Management configuration that determines which models are
available and how Gateway traffic is routed. It requires `OWLVIGIL_API_KEY`.
Gateway invocation itself still uses `OWLVIGIL_GATEWAY_KEY`; see
[Gateway](gateway.md).

## Discover models and routes

Use `ListModels` and `GetModel` to present available models. `ListRoutes` and
`GetRoute` inspect configured routing rules. All four are read-only and list
methods use `management.ListOptions`. All four require
`owlvigil.WithWorkspaceID(workspaceID)`. Use `ListRoutesWithFilters` or
`GetRouteWithFilters` when selecting a key or model explicitly.

Before changing a route, use `PreviewRoute` with a
`management.PreviewRouteRequest`. A preview evaluates a proposed route without
making it active, which lets an operator inspect the selected provider and
model before traffic changes.

## Providers

Providers are workspace-scoped upstream credentials. Use `ListProviders`,
`GetProvider`, `CreateProvider`, `UpdateProvider`, and `DeleteProvider`.

```go
provider, _, err := client.CreateProvider(ctx, &management.CreateProviderRequest{
    WorkspaceID:  workspaceID,
    Name:         "primary-openai",
    Type:         "openai",
    APIKey:       os.Getenv("UPSTREAM_PROVIDER_API_KEY"),
    DefaultModel: "gpt-4o-mini",
}, owlvigil.WithIdempotencyKey("provider-primary-openai"))
if err != nil {
    return err
}
_ = provider
```

Provider `APIKey` is a secret. Read it only from a secret store or process
environment, never log the request, and never put it in source control. Create
and update calls are mutations; use a stable idempotency key for retries.
Deleting a provider can make routes unusable, so preview and migrate routes
first.

## Gateway keys and policies

`ListGatewayKeys`, `GetGatewayKey`, `CreateGatewayKey`, `UpdateGatewayKey`,
`RotateGatewayKey`, `EnableGatewayKey`, `DisableGatewayKey`, and
`DeleteGatewayKey` manage credentials used by callers of the Gateway API.
Except for create, these routes require
`owlvigil.WithWorkspaceID(workspaceID)`; list also requires it.
The key `Secret` is returned only when the service makes it visible. Store it
once, then discard it from application memory and logs.
The `WithResult` action variants return the updated key or delete
confirmation while the original signatures retain compatibility.

`GetGatewayPolicies` retrieves policy configuration for a key.
`PreviewPolicyEffect` evaluates a proposed policy effect, and
`UpdateGatewayPolicy` changes selected fields. Preview before every update and
keep the returned request ID with the change record. Reads and updates require
`owlvigil.WithWorkspaceID(workspaceID)`; previews carry `workspace_id` in the
request body.

## Safe rollout sequence

1. Discover a model and current route with `ListModels` and `ListRoutes`.
2. Create or update the provider with secret-store data.
3. Call `PreviewRoute` or `PreviewPolicyEffect`.
4. Apply the approved write with an idempotency key.
5. Verify the result through a read-only `GetProvider`, `GetRoute`, or
   `GetGatewayPolicies` call.
6. Rotate or disable old Gateway keys only after dependent callers use the
   replacement.
