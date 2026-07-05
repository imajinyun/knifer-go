# vhan: Han text adapter helpers

`vhan` provides provider-neutral Chinese-to-pinyin conversion helpers. It defines a small interface for callers to inject their own pinyin provider while keeping `knifer-go` free of dictionary, tokenizer, network-client, and credential dependencies.

## When to use

Use `vhan` when application code needs a stable internal contract for pinyin conversion or initials extraction, but dictionary choice, polyphone handling, phrase ranking, segmentation, and provider-specific behavior belong to the application boundary.

Use a dedicated NLP or pinyin library directly when you need built-in dictionaries, word segmentation, polyphone disambiguation, locale-specific romanization, or streaming text processing that is not part of the `vhan` MVP.

## Golden path APIs

The first-choice API set for this facade is kept in sync with `ai-context.json` and the generated tools catalog.

- `Convert`
- `New`
- `WithProvider`
- `Initials`

## Which helper should I use?

Choose helpers by whether you want a reusable client, a one-off provider call, or request validation.

| Need | Use | Notes |
| --- | --- | --- |
| Configure a reusable pinyin adapter | `New`, `WithProvider`, `Client.Convert`, `Client.Initials` | Good for application services that share one provider instance. |
| Convert Chinese text to pinyin once | `Convert` with a `Provider` | Keeps provider choice explicit at the call site. |
| Extract initials once | `Initials` with a `Provider` | Use for search keys, short labels, or display helpers when provider behavior is acceptable. |
| Validate request shape | `ConvertRequest.Validate`, `InitialsRequest.Validate` | Enforces non-blank text, NUL rejection, tone style, and optional input limits before provider work. |
| Control output shape | `Separator`, `ToneStyleDefault`, `ToneStylePlain`, `ToneStyleNumber`, `ToneStyleMark` | Provider remains responsible for dictionary, segmentation, and polyphone behavior. |

## Han adapter safety checklist

- Treat input text as potentially sensitive; do not log raw names, addresses, or free-form user text unless policy permits it.
- Set `MaxInputRunes` for user-controlled text so providers cannot receive unbounded payloads.
- Keep provider choice, dictionary version, and polyphone behavior outside the facade and visible in application wiring.
- Use context cancellation for providers that may call remote services or expensive local NLP libraries.
- Test with fake providers to keep examples deterministic and avoid hidden dictionary or network dependencies.
- Preserve defensive-copy expectations: callers and providers should not rely on shared slices or maps.

## When not to use vhan

- Use a dedicated pinyin or NLP library directly when you need built-in dictionaries, segmentation, polyphone disambiguation, phrase ranking, or locale-specific romanization.
- Use provider APIs directly when provider-specific options, batch operations, streaming, model loading, or diagnostics are central to the workflow.
- Avoid treating initials or pinyin output as stable identifiers unless provider behavior, dictionary version, and normalization rules are part of the data contract.
- Avoid sending unbounded or sensitive user text to remote providers without input limits, cancellation, logging policy, and privacy review.
- Use simple hand-written mappings when the conversion domain is tiny, fixed, and easier to review than a provider adapter.

## Provider injection

`vhan` has no built-in pinyin provider. It does not import dictionaries, read environment variables, open network connections, or read local files. Tests and applications provide behavior by implementing `Provider`.

```go
package main

import (
	"context"
	"fmt"

	"github.com/imajinyun/knifer-go/vhan"
)

type pinyinProvider struct{}

func (pinyinProvider) Convert(ctx context.Context, request vhan.ConvertRequest) (vhan.ConvertResponse, error) {
	return vhan.ConvertResponse{
		Text:   request.Text,
		Output: "zhong guo",
		Tokens: []vhan.Token{{Text: "中", Syllables: []string{"zhong"}}, {Text: "国", Syllables: []string{"guo"}}},
	}, nil
}

func (pinyinProvider) Initials(ctx context.Context, request vhan.InitialsRequest) (vhan.InitialsResponse, error) {
	return vhan.InitialsResponse{Text: request.Text, Output: "zg", Initials: []string{"z", "g"}}, nil
}

func main() {
	client := vhan.New(vhan.WithProvider(pinyinProvider{}))
	response, err := client.Convert(context.Background(), vhan.ConvertRequest{Text: "中国", Separator: " ", ToneStyle: vhan.ToneStylePlain})
	if err != nil {
		panic(err)
	}
	fmt.Println(response.Output)
}
```

