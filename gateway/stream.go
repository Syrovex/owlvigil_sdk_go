package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/Syrovex/owlvigil_sdk_go/internal/owlvigilhttp"
	"github.com/Syrovex/owlvigil_sdk_go/internal/sse"

	owlvigil "github.com/Syrovex/owlvigil_sdk_go"
)

// StreamEvent is one decoded server-sent event from a Gateway stream.
type StreamEvent struct {
	Event string          `json:"event"`
	Data  json.RawMessage `json:"data"`
	Raw   string          `json:"raw"`
}

// Stream iterates over server-sent events returned by Gateway streaming calls.
type Stream struct {
	body    io.Closer
	scanner *sse.Scanner
	current StreamEvent
	err     error
}

func (c *Client) newStream(ctx context.Context, method, endpoint string, body any, opts ...owlvigil.RequestOption) (*Stream, error) {
	req, err := c.http.NewStreamRequest(ctx, method, endpoint, body, opts...)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "text/event-stream")
	resp, err := c.http.StreamHTTPClient().Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// DecodeResponse closes resp.Body via defer, so this is safe
		_, err := owlvigilhttp.DecodeResponse(resp, nil, c.http.RequestSecrets(req)...)
		if err != nil {
			return nil, err
		}
		// This path should not be reached (DecodeResponse returns error for non-2xx)
		// but close body defensively
		resp.Body.Close()
		return nil, fmt.Errorf("owlvigil: unexpected stream response status %d", resp.StatusCode)
	}
	return &Stream{
		body:    resp.Body,
		scanner: sse.NewScanner(resp.Body),
	}, nil
}

// Next advances the stream to the next event.
func (s *Stream) Next() bool {
	if s.err != nil {
		return false
	}
	event, ok, err := s.scanner.Next()
	if err != nil {
		s.err = err
		return false
	}
	if !ok {
		return false
	}
	s.current = StreamEvent{
		Event: event.Event,
		Raw:   event.Data,
		Data:  json.RawMessage(event.Data),
	}
	return true
}

// Current returns the most recent stream event.
func (s *Stream) Current() StreamEvent {
	return s.current
}

// Err returns the terminal stream error, if any.
func (s *Stream) Err() error {
	return s.err
}

// Close closes the underlying response body.
func (s *Stream) Close() error {
	if s.body == nil {
		return nil
	}
	return s.body.Close()
}
