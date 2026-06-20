package tokenize

import (
	"context"
	"testing"
)

type benchmarkProvider struct{}

func (benchmarkProvider) Tokenize(ctx context.Context, request TokenizeRequest) (TokenizeResponse, error) {
	return TokenizeResponse{Text: request.Text, Tokens: []Token{{Text: "南京", Start: 0, End: 2}, {Text: "长江大桥", Start: 3, End: 7}}}, nil
}

func (benchmarkProvider) Keywords(ctx context.Context, request KeywordsRequest) (KeywordsResponse, error) {
	return KeywordsResponse{Text: request.Text, Keywords: []Keyword{{Text: "南京", Score: 0.9}, {Text: "长江大桥", Score: 0.8}}}, nil
}

func BenchmarkClientTokenize(b *testing.B) {
	client := New(WithProvider(benchmarkProvider{}))
	request := TokenizeRequest{Text: "南京市长江大桥", Mode: ModePrecise, MaxInputRunes: 16, MaxTokens: 8}
	b.ReportAllocs()
	for b.Loop() {
		_, _ = client.Tokenize(context.Background(), request)
	}
}

func BenchmarkClientKeywords(b *testing.B) {
	client := New(WithProvider(benchmarkProvider{}))
	request := KeywordsRequest{Text: "南京市长江大桥", Limit: 2, MaxInputRunes: 16}
	b.ReportAllocs()
	for b.Loop() {
		_, _ = client.Keywords(context.Background(), request)
	}
}
