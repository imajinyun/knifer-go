package pinyin

import (
	"context"
	"testing"
)

type benchmarkProvider struct{}

func (benchmarkProvider) Convert(ctx context.Context, request ConvertRequest) (ConvertResponse, error) {
	return ConvertResponse{Text: request.Text, Output: "zhong guo", Tokens: []Token{{Text: "中", Syllables: []string{"zhong"}}, {Text: "国", Syllables: []string{"guo"}}}}, nil
}

func (benchmarkProvider) Initials(ctx context.Context, request InitialsRequest) (InitialsResponse, error) {
	return InitialsResponse{Text: request.Text, Output: "zg", Initials: []string{"z", "g"}}, nil
}

func BenchmarkClientConvert(b *testing.B) {
	client := New(WithProvider(benchmarkProvider{}))
	request := ConvertRequest{Text: "中国", Separator: " ", ToneStyle: ToneStylePlain, MaxInputRunes: 16}
	b.ReportAllocs()
	for b.Loop() {
		_, _ = client.Convert(context.Background(), request)
	}
}

func BenchmarkClientInitials(b *testing.B) {
	client := New(WithProvider(benchmarkProvider{}))
	request := InitialsRequest{Text: "中国", MaxInputRunes: 16}
	b.ReportAllocs()
	for b.Loop() {
		_, _ = client.Initials(context.Background(), request)
	}
}
