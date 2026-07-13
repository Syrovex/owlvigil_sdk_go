package gateway

import (
	"context"
	"net/http"

	owlvigil "github.com/owlvigil/owlvigil-go"
)

// Message is an OpenAI-compatible chat message.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatCompletionRequest is the request body for chat completions.
type ChatCompletionRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature *float64  `json:"temperature,omitempty"`
	MaxTokens   *int      `json:"max_tokens,omitempty"`
	Stream      bool      `json:"stream,omitempty"`
}

// ChatCompletionResponse is an OpenAI-compatible chat completion response.
type ChatCompletionResponse struct {
	ID      string                 `json:"id"`
	Object  string                 `json:"object"`
	Created int64                  `json:"created"`
	Model   string                 `json:"model"`
	Choices []ChatCompletionChoice `json:"choices"`
	Usage   *Usage                 `json:"usage,omitempty"`
}

// ChatCompletionChoice is one generated chat completion choice.
type ChatCompletionChoice struct {
	Index        int      `json:"index"`
	Message      *Message `json:"message,omitempty"`
	FinishReason string   `json:"finish_reason,omitempty"`
}

// Usage contains token usage returned by Gateway model endpoints.
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// CreateChatCompletion calls POST /v1/chat/completions.
func (c *Client) CreateChatCompletion(ctx context.Context, req *ChatCompletionRequest, opts ...owlvigil.RequestOption) (*ChatCompletionResponse, *owlvigil.ResponseMeta, error) {
	var out ChatCompletionResponse
	meta, err := c.http.Do(ctx, http.MethodPost, "/v1/chat/completions", nil, req, &out, opts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// CreateChatCompletionStream calls POST /v1/chat/completions with stream=true.
func (c *Client) CreateChatCompletionStream(ctx context.Context, req *ChatCompletionRequest, opts ...owlvigil.RequestOption) (*Stream, error) {
	copyReq := *req
	copyReq.Stream = true
	return c.newStream(ctx, http.MethodPost, "/v1/chat/completions", &copyReq, opts...)
}
