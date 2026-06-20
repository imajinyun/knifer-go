# vai: AI adapter helpers

`vai` provides provider-neutral chat and embedding helpers. It defines small interfaces for callers to inject their own AI providers while keeping `go-knifer` free of provider SDK dependencies.

## When to use

Use `vai` when application code needs a stable internal contract for chat or embedding calls, but provider selection belongs to the application boundary.

Use a provider SDK directly when you need provider-specific streaming, tool calls, retry policy, authentication flows, or advanced request fields that are not part of the `vai` MVP.

## Provider injection

`vai` has no built-in network provider. It does not read API keys, create HTTP clients, or use environment variables. Tests and applications provide behavior by implementing `ChatProvider` or `EmbeddingProvider`.

```go
package main

import (
	"context"
	"fmt"

	"github.com/imajinyun/go-knifer/vai"
)

type chatProvider struct{}

func (chatProvider) Chat(ctx context.Context, request vai.ChatRequest) (vai.ChatResponse, error) {
	return vai.ChatResponse{Message: vai.Message{Role: vai.RoleAssistant, Content: "hello gopher"}}, nil
}

func main() {
	client := vai.New(vai.WithChatProvider(chatProvider{}))
	response, err := client.Chat(context.Background(), vai.ChatRequest{
		Model:    "example-chat",
		Messages: []vai.Message{{Role: vai.RoleUser, Content: "hello"}},
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(response.Message.Content)
}
```

## Chat example

For one-off calls, use `vai.Chat` with an injected provider:

```go
response, err := vai.Chat(context.Background(), chatProvider{}, vai.ChatRequest{
	Model:    "example-chat",
	Messages: []vai.Message{{Role: vai.RoleUser, Content: "ping"}},
})
if err != nil {
	panic(err)
}
fmt.Println(response.Message.Role, response.Message.Content)
```

`ChatRequest.Validate` requires a model, at least one message, and non-empty message roles and content.

## Embedding example

```go
type embeddingProvider struct{}

func (embeddingProvider) Embed(ctx context.Context, request vai.EmbeddingRequest) (vai.EmbeddingResponse, error) {
	return vai.EmbeddingResponse{Vectors: [][]float32{{0.1, 0.2, 0.3}}}, nil
}

response, err := vai.Embed(context.Background(), embeddingProvider{}, vai.EmbeddingRequest{
	Model: "example-embedding",
	Input: []string{"hello"},
})
if err != nil {
	panic(err)
}
fmt.Println(len(response.Vectors), len(response.Vectors[0]))
```

`EmbeddingRequest.Validate` requires a model and at least one non-blank input string.

## Redaction and security boundary

`vai` does not log prompts, request metadata, model output, provider headers, or API keys. If callers need diagnostic text, redact obvious secret-like fields first:

```go
fmt.Println(vai.Redact("api key sk-test secret"))
// Output: api key [REDACTED] [REDACTED]
```

The redaction helper is intentionally small. It is not a substitute for a full data-loss-prevention policy.

## Out of scope

- Provider SDKs and network clients.
- API key discovery or environment-variable loading.
- Streaming responses.
- Tool or function calling.
- Retry, rate limiting, tracing, or logging middleware.
- Vector storage or retrieval-augmented generation.

## Validation

Focused checks:

```bash
go test ./internal/ai ./vai
go test -bench=. -benchmem -run=^$ ./internal/ai ./vai
```

Governance checks for public API and catalog changes:

```bash
UPDATE_API=1 make api-check
make docs-gen
make docs-check
make tools-check
make agent-check
make agent-security-check
```
