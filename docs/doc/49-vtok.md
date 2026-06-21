# vtok: tokenization adapter helpers

`vtok` provides provider-neutral text tokenization and keyword extraction helpers. It defines a small interface for callers to inject their own tokenizer provider while keeping `go-knifer` free of dictionary, segmentation, ranking, network-client, and credential dependencies.

## When to use

Use `vtok` when application code needs a stable internal contract for text tokenization or keyword extraction, but dictionary choice, segmentation behavior, ranking, stop-word handling, and provider-specific behavior belong to the application boundary.

Use a dedicated NLP or tokenizer library directly when you need built-in dictionaries, language detection, model loading, synonym expansion, stop-word management, or streaming text processing that is not part of the `vtok` MVP.

## Which helper should I use?

Choose helpers by whether you want a reusable client, a one-off provider call, tokenization, keyword extraction, or request validation.

| Need | Use | Notes |
| --- | --- | --- |
| Configure a reusable tokenizer adapter | `New`, `WithProvider`, `Client.Tokenize`, `Client.Keywords` | Good for application services that share one provider instance. |
| Tokenize text once | `Tokenize` with a `Provider` | Keeps dictionary and segmentation provider explicit at the call site. |
| Extract keywords once | `Keywords` with a `Provider` | Provider owns ranking, stop-word handling, and score meaning. |
| Validate tokenization requests | `TokenizeRequest.Validate` | Enforces non-blank text, NUL rejection, valid mode, input limits, and token limits. |
| Validate keyword requests | `KeywordsRequest.Validate` | Enforces non-blank text, NUL rejection, keyword limit, and input limits. |
| Select tokenization mode | `ModeDefault`, `ModePrecise`, `ModeSearch`, `ModeFull` | Mode semantics are provider-defined beyond request validation. |

## Tokenization adapter safety checklist

- Treat source text as potentially sensitive; avoid logging raw user input, documents, messages, or queries.
- Set `MaxInputRunes`, `MaxTokens`, and `Limit` for user-controlled or large text.
- Keep provider choice, dictionary version, stop-word policy, and ranking behavior visible in application wiring.
- Use context cancellation for providers that call remote services, load models, or perform expensive local NLP work.
- Test with fake providers to keep examples deterministic and independent of dictionary/model versions.
- Do not treat keyword scores as comparable across providers unless the provider contract guarantees it.

## Related packages

- Use `vhan` when Chinese text should be converted to pinyin or initials instead of tokenized.
- Use `vstr` when simple string normalization, splitting, or similarity checks are sufficient.
- Use `vai` when tokenization or keyword extraction is delegated to an AI/provider adapter.

## When not to use vtok

- Use a dedicated tokenizer, search, or NLP library directly when you need dictionaries, language detection, synonym expansion, stop-word management, model loading, or streaming text processing.
- Use provider APIs directly when provider-specific ranking, batch operations, diagnostics, or model configuration are central to the workflow.
- Avoid using token boundaries or keyword scores as durable identifiers unless provider behavior and versioning are part of the data contract.
- Avoid sending unbounded or sensitive documents to remote providers without input limits, cancellation, logging policy, and privacy review.
- Use simpler `strings` or `regexp` helpers when the task is fixed delimiter splitting, exact matching, or small deterministic extraction.

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

## Benchmarks and trade-offs

Benchmark facade overhead with fake providers and benchmark real tokenizers separately:

```bash
go test -bench=. -benchmem -run=^$ ./internal/tokenize ./vtok
```

`vtok` measures validation, defensive copying, client dispatch, and provider-interface overhead. Real tokenization and keyword extraction cost depends on the injected provider's dictionaries, ranking algorithms, model lifecycle, and caching.

Provider neutrality keeps `go-knifer` free of heavy NLP dependencies, but applications own provider lifecycle, privacy controls, dictionary/model updates, and observability.

## FAQ

### Does vtok include a tokenizer dictionary?

No. `vtok` defines the adapter contract and validation rules. Applications inject a provider for segmentation, ranking, stop words, and model behavior.

### How should I test code using vtok?

Inject a fake `Provider` that returns deterministic token and keyword responses. This keeps tests independent of local dictionaries, remote services, and model versions.

### Are keyword scores portable across providers?

Not necessarily. Treat scores as provider-defined unless your application owns and documents a normalized scoring contract.

## Out of scope

- Built-in tokenizer dictionaries, ranking algorithms, or NLP libraries.
- Language detection, normalization, stop-word filtering, or synonym expansion.
- Network-backed providers, credential discovery, or local dictionary file loading.
- Streaming tokenization, model lifecycle, logging, metrics, tracing, cache lifecycle, or provider pooling.

## Validation

Focused checks:

```bash
go test ./internal/tokenize ./vtok
```

Governance checks for public API and catalog changes:

```bash
UPDATE_API=1 make api-check
make docs-gen
make docs-check
make tools-check
make agent-check
```
