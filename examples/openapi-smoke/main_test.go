package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	owlvigil "github.com/Syrovex/owlvigil_sdk_go"
	"github.com/Syrovex/owlvigil_sdk_go/management"
)

func TestOAuthEnabled(t *testing.T) {
	tests := []struct {
		name         string
		accessToken  string
		clientID     string
		clientSecret string
		want         bool
	}{
		{name: "access token", accessToken: "token", want: true},
		{name: "client credentials", clientID: "client", clientSecret: "secret", want: true},
		{name: "no OAuth credentials", want: false},
		{name: "client ID only", clientID: "client", want: false},
		{name: "client secret only", clientSecret: "secret", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := oauthEnabled(tt.accessToken, tt.clientID, tt.clientSecret); got != tt.want {
				t.Errorf("oauthEnabled(%q, %q, %q) = %t, want %t", tt.accessToken, tt.clientID, tt.clientSecret, got, tt.want)
			}
		})
	}
}

func TestWriteSmokeEnabled(t *testing.T) {
	tests := []struct {
		value string
		want  bool
	}{
		{value: "1", want: true},
		{value: "true", want: false},
		{value: "", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			if got := writeSmokeEnabled(tt.value); got != tt.want {
				t.Errorf("writeSmokeEnabled(%q) = %t, want %t", tt.value, got, tt.want)
			}
		})
	}
}

func TestStripePaymentMethodID(t *testing.T) {
	tests := []struct {
		value string
		want  string
	}{
		{value: "pm_card_visa", want: "pm_card_visa"},
		{value: " pm_card_visa ", want: "pm_card_visa"},
		{value: "", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			if got := stripePaymentMethodID(tt.value); got != tt.want {
				t.Errorf("stripePaymentMethodID(%q) = %q, want %q", tt.value, got, tt.want)
			}
		})
	}
}

func TestRunner_DoesNotCreateSmokeResourcesWhenWritesAreDisabled(t *testing.T) {
	var mutations atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			mutations.Add(1)
		}
		if req.URL.Path == "/webhook-event-types" {
			_, _ = w.Write([]byte(`[]`))
			return
		}
		if req.URL.Path == "/webhook-events" {
			_, _ = w.Write([]byte(`{"items":[{"id":"1"}],"page_info":{}}`))
			return
		}
		_, _ = w.Write([]byte(`{"items":[],"page_info":{}}`))
	}))
	defer server.Close()

	r := &runner{
		ctx: context.Background(),
		management: management.NewClient(
			owlvigil.WithBaseURL(server.URL),
			owlvigil.WithoutRetry(),
		),
		workspaceID: 1,
		writes:      false,
	}
	r.runGateway()
	r.runWebhooks()

	if got := mutations.Load(); got != 0 {
		t.Fatalf("mutating smoke requests = %d, want 0 when OWLVIGIL_SMOKE_WRITES is disabled", got)
	}
}

func TestRunner_CleanupContextOutlivesRunContext(t *testing.T) {
	runCtx, cancelRun := context.WithCancel(context.Background())
	cancelRun()
	r := &runner{ctx: runCtx}

	cleanupCtx, cancelCleanup := r.cleanupContext()
	defer cancelCleanup()

	if err := cleanupCtx.Err(); err != nil {
		t.Fatalf("cleanup context error = %v, want usable context", err)
	}
	deadline, ok := cleanupCtx.Deadline()
	if !ok {
		t.Fatal("cleanup context has no deadline")
	}
	if remaining := time.Until(deadline); remaining <= 0 || remaining > cleanupTimeout {
		t.Fatalf("cleanup deadline remaining = %s, want (0, %s]", remaining, cleanupTimeout)
	}
}
