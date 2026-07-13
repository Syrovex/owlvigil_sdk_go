package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	owlvigil "github.com/owlvigil/owlvigil-go"
	"github.com/owlvigil/owlvigil-go/examples/internal/envfile"
	oauth2 "github.com/owlvigil/owlvigil-go/oauth2"
)

const defaultScope = "workspace:read profile:read"

func main() {
	if err := envfile.Load(); err != nil {
		log.Fatal(err)
	}
	clientID, err := envfile.Required("OWLVIGIL_CLIENT_ID")
	if err != nil {
		log.Fatal(err)
	}
	clientSecret, err := envfile.Required("OWLVIGIL_CLIENT_SECRET")
	if err != nil {
		log.Fatal(err)
	}
	scope := envDefault("OWLVIGIL_SCOPE", defaultScope)

	client := oauth2.NewClient(owlvigil.WithEnvironmentFromEnv())
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	token, err := client.ClientCredentials(ctx, oauth2.ClientCredentialsRequest{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       strings.Fields(scope),
	})
	if err != nil {
		log.Fatalf("client credentials exchange failed: %v", err)
	}

	fmt.Printf("OAuth base URL: %s\n", client.BaseURL())
	fmt.Printf("token_type: %s\n", token.TokenType)
	fmt.Printf("expires_in: %d\n", token.ExpiresIn)
	fmt.Printf("scope: %s\n", token.Scope)
	fmt.Printf("access_token: %s\n", maskToken(token.AccessToken))
	fmt.Printf("refresh_token: %s\n", maskToken(token.RefreshToken))
}

func envDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func maskToken(token string) string {
	token = strings.TrimSpace(token)
	if token == "" {
		return ""
	}
	if len(token) <= 16 {
		return token[:4] + "..."
	}
	return token[:8] + "..." + token[len(token)-6:]
}
