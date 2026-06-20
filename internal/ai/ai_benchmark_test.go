package ai

import (
	"context"
	"testing"
)

type benchmarkChatProvider struct{}

func (benchmarkChatProvider) Chat(ctx context.Context, request ChatRequest) (ChatResponse, error) {
	return ChatResponse{Message: Message{Role: RoleAssistant, Content: "ok"}}, nil
}

type benchmarkEmbeddingProvider struct{}

func (benchmarkEmbeddingProvider) Embed(ctx context.Context, request EmbeddingRequest) (EmbeddingResponse, error) {
	return EmbeddingResponse{Vectors: [][]float32{{1, 2, 3}}}, nil
}

func BenchmarkClientChat(b *testing.B) {
	b.ReportAllocs()
	client := New(WithChatProvider(benchmarkChatProvider{}))
	request := ChatRequest{Model: "model-a", Messages: []Message{{Role: RoleUser, Content: "hello"}}}
	for b.Loop() {
		if _, err := client.Chat(context.Background(), request); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkClientEmbed(b *testing.B) {
	b.ReportAllocs()
	client := New(WithEmbeddingProvider(benchmarkEmbeddingProvider{}))
	request := EmbeddingRequest{Model: "embed-a", Input: []string{"hello"}}
	for b.Loop() {
		if _, err := client.Embed(context.Background(), request); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRedact(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_ = Redact("token sk-test password secret")
	}
}
