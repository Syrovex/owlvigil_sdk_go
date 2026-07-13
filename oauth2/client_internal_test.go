package oauth2

import (
	"net/http"
	"testing"
)

func TestHTTPClientDefaultFallback(t *testing.T) {
	t.Parallel()

	client := &Client{}
	if client.httpClient() != http.DefaultClient {
		t.Fatalf("httpClient did not fall back to http.DefaultClient")
	}
}
