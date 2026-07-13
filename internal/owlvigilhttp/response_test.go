package owlvigilhttp

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
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
