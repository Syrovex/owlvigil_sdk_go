package owlvigilhttp

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	owlvigil "github.com/Syrovex/owlvigil_sdk_go"
)

func TestDecodeResponseRawJSON(t *testing.T) {
	t.Parallel()

	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"X-Request-Id": []string{"req_header"}},
		Body:       ioNopCloser(`{"id":"raw"}`),
	}
	var out struct {
		ID string `json:"id"`
	}
	meta, err := DecodeResponse(resp, &out)
	if err != nil {
		t.Fatalf("DecodeResponse returned error: %v", err)
	}
	if out.ID != "raw" || meta.RequestID != "req_header" {
		t.Fatalf("out=%+v meta=%+v", out, meta)
	}
}

func TestDecodeResponseUsesOpenAPIRequestIDHeader(t *testing.T) {
	t.Parallel()

	resp := &http.Response{
		StatusCode: http.StatusNoContent,
		Header:     http.Header{"Ow-Request-Id": []string{"req_openapi_header"}},
		Body:       ioNopCloser(""),
	}
	meta, err := DecodeResponse(resp, nil)
	if err != nil {
		t.Fatalf("DecodeResponse returned error: %v", err)
	}
	if meta.RequestID != "req_openapi_header" {
		t.Fatalf("meta.RequestID = %q, want %q", meta.RequestID, "req_openapi_header")
	}
}

func TestDecodeResponseEnvelopeNullAndErrorFallback(t *testing.T) {
	t.Parallel()

	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"X-Request-Id": []string{"req_header"}},
		Body:       ioNopCloser(`{"request_id":"req_env","code":"ok","message":"done","data":null}`),
	}
	var out struct {
		ID string `json:"id"`
	}
	meta, err := DecodeResponse(resp, &out)
	if err != nil {
		t.Fatalf("DecodeResponse returned error: %v", err)
	}
	if meta.RequestID != "req_env" || out.ID != "" {
		t.Fatalf("meta=%+v out=%+v", meta, out)
	}

	resp = &http.Response{
		StatusCode: http.StatusTeapot,
		Body:       ioNopCloser(strings.Repeat("x", 5000)),
	}
	_, err = DecodeResponse(resp, nil)
	var apiErr interface {
		error
	}
	if !errors.As(err, &apiErr) {
		t.Fatalf("err = %T, want API error", err)
	}
	if !strings.Contains(err.Error(), "HTTP_418") {
		t.Fatalf("err = %v", err)
	}
}

func TestDecodeResponseOpenAICompatibleError(t *testing.T) {
	t.Parallel()

	const secret = "sk-sensitive"
	resp := &http.Response{
		StatusCode: http.StatusForbidden,
		Header:     http.Header{"X-Request-Id": []string{"req_gateway"}},
		Body: ioNopCloser(`{
			"error": {
				"message": "budget limit exceeded for sk-sensitive",
				"type": "budget_limit_exceeded",
				"code": "budget_limit_exceeded"
			}
		}`),
	}

	meta, err := DecodeResponse(resp, nil, secret)
	var apiErr *owlvigil.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("DecodeResponse(OpenAI error) error = %T, want *owlvigil.APIError", err)
	}
	if meta.RequestID != "req_gateway" {
		t.Errorf("DecodeResponse(OpenAI error) request ID = %q, want %q", meta.RequestID, "req_gateway")
	}
	if apiErr.Code != "budget_limit_exceeded" {
		t.Errorf("DecodeResponse(OpenAI error) code = %q, want %q", apiErr.Code, "budget_limit_exceeded")
	}
	if !strings.Contains(apiErr.Message, "budget limit exceeded") {
		t.Errorf("DecodeResponse(OpenAI error) message = %q, want gateway error message", apiErr.Message)
	}
	if strings.Contains(apiErr.Message, secret) {
		t.Errorf("DecodeResponse(OpenAI error) message = %q, want secret redacted", apiErr.Message)
	}
	if strings.Contains(apiErr.Body, secret) {
		t.Errorf("DecodeResponse(OpenAI error) body = %q, want secret redacted", apiErr.Body)
	}
}

func TestDecodeResponseReadError(t *testing.T) {
	t.Parallel()

	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       errorReadCloser{err: errors.New("read failed")},
	}
	_, err := DecodeResponse(resp, nil)
	if err == nil || !strings.Contains(err.Error(), "read failed") {
		t.Fatalf("err = %v", err)
	}
}

func ioNopCloser(s string) *readCloser {
	return &readCloser{Reader: strings.NewReader(s)}
}

type readCloser struct {
	*strings.Reader
}

func (r *readCloser) Close() error { return nil }

type errorReadCloser struct {
	err error
}

func (r errorReadCloser) Read([]byte) (int, error) {
	return 0, r.err
}

func (r errorReadCloser) Close() error { return nil }

var _ io.ReadCloser = errorReadCloser{}
