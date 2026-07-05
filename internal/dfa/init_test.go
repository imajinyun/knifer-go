package dfa

import (
	"reflect"
	"sync"
	"testing"
)

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

func TestPackageLevelQueryHelpers(t *testing.T) {
	InitString("alpha,beta,alphabet", DefaultSeparator)
	if !IsInited() {
		t.Fatal("InitString should mark package matcher as initialized")
	}
	if !ContainsAny(map[string]any{"text": "contains beta"}) {
		t.Fatal("ContainsAny should match JSON text")
	}
	first, ok := GetFoundFirst("say alpha now")
	if !ok || first.Word != "alpha" || first.String() != "alpha" {
		t.Fatalf("GetFoundFirst = %#v ok=%v", first, ok)
	}
	firstAny, ok := GetFoundFirstAny(map[string]string{"text": "say beta"})
	if !ok || firstAny.Word != "beta" {
		t.Fatalf("GetFoundFirstAny = %#v ok=%v", firstAny, ok)
	}
	all := GetFoundAll("alpha beta")
	if len(all) != 2 || all[0].Word != "alpha" || all[1].Word != "beta" {
		t.Fatalf("GetFoundAll = %#v", all)
	}
	mode := GetFoundAllMode("alphabet", true, true)
	if got := []string{mode[0].Word, mode[1].Word}; !reflect.DeepEqual(got, []string{"alpha", "alphabet"}) {
		t.Fatalf("GetFoundAllMode dense+greedy = %#v", mode)
	}
	allAny := GetFoundAllAny(map[string]string{"text": "alpha beta"})
	if len(allAny) != 2 {
		t.Fatalf("GetFoundAllAny = %#v", allAny)
	}
	if got := FilterMode("alpha beta", true, func(word FoundWord) string { return "<" + word.Word + ">" }); got != "<alphabet>a" {
		t.Fatalf("FilterMode = %q", got)
	}
}

func TestPackageSetCharFilterAndAsyncRunnerFallback(t *testing.T) {
	Init([]string{"a-b"})
	if !Contains("ab") {
		t.Fatal("default matcher should ignore stop runes")
	}
	SetCharFilter(func(r rune) bool { return r != '-' })
	if !Contains("ab") {
		t.Fatal("SetCharFilter should update package matcher filter")
	}
	SetCharFilter(nil)
	if !Contains("ab") {
		t.Fatal("SetCharFilter(nil) should leave matcher usable")
	}

	ConfigureAsyncRunner(func(func()) {})
	ConfigureAsyncRunner(nil)
	done := make(chan struct{})
	ConfigureAsyncRunner(func(fn func()) {
		fn()
		close(done)
	})
	t.Cleanup(ResetAsyncRunner)
	InitAsync([]string{"async-fallback"})
	<-done
	if !Contains("async fallback") {
		t.Fatal("InitAsync should use configured deterministic runner")
	}
}

func TestPackageMatcherConcurrentInitAndQuery(t *testing.T) {
	ResetAsyncRunner()
	t.Cleanup(ResetAsyncRunner)

	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				Init([]string{"alpha", "beta", "secret"})
				SetCharFilter(func(r rune) bool { return r != '-' })
				InitStringWithOptions("a-b,runner", DefaultSeparator, WithCharFilter(func(r rune) bool { return r != '-' }))
			}
		}()
	}
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				ConfigureAsyncRunner(func(fn func()) { fn() })
				InitAsync([]string{"async"})
				InitStringAsync("runner", DefaultSeparator)
				ResetAsyncRunner()
			}
		}()
	}
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				_ = IsInited()
				_ = Contains("alpha secret")
				_, _ = GetFoundFirst("alpha secret")
				_ = GetFoundAll("alpha beta")
				_ = GetFoundAllMode("alphabet", true, true)
				_ = Filter("alpha beta")
			}
		}()
	}
	wg.Wait()
}
