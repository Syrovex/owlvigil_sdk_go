// Package webhook verifies inbound OwlVigil webhook signatures.
package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

var (
	// ErrMissingSignature indicates the signature header is absent.
	ErrMissingSignature = errors.New("owlvigil webhook: missing signature")
	// ErrInvalidSignature indicates the signature header is malformed or does not match.
	ErrInvalidSignature = errors.New("owlvigil webhook: invalid signature")
	// ErrMissingSecret indicates the webhook signing secret is absent.
	ErrMissingSecret = errors.New("owlvigil webhook: missing secret")
	// ErrStaleTimestamp indicates the signature timestamp is outside tolerance.
	ErrStaleTimestamp = errors.New("owlvigil webhook: stale timestamp")
)

// VerifyOptions configures webhook signature verification.
type VerifyOptions struct {
	Tolerance time.Duration
	Now       func() time.Time
}

// VerifySignature verifies an OwlVigil webhook signature header.
func VerifySignature(payload []byte, header string, secret string, opts VerifyOptions) error {
	if header == "" {
		return ErrMissingSignature
	}
	if strings.TrimSpace(secret) == "" {
		return ErrMissingSecret
	}
	if opts.Tolerance == 0 {
		opts.Tolerance = 5 * time.Minute
	}
	if opts.Now == nil {
		opts.Now = time.Now
	}
	timestamp, signature, err := parseHeader(header)
	if err != nil {
		return err
	}
	eventTime := time.Unix(timestamp, 0)
	if delta := opts.Now().Sub(eventTime); delta > opts.Tolerance || delta < -opts.Tolerance {
		return ErrStaleTimestamp
	}
	expected := computeSignature(payload, timestamp, secret)
	got, err := hex.DecodeString(signature)
	if err != nil {
		return ErrInvalidSignature
	}
	if !hmac.Equal(got, expected) {
		return ErrInvalidSignature
	}
	return nil
}

// SignPayload signs a payload using OwlVigil's webhook signature format.
func SignPayload(payload []byte, timestamp int64, secret string) string {
	return fmt.Sprintf("t=%d,v1=%s", timestamp, hex.EncodeToString(computeSignature(payload, timestamp, secret)))
}

func computeSignature(payload []byte, timestamp int64, secret string) []byte {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(strconv.FormatInt(timestamp, 10)))
	mac.Write([]byte("."))
	mac.Write(payload)
	return mac.Sum(nil)
}

func parseHeader(header string) (int64, string, error) {
	var timestamp int64
	var signature string
	for _, part := range strings.Split(header, ",") {
		key, value, ok := strings.Cut(strings.TrimSpace(part), "=")
		if !ok {
			continue
		}
		switch key {
		case "t":
			parsed, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return 0, "", ErrInvalidSignature
			}
			timestamp = parsed
		case "v1":
			signature = value
		}
	}
	if timestamp == 0 || signature == "" {
		return 0, "", ErrInvalidSignature
	}
	return timestamp, signature, nil
}
