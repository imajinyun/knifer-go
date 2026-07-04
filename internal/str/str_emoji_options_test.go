package str

import (
	"testing"
)

func TestWithEmojiMatcher(t *testing.T) {
	customMatcherCalled := false
	opts := []EmojiOption{WithEmojiMatcher(func(s string) bool {
		customMatcherCalled = true
		return s == "custom"
	})}
	if got := ContainsEmojiWithOptions("hello", opts...); got {
		t.Fatal("ContainsEmojiWithOptions with custom matcher should return false")
	}
	if !customMatcherCalled {
		t.Fatal("custom matcher was not called")
	}

	// WithEmojiMatcher as a standalone option function (coverage for the function itself)
	opt := WithEmojiMatcher(func(s string) bool { return len(s) > 0 })
	if opt == nil {
		t.Fatal("WithEmojiMatcher returned nil")
	}
}

func TestNilEmojiMatcherDoesNotClearPreviousMatcher(t *testing.T) {
	customMatcherCalled := false
	if !ContainsEmojiWithOptions("custom", WithEmojiMatcher(func(s string) bool {
		customMatcherCalled = true
		return s == "custom"
	}), WithEmojiMatcher(nil)) {
		t.Fatal("nil WithEmojiMatcher cleared previous matcher")
	}
	if !customMatcherCalled {
		t.Fatal("custom matcher was not called")
	}
}

func TestWithEmojiReplacer(t *testing.T) {
	customReplacerCalled := false
	opts := []EmojiOption{WithEmojiReplacer(func(s string) string {
		customReplacerCalled = true
		return "[emoji]"
	})}
	if got := RemoveEmojiWithOptions("hi😀", opts...); got != "[emoji]" {
		t.Fatalf("RemoveEmojiWithOptions with custom replacer = %q, want [emoji]", got)
	}
	if !customReplacerCalled {
		t.Fatal("custom replacer was not called")
	}

	opt := WithEmojiReplacer(func(s string) string { return "" })
	if opt == nil {
		t.Fatal("WithEmojiReplacer returned nil")
	}
}

func TestNilEmojiReplacerDoesNotClearPreviousReplacer(t *testing.T) {
	customReplacerCalled := false
	got := RemoveEmojiWithOptions("hi", WithEmojiReplacer(func(string) string {
		customReplacerCalled = true
		return "custom"
	}), WithEmojiReplacer(nil))
	if got != "custom" {
		t.Fatalf("nil WithEmojiReplacer cleared previous replacer: got %q", got)
	}
	if !customReplacerCalled {
		t.Fatal("custom replacer was not called")
	}
}
