package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/owlvigil/owlvigil-go/webhook"
)

func main() {
	if err := run(os.Stdout); err != nil {
		log.Fatal(err)
	}
}

func run(output io.Writer) error {
	payload := []byte(`{"id":"evt_1","type":"request.completed"}`)
	timestamp := time.Now().Unix()
	secret := "whsec_example"
	header := webhook.SignPayload(payload, timestamp, secret)

	if err := webhook.VerifySignature(payload, header, secret, webhook.VerifyOptions{}); err != nil {
		return err
	}
	_, err := fmt.Fprintln(output, "verified")
	return err
}
