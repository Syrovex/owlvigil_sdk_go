package webhook_test

import (
	"strings"
	"testing"
	"time"

	"github.com/Syrovex/owlvigil_sdk_go/webhook"
)

func TestVerifySignature_ServiceHeaders(t *testing.T) {
	payload := []byte(`{"id":"evt_1","type":"webhook.test"}`)
	secret := "whsec_test_only"
	timestamp := time.Unix(1_700_000_000, 0)
	combined := webhook.SignPayload(payload, timestamp.Unix(), secret)
	parts := strings.SplitN(combined, ",", 2)
	if len(parts) != 2 {
		t.Fatalf("SignPayload() = %q, want timestamp and signature parts", combined)
	}

	timestampHeader := strings.TrimPrefix(parts[0], "t=")
	signatureHeader := parts[1]
	gotHeader := "t=" + timestampHeader + "," + signatureHeader
	err := webhook.VerifySignature(payload, gotHeader, secret, webhook.VerifyOptions{
		Now: func() time.Time { return timestamp },
	})
	if err != nil {
		t.Fatalf("VerifySignature() error = %v, want nil", err)
	}
}
