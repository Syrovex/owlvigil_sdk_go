package owlvigil_test

import (
	"strings"
	"testing"

	owlvigil "github.com/Syrovex/owlvigil_sdk_go"
)

func TestAPIErrorError(t *testing.T) {
	t.Parallel()

	err := (&owlvigil.APIError{
		StatusCode: 400,
		RequestID:  "req_123",
		Code:       "invalid_request",
		Message:    "invalid request",
	}).Error()

	for _, part := range []string{"status=400", "code=invalid_request", "request_id=req_123", "invalid request"} {
		if !strings.Contains(err, part) {
			t.Fatalf("Error() = %q, missing %q", err, part)
		}
	}

	withoutRequestID := (&owlvigil.APIError{
		StatusCode: 500,
		Code:       "server_error",
		Message:    "failed",
	}).Error()
	if strings.Contains(withoutRequestID, "request_id=") {
		t.Fatalf("Error() = %q, should not include request_id", withoutRequestID)
	}

	var nilErr *owlvigil.APIError
	if nilErr.Error() != "<nil>" {
		t.Fatalf("nil APIError Error() = %q", nilErr.Error())
	}
}

func TestOAuthErrorError(t *testing.T) {
	t.Parallel()

	err := (&owlvigil.OAuthError{
		StatusCode:       401,
		ErrorCode:        "invalid_client",
		ErrorDescription: "invalid client",
	}).Error()

	if !strings.Contains(err, "invalid_client") || !strings.Contains(err, "invalid client") {
		t.Fatalf("Error() = %q", err)
	}

	withoutDescription := (&owlvigil.OAuthError{
		StatusCode: 400,
		ErrorCode:  "invalid_request",
	}).Error()
	if strings.Contains(withoutDescription, "description=") {
		t.Fatalf("Error() = %q, should not include description", withoutDescription)
	}

	var nilErr *owlvigil.OAuthError
	if nilErr.Error() != "<nil>" {
		t.Fatalf("nil OAuthError Error() = %q", nilErr.Error())
	}
}
