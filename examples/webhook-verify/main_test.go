package main

import (
	"bytes"
	"testing"
)

func TestRunVerifiesSampleWebhookAndWritesConfirmation(t *testing.T) {
	var output bytes.Buffer
	if err := run(&output); err != nil {
		t.Fatalf("run() error = %v", err)
	}
	if got, want := output.String(), "verified\n"; got != want {
		t.Errorf("run() output = %q, want %q", got, want)
	}
}
