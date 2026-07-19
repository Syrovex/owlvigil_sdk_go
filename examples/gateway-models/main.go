package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	owlvigil "github.com/Syrovex/owlvigil_sdk_go"
	"github.com/Syrovex/owlvigil_sdk_go/examples/internal/envfile"
	"github.com/Syrovex/owlvigil_sdk_go/gateway"
)

func main() {
	if err := envfile.Load(); err != nil {
		log.Fatal(err)
	}
	client := gateway.NewClient(owlvigil.WithAPIKey(os.Getenv("OWLVIGIL_GATEWAY_KEY")))
	if err := run(context.Background(), client, os.Stdout); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context, client *gateway.Client, output io.Writer) error {
	models, _, err := client.ListModels(ctx)
	if err != nil {
		return err
	}
	for _, model := range models.Data {
		if _, err := fmt.Fprintln(output, model.ID); err != nil {
			return err
		}
	}
	return nil
}
