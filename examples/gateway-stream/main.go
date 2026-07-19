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

	stream, err := client.CreateChatCompletionStream(context.Background(), &gateway.ChatCompletionRequest{
		Model: "gpt-4o-mini",
		Messages: []gateway.Message{
			{Role: "user", Content: "Count to three."},
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	defer stream.Close()

	for stream.Next() {
		event := stream.Current()
		fmt.Println(event.Event, string(event.Data))
	}
	if err := stream.Err(); err != nil {
		log.Fatal(err)
	}
}
