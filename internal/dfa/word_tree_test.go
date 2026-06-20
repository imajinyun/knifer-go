package dfa

import (
	"reflect"
	"testing"
)

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

func TestWordTreeOptions(t *testing.T) {
	tree := NewWordTreeWithOptions(WithCharFilter(func(r rune) bool { return r != '-' })).AddWord("t-io")
	if got := tree.MatchAll("tio"); !reflect.DeepEqual(got, []string{"tio"}) {
		t.Fatalf("NewWordTreeWithOptions MatchAll() = %#v", got)
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

func TestWordTreeBoundaryMethodsAndNilSafety(t *testing.T) {
	var nilTree *WordTree
	if !nilTree.IsEmpty() || nilTree.IsMatch("anything") {
		t.Fatal("nil WordTree should be empty and never match")
	}
	if got, ok := nilTree.Match("anything"); ok || got != "" {
		t.Fatalf("nil Match = %q ok=%v, want empty false", got, ok)
	}
	if got := nilTree.MatchAllWords("anything", -1, false, false); got != nil {
		t.Fatalf("nil MatchAllWords = %#v, want nil", got)
	}

	tree := NewWordTreeWithOptions(nil, WithCharFilter(nil))
	if !tree.IsEmpty() {
		t.Fatal("new WordTree should be empty")
	}
	tree.AddWords("", "重复", "重复")
	if tree.IsEmpty() {
		t.Fatal("tree with a word should not be empty")
	}
	found, ok := tree.MatchWord("这里有重复词")
	if !ok || found.String() != "重复" {
		t.Fatalf("MatchWord/String = %#v ok=%v", found, ok)
	}
	if got, ok := tree.Match("这里有重复词"); !ok || got != "重复" {
		t.Fatalf("Match = %q ok=%v", got, ok)
	}
	tree.SetCharFilter(nil).SetCharFilter(func(r rune) bool { return r != '_' })
	tree.AddWord("a_b")
	if got := tree.MatchAll("ab"); !reflect.DeepEqual(got, []string{"ab"}) {
		t.Fatalf("SetCharFilter custom match = %#v", got)
	}

	blank := NewWordTree().AddWord("")
	if !blank.IsEmpty() || blank.IsMatch("anything") {
		t.Fatal("empty word should not create a matching tree")
	}
}

func TestWordTreeFilterBoundaries(t *testing.T) {
	tree := NewWordTree().AddWords("秘密", "密")
	if got := tree.Filter("", true, nil); got != "" {
		t.Fatalf("Filter empty = %q", got)
	}
	if got := tree.Filter("公开信息", true, nil); got != "公开信息" {
		t.Fatalf("Filter without match = %q", got)
	}
	if got := tree.Filter("秘密和密", false, func(word FoundWord) string {
		return "[" + word.Word + "]"
	}); got != "[秘密]和[密]" {
		t.Fatalf("Filter non-greedy custom processor = %q", got)
	}
	if got := tree.Filter("秘密和密", true, func(word FoundWord) string {
		return "[" + word.Word + "]"
	}); got != "[秘密]和[密]" {
		t.Fatalf("Filter greedy custom processor = %q", got)
	}
}
