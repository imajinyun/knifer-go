package vai_test

import (
	"context"
	"fmt"

	"github.com/imajinyun/go-knifer/vai"
)

type exampleChatProvider struct{}

func (exampleChatProvider) Chat(ctx context.Context, request vai.ChatRequest) (vai.ChatResponse, error) {
	return vai.ChatResponse{Message: vai.Message{Role: vai.RoleAssistant, Content: "hello gopher"}}, nil
}

type exampleEmbeddingProvider struct{}

func (exampleEmbeddingProvider) Embed(ctx context.Context, request vai.EmbeddingRequest) (vai.EmbeddingResponse, error) {
	return vai.EmbeddingResponse{Vectors: [][]float32{{0.1, 0.2, 0.3}}}, nil
}

func ExampleNew() {
	client := vai.New(vai.WithChatProvider(exampleChatProvider{}))
	response, _ := client.Chat(context.Background(), vai.ChatRequest{
		Model:    "example-chat",
		Messages: []vai.Message{{Role: vai.RoleUser, Content: "hello"}},
	})
	fmt.Println(response.Message.Content)
	// Output: hello gopher
}

func ExampleWithChatProvider() {
	client := vai.New(vai.WithChatProvider(exampleChatProvider{}))
	response, _ := client.Chat(context.Background(), vai.ChatRequest{
		Model:    "example-chat",
		Messages: []vai.Message{{Role: vai.RoleUser, Content: "hello"}},
	})
	fmt.Println(response.Message.Role)
	// Output: assistant
}

func ExampleWithEmbeddingProvider() {
	client := vai.New(vai.WithEmbeddingProvider(exampleEmbeddingProvider{}))
	response, _ := client.Embed(context.Background(), vai.EmbeddingRequest{
		Model: "example-embedding",
		Input: []string{"hello"},
	})
	fmt.Println(len(response.Vectors))
	// Output: 1
}

func ExampleChat() {
	response, _ := vai.Chat(context.Background(), exampleChatProvider{}, vai.ChatRequest{
		Model:    "example-chat",
		Messages: []vai.Message{{Role: vai.RoleUser, Content: "ping"}},
	})
	fmt.Println(response.Message.Role, response.Message.Content)
	// Output: assistant hello gopher
}

func ExampleEmbed() {
	response, _ := vai.Embed(context.Background(), exampleEmbeddingProvider{}, vai.EmbeddingRequest{
		Model: "example-embedding",
		Input: []string{"hello"},
	})
	fmt.Println(len(response.Vectors), len(response.Vectors[0]))
	// Output: 1 3
}

func ExampleRedact() {
	fmt.Println(vai.Redact("api key sk-test secret"))
	// Output: api key [REDACTED] [REDACTED]
}
