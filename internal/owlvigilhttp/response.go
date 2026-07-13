package owlvigilhttp

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	owlvigil "github.com/owlvigil/owlvigil-go"
)

const (
	// maxErrorBodySize limits error body truncation to prevent excessive memory usage.
	maxErrorBodySize = 4096
)

type rawEnvelope struct {
	RequestID string          `json:"request_id"`
	Code      string          `json:"code"`
	Message   string          `json:"message"`
	Data      json.RawMessage `json:"data"`
}

// DecodeResponse decodes an API response into out.
func DecodeResponse(resp *http.Response, out any, secrets ...string) (*owlvigil.ResponseMeta, error) {
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	meta := &owlvigil.ResponseMeta{RequestID: resp.Header.Get("X-Request-Id")}
	var env rawEnvelope
	hasEnvelope := json.Unmarshal(body, &env) == nil && (env.RequestID != "" || env.Code != "" || env.Message != "")
	if hasEnvelope {
		if env.RequestID != "" {
			meta.RequestID = env.RequestID
		}
		meta.Code = env.Code
		meta.Message = env.Message
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg := env.Message
		code := env.Code
		if msg == "" {
			msg = http.StatusText(resp.StatusCode)
		}
		if code == "" {
			code = fmt.Sprintf("HTTP_%d", resp.StatusCode)
		}
		return meta, &owlvigil.APIError{
			StatusCode: resp.StatusCode,
			RequestID:  meta.RequestID,
			Code:       code,
			Message:    Redact(msg, secrets...),
			Body:       Redact(truncateBody(body), secrets...),
		}
	}

	if out == nil || len(body) == 0 {
		return meta, nil
	}
	if hasEnvelope && env.Data != nil {
		if string(env.Data) == "null" {
			return meta, nil
		}
		return meta, json.Unmarshal(env.Data, out)
	}
	return meta, json.Unmarshal(body, out)
}

func truncateBody(body []byte) string {
	text := strings.TrimSpace(string(body))
	if len(text) > maxErrorBodySize {
		return text[:maxErrorBodySize]
	}
	return text
}
