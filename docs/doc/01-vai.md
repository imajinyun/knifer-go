# vai: AI adapter helpers

`vai` provides provider-neutral chat and embedding helpers. It defines small interfaces for callers to inject their own AI providers while keeping `knifer-go` free of provider SDK dependencies.

## Golden path APIs

The first-choice API set for this facade is kept in sync with `ai-context.json` and the generated tools catalog.

- `Chat`
- `New`
- `Redact`
- `Embed`
- `WithChatProvider`

## Which helper should I use?

| Goal | Start with | Notes |
| --- | --- | --- |
| Keep a reusable AI adapter | `New` with `WithChatProvider` and/or `WithEmbeddingProvider` | Use when application code sends multiple requests through the same provider boundary. |
| Send one chat request | `Chat` | Validates model and messages, then delegates to the injected chat provider. |
| Generate embeddings once | `Embed` | Validates model and non-blank input strings, then delegates to the injected embedding provider. |
| Represent chat roles | `RoleSystem`, `RoleUser`, `RoleAssistant` | Keeps provider-specific role strings out of callers. |
| Carry provider metadata | `Usage`, `ProviderMetadata` | Use for token accounting and low-cardinality provider details when available. |
| Redact diagnostic text | `Redact` | Small helper for examples and logs; not a full DLP system. |
| Handle stable errors | `ErrInvalidChatRequest`, `ErrInvalidEmbeddingRequest`, `ErrMissingChatProvider`, `ErrMissingEmbeddingProvider` | Use `errors.Is` for request-validation or missing-provider branches. |

## AI adapter safety checklist

- Always inject providers; `vai` does not read API keys, create HTTP clients, or call external services by itself.
- Pass contexts with deadlines for provider calls so request paths can cancel expensive model work.
- Validate and redact prompts, metadata, and outputs before logging. `Redact` only catches obvious secret-like tokens.
- Keep retry, rate limiting, tracing, billing controls, and provider-specific safety settings in the provider layer.
- Bound prompt and embedding input sizes before calling providers when costs or latency matter.
- Do not treat provider output as trusted code, SQL, shell, HTML, or policy decisions without downstream validation.

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

	"github.com/imajinyun/knifer-go/vai"
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

## When not to use vai

- Use a provider SDK directly when you need streaming, tool/function calling, structured-output schemas, multimodal inputs, retries, or provider-specific request fields.
- Use a RAG/vector-store library when the task includes retrieval, chunking, indexing, or vector database operations.
- Use a policy/safety gateway when prompts and outputs require centralized moderation, audit, or data-loss-prevention controls.
- Avoid provider-neutral wrappers when provider-specific features are the main requirement and hiding them would make behavior unclear.

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

## Related packages

- Use `vhttp` or `vresty` when an AI provider adapter needs HTTP transport helpers.
- Use `vjson` when provider payloads need JSON formatting, inspection, or fixture generation.
- Use `verr` and `vlog` when adapter failures need structured error handling and diagnostics.

## Benchmarks and trade-offs

- Provider-neutral validation and defensive request shape checks are tiny compared with real model latency, but they keep tests and adapter contracts stable.
- The short `Chat` and `Embed` helpers are concise for one-off calls. Reuse `New(...)` when a service shares provider configuration.
- Keeping SDKs out of knifer-go avoids dependency bloat, but applications must implement provider adapters and document provider-specific behavior.
- Redaction is intentionally conservative and lightweight; comprehensive DLP belongs outside the facade.

## FAQ

### Does `vai` call OpenAI, Doubao, or another provider directly?

No. It has no built-in provider SDK or network client. Applications inject `ChatProvider` and `EmbeddingProvider` implementations.

### Where should API keys be loaded?

At the application/provider boundary. `vai` does not read environment variables or credentials.

### Is `Redact` enough for production logging?

No. It is a small helper for obvious secret-like tokens. Production systems should use a project-specific redaction and DLP policy.

### Can I use `vai` for streaming or tool calls?

Not through the current facade. Use the provider SDK directly or build an application adapter with a richer contract.
