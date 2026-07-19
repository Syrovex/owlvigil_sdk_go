package gateway

import (
	"context"
	"net/http"

	owlvigil "github.com/Syrovex/owlvigil_sdk_go"
)

// EmbeddingsRequest is the request body for embedding generation.
type EmbeddingsRequest struct {
	Model string `json:"model"`
	Input any    `json:"input"`
}

// EmbeddingsResponse is returned by CreateEmbeddings.
type EmbeddingsResponse struct {
	Object string          `json:"object"`
	Data   []EmbeddingData `json:"data"`
	Model  string          `json:"model"`
	Usage  *Usage          `json:"usage,omitempty"`
}

// EmbeddingData is one embedding vector in an embeddings response.
type EmbeddingData struct {
	Object    string    `json:"object"`
	Index     int       `json:"index"`
	Embedding []float64 `json:"embedding"`
}

// CreateEmbeddings calls POST /v1/embeddings.
func (c *Client) CreateEmbeddings(ctx context.Context, req *EmbeddingsRequest, opts ...owlvigil.RequestOption) (*EmbeddingsResponse, *owlvigil.ResponseMeta, error) {
	var out EmbeddingsResponse
	meta, err := c.http.Do(ctx, http.MethodPost, "/v1/embeddings", nil, req, &out, opts...)
	if err != nil {
		return nil, meta, err
	}
	return &out, meta, nil
}
