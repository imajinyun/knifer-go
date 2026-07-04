package ai

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

type fakeChatProvider struct {
	requests []ChatRequest
	response ChatResponse
	err      error
}

func (p *fakeChatProvider) Chat(ctx context.Context, request ChatRequest) (ChatResponse, error) {
	select {
	case <-ctx.Done():
		return ChatResponse{}, ctx.Err()
	default:
	}
	p.requests = append(p.requests, request)
	return p.response, p.err
}

type fakeEmbeddingProvider struct {
	requests []EmbeddingRequest
	response EmbeddingResponse
	err      error
}

func (p *fakeEmbeddingProvider) Embed(ctx context.Context, request EmbeddingRequest) (EmbeddingResponse, error) {
	select {
	case <-ctx.Done():
		return EmbeddingResponse{}, ctx.Err()
	default:
	}
	p.requests = append(p.requests, request)
	return p.response, p.err
}

func TestClientChatUsesProviderAndClonesRequest(t *testing.T) {
	provider := &fakeChatProvider{response: ChatResponse{
		Message:  Message{Role: RoleAssistant, Content: "hi"},
		Usage:    Usage{InputTokens: 1, OutputTokens: 2, TotalTokens: 3},
		Provider: ProviderMetadata{Name: "fake", Model: "model-a", TraceID: "trace-1"},
	}}
	client := New(WithChatProvider(provider))
	request := ChatRequest{
		Model:    "model-a",
		Messages: []Message{{Role: RoleUser, Content: "hello"}},
		Metadata: map[string]string{"trace": "one"},
	}

	response, err := client.Chat(context.Background(), request)
	if err != nil {
		t.Fatalf("Chat returned error: %v", err)
	}
	if response.Message.Content != "hi" || response.Usage.TotalTokens != 3 || response.Provider.TraceID != "trace-1" {
		t.Fatalf("Chat response = %+v", response)
	}
	request.Messages[0].Content = "changed"
	request.Metadata["trace"] = "changed"
	if provider.requests[0].Messages[0].Content != "hello" || provider.requests[0].Metadata["trace"] != "one" {
		t.Fatalf("provider request was not cloned: %+v", provider.requests[0])
	}
}

func TestNilProviderOptionsDoNotOverwriteConfiguredProviders(t *testing.T) {
	chatProvider := &fakeChatProvider{response: ChatResponse{Message: Message{Role: RoleAssistant, Content: "hi"}}}
	embeddingProvider := &fakeEmbeddingProvider{response: EmbeddingResponse{Vectors: [][]float32{{1}}}}
	client := New(
		WithChatProvider(chatProvider),
		WithChatProvider(nil),
		WithEmbeddingProvider(embeddingProvider),
		WithEmbeddingProvider(nil),
	)

	if _, err := client.Chat(context.Background(), ChatRequest{Model: "model-a", Messages: []Message{{Role: RoleUser, Content: "hello"}}}); err != nil {
		t.Fatalf("Chat with nil overwrite option error = %v", err)
	}
	if _, err := client.Embed(context.Background(), EmbeddingRequest{Model: "embed-a", Input: []string{"hello"}}); err != nil {
		t.Fatalf("Embed with nil overwrite option error = %v", err)
	}
}

func TestClientChatRequiresProvider(t *testing.T) {
	client := New()
	_, err := client.Chat(context.Background(), ChatRequest{Model: "model-a", Messages: []Message{{Role: RoleUser, Content: "hello"}}})
	if !errors.Is(err, ErrMissingChatProvider) {
		t.Fatalf("Chat error = %v, want ErrMissingChatProvider", err)
	}
}

func TestClientChatValidatesBeforeProvider(t *testing.T) {
	provider := &fakeChatProvider{}
	client := New(WithChatProvider(provider))
	_, err := client.Chat(context.Background(), ChatRequest{Model: "model-a"})
	if !errors.Is(err, ErrInvalidChatRequest) {
		t.Fatalf("Chat error = %v, want ErrInvalidChatRequest", err)
	}
	if len(provider.requests) != 0 {
		t.Fatalf("provider was called for invalid request: %d", len(provider.requests))
	}
}

func TestClientChatPropagatesContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	client := New(WithChatProvider(&fakeChatProvider{}))
	_, err := client.Chat(ctx, ChatRequest{Model: "model-a", Messages: []Message{{Role: RoleUser, Content: "hello"}}})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Chat error = %v, want context.Canceled", err)
	}
}

func TestClientChatPropagatesProviderError(t *testing.T) {
	providerErr := errors.New("provider failed")
	client := New(WithChatProvider(&fakeChatProvider{err: providerErr}))
	_, err := client.Chat(context.Background(), ChatRequest{Model: "model-a", Messages: []Message{{Role: RoleUser, Content: "hello"}}})
	if !errors.Is(err, providerErr) {
		t.Fatalf("Chat error = %v, want provider error", err)
	}
}

func TestClientEmbedUsesProviderAndClonesRequestAndResponse(t *testing.T) {
	provider := &fakeEmbeddingProvider{response: EmbeddingResponse{
		Vectors:  [][]float32{{1, 2}},
		Usage:    Usage{InputTokens: 1, TotalTokens: 1},
		Provider: ProviderMetadata{Name: "fake", Model: "embed-a"},
	}}
	client := New(WithEmbeddingProvider(provider))
	request := EmbeddingRequest{
		Model:    "embed-a",
		Input:    []string{"hello"},
		Metadata: map[string]string{"trace": "one"},
	}

	response, err := client.Embed(context.Background(), request)
	if err != nil {
		t.Fatalf("Embed returned error: %v", err)
	}
	request.Input[0] = "changed"
	request.Metadata["trace"] = "changed"
	provider.response.Vectors[0][0] = 99
	if provider.requests[0].Input[0] != "hello" || provider.requests[0].Metadata["trace"] != "one" {
		t.Fatalf("provider request was not cloned: %+v", provider.requests[0])
	}
	if !reflect.DeepEqual(response.Vectors, [][]float32{{1, 2}}) {
		t.Fatalf("Embed response vectors = %#v", response.Vectors)
	}
}

func TestClientEmbedRequiresProvider(t *testing.T) {
	client := New()
	_, err := client.Embed(context.Background(), EmbeddingRequest{Model: "embed-a", Input: []string{"hello"}})
	if !errors.Is(err, ErrMissingEmbeddingProvider) {
		t.Fatalf("Embed error = %v, want ErrMissingEmbeddingProvider", err)
	}
}

func TestClientEmbedValidatesBeforeProvider(t *testing.T) {
	provider := &fakeEmbeddingProvider{}
	client := New(WithEmbeddingProvider(provider))
	_, err := client.Embed(context.Background(), EmbeddingRequest{Model: "embed-a"})
	if !errors.Is(err, ErrInvalidEmbeddingRequest) {
		t.Fatalf("Embed error = %v, want ErrInvalidEmbeddingRequest", err)
	}
	if len(provider.requests) != 0 {
		t.Fatalf("provider was called for invalid request: %d", len(provider.requests))
	}
}

func TestRedactReplacesSensitiveValues(t *testing.T) {
	got := Redact("token sk-test password secret")
	want := "token [REDACTED] [REDACTED] [REDACTED]"
	if got != want {
		t.Fatalf("Redact() = %q, want %q", got, want)
	}
}
