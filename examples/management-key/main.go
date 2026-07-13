package main

import (
	"context"
	"fmt"
	"log"

	owlvigil "github.com/owlvigil/owlvigil-go"
	"github.com/owlvigil/owlvigil-go/examples/internal/envfile"
	"github.com/owlvigil/owlvigil-go/management"
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

	key, _, err := client.CreateGatewayKey(
		context.Background(),
		&management.CreateGatewayKeyRequest{
			Name:   "production",
			Scopes: []string{"gateway:invoke"},
		},
		owlvigil.WithIdempotencyKey("create-production-key-001"),
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(key.ID, key.Name)
}
