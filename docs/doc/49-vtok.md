# vtok: tokenization adapter helpers

`vtok` provides provider-neutral text tokenization and keyword extraction helpers. It defines a small interface for callers to inject their own tokenizer provider while keeping `go-knifer` free of dictionary, segmentation, ranking, network-client, and credential dependencies.

## When to use

Use `vtok` when application code needs a stable internal contract for text tokenization or keyword extraction, but dictionary choice, segmentation behavior, ranking, stop-word handling, and provider-specific behavior belong to the application boundary.

Use a dedicated NLP or tokenizer library directly when you need built-in dictionaries, language detection, model loading, synonym expansion, stop-word management, or streaming text processing that is not part of the `vtok` MVP.

## Provider injection

`vtok` has no built-in tokenizer provider. It does not import dictionaries, read environment variables, open network connections, or read local files. Tests and applications provide behavior by implementing `Provider`.

```go
package main

import (
	"context"
	"fmt"

	"github.com/imajinyun/go-knifer/vtok"
)

type tokenizeProvider struct{}

func (tokenizeProvider) Tokenize(ctx context.Context, request vtok.TokenizeRequest) (vtok.TokenizeResponse, error) {
	return vtok.TokenizeResponse{
		Text: request.Text,
		Tokens: []vtok.Token{
			{Text: "南京", Start: 0, End: 2, Position: 0},
			{Text: "长江大桥", Start: 3, End: 7, Position: 1},
		},
	}, nil
}

func (tokenizeProvider) Keywords(ctx context.Context, request vtok.KeywordsRequest) (vtok.KeywordsResponse, error) {
	return vtok.KeywordsResponse{
		Text:     request.Text,
		Keywords: []vtok.Keyword{{Text: "南京", Score: 0.9}, {Text: "长江大桥", Score: 0.8}},
	}, nil
}

func main() {
	client := vtok.New(vtok.WithProvider(tokenizeProvider{}))
	response, err := client.Tokenize(context.Background(), vtok.TokenizeRequest{Text: "南京市长江大桥", Mode: vtok.ModePrecise})
	if err != nil {
		panic(err)
	}
	fmt.Println(response.Tokens[0].Text)
}
```

## Tokenization example

For one-off calls, use `vtok.Tokenize` with an injected provider:

```go
response, err := vtok.Tokenize(context.Background(), tokenizeProvider{}, vtok.TokenizeRequest{
	Text:          "南京市长江大桥",
	Mode:          vtok.ModePrecise,
	MaxInputRunes: 64,
	MaxTokens:     16,
})
if err != nil {
	panic(err)
}
fmt.Println(response.Tokens[0].Text)
```

`TokenizeRequest.Validate` requires non-blank text, rejects NUL bytes, accepts `ModeDefault`, `ModePrecise`, `ModeSearch`, and `ModeFull`, and requires `MaxInputRunes` and `MaxTokens` to be non-negative. When `MaxInputRunes` is greater than zero, input text whose rune count exceeds the limit is rejected with `ErrInputLimitExceeded`. When `MaxTokens` is greater than zero, provider responses whose token count exceeds the limit are rejected with `ErrTokenLimitExceeded`.

## Keyword example

```go
response, err := vtok.Keywords(context.Background(), tokenizeProvider{}, vtok.KeywordsRequest{
	Text:          "南京市长江大桥",
	Limit:         2,
	MaxInputRunes: 64,
})
if err != nil {
	panic(err)
}
fmt.Println(response.Keywords[0].Text)
```

`KeywordsRequest.Validate` requires non-blank text, rejects NUL bytes, and requires `Limit` and `MaxInputRunes` to be non-negative.

## Safety boundary

`vtok` treats input text, tokenizer mode, keyword limits, and token limits as validation inputs only. It does not load dictionaries, segment text, rank keywords, normalize language variants, call remote services, read local files, log input text, or maintain hidden global provider state. Providers are responsible for dictionary selection, segmentation, ranking, stop-word behavior, synonym expansion, caching, tracing, and deployment-specific privacy controls.

Requests and responses are defensively copied around provider calls so callers and providers can mutate their own values without sharing slices or maps unexpectedly.

## Out of scope

- Built-in tokenizer dictionaries, ranking algorithms, or NLP libraries.
- Language detection, normalization, stop-word filtering, or synonym expansion.
- Network-backed providers, credential discovery, or local dictionary file loading.
- Streaming tokenization, model lifecycle, logging, metrics, tracing, cache lifecycle, or provider pooling.

## Validation

Focused checks:

```bash
go test ./internal/tokenize ./vtok
go test -bench=. -benchmem -run=^$ ./internal/tokenize ./vtok
```

Governance checks for public API and catalog changes:

```bash
UPDATE_API=1 make api-check
make docs-gen
make docs-check
make tools-check
make agent-check
```
