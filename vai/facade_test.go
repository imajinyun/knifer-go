package vai_test

import (
	"context"
	"testing"

	"github.com/imajinyun/go-knifer/vai"
)

type chatProviderFunc func(context.Context, vai.ChatRequest) (vai.ChatResponse, error)

func (f chatProviderFunc) Chat(ctx context.Context, request vai.ChatRequest) (vai.ChatResponse, error) {
	return f(ctx, request)
}

type embeddingProviderFunc func(context.Context, vai.EmbeddingRequest) (vai.EmbeddingResponse, error)

func (f embeddingProviderFunc) Embed(ctx context.Context, request vai.EmbeddingRequest) (vai.EmbeddingResponse, error) {
	return f(ctx, request)
}

func TestChatFacade(t *testing.T) {
	provider := chatProviderFunc(func(ctx context.Context, request vai.ChatRequest) (vai.ChatResponse, error) {
		return vai.ChatResponse{Message: vai.Message{Role: vai.RoleAssistant, Content: "pong"}}, nil
	})
	response, err := vai.Chat(context.Background(), provider, vai.ChatRequest{
		Model:    "model-a",
		Messages: []vai.Message{{Role: vai.RoleUser, Content: "ping"}},
	})
	if err != nil {
		t.Fatalf("Chat returned error: %v", err)
	}
	if response.Message.Content != "pong" {
		t.Fatalf("Chat response = %+v", response)
	}
}

func TestEmbedFacade(t *testing.T) {
	provider := embeddingProviderFunc(func(ctx context.Context, request vai.EmbeddingRequest) (vai.EmbeddingResponse, error) {
		return vai.EmbeddingResponse{Vectors: [][]float32{{1, 2, 3}}}, nil
	})
	response, err := vai.Embed(context.Background(), provider, vai.EmbeddingRequest{
		Model: "embed-a",
		Input: []string{"hello"},
	})
	if err != nil {
		t.Fatalf("Embed returned error: %v", err)
	}
	if len(response.Vectors) != 1 || len(response.Vectors[0]) != 3 {
		t.Fatalf("Embed response = %+v", response)
	}
}