## Conversion example

For one-off calls, use `vhan.Convert` with an injected provider:

```go
response, err := vhan.Convert(context.Background(), pinyinProvider{}, vhan.ConvertRequest{
	Text:          "中国",
	Separator:     " ",
	ToneStyle:     vhan.ToneStylePlain,
	MaxInputRunes: 64,
})
if err != nil {
	panic(err)
}
fmt.Println(response.Output)
```

`ConvertRequest.Validate` requires non-blank text, rejects NUL bytes, accepts `ToneStyleDefault`, `ToneStylePlain`, `ToneStyleNumber`, and `ToneStyleMark`, and requires `MaxInputRunes` to be non-negative. When `MaxInputRunes` is greater than zero, input text whose rune count exceeds the limit is rejected with `ErrInputLimitExceeded`.

## Initials example

```go
response, err := vhan.Initials(context.Background(), pinyinProvider{}, vhan.InitialsRequest{
	Text:          "中国",
	MaxInputRunes: 64,
})
if err != nil {
	panic(err)
}
fmt.Println(response.Output)
```

`InitialsRequest.Validate` requires non-blank text, rejects NUL bytes, and enforces the optional input rune limit.

## Safety boundary

`vhan` treats input text, tone style, separators, and input size as validation inputs only. It does not load dictionaries, tokenize text, normalize language variants, resolve polyphones, call remote services, read local files, log input text, or maintain hidden global provider state. Providers are responsible for dictionary selection, phrase ranking, heteronym behavior, segmentation, caching, tracing, and deployment-specific privacy controls.

Requests and responses are defensively copied around provider calls so callers and providers can mutate their own values without sharing slices or maps unexpectedly.

## Related packages

- Use `vtok` when Chinese or mixed-language text needs tokenization or keyword extraction rather than pinyin conversion.
- Use `vstr` when the task is general string normalization, trimming, or similarity scoring.
- Use `vai` when pinyin conversion is delegated to an AI/provider adapter with request validation.

## Benchmarks and trade-offs

Benchmark the facade with fake providers and benchmark real providers separately:

```bash
go test -bench=. -benchmem -run=^$ ./internal/pinyin ./vhan
```

`vhan` measures only validation, defensive copying, client dispatch, and provider-interface overhead. Real pinyin quality and cost depend on the injected provider's dictionaries, segmentation, caching, and polyphone logic.

Provider neutrality keeps `knifer-go` small and deterministic, but it means applications own provider lifecycle, privacy controls, dictionary updates, and observability.

## FAQ

### Does vhan include a pinyin dictionary?

No. `vhan` defines the adapter contract and validation rules. Applications inject a provider that implements dictionary and conversion behavior.

### How should I test code using vhan?

Inject a fake `Provider` that returns deterministic `ConvertResponse` and `InitialsResponse` values. This avoids network, dictionary, locale, and model-version dependencies.

### Where should privacy controls live?

At the provider and application boundary. `vhan` validates and copies requests, but providers own logging, tracing, caching, and any external service calls.

## Out of scope

- Built-in pinyin dictionaries or NLP libraries.
- Chinese word segmentation, tokenization, or phrase ranking.
- Polyphone disambiguation beyond provider-defined behavior.
- Network-backed providers, credential discovery, or local dictionary file loading.
- Logging, metrics, tracing, cache lifecycle, or provider pooling.

## Validation

Focused checks:

```bash
go test ./internal/pinyin ./vhan
```

Governance checks for public API and catalog changes:

```bash
UPDATE_API=1 make api-check
make docs-gen
make docs-check
make tools-check
make agent-check
```
