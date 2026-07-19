package main

import (
	"context"
	"fmt"
	"html"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	owlvigil "github.com/Syrovex/owlvigil_sdk_go"
	"github.com/Syrovex/owlvigil_sdk_go/examples/internal/envfile"
	oauth2 "github.com/Syrovex/owlvigil_sdk_go/oauth2"
)

func main() {
	if err := envfile.Load(); err != nil {
		log.Fatal(err)
	}
	addr := envDefault("OAUTH_CALLBACK_ADDR", "localhost:3000")
	path := envDefault("OAUTH_CALLBACK_PATH", "/oauth/callback")
	redirectURI := envDefault("OWLVIGIL_REDIRECT_URI", "http://localhost:3000/oauth/callback")
	clientID, err := envfile.Required("OWLVIGIL_CLIENT_ID")
	if err != nil {
		log.Fatal(err)
	}
	clientSecret := os.Getenv("OWLVIGIL_CLIENT_SECRET")
	codeVerifier := os.Getenv("OWLVIGIL_CODE_VERIFIER")
	scope := envDefault("OWLVIGIL_SCOPE", "workspace:read profile:read")
	state := envDefault("OWLVIGIL_STATE", "test_state")
	printTokens := os.Getenv("OWLVIGIL_PRINT_TOKENS") == "true"
	if clientID == "" {
		log.Fatal("set OWLVIGIL_CLIENT_ID before starting this server")
	}
	oauthClient := oauth2.NewClient(owlvigil.WithEnvironmentFromEnv())
	authURL, err := oauthClient.AuthorizationURL(oauth2.AuthCodeOptions{
		ClientID:    clientID,
		RedirectURI: redirectURI,
		Scopes:      strings.Fields(scope),
		State:       state,
	})
	if err != nil {
		log.Fatalf("build authorize URL: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		state := r.URL.Query().Get("state")
		oauthErr := r.URL.Query().Get("error")
		errorDescription := r.URL.Query().Get("error_description")
		tokenResult := exchangeResult{Skipped: true, Message: "Set OWLVIGIL_CLIENT_ID and OWLVIGIL_CLIENT_SECRET before starting this server to auto-exchange codes."}

		log.Printf("OAuth callback received: code=%q state=%q error=%q", code, state, oauthErr)
		if code != "" && oauthErr == "" && clientID != "" && clientSecret != "" {
			ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
			defer cancel()
			token, err := oauthClient.Exchange(ctx, oauth2.TokenExchangeRequest{
				ClientID:     clientID,
				ClientSecret: clientSecret,
				Code:         code,
				RedirectURI:  redirectURI,
				CodeVerifier: codeVerifier,
			})
			if err != nil {
				tokenResult = exchangeResult{Message: err.Error()}
				log.Printf("OAuth token exchange failed: %v", err)
			} else if token == nil {
				tokenResult = exchangeResult{Message: "token endpoint returned an empty response"}
			} else {
				tokenResult = exchangeResult{
					OK:           true,
					AccessToken:  maskToken(token.AccessToken),
					RefreshToken: maskToken(token.RefreshToken),
					TokenType:    token.TokenType,
					ExpiresIn:    token.ExpiresIn,
					Scope:        token.Scope,
				}
				log.Printf("OAuth token exchange succeeded: token_type=%q expires_in=%d scope=%q access_token=%s refresh_token=%s",
					token.TokenType,
					token.ExpiresIn,
					token.Scope,
					maskToken(token.AccessToken),
					maskToken(token.RefreshToken),
				)
				if printTokens {
					log.Printf("OWLVIGIL_ACCESS_TOKEN=%s", token.AccessToken)
					log.Printf("OWLVIGIL_REFRESH_TOKEN=%s", token.RefreshToken)
				}
			}
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if oauthErr != "" || (!tokenResult.OK && !tokenResult.Skipped) {
			w.WriteHeader(http.StatusBadRequest)
		}
		_, _ = fmt.Fprintf(w, `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>OwlVigil OAuth Callback</title>
  <style>
    body { margin: 0; font-family: system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; background: #f7f7f8; color: #171717; }
    main { max-width: 760px; margin: 8vh auto; padding: 32px; background: white; border: 1px solid #e5e5e5; border-radius: 8px; }
    h1 { margin: 0 0 20px; font-size: 24px; }
    dl { display: grid; grid-template-columns: 140px 1fr; gap: 12px 16px; }
    dt { font-weight: 650; color: #525252; }
    dd { margin: 0; min-width: 0; }
    code { display: block; padding: 12px; background: #f1f5f9; border-radius: 6px; overflow-wrap: anywhere; }
    .ok { color: #047857; }
    .error { color: #b91c1c; }
    section { margin-top: 28px; }
  </style>
</head>
<body>
  <main>
    <h1 class="%s">%s</h1>
    <dl>
      <dt>code</dt><dd><code>%s</code></dd>
      <dt>state</dt><dd><code>%s</code></dd>
      <dt>error</dt><dd><code>%s</code></dd>
      <dt>description</dt><dd><code>%s</code></dd>
    </dl>
    <section>
      <h1 class="%s">%s</h1>
      <dl>
        <dt>access_token</dt><dd><code>%s</code></dd>
        <dt>refresh_token</dt><dd><code>%s</code></dd>
        <dt>token_type</dt><dd><code>%s</code></dd>
        <dt>expires_in</dt><dd><code>%d</code></dd>
        <dt>scope</dt><dd><code>%s</code></dd>
        <dt>message</dt><dd><code>%s</code></dd>
      </dl>
    </section>
  </main>
</body>
</html>`,
			statusClass(oauthErr),
			statusTitle(oauthErr),
			html.EscapeString(code),
			html.EscapeString(state),
			html.EscapeString(oauthErr),
			html.EscapeString(errorDescription),
			tokenResult.statusClass(),
			html.EscapeString(tokenResult.statusTitle()),
			html.EscapeString(tokenResult.AccessToken),
			html.EscapeString(tokenResult.RefreshToken),
			html.EscapeString(tokenResult.TokenType),
			tokenResult.ExpiresIn,
			html.EscapeString(tokenResult.Scope),
			html.EscapeString(tokenResult.Message),
		)
	})

	log.Printf("Listening on http://%s%s", addr, path)
	log.Printf("OAuth base URL: %s", oauthClient.BaseURL())
	log.Printf("Token auto-exchange enabled: %t", clientID != "" && clientSecret != "")
	log.Printf("Full token logging enabled: %t", printTokens)
	log.Printf("Open this authorize URL:\n%s", authURL)
	log.Fatal(http.ListenAndServe(addr, mux))
}

type exchangeResult struct {
	OK           bool
	Skipped      bool
	Message      string
	AccessToken  string
	RefreshToken string
	TokenType    string
	ExpiresIn    int
	Scope        string
}

func (r exchangeResult) statusClass() string {
	if r.OK {
		return "ok"
	}
	if r.Skipped {
		return ""
	}
	return "error"
}

func (r exchangeResult) statusTitle() string {
	if r.OK {
		return "Token exchange succeeded"
	}
	if r.Skipped {
		return "Token exchange skipped"
	}
	return "Token exchange failed"
}

func envDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func statusClass(oauthErr string) string {
	if oauthErr != "" {
		return "error"
	}
	return "ok"
}

func statusTitle(oauthErr string) string {
	if oauthErr != "" {
		return "Authorization failed"
	}
	return "Authorization code received"
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
