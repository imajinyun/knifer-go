package vdfa

import (
	"testing"
)

func TestFacadeContainsAny(t *testing.T) {
	InitString("secret", DefaultSeparator)
	if !ContainsAny(struct {
		Text string `json:"text"`
	}{Text: "has secret inside"}) {
		t.Fatal("ContainsAny should find 'secret' in JSON")
	}
	if ContainsAny(struct {
		Text string `json:"text"`
	}{Text: "clean"}) {
		t.Fatal("ContainsAny should not find word in clean text")
	}
}

func TestFacadeGetFoundFirstWithOptions(t *testing.T) {
	tree := NewWordTree().AddWords("secret")
	got, ok := GetFoundFirstWithOptions("this is a secret message", WithMatcher(tree))
	if !ok || got.Word != "secret" {
		t.Fatalf("GetFoundFirstWithOptions = %q, %v", got.Word, ok)
	}
	_, ok = GetFoundFirstWithOptions("clean text", WithMatcher(tree))
	if ok {
		t.Fatal("GetFoundFirstWithOptions should not find word in clean text")
	}
}

func TestFacadeGetFoundFirstAny(t *testing.T) {
	InitString("secret", DefaultSeparator)
	got, ok := GetFoundFirstAny(struct {
		Text string `json:"text"`
	}{Text: "a secret"})
	if !ok || got.Word != "secret" {
		t.Fatalf("GetFoundFirstAny = %q, %v", got.Word, ok)
	}
}

func TestFacadeGetFoundAllWithOptions(t *testing.T) {
	tree := NewWordTree().AddWords("secret", "message")
	got := GetFoundAllWithOptions("this secret contains a message", WithMatcher(tree))
	if len(got) != 2 {
		t.Fatalf("GetFoundAllWithOptions found %d, want 2", len(got))
	}
}

func TestFacadeGetFoundAllModeWithOptions(t *testing.T) {
	tree := NewWordTree().AddWords("ab", "abc")
	got := GetFoundAllModeWithOptions("abc", true, true, WithMatcher(tree))
	if len(got) == 0 {
		t.Fatal("GetFoundAllModeWithOptions should find at least one match")
	}
}

func TestFacadeGetFoundAllAny(t *testing.T) {
	InitString("secret", DefaultSeparator)
	got := GetFoundAllAny(struct {
		Text string `json:"text"`
	}{Text: "a secret message"})
	if len(got) == 0 {
		t.Fatal("GetFoundAllAny should find at least one match")
	}
}

func TestFacadeFilterModeWithOptions(t *testing.T) {
	tree := NewWordTree().AddWords("secret")
	got := FilterModeWithOptions("a secret", true, DefaultProcessor, WithMatcher(tree))
	if got != "a ******" {
		t.Fatalf("FilterModeWithOptions = %q, want %q", got, "a ******")
	}
}
