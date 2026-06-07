package dfa

import (
	"reflect"
	"testing"
)

const sampleText = "我有一颗$大土^豆，刚出锅的"

func TestWordTreeMatchModes(t *testing.T) {
	tree := buildTestTree()
	tests := []struct {
		name    string
		density bool
		greed   bool
		want    []string
	}{
		{name: "standard", want: []string{"大", "土^豆", "刚出锅"}},
		{name: "density", density: true, want: []string{"大", "土^豆", "刚出锅", "出锅"}},
		{name: "greed", greed: true, want: []string{"大", "土^豆", "刚出锅"}},
		{name: "density greed", density: true, greed: true, want: []string{"大", "大土^豆", "土^豆", "刚出锅", "出锅"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tree.MatchAllMode(sampleText, -1, tt.density, tt.greed)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("MatchAllMode() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestWordTreeFoundWordMetadata(t *testing.T) {
	tree := NewWordTree().AddWords("赵", "赵阿", "赵阿三")
	got := tree.MatchAllWords("赵阿三在做什么", -1, true, true)
	if len(got) != 3 {
		t.Fatalf("len = %d", len(got))
	}
	assertFoundWord(t, got[0], "赵", "赵", 0, 0)
	assertFoundWord(t, got[1], "赵阿", "赵阿", 0, 1)
	assertFoundWord(t, got[2], "赵阿三", "赵阿三", 0, 2)
}

func TestStopRunes(t *testing.T) {
	tree := NewWordTree().AddWord("tio")
	got := tree.MatchAll("AAAAAAAt-ioBBBBBBB")
	want := []string{"t-io"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("MatchAll() = %#v, want %#v", got, want)
	}
}

func TestWordTreeOptions(t *testing.T) {
	tree := NewWordTreeWithOptions(WithCharFilter(func(r rune) bool { return r != '-' })).AddWord("t-io")
	if got := tree.MatchAll("tio"); !reflect.DeepEqual(got, []string{"tio"}) {
		t.Fatalf("NewWordTreeWithOptions MatchAll() = %#v", got)
	}
}

func TestInitWithOptions(t *testing.T) {
	InitWithOptions([]string{"t-io"}, WithCharFilter(func(r rune) bool { return r != '-' }))
	if !Contains("tio") {
		t.Fatal("InitWithOptions should apply custom char filter")
	}
	InitStringWithOptions("a-b", 0, WithCharFilter(func(r rune) bool { return r != '-' }))
	if !Contains("ab") {
		t.Fatal("InitStringWithOptions should apply custom char filter")
	}
}

func TestAsyncRunnerCanBeConfiguredAndReset(t *testing.T) {
	ResetAsyncRunner()
	t.Cleanup(ResetAsyncRunner)

	runs := 0
	ConfigureAsyncRunner(func(fn func()) {
		runs++
		fn()
	})
	InitAsync([]string{"async"})
	if runs != 1 || !Contains("async word") {
		t.Fatalf("InitAsync with configured runner runs=%d contains=%v", runs, Contains("async word"))
	}
	InitStringAsync("runner", DefaultSeparator)
	if runs != 2 || !Contains("runner word") {
		t.Fatalf("InitStringAsync with configured runner runs=%d contains=%v", runs, Contains("runner word"))
	}
	InitStringAsyncWithOptions("a-b", DefaultSeparator, WithCharFilter(func(r rune) bool { return r != '-' }))
	if runs != 3 || !Contains("ab") {
		t.Fatalf("InitStringAsyncWithOptions with configured runner runs=%d contains=%v", runs, Contains("ab"))
	}

	ResetAsyncRunner()
	Init([]string{"reset"})
	if !Contains("reset") {
		t.Fatal("ResetAsyncRunner should preserve synchronous Init behavior")
	}
}

func TestMatcherOptionsBypassPackageMatcher(t *testing.T) {
	Init([]string{"global"})
	matcher := NewWordTree().AddWords("local")
	if ContainsWithOptions("local", WithMatcher(matcher)) != true {
		t.Fatal("ContainsWithOptions should use provided matcher")
	}
	if Contains("local") {
		t.Fatal("per-call matcher should not mutate package matcher")
	}
	if got := FilterWithOptions("a local word", WithMatcher(matcher)); got != "a ***** word" {
		t.Fatalf("FilterWithOptions = %q", got)
	}
	found, ok := GetFoundFirstWithOptions("a local word", WithMatcherWords([]string{"local"}))
	if !ok || found.Word != "local" {
		t.Fatalf("GetFoundFirstWithOptions = %#v ok=%v", found, ok)
	}
}

func TestFilterAnyWithMatcherOptions(t *testing.T) {
	type payload struct {
		Text string `json:"text"`
	}
	Init([]string{"global"})
	got, err := FilterAnyWithOptions(payload{Text: "a local"}, true, nil, WithMatcherWords([]string{"local"}))
	if err != nil {
		t.Fatalf("FilterAnyWithOptions: %v", err)
	}
	if got.Text != "a *****" || Contains("local") {
		t.Fatalf("FilterAnyWithOptions = %#v globalContainsLocal=%v", got, Contains("local"))
	}
}

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

func TestAddWordWithFilteredRune(t *testing.T) {
	tree := NewWordTree().AddWord("hello(")
	if got := tree.MatchAllLimit("hello", -1); !reflect.DeepEqual(got, []string{"hello"}) {
		t.Fatalf("trailing filtered rune match = %#v", got)
	}

	tree = NewWordTree().AddWord("he(llo")
	if got := tree.MatchAllLimit("hello", -1); !reflect.DeepEqual(got, []string{"hello"}) {
		t.Fatalf("middle filtered rune match = %#v", got)
	}
}

func TestClear(t *testing.T) {
	tree := NewWordTree().AddWord("黑")
	if !contains(tree.MatchAll("黑大衣"), "黑") {
		t.Fatalf("expected initial match")
	}
	tree.Clear()
	tree.AddWords("黑大衣", "红色大衣")
	if !contains(tree.MatchAll("黑大衣"), "黑大衣") {
		t.Fatalf("expected 黑大衣 after clear")
	}
	if contains(tree.MatchAll("黑大衣"), "黑") {
		t.Fatalf("did not expect stale 黑 match")
	}
	if !contains(tree.MatchAll("红色大衣"), "红色大衣") {
		t.Fatalf("expected 红色大衣")
	}
}

func TestFilter(t *testing.T) {
	Init([]string{"大", "大土豆", "土豆", "刚出锅", "出锅"})
	got := Filter(sampleText)
	want := "我有一颗$****，***的"
	if got != want {
		t.Fatalf("Filter() = %q, want %q", got, want)
	}
}

func TestFilterGreedyLongest(t *testing.T) {
	Init([]string{"赵", "赵阿", "赵阿三"})
	got := Filter("赵阿三在做什么。")
	want := "***在做什么。"
	if got != want {
		t.Fatalf("Filter() = %q, want %q", got, want)
	}
}

func TestFilterAny(t *testing.T) {
	type payload struct {
		Text string `json:"text"`
		Num  int    `json:"num"`
	}
	Init([]string{"大", "大土豆", "土豆", "刚出锅", "出锅"})
	got, err := FilterAny(payload{Text: sampleText, Num: 100}, true, nil)
	if err != nil {
		t.Fatalf("FilterAny() error = %v", err)
	}
	if got.Text != "我有一颗$****，***的" || got.Num != 100 {
		t.Fatalf("FilterAny() = %#v", got)
	}
}

func TestFilterDoesNotMatchInsideDigits(t *testing.T) {
	Init([]string{"12宝宝龙", "34皮卡丘"})
	text := "creator_user_id=2000907612345839744"
	if got := Filter(text); got != text {
		t.Fatalf("Filter() = %q, want %q", got, text)
	}
}

func TestCustomProcessor(t *testing.T) {
	tree := NewWordTree().AddWords("bad")
	got := tree.Filter("a bad word", true, func(word FoundWord) string {
		return "[" + word.Word + "]"
	})
	if got != "a [bad] word" {
		t.Fatalf("custom filter = %q", got)
	}
}

func buildTestTree() *WordTree {
	return NewWordTree().AddWords("大", "大土豆", "土豆", "刚出锅", "出锅")
}

func assertFoundWord(t *testing.T, got FoundWord, word, found string, start, end int) {
	t.Helper()
	if got.Word != word || got.FoundWord != found || got.Start != start || got.End != end {
		t.Fatalf("FoundWord = %#v, want word=%q found=%q start=%d end=%d", got, word, found, start, end)
	}
}

func contains(values []string, value string) bool {
	for _, v := range values {
		if v == value {
			return true
		}
	}
	return false
}
