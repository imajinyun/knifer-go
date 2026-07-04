package vai_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	knifer "github.com/imajinyun/knifer-go"
	"github.com/imajinyun/knifer-go/vai"
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

func TestFacadeExposesErrors(t *testing.T) {
	_, err := vai.Chat(context.Background(), nil, vai.ChatRequest{
		Model:    "model-a",
		Messages: []vai.Message{{Role: vai.RoleUser, Content: "ping"}},
	})
	if !errors.Is(err, vai.ErrMissingChatProvider) {
		t.Fatalf("Chat error = %v, want ErrMissingChatProvider", err)
	}

	_, err = vai.Embed(context.Background(), embeddingProviderFunc(func(context.Context, vai.EmbeddingRequest) (vai.EmbeddingResponse, error) {
		t.Fatal("provider should not be called for invalid request")
		return vai.EmbeddingResponse{}, nil
	}), vai.EmbeddingRequest{Model: "embed-a"})
	if !errors.Is(err, vai.ErrInvalidEmbeddingRequest) {
		t.Fatalf("Embed error = %v, want ErrInvalidEmbeddingRequest", err)
	}
}

func TestFacadeProviderErrorContract(t *testing.T) {
	cause := errors.New("provider unavailable")
	secret := "sk-test-secret"
	_, err := vai.Chat(context.Background(), chatProviderFunc(func(context.Context, vai.ChatRequest) (vai.ChatResponse, error) {
		return vai.ChatResponse{}, cause
	}), vai.ChatRequest{
		Model:    "model-a",
		Messages: []vai.Message{{Role: vai.RoleUser, Content: secret}},
	})
	if !errors.Is(err, cause) {
		t.Fatalf("Chat error = %v, want provider cause", err)
	}
	if !errors.Is(err, knifer.ErrCodeProviderFailure) {
		t.Fatalf("Chat error = %v, want ErrCodeProviderFailure", err)
	}
	if strings.Contains(err.Error(), secret) || strings.Contains(err.Error(), "secret") {
		t.Fatalf("Chat error leaked prompt secret: %v", err)
	}
}
