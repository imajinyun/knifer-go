package dfa

import (
	"errors"
	"testing"
)

func TestAnyHelpersUseJSONProviders(t *testing.T) {
	tree := NewWordTree().AddWords("secret")
	marshalCalls := 0
	unmarshalCalls := 0
	marshal := func(any) ([]byte, error) {
		marshalCalls++
		return []byte(`{"text":"secret"}`), nil
	}
	unmarshal := func(_ []byte, dst any) error {
		unmarshalCalls++
		dst.(*struct {
			Text string `json:"text"`
		}).Text = "***"
		return nil
	}
	if !ContainsAnyWithOptions(struct{}{}, WithMatcher(tree), WithJSONMarshal(marshal)) {
		t.Fatal("ContainsAnyWithOptions should use marshal provider")
	}
	if got := GetFoundAllAnyWithOptions(struct{}{}, WithMatcher(tree), WithJSONMarshal(marshal)); len(got) != 1 || got[0].Word != "secret" {
		t.Fatalf("GetFoundAllAnyWithOptions = %#v", got)
	}
	got, err := FilterAnyWithOptions(struct {
		Text string `json:"text"`
	}{}, true, nil,
		WithMatcher(tree), WithJSONMarshal(marshal), WithJSONUnmarshal(unmarshal))
	if err != nil {
		t.Fatalf("FilterAnyWithOptions: %v", err)
	}
	if got.Text != "***" || marshalCalls != 3 || unmarshalCalls != 1 {
		t.Fatalf("providers got=%+v marshalCalls=%d unmarshalCalls=%d", got, marshalCalls, unmarshalCalls)
	}
}

func TestAnyHelpersProviderErrorContracts(t *testing.T) {
	tree := NewWordTree().AddWords("secret")
	sentinel := errors.New("json provider failed")
	failingMarshal := func(any) ([]byte, error) { return nil, sentinel }
	if ContainsAnyWithOptions(struct{}{}, WithMatcher(tree), WithJSONMarshal(failingMarshal)) {
		t.Fatal("ContainsAnyWithOptions should not match when marshal provider fails")
	}
	if got, ok := GetFoundFirstAnyWithOptions(struct{}{}, WithMatcher(tree), WithJSONMarshal(failingMarshal)); ok || got.String() != "" {
		t.Fatalf("GetFoundFirstAnyWithOptions marshal failure = %#v ok=%v", got, ok)
	}
	if got := GetFoundAllAnyWithOptions(struct{}{}, WithMatcher(tree), WithJSONMarshal(failingMarshal)); len(got) != 0 {
		t.Fatalf("GetFoundAllAnyWithOptions marshal failure = %#v", got)
	}

	_, err := FilterAnyWithOptions(struct{}{}, true, nil, WithMatcher(tree), WithJSONMarshal(failingMarshal))
	if !errors.Is(err, sentinel) {
		t.Fatalf("FilterAnyWithOptions marshal failure err = %v, want sentinel", err)
	}
	_, err = FilterAnyWithOptions(struct {
		Text string `json:"text"`
	}{Text: "secret"}, true, nil, WithMatcher(tree), WithJSONUnmarshal(func([]byte, any) error { return sentinel }))
	if !errors.Is(err, sentinel) {
		t.Fatalf("FilterAnyWithOptions unmarshal failure err = %v, want sentinel", err)
	}

	if got := jsonTextWithMarshal(map[string]string{"text": "secret"}, nil); got == "" {
		t.Fatal("jsonTextWithMarshal nil provider should fall back to encoding/json")
	}
	if got := jsonTextWithMarshal("raw secret", failingMarshal); got != "raw secret" {
		t.Fatalf("jsonTextWithMarshal string = %q", got)
	}
}
