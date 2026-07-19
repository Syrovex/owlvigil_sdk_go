package main

import (
	"context"
	"fmt"
	"log"
	"os"

	owlvigil "github.com/Syrovex/owlvigil_sdk_go"
	"github.com/Syrovex/owlvigil_sdk_go/examples/internal/envfile"
	"github.com/Syrovex/owlvigil_sdk_go/management"
)

func main() {
	if err := envfile.Load(); err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()

	// Example 1: Production environment (default)
	fmt.Println("=== Production Environment (Default) ===")
	prodClient := management.NewClient(
		owlvigil.WithAPIKey(os.Getenv("OWLVIGIL_API_KEY")),
	)
	fmt.Printf("Base URL: %s\n\n", prodClient.BaseURL())

	// Example 2: Staging environment
	fmt.Println("=== Staging Environment ===")
	stagingClient := management.NewClient(
		owlvigil.WithEnvironment(owlvigil.EnvironmentStaging),
		owlvigil.WithAPIKey(os.Getenv("OWLVIGIL_STAGING_API_KEY")),
	)
	fmt.Printf("Base URL: %s\n\n", stagingClient.BaseURL())

	// Example 3: Local development environment
	fmt.Println("=== Local Development Environment ===")
	localClient := management.NewClient(
		owlvigil.WithEnvironment(owlvigil.EnvironmentLocal),
		owlvigil.WithAPIKey(os.Getenv("OWLVIGIL_LOCAL_API_KEY")),
	)
	fmt.Printf("Base URL: %s\n\n", localClient.BaseURL())

	// Example 4: Using environment variable
	fmt.Println("=== Environment from OWLVIGIL_ENV ===")
	// Set: export OWLVIGIL_ENV=staging
	envClient := management.NewClient(
		owlvigil.WithEnvironmentFromEnv(),
		owlvigil.WithAPIKey(os.Getenv("OWLVIGIL_API_KEY")),
	)
	fmt.Printf("Base URL: %s\n", envClient.BaseURL())
	fmt.Printf("OWLVIGIL_ENV: %s\n\n", os.Getenv("OWLVIGIL_ENV"))

	// Example 5: Test API call in staging (if an API key is available)
	if stagingAPIKey := os.Getenv("OWLVIGIL_STAGING_API_KEY"); stagingAPIKey != "" {
		fmt.Println("=== Testing Staging API Call ===")
		workspaces, _, err := stagingClient.ListWorkspaces(ctx, management.ListOptions{Limit: 5})
		if err != nil {
			log.Printf("Staging API call failed: %v\n", err)
		} else {
			fmt.Printf("Found %d workspaces in staging environment\n", len(workspaces.Items))
		}
	}
}
