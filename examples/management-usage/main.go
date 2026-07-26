package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"

	owlvigil "github.com/Syrovex/owlvigil_sdk_go"
	"github.com/Syrovex/owlvigil_sdk_go/examples/internal/envfile"
	"github.com/Syrovex/owlvigil_sdk_go/management"
)

func main() {
	if err := envfile.Load(); err != nil {
		log.Fatal(err)
	}
	apiKey, err := envfile.Required("OWLVIGIL_API_KEY")
	if err != nil {
		log.Fatal(err)
	}

	client := management.NewClient(owlvigil.WithAPIKey(apiKey))
	if err := run(context.Background(), client, os.Getenv("OWLVIGIL_WORKSPACE_ID"), os.Stdout); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context, client *management.Client, configuredWorkspaceID string, output io.Writer) error {
	workspaces, _, err := client.ListWorkspaces(ctx, management.ListOptions{Limit: 1})
	if err != nil {
		return err
	}
	if len(workspaces.Items) == 0 {
		return fmt.Errorf("no workspaces available to the API key")
	}
	workspaceID := workspaces.Items[0].ID
	if configuredWorkspaceID != "" {
		workspaceID, err = strconv.ParseInt(configuredWorkspaceID, 10, 64)
		if err != nil {
			return fmt.Errorf("parse OWLVIGIL_WORKSPACE_ID: %w", err)
		}
	}

	summary, _, err := client.GetUsageSummary(ctx, owlvigil.WithWorkspaceID(workspaceID))
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(output, "workspace_id=%d requests=%d tokens=%d cost=%v\n", workspaceID, summary.Requests, summary.Tokens, summary.Cost)
	return err
}
