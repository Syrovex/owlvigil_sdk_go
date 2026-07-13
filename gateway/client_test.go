package gateway_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	owlvigil "github.com/owlvigil/owlvigil-go"
	"github.com/owlvigil/owlvigil-go/gateway"
)

func TestClientDefaults(t *testing.T) {
	t.Parallel()

	client := gateway.NewClient()

	if client.BaseURL() != owlvigil.DefaultGatewayBaseURL {
		t.Fatalf("BaseURL = %q, want %q", client.BaseURL(), owlvigil.DefaultGatewayBaseURL)
	}
}

func TestGatewayClientMethods(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		call          func(context.Context, *gateway.Client) error
		expectedPath  string
		expectedAuth  string
		expectedModel string
	}{
		{
			name: "chat completions",
			call: func(ctx context.Context, client *gateway.Client) error {
				_, _, err := client.CreateChatCompletion(ctx, &gateway.ChatCompletionRequest{
					Model:    "gpt-4o-mini",
					Messages: []gateway.Message{{Role: "user", Content: "hi"}},
				})
				return err
			},
			expectedPath:  "/v1/chat/completions",
			expectedAuth:  "Bearer ov_sk_test",
			expectedModel: "gpt-4o-mini",
		},
		{
			name: "responses",
			call: func(ctx context.Context, client *gateway.Client) error {
				_, _, err := client.CreateResponse(ctx, &gateway.ResponseRequest{Model: "gpt-4o-mini", Input: "hi"})
				return err
			},
			expectedPath:  "/v1/responses",
			expectedAuth:  "Bearer ov_sk_test",
			expectedModel: "gpt-4o-mini",
		},
		{
			name: "embeddings",
			call: func(ctx context.Context, client *gateway.Client) error {
				_, _, err := client.CreateEmbeddings(ctx, &gateway.EmbeddingsRequest{Model: "text-embedding-3-small", Input: "hi"})
				return err
			},
			expectedPath:  "/v1/embeddings",
			expectedAuth:  "Bearer ov_sk_test",
			expectedModel: "text-embedding-3-small",
		},
		{
			name: "anthropic message",
			call: func(ctx context.Context, client *gateway.Client) error {
				_, _, err := client.CreateAnthropicMessage(ctx, &gateway.AnthropicMessageRequest{
					Model:    "claude-3-5-sonnet",
					Messages: []gateway.AnthropicMessage{{Role: "user", Content: "hi"}},
				})
				return err
			},
			expectedPath:  "/anthropic/v1/messages",
			expectedAuth:  "Bearer ov_sk_test",
			expectedModel: "claude-3-5-sonnet",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != tt.expectedPath {
					t.Fatalf("path = %q, want %q", r.URL.Path, tt.expectedPath)
				}
				if got := r.Header.Get("Authorization"); got != tt.expectedAuth {
					t.Fatalf("Authorization = %q", got)
				}
				var payload struct {
					Model string `json:"model"`
				}
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					t.Fatalf("decode request: %v", err)
				}
				if payload.Model != tt.expectedModel {
					t.Fatalf("model = %q, want %q", payload.Model, tt.expectedModel)
				}
				_, _ = w.Write([]byte(`{"id":"ok","object":"test","model":"` + payload.Model + `","data":[]}`))
			}))
			defer server.Close()

			client := gateway.NewClient(
				owlvigil.WithBaseURL(server.URL),
				owlvigil.WithAPIKey("ov_sk_test"),
				owlvigil.WithoutRetry(),
			)
			if err := tt.call(context.Background(), client); err != nil {
				t.Fatalf("call returned error: %v", err)
			}
		})
	}
}

func TestListAndGetModels(t *testing.T) {
	t.Parallel()

	paths := make(chan string, 2)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths <- r.URL.Path
		switch r.URL.Path {
		case "/v1/models":
			_, _ = w.Write([]byte(`{"object":"list","data":[{"id":"model-a"}]}`))
		case "/v1/models/model-a":
			_, _ = w.Write([]byte(`{"id":"model-a","object":"model"}`))
		default:
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
	}))
	defer server.Close()

	client := gateway.NewClient(owlvigil.WithBaseURL(server.URL), owlvigil.WithAPIKey("ov_sk_test"))
	models, _, err := client.ListModels(context.Background())
	if err != nil {
		t.Fatalf("ListModels returned error: %v", err)
	}
	if len(models.Data) != 1 || models.Data[0].ID != "model-a" {
		t.Fatalf("models = %+v", models)
	}
	model, _, err := client.GetModel(context.Background(), "model-a")
	if err != nil {
		t.Fatalf("GetModel returned error: %v", err)
	}
	if model.ID != "model-a" {
		t.Fatalf("model = %+v", model)
	}
}
