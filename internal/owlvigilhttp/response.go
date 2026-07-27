package owlvigilhttp

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	owlvigil "github.com/Syrovex/owlvigil_sdk_go"
)

const (
	// maxResponseBodySize bounds successful JSON responses at 16 MiB.
	maxResponseBodySize = 16 << 20
	// maxErrorResponseBodySize bounds error responses before decoding.
	maxErrorResponseBodySize = 64 << 10
	// maxErrorBodySize limits error body truncation to prevent excessive memory usage.
	maxErrorBodySize = 4096
)

type rawEnvelope struct {
	RequestID string          `json:"request_id"`
	Code      string          `json:"code"`
	Message   string          `json:"message"`
	Data      json.RawMessage `json:"data"`
}

type rawOpenAIErrorEnvelope struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error"`
}

// DecodeResponse decodes an API response into out.
func DecodeResponse(resp *http.Response, out any, secrets ...string) (*owlvigil.ResponseMeta, error) {
	defer resp.Body.Close()
	bodyLimit := int64(maxResponseBodySize)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyLimit = maxErrorResponseBodySize
	}
	body, err := ReadAllLimited(resp.Body, bodyLimit)
	if err != nil {
		return nil, err
	}

	requestID := resp.Header.Get("OW-Request-Id")
	if requestID == "" {
		// Retain compatibility with pre-facade services that used the X prefix.
		requestID = resp.Header.Get("X-Request-Id")
	}
	meta := &owlvigil.ResponseMeta{RequestID: requestID}
	var env rawEnvelope
	hasEnvelope := json.Unmarshal(body, &env) == nil && (env.RequestID != "" || env.Code != "" || env.Message != "")
	if hasEnvelope {
		if env.RequestID != "" {
			meta.RequestID = env.RequestID
		}
		meta.Code = Redact(env.Code, secrets...)
		meta.Message = Redact(env.Message, secrets...)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg := meta.Message
		code := meta.Code
		if !hasEnvelope {
			var openAIError rawOpenAIErrorEnvelope
			if json.Unmarshal(body, &openAIError) == nil {
				msg = openAIError.Error.Message
				code = openAIError.Error.Code
				if code == "" {
					code = openAIError.Error.Type
				}
			}
		}
		if msg == "" {
			msg = http.StatusText(resp.StatusCode)
		}
		if code == "" {
			code = fmt.Sprintf("HTTP_%d", resp.StatusCode)
		}
		return meta, &owlvigil.APIError{
			StatusCode: resp.StatusCode,
			RequestID:  meta.RequestID,
			Code:       Redact(code, secrets...),
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

// ReadAllLimited reads at most limit bytes and reports an error when the
// response contains more data.
func ReadAllLimited(reader io.Reader, limit int64) ([]byte, error) {
	if limit < 0 {
		return nil, fmt.Errorf("owlvigil: response body limit must not be negative")
	}
	body, err := io.ReadAll(io.LimitReader(reader, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(body)) > limit {
		return nil, fmt.Errorf("owlvigil: response body exceeds %d byte limit", limit)
	}
	return body, nil
}

func truncateBody(body []byte) string {
	text := strings.TrimSpace(string(body))
	if len(text) > maxErrorBodySize {
		return text[:maxErrorBodySize]
	}
	return text
}
