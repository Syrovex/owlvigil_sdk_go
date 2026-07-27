package gateway

import (
	"context"
	"errors"
	"net/http"

	owlvigil "github.com/Syrovex/owlvigil_sdk_go"
)

// ResponseRequest is the request body for the OpenAI-compatible Responses API.
type ResponseRequest struct {
	Model  string `json:"model"`
	Input  any    `json:"input"`
	Stream bool   `json:"stream,omitempty"`
}

// Response is returned by the OpenAI-compatible Responses API.
type Response struct {
	ID     string `json:"id"`
	Object string `json:"object"`
	Model  string `json:"model"`
	Output any    `json:"output,omitempty"`
}

// CreateResponse calls POST /v1/responses.
func (c *Client) CreateResponse(ctx context.Context, req *ResponseRequest, opts ...owlvigil.RequestOption) (*Response, *owlvigil.ResponseMeta, error) {
	var out Response
	meta, err := c.http.Do(ctx, http.MethodPost, "/v1/responses", nil, req, &out, opts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}

// CreateResponseStream calls POST /v1/responses with stream=true.
func (c *Client) CreateResponseStream(ctx context.Context, req *ResponseRequest, opts ...owlvigil.RequestOption) (*Stream, error) {
	if req == nil {
		return nil, errors.New("owlvigil: response stream request is nil")
	}
	copyReq := *req
	copyReq.Stream = true
	return c.newStream(ctx, http.MethodPost, "/v1/responses", &copyReq, opts...)
}
