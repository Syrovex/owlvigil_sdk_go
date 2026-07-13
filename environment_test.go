package owlvigil_test

import (
	"testing"

	owlvigil "github.com/owlvigil/owlvigil-go"
)

func TestWithEnvironment(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		baseURL     string
		environment owlvigil.Environment
		want        string
	}{
		{
			name:        "gateway production",
			baseURL:     owlvigil.DefaultGatewayBaseURL,
			environment: owlvigil.EnvironmentProduction,
			want:        "https://api.owlvigil.com",
		},
		{
			name:        "gateway staging",
			baseURL:     owlvigil.DefaultGatewayBaseURL,
			environment: owlvigil.EnvironmentStaging,
			want:        "https://staging.owlvigil.com",
		},
		{
			name:        "gateway local",
			baseURL:     owlvigil.DefaultGatewayBaseURL,
			environment: owlvigil.EnvironmentLocal,
			want:        "http://localhost:8080",
		},
		{
			name:        "management production",
			baseURL:     owlvigil.DefaultManagementBaseURL,
			environment: owlvigil.EnvironmentProduction,
			want:        "https://api.owlvigil.com/open/v1",
		},
		{
			name:        "management staging",
			baseURL:     owlvigil.DefaultManagementBaseURL,
			environment: owlvigil.EnvironmentStaging,
			want:        "https://staging.owlvigil.com/open/v1",
		},
		{
			name:        "management local",
			baseURL:     owlvigil.DefaultManagementBaseURL,
			environment: owlvigil.EnvironmentLocal,
			want:        "http://localhost:8081/open/v1",
		},
		{
			name:        "oauth production",
			baseURL:     owlvigil.DefaultOAuthBaseURL,
			environment: owlvigil.EnvironmentProduction,
			want:        "https://open.owlvigil.com",
		},
		{
			name:        "oauth staging",
			baseURL:     owlvigil.DefaultOAuthBaseURL,
			environment: owlvigil.EnvironmentStaging,
			want:        "https://openstaging.owlvigil.com",
		},
		{
			name:        "oauth local",
			baseURL:     owlvigil.DefaultOAuthBaseURL,
			environment: owlvigil.EnvironmentLocal,
			want:        "http://localhost:8081",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := owlvigil.DefaultConfig(tt.baseURL)
			opt := owlvigil.WithEnvironment(tt.environment)
			opt(&config)

			if config.BaseURL != tt.want {
				t.Errorf("WithEnvironment() BaseURL = %v, want %v", config.BaseURL, tt.want)
			}
		})
	}
}

func TestWithEnvironmentFromEnv(t *testing.T) {
	// Remove t.Parallel() because we use t.Setenv

	tests := []struct {
		name    string
		baseURL string
		envVar  string
		want    string
	}{
		{
			name:    "empty env defaults to production",
			baseURL: owlvigil.DefaultGatewayBaseURL,
			envVar:  "",
			want:    "https://api.owlvigil.com",
		},
		{
			name:    "staging env",
			baseURL: owlvigil.DefaultGatewayBaseURL,
			envVar:  "staging",
			want:    "https://staging.owlvigil.com",
		},
		{
			name:    "local env",
			baseURL: owlvigil.DefaultGatewayBaseURL,
			envVar:  "local",
			want:    "http://localhost:8080",
		},
		{
			name:    "production env explicit",
			baseURL: owlvigil.DefaultGatewayBaseURL,
			envVar:  "production",
			want:    "https://api.owlvigil.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			if tt.envVar != "" {
				t.Setenv("OWLVIGIL_ENV", tt.envVar)
			}

			config := owlvigil.DefaultConfig(tt.baseURL)
			opt := owlvigil.WithEnvironmentFromEnv()
			opt(&config)

			if config.BaseURL != tt.want {
				t.Errorf("WithEnvironmentFromEnv() BaseURL = %v, want %v", config.BaseURL, tt.want)
			}
		})
	}
}

func TestEnvironmentWithBaseURL(t *testing.T) {
	t.Parallel()

	// Test that WithBaseURL overrides WithEnvironment when called after
	config := owlvigil.DefaultConfig(owlvigil.DefaultGatewayBaseURL)

	// Apply environment first
	owlvigil.WithEnvironment(owlvigil.EnvironmentStaging)(&config)
	if config.BaseURL != "https://staging.owlvigil.com" {
		t.Errorf("After WithEnvironment, BaseURL = %v, want %v", config.BaseURL, "https://staging.owlvigil.com")
	}

	// Apply custom BaseURL (should override)
	owlvigil.WithBaseURL("https://custom.example.com")(&config)
	if config.BaseURL != "https://custom.example.com" {
		t.Errorf("After WithBaseURL, BaseURL = %v, want %v", config.BaseURL, "https://custom.example.com")
	}
}
