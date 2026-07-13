# Multi-Environment Support Example

This example demonstrates how to use different environments with the OwlVigil Go SDK.

## Usage

### 1. Production Environment (Default)

```go
client := management.NewClient(
    owlvigil.WithAPIKey(apiKey),
)
// Uses: https://api.owlvigil.com/open/v1
```

### 2. Staging Environment

```go
client := management.NewClient(
    owlvigil.WithEnvironment(owlvigil.EnvironmentStaging),
    owlvigil.WithAPIKey(stagingAPIKey),
)
// Uses: https://open-staging.owlvigil.com/open/v1
```

### 3. Local Development Environment

```go
client := management.NewClient(
    owlvigil.WithEnvironment(owlvigil.EnvironmentLocal),
    owlvigil.WithAPIKey(devAPIKey),
)
// Uses: http://localhost:8081/open/v1
```

### 4. Using Environment Variable

```bash
export OWLVIGIL_ENV=staging
export OWLVIGIL_API_KEY=your_staging_api_key
```

```go
client := management.NewClient(
    owlvigil.WithEnvironmentFromEnv(),
    owlvigil.WithAPIKey(os.Getenv("OWLVIGIL_API_KEY")),
)
```

## Running the Example

```bash
# Production (default)
go run main.go

# Staging
export OWLVIGIL_ENV=staging
export OWLVIGIL_STAGING_TOKEN=your_staging_token
go run main.go

# Local
export OWLVIGIL_ENV=local
go run main.go
```

## Environment URLs

| Environment | Gateway API | Management API | OAuth API |
|-------------|-------------|----------------|-----------|
| Production  | `api.owlvigil.com` | `api.owlvigil.com/open/v1` | `open.owlvigil.com` |
| Staging     | `apistaging.owlvigil.com` | `openstaging.owlvigil.com/open/v1` | `openstaging.owlvigil.com` |
| Local       | `localhost:8080` | `localhost:8081/open/v1` | `localhost:8081` |
