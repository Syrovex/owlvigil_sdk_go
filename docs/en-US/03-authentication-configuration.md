# Authentication and Client Configuration

## Authentication

OwlVigil has two public calling surfaces:

| Surface | Default URL | Auth |
| --- | --- | --- |
| Gateway model calls | `https://gateway.owlvigil.com` | `Authorization: Bearer ov_sk_xxx` |
| OpenAPI Management | `https://api.owlvigil.com/v1` | `Authorization: Bearer <OWLVIGIL_API_KEY>` |

Gateway keys are for model invocation. Management operations such as listing workspaces, creating Gateway keys, reading usage, and managing webhooks require a scoped service-account API key.

Static tokens:

<!-- evidence: option_test.go, environment_test.go -->
```go
gatewayClient := gateway.NewClient(owlvigil.WithAPIKey(os.Getenv("OWLVIGIL_GATEWAY_KEY")))
managementClient := management.NewClient(owlvigil.WithAPIKey(os.Getenv("OWLVIGIL_API_KEY")))

```

The SDK sets `User-Agent: owlvigil-go/<version>` by default.

### Select the right credential

Use a Gateway key only with `gateway.NewClient` and a scoped service-account
API key with `management.NewClient`. A `403` normally means a
valid credential lacks the required workspace or scope, while `401` usually
means the credential is absent, disabled, expired, or sent to the wrong API
surface.

### Rotation and storage

Store only variable names and safe placeholders in `.env.example`; `.env` is
local and ignored by Git. Gateway-key creation and rotation can expose a
one-time secret: write it directly to a secret store, update consumers, then
disable the old key. Never send Management API keys, Gateway keys, upstream
provider credentials, or webhook signing secrets to logs, analytics, issue trackers, or browser
clients.

## Multi-Environment Support

The SDK supports multiple environments for development, testing, and production.

### Available Environments

#### Production (Default)
<!-- evidence: option_test.go, environment_test.go -->
```go
client := management.NewClient(
	owlvigil.WithAPIKey(apiKey),
)

```
- Gateway: `https://gateway.owlvigil.com`
- Management: `https://api.owlvigil.com/v1`

#### Staging
<!-- evidence: option_test.go, environment_test.go -->
```go
client := management.NewClient(
	owlvigil.WithEnvironment(owlvigil.EnvironmentStaging),
	owlvigil.WithAPIKey(stagingAPIKey),
)

```
- Gateway: `https://staginggateway.owlvigil.com`
- Management: `https://stagingapi.owlvigil.com/v1`

#### Local Development
<!-- evidence: option_test.go, environment_test.go -->
```go
client := management.NewClient(
	owlvigil.WithEnvironment(owlvigil.EnvironmentLocal),
	owlvigil.WithAPIKey(devAPIKey),
)

```
- Gateway: `http://localhost:8080`
- Management: `http://localhost:8081/v1`

### Using Environment Variables

Set the `OWLVIGIL_ENV` environment variable:

<!-- evidence: environment_test.go, examples/multi-environment/main.go -->
```bash
# Staging
export OWLVIGIL_ENV=staging
export OWLVIGIL_API_KEY=staging_api_key_xxx

# Local
export OWLVIGIL_ENV=local
export OWLVIGIL_API_KEY=dev_api_key_xxx
```

Then use `WithEnvironmentFromEnv()`:

<!-- evidence: option_test.go, environment_test.go -->
```go
client := management.NewClient(
	owlvigil.WithEnvironmentFromEnv(),
	owlvigil.WithAPIKey(os.Getenv("OWLVIGIL_API_KEY")),
)

```

### Custom URLs

You can still use custom URLs with `WithBaseURL()`:

<!-- evidence: option_test.go, environment_test.go -->
```go
client := management.NewClient(
	owlvigil.WithBaseURL("https://custom.example.com/v1"),
	owlvigil.WithAPIKey(apiKey),
)

```

**Note**: If you use both `WithEnvironment()` and `WithBaseURL()`, call `WithEnvironment()` first, otherwise `WithBaseURL()` will override the environment setting.

### Best Practices

#### 1. Separate API Keys per Environment

<!-- evidence: environment_test.go, examples/multi-environment/main.go -->
```bash
# Production
OWLVIGIL_PROD_API_KEY=prod_api_key_xxx

# Staging
OWLVIGIL_STAGING_API_KEY=staging_api_key_xxx

# Local
OWLVIGIL_LOCAL_API_KEY=local_api_key_xxx
```

#### 2. CI/CD Configuration

<!-- evidence: .github/workflows/test.yml -->
```yaml
# .github/workflows/test.yml
env:
  OWLVIGIL_ENV: staging
  OWLVIGIL_API_KEY: ${{ secrets.STAGING_API_KEY }}
```

#### 3. Environment-Specific Clients

<!-- evidence: option_test.go, environment_test.go -->
```go
func NewManagementClient(env owlvigil.Environment) (*management.Client, error) {
	var apiKey string

	switch env {
	case owlvigil.EnvironmentProduction:
		apiKey = os.Getenv("OWLVIGIL_PROD_API_KEY")
	case owlvigil.EnvironmentStaging:
		apiKey = os.Getenv("OWLVIGIL_STAGING_API_KEY")
	case owlvigil.EnvironmentLocal:
		apiKey = os.Getenv("OWLVIGIL_LOCAL_API_KEY")
	default:
		return nil, fmt.Errorf("unsupported OwlVigil environment %q", env)
	}
	if apiKey == "" {
		return nil, fmt.Errorf("OwlVigil API key is required for %q", env)
	}

	return management.NewClient(
		owlvigil.WithEnvironment(env),
		owlvigil.WithAPIKey(apiKey),
	), nil
}
```

### Examples

Environment selection transforms the built-in OwlVigil URLs only. If you use a
private proxy or a `httptest` server, pass `WithBaseURL` explicitly. When both
are needed, apply `WithEnvironment` before `WithBaseURL`, because the latter is
the final override.

The SDK reads `OWLVIGIL_ENV` only when `WithEnvironmentFromEnv()` is supplied.
It does not implicitly read credentials from `.env`; runnable examples load the
file for convenience and preserve already-exported shell values. Keep
production, staging, and local credentials separate so a local test cannot
write to a production workspace.

See [examples/multi-environment](../../examples/multi-environment/) for complete examples.

Call `BaseURL` when you need to inspect the effective client base URL; do not depend on unexported configuration fields.
