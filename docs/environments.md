# Multi-Environment Support

The SDK supports multiple environments for development, testing, and production.

## Available Environments

### Production (Default)
```go
client := management.NewClient(
    owlvigil.WithAPIKey(apiKey),
)
```
- Gateway: `https://api.owlvigil.com`
- Management: `https://api.owlvigil.com/open/v1`
- OAuth: `https://open.owlvigil.com`

### Staging
```go
client := management.NewClient(
    owlvigil.WithEnvironment(owlvigil.EnvironmentStaging),
    owlvigil.WithAPIKey(stagingAPIKey),
)
```
- Gateway: `https://apistaging.owlvigil.com`
- Management: `https://openstaging.owlvigil.com/open/v1`
- OAuth: `https://openstaging.owlvigil.com`

### Local Development
```go
client := management.NewClient(
    owlvigil.WithEnvironment(owlvigil.EnvironmentLocal),
    owlvigil.WithAPIKey(devAPIKey),
)
```
- Gateway: `http://localhost:8080`
- Management: `http://localhost:8081/open/v1`
- OAuth: `http://localhost:8081`

## Using Environment Variables

Set the `OWLVIGIL_ENV` environment variable:

```bash
# Staging
export OWLVIGIL_ENV=staging
export OWLVIGIL_API_KEY=staging_api_key_xxx

# Local
export OWLVIGIL_ENV=local
export OWLVIGIL_API_KEY=dev_api_key_xxx
```

Then use `WithEnvironmentFromEnv()`:

```go
client := management.NewClient(
    owlvigil.WithEnvironmentFromEnv(),
    owlvigil.WithAPIKey(os.Getenv("OWLVIGIL_API_KEY")),
)
```

## Custom URLs

You can still use custom URLs with `WithBaseURL()`:

```go
client := management.NewClient(
    owlvigil.WithBaseURL("https://custom.example.com/v1"),
    owlvigil.WithAPIKey(apiKey),
)
```

**Note**: If you use both `WithEnvironment()` and `WithBaseURL()`, call `WithEnvironment()` first, otherwise `WithBaseURL()` will override the environment setting.

## Best Practices

### 1. Separate API Keys per Environment

```bash
# Production
OWLVIGIL_PROD_API_KEY=prod_api_key_xxx

# Staging
OWLVIGIL_STAGING_API_KEY=staging_api_key_xxx

# Local
OWLVIGIL_LOCAL_API_KEY=local_api_key_xxx
```

### 2. CI/CD Configuration

```yaml
# .github/workflows/test.yml
env:
  OWLVIGIL_ENV: staging
  OWLVIGIL_API_KEY: ${{ secrets.STAGING_API_KEY }}
```

### 3. Environment-Specific Clients

```go
func NewClientForEnv(env string) *management.Client {
    var apiKey string
    var environment owlvigil.Environment

    switch env {
    case "production":
        apiKey = os.Getenv("OWLVIGIL_PROD_API_KEY")
        environment = owlvigil.EnvironmentProduction
    case "staging":
        apiKey = os.Getenv("OWLVIGIL_STAGING_API_KEY")
        environment = owlvigil.EnvironmentStaging
    case "local":
        apiKey = os.Getenv("OWLVIGIL_LOCAL_API_KEY")
        environment = owlvigil.EnvironmentLocal
    default:
        apiKey = os.Getenv("OWLVIGIL_API_KEY")
        environment = owlvigil.EnvironmentProduction
    }

    return management.NewClient(
        owlvigil.WithEnvironment(environment),
        owlvigil.WithAPIKey(apiKey),
    )
}
```

## Examples

See [examples/multi-environment](../examples/multi-environment/) for complete examples.
