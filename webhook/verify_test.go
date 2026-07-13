package webhook_test

import (
	"errors"
	"testing"
	"time"

	"github.com/owlvigil/owlvigil-go/webhook"
)

func TestVerifySignature(t *testing.T) {
	t.Parallel()

	payload := []byte(`{"id":"evt_1"}`)
	now := time.Unix(1000, 0)
	header := webhook.SignPayload(payload, now.Unix(), "whsec_test")

	err := webhook.VerifySignature(payload, header, "whsec_test", webhook.VerifyOptions{
		Now: func() time.Time { return now },
	})
	if err != nil {
		t.Fatalf("VerifySignature returned error: %v", err)
	}
}

func TestVerifySignatureFailures(t *testing.T) {
	t.Parallel()

	payload := []byte(`{"id":"evt_1"}`)
	now := time.Unix(1000, 0)

	tests := []struct {
		name     string
		header   string
		secret   string
		now      time.Time
		expected error
	}{
		{name: "missing header", header: "", secret: "whsec_test", now: now, expected: webhook.ErrMissingSignature},
		{name: "missing secret", header: webhook.SignPayload(payload, now.Unix(), "whsec_test"), secret: "", now: now, expected: webhook.ErrMissingSecret},
		{name: "invalid header", header: "v1=abc", secret: "whsec_test", now: now, expected: webhook.ErrInvalidSignature},
		{name: "stale timestamp", header: webhook.SignPayload(payload, now.Add(-10*time.Minute).Unix(), "whsec_test"), secret: "whsec_test", now: now, expected: webhook.ErrStaleTimestamp},
		{name: "wrong secret", header: webhook.SignPayload(payload, now.Unix(), "whsec_test"), secret: "wrong_secret", now: now, expected: webhook.ErrInvalidSignature},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := webhook.VerifySignature(payload, tt.header, tt.secret, webhook.VerifyOptions{
				Now: func() time.Time { return tt.now },
			})
			if !errors.Is(err, tt.expected) {
				t.Fatalf("err = %v, want %v", err, tt.expected)
			}
		})
	}
}
