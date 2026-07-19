package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	owlvigil "github.com/Syrovex/owlvigil_sdk_go"
	"github.com/Syrovex/owlvigil_sdk_go/examples/internal/envfile"
	oauth2 "github.com/Syrovex/owlvigil_sdk_go/oauth2"
)

const (
	testClientID     = ""
	testClientSecret = ""
	testScope        = "workspace:read profile:read"
	testAccessToken  = "" // Optional local testing fallback for OWLVIGIL_ACCESS_TOKEN.
	testRefreshToken = "" // Optional local testing fallback for OWLVIGIL_REFRESH_TOKEN.
)

func main() {
	if err := envfile.Load(); err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	oauthClient := oauth2.NewClient(owlvigil.WithEnvironmentFromEnv())
	accessToken := envDefault("OWLVIGIL_ACCESS_TOKEN", testAccessToken)
	refreshToken := envDefault("OWLVIGIL_REFRESH_TOKEN", testRefreshToken)
	clientID := envDefault("OWLVIGIL_CLIENT_ID", testClientID)
	clientSecret := envDefault("OWLVIGIL_CLIENT_SECRET", testClientSecret)
	scope := envDefault("OWLVIGIL_SCOPE", testScope)

	fmt.Printf("OAuth base URL: %s\n", oauthClient.BaseURL())

	if accessToken == "" {
		if clientID == "" || clientSecret == "" {
			log.Fatal("set OWLVIGIL_ACCESS_TOKEN, or set OWLVIGIL_CLIENT_ID and OWLVIGIL_CLIENT_SECRET for client_credentials")
		}
		fmt.Println("mode: client_credentials")
		token, err := oauthClient.ClientCredentials(ctx, oauth2.ClientCredentialsRequest{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			Scopes:       strings.Fields(scope),
		})
		if err != nil {
			log.Fatalf("client_credentials failed: %v", err)
		}
		accessToken = token.AccessToken
		fmt.Printf("client_credentials: ok token_type=%s expires_in=%d scope=%q access_token=%s\n",
			token.TokenType,
			token.ExpiresIn,
			token.Scope,
			maskToken(token.AccessToken),
		)
	} else {
		fmt.Println("mode: provided access token")
		fmt.Printf("using provided access token: %s\n", maskToken(accessToken))
	}

	userinfo, err := oauthClient.UserInfo(ctx, accessToken)
	if err != nil {
		log.Printf("userinfo: failed: %v", err)
	} else {
		fmt.Printf("userinfo: ok sub=%q email=%q name=%q\n", userinfo.Subject, userinfo.Email, userinfo.Name)
	}

	if refreshToken != "" {
		refreshed, err := oauthClient.Refresh(ctx, oauth2.RefreshTokenRequest{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RefreshToken: refreshToken,
		})
		if err != nil {
			log.Printf("refresh: failed: %v", err)
		} else {
			fmt.Printf("refresh: ok token_type=%s expires_in=%d access_token=%s refresh_token=%s\n",
				refreshed.TokenType,
				refreshed.ExpiresIn,
				maskToken(refreshed.AccessToken),
				maskToken(refreshed.RefreshToken),
			)
		}
	} else {
		fmt.Println("refresh: skipped; this run has no OWLVIGIL_REFRESH_TOKEN. client_credentials tokens do not include refresh tokens.")
	}

	if os.Getenv("OWLVIGIL_REVOKE_TOKEN") == "true" {
		if err := oauthClient.Revoke(ctx, accessToken); err != nil {
			log.Printf("revoke access token: failed: %v", err)
		} else {
			fmt.Println("revoke access token: ok")
		}
	} else {
		fmt.Println("revoke: skipped; set OWLVIGIL_REVOKE_TOKEN=true to revoke the access token")
	}
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
