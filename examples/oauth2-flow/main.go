package main

import (
	"context"
	"fmt"
	"log"
	"os"

	owlvigil "github.com/owlvigil/owlvigil-go"
	"github.com/owlvigil/owlvigil-go/examples/internal/envfile"
	oauth2 "github.com/owlvigil/owlvigil-go/oauth2"
)

func main() {
	if err := envfile.Load(); err != nil {
		log.Fatal(err)
	}
	clientID, err := envfile.Required("OWLVIGIL_CLIENT_ID")
	if err != nil {
		log.Fatal(err)
	}
	client := oauth2.NewClient(owlvigil.WithEnvironmentFromEnv())

	authURL, err := client.AuthorizationURL(oauth2.AuthCodeOptions{
		ClientID:    clientID,
		RedirectURI: "https://app.example.com/callback",
		Scopes:      []string{"workspace:read", "gateway:write"},
		State:       "state_123",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(authURL)

	code := os.Getenv("OWLVIGIL_AUTH_CODE")
	if code == "" {
		return
	}
	clientSecret, err := envfile.Required("OWLVIGIL_CLIENT_SECRET")
	if err != nil {
		log.Fatal(err)
	}
	token, err := client.Exchange(context.Background(), oauth2.TokenExchangeRequest{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Code:         code,
		RedirectURI:  "https://app.example.com/callback",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("OAuth exchange succeeded: token_type=%s expires_in=%d. Open API management calls require OWLVIGIL_API_KEY.\n", token.TokenType, token.ExpiresIn)
}
