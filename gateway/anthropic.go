package gateway

import (
	"context"
	"net/http"

	owlvigil "github.com/Syrovex/owlvigil_sdk_go"
)

// AnthropicMessageRequest is the request body for Anthropic-compatible messages.
type AnthropicMessageRequest struct {
	Model     string             `json:"model"`
	Messages  []AnthropicMessage `json:"messages"`
	MaxTokens int                `json:"max_tokens,omitempty"`
	Stream    bool               `json:"stream,omitempty"`
}

// AnthropicMessage is an Anthropic-compatible message.
type AnthropicMessage struct {
	Role    string `json:"role"`
	Content any    `json:"content"`
}

// AnthropicMessageResponse is returned by Anthropic-compatible message calls.
type AnthropicMessageResponse struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Role    string `json:"role"`
	Model   string `json:"model"`
	Content any    `json:"content"`
}

// CreateAnthropicMessage calls POST /anthropic/v1/messages.
func (c *Client) CreateAnthropicMessage(ctx context.Context, req *AnthropicMessageRequest, opts ...owlvigil.RequestOption) (*AnthropicMessageResponse, *owlvigil.ResponseMeta, error) {
	var out AnthropicMessageResponse
	meta, err := c.http.Do(ctx, http.MethodPost, "/anthropic/v1/messages", nil, req, &out, opts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}
