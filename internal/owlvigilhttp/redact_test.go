package owlvigilhttp

import (
	"strings"
	"testing"
)

func TestRedact(t *testing.T) {
	t.Parallel()

	const secret = "ov_sk_1234567890"
	got := Redact("failed with ov_sk_1234567890", secret)

	if strings.Contains(got, secret) {
		t.Fatalf("redacted text still contains secret: %q", got)
	}
	if !strings.Contains(got, "ov_s...7890") {
		t.Fatalf("redacted text = %q", got)
	}

	short := Redact("secret abc123 should be masked, tiny abc ignored", "abc123", "abc")
	if !strings.Contains(short, "secret **** should be masked") {
		t.Fatalf("short redaction = %q", short)
	}
	if !strings.Contains(short, "tiny abc ignored") {
		t.Fatalf("tiny secret should not be redacted: %q", short)
	}
}
