package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	owlvigil "github.com/Syrovex/owlvigil_sdk_go"
	"github.com/Syrovex/owlvigil_sdk_go/gateway"
)

func main() {
	if err := run(); err != nil {
		slog.Error("OwlVigil quickstart failed", "error", err)
		os.Exit(1)
	}
}

func run() error {
	key := os.Getenv("OWLVIGIL_GATEWAY_KEY")
	if key == "" {
		return errors.New("OWLVIGIL_GATEWAY_KEY is required")
	}

	client := gateway.NewClient(
		owlvigil.WithAPIKey(key),
		owlvigil.WithTimeout(30*time.Second),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 35*time.Second)
	defer cancel()

	models, meta, err := client.ListModels(ctx)
	if err != nil {
		var apiErr *owlvigil.APIError
		if errors.As(err, &apiErr) {
			return fmt.Errorf("list Gateway models: status=%d code=%s request_id=%s message=%s",
				apiErr.StatusCode, apiErr.Code, apiErr.RequestID, apiErr.Message)
		}
		return fmt.Errorf("list Gateway models: %w", err)
	}

	fmt.Printf("request_id=%s models=%d\n", meta.RequestID, len(models.Data))
	for _, model := range models.Data {
		fmt.Println(model.ID)
	}
	return nil
}
