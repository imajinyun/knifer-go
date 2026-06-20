# vhan: Han text adapter helpers

`vhan` provides provider-neutral Chinese-to-pinyin conversion helpers. It defines a small interface for callers to inject their own pinyin provider while keeping `go-knifer` free of dictionary, tokenizer, network-client, and credential dependencies.

## When to use

Use `vhan` when application code needs a stable internal contract for pinyin conversion or initials extraction, but dictionary choice, polyphone handling, phrase ranking, segmentation, and provider-specific behavior belong to the application boundary.

Use a dedicated NLP or pinyin library directly when you need built-in dictionaries, word segmentation, polyphone disambiguation, locale-specific romanization, or streaming text processing that is not part of the `vhan` MVP.

## Provider injection

`vhan` has no built-in pinyin provider. It does not import dictionaries, read environment variables, open network connections, or read local files. Tests and applications provide behavior by implementing `Provider`.

```go
package main

import (
	"context"
	"fmt"

	"github.com/imajinyun/go-knifer/vhan"
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
go test -bench=. -benchmem -run=^$ ./internal/pinyin ./vhan
```

Governance checks for public API and catalog changes:

```bash
UPDATE_API=1 make api-check
make docs-gen
make docs-check
make tools-check
make agent-check
```
