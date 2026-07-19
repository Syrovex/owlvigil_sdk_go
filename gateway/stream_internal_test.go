package gateway

import (
	"errors"
	"strings"
	"testing"

	"github.com/Syrovex/owlvigil_sdk_go/internal/sse"
)

func TestStreamErrorAndNilClose(t *testing.T) {
	t.Parallel()

	stream := &Stream{scanner: sse.NewScanner(errorReader{err: errors.New("stream failed")})}
	if stream.Next() {
		t.Fatalf("Next = true")
	}
	if stream.Err() == nil || !strings.Contains(stream.Err().Error(), "stream failed") {
		t.Fatalf("Err = %v", stream.Err())
	}
	if stream.Next() {
		t.Fatalf("Next after error = true")
	}

	empty := &Stream{}
	if err := empty.Close(); err != nil {
		t.Fatalf("Close = %v", err)
	}
}

type errorReader struct {
	err error
}

func (r errorReader) Read([]byte) (int, error) {
	return 0, r.err
}
