package sse

import (
	"errors"
	"io"
	"strings"
	"testing"
)

func TestScannerNext(t *testing.T) {
	t.Parallel()

	scanner := NewScanner(strings.NewReader(": keepalive\nid: evt_1\nevent: message\ndata: {\"a\":1}\ndata: {\"b\":2}\n\n"))

	event, ok, err := scanner.Next()
	if err != nil {
		t.Fatalf("Next returned error: %v", err)
	}
	if !ok {
		t.Fatalf("Next ok = false")
	}
	if event.ID != "evt_1" || event.Event != "message" || event.Data != "{\"a\":1}\n{\"b\":2}" {
		t.Fatalf("event = %+v", event)
	}
}

func TestScannerNextEOFAndBlankEvents(t *testing.T) {
	t.Parallel()

	scanner := NewScanner(strings.NewReader("\n\nid: evt_2\ndata: done"))

	event, ok, err := scanner.Next()
	if err != nil {
		t.Fatalf("Next returned error: %v", err)
	}
	if !ok || event.ID != "evt_2" || event.Data != "done" {
		t.Fatalf("event=%+v ok=%t", event, ok)
	}
	if event, ok, err := scanner.Next(); err != nil || ok || event != (Event{}) {
		t.Fatalf("second Next = %+v, %t, %v", event, ok, err)
	}
}

func TestScannerNextReadError(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("read failed")
	scanner := NewScanner(errorReader{err: wantErr})

	_, ok, err := scanner.Next()
	if ok {
		t.Fatalf("ok = true")
	}
	if !errors.Is(err, wantErr) {
		t.Fatalf("err = %v, want %v", err, wantErr)
	}
}

type errorReader struct {
	err error
}

func (r errorReader) Read([]byte) (int, error) {
	return 0, r.err
}

var _ io.Reader = errorReader{}
