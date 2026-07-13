package sse

import (
	"bufio"
	"bytes"
	"io"
	"strings"
)

// Event is a parsed server-sent event.
type Event struct {
	Event string
	Data  string
	ID    string
}

// Scanner parses server-sent events from an io.Reader.
type Scanner struct {
	scanner *bufio.Scanner
}

// NewScanner creates an SSE scanner.
func NewScanner(r io.Reader) *Scanner {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	return &Scanner{scanner: scanner}
}

// Next reads the next SSE event.
func (s *Scanner) Next() (Event, bool, error) {
	var event Event
	var data bytes.Buffer
	for s.scanner.Scan() {
		line := s.scanner.Text()
		if line == "" {
			if data.Len() == 0 && event.Event == "" && event.ID == "" {
				continue
			}
			event.Data = strings.TrimSuffix(data.String(), "\n")
			return event, true, nil
		}
		if strings.HasPrefix(line, ":") {
			continue
		}
		key, value, found := strings.Cut(line, ":")
		if found {
			value = strings.TrimPrefix(value, " ")
		}
		switch key {
		case "event":
			event.Event = value
		case "data":
			data.WriteString(value)
			data.WriteByte('\n')
		case "id":
			event.ID = value
		}
	}
	if err := s.scanner.Err(); err != nil {
		return Event{}, false, err
	}
	if data.Len() > 0 || event.Event != "" || event.ID != "" {
		event.Data = strings.TrimSuffix(data.String(), "\n")
		return event, true, nil
	}
	return Event{}, false, nil
}
