package main

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	owlvigil "github.com/Syrovex/owlvigil_sdk_go"
	"github.com/Syrovex/owlvigil_sdk_go/examples/internal/envfile"
	"github.com/Syrovex/owlvigil_sdk_go/examples/internal/fullsmoke"
	"github.com/Syrovex/owlvigil_sdk_go/management"
	oauth2 "github.com/Syrovex/owlvigil_sdk_go/oauth2"
)

func main() {
	if err := envfile.Load(); err != nil {
		log.Fatal(err)
	}
	apiKey, err := envfile.Required("OWLVIGIL_API_KEY")
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	environment := smokeEnvironment(os.Getenv("OWLVIGIL_ENV"))
	managementClient := management.NewClient(
		owlvigil.WithEnvironment(environment),
		owlvigil.WithAPIKey(apiKey),
		owlvigil.WithoutRetry(),
	)
	oauthClient := oauth2.NewClient(owlvigil.WithEnvironment(environment))

	if err := fullsmoke.Run(ctx, fullsmoke.Config{
		Management: managementClient,
		OAuth:      oauthClient,
		Writes:     smokeEnabled(os.Getenv("OWLVIGIL_SMOKE_WRITES")),
		RequireAll: smokeEnabled(os.Getenv("OWLVIGIL_SMOKE_REQUIRE_ALL")),
	}); err != nil {
		log.Fatal(err)
	}
}

func smokeEnvironment(value string) owlvigil.Environment {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "staging":
		return owlvigil.EnvironmentStaging
	case "local":
		return owlvigil.EnvironmentLocal
	default:
		return owlvigil.EnvironmentProduction
	}
}

func smokeEnabled(value string) bool {
	return strings.TrimSpace(value) == "1"
}
