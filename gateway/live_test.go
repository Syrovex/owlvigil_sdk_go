package gateway_test

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	owlvigil "github.com/owlvigil/owlvigil-go"
	"github.com/owlvigil/owlvigil-go/gateway"
)

func TestLiveGatewayReadOnly(t *testing.T) {
	if os.Getenv("OWLVIGIL_LIVE_TEST") != "1" {
		t.Skip("set OWLVIGIL_LIVE_TEST=1 with OWLVIGIL_GATEWAY_KEY to run against the live Gateway API")
	}
	apiKey := strings.TrimSpace(os.Getenv("OWLVIGIL_GATEWAY_KEY"))
	if apiKey == "" {
		t.Skip("OWLVIGIL_GATEWAY_KEY is not configured")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)
	client := gateway.NewClient(
		owlvigil.WithEnvironmentFromEnv(),
		owlvigil.WithAPIKey(apiKey),
		owlvigil.WithoutRetry(),
	)
	if os.Getenv("OWLVIGIL_ENV") == "staging" && client.BaseURL() != "https://staging.owlvigil.com" {
		t.Fatalf("Gateway staging BaseURL = %q, want %q", client.BaseURL(), "https://staging.owlvigil.com")
	}
	models, _, err := client.ListModels(ctx)
	if err != nil {
		t.Fatalf("ListModels() error = %v", err)
	}
	if len(models.Data) == 0 {
		t.Fatal("ListModels() returned no available models")
	}
}
