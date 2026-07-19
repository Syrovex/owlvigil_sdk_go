package main

import (
	"context"
	"fmt"
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

	resp, meta, err := client.CreateChatCompletion(context.Background(), &gateway.ChatCompletionRequest{
		Model: "gpt-4o-mini",
		Messages: []gateway.Message{
			{Role: "user", Content: "Say hello from OwlVigil."},
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(meta.RequestID, resp.ID)
}
