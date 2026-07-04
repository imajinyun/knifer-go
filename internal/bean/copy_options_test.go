package bean

import (
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestCopyPropertiesWithParserOptions(t *testing.T) {
	src := map[string]any{
		"age":   "custom-int",
		"admin": "custom-bool",
		"score": "custom-float",
		"quota": "custom-uint",
	}
	type target struct {
		Age   int
		Admin bool
		Score float64
		Quota uint
	}
	var dst target
	var intCalled, boolCalled, floatCalled, uintCalled int
	err := CopyProperties(src, &dst,
		WithIntParser(func(text string, base, bits int) (int64, error) {
			intCalled++
			if text == "custom-int" {
				return 42, nil
			}
			return strconv.ParseInt(text, base, bits)
		}),
		WithBoolParser(func(text string) (bool, error) {
			boolCalled++
			return text == "custom-bool", nil
		}),
		WithFloatParser(func(text string, bits int) (float64, error) {
			floatCalled++
			if text == "custom-float" {
				return 9.5, nil
			}
			return strconv.ParseFloat(text, bits)
		}),
		WithUintParser(func(text string, base, bits int) (uint64, error) {
			uintCalled++
			if text == "custom-uint" {
				return 7, nil
			}
			return strconv.ParseUint(text, base, bits)
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	if dst != (target{Age: 42, Admin: true, Score: 9.5, Quota: 7}) {
		t.Fatalf("CopyProperties dst = %+v", dst)
	}
	if intCalled != 1 || boolCalled != 1 || floatCalled != 1 || uintCalled != 1 {
		t.Fatalf("parser calls int=%d bool=%d float=%d uint=%d", intCalled, boolCalled, floatCalled, uintCalled)
	}
}

func TestDecodeWithDecodeHook(t *testing.T) {
	type target struct {
		Created time.Time
	}
	var dst target
	called := 0
	err := Decode(map[string]any{"created": "2026-06-22"}, &dst,
		WithDecodeHook(func(from, to reflect.Type, value any) (any, error) {
			called++
			if from.Kind() == reflect.String && to == reflect.TypeOf(time.Time{}) {
				return time.Parse(time.DateOnly, value.(string))
			}
			return value, nil
		}),
	)
	if err != nil {
		t.Fatalf("Decode() with hook error = %v", err)
	}
	if called != 1 {
		t.Fatalf("hook calls = %d, want 1", called)
	}
	if got := dst.Created.Format(time.DateOnly); got != "2026-06-22" {
		t.Fatalf("Created = %q", got)
	}
}

func TestNilDecodeHookOptionDoesNotClearPreviousHook(t *testing.T) {
	type target struct {
		Created time.Time
	}
	var dst target
	called := 0
	err := Decode(map[string]any{"created": "2026-06-22"}, &dst,
		WithDecodeHook(func(from, to reflect.Type, value any) (any, error) {
			called++
			if from.Kind() == reflect.String && to == reflect.TypeOf(time.Time{}) {
				return time.Parse(time.DateOnly, value.(string))
			}
			return value, nil
		}),
		WithDecodeHook(nil),
	)
	if err != nil {
		t.Fatalf("Decode() with nil hook after custom hook error = %v", err)
	}
	if called != 1 {
		t.Fatalf("hook calls = %d, want 1", called)
	}
	if got := dst.Created.Format(time.DateOnly); got != "2026-06-22" {
		t.Fatalf("Created = %q", got)
	}
}

func TestDecodeWithComposedDecodeHooks(t *testing.T) {
	type target struct {
		Created time.Time
		Delay   time.Duration
	}
	var dst target
	err := Decode(map[string]any{"created": "2026-06-22", "delay": "150ms"}, &dst,
		WithDecodeHook(ComposeDecodeHook(
			StringToTimeHook(time.DateOnly),
			StringToDurationHook(),
		)),
	)
	if err != nil {
		t.Fatalf("Decode() with composed hooks error = %v", err)
	}
	if got := dst.Created.Format(time.DateOnly); got != "2026-06-22" {
		t.Fatalf("Created = %q", got)
	}
	if dst.Delay != 150*time.Millisecond {
		t.Fatalf("Delay = %v, want 150ms", dst.Delay)
	}
}

func TestDecodeRejectsUnsafeReflectNumericConversion(t *testing.T) {
	type source struct {
		Count int16
	}
	type target struct {
		Count int8
	}

	var dst target
	err := CopyProperties(source{Count: 128}, &dst, WithWeaklyTyped(false))
	if err == nil || !strings.Contains(err.Error(), "integer overflow") {
		t.Fatalf("CopyProperties() error = %v, want integer overflow", err)
	}
	assertBeanInvalidInput(t, err)

	err = CopyProperties(source{Count: 127}, &dst, WithWeaklyTyped(false))
	if err != nil {
		t.Fatalf("CopyProperties() safe conversion error = %v", err)
	}
	if dst.Count != 127 {
		t.Fatalf("Count = %d, want 127", dst.Count)
	}
}

func TestDecodeHookRejectsUnsafeReflectNumericConversion(t *testing.T) {
	type target struct {
		Count uint8
	}
	var dst target
	err := Decode(map[string]any{"count": "ignored"}, &dst,
		WithDecodeHook(func(from, to reflect.Type, value any) (any, error) {
			if to.Kind() == reflect.Uint8 {
				return int8(-1), nil
			}
			return value, nil
		}),
	)
	if err == nil || !strings.Contains(err.Error(), "negative value") {
		t.Fatalf("Decode() error = %v, want negative value", err)
	}
	assertBeanInvalidInput(t, err)
}

func TestDecodeReportsNestedFieldPath(t *testing.T) {
	type item struct {
		Count int
	}
	type target struct {
		Items []item
	}
	var dst target
	err := Decode(map[string]any{"items": []any{map[string]any{"count": "bad"}}}, &dst)
	if err == nil {
		t.Fatal("Decode() error = nil, want nested conversion error")
	}
	for _, want := range []string{"bean: set field Items", "index 0", "bean: set field Count"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("Decode() error = %q, want path fragment %q", err.Error(), want)
		}
	}
}

func TestWeaklyTypedDisabled(t *testing.T) {
	var dst targetProfile
	err := CopyProperties(map[string]any{"age": "42"}, &dst, WithWeaklyTyped(false))
	if err == nil {
		t.Fatal("expected strict assignment error")
	}
}

func TestOptionFallbacksAndIgnoreZero(t *testing.T) {
	type target struct {
		Name  string
		Age   int
		Admin bool
		Score float64
		Quota uint
	}

	var dst target
	err := CopyProperties(map[string]any{
		"Name":  "alice",
		"Age":   "42",
		"Admin": "yes",
		"Score": "9.5",
		"Quota": "7",
	}, &dst,
		func(o *Options) {
			o.ParseBool = nil
			o.ParseInt = nil
			o.ParseUint = nil
			o.ParseFloat = nil
		},
	)
	if err != nil {
		t.Fatalf("CopyProperties() fallback parser error = %v", err)
	}
	if dst != (target{Name: "alice", Age: 42, Admin: true, Score: 9.5, Quota: 7}) {
		t.Fatalf("fallback parser dst = %+v", dst)
	}

	dst = target{Name: "kept", Age: 11, Admin: true, Score: 1.5, Quota: 3}
	err = CopyProperties(target{}, &dst, WithIgnoreZero(true))
	if err != nil {
		t.Fatalf("CopyProperties() ignore zero error = %v", err)
	}
	if dst != (target{Name: "kept", Age: 11, Admin: true, Score: 1.5, Quota: 3}) {
		t.Fatalf("ignore zero dst = %+v", dst)
	}
}

func TestNilParserOptionsKeepDefaults(t *testing.T) {
	options := NewOptions()
	WithBoolParser(nil)(&options)
	WithIntParser(nil)(&options)
	WithUintParser(nil)(&options)
	WithFloatParser(nil)(&options)

	if got, err := options.ParseBool("on"); err != nil || !got {
		t.Fatalf("ParseBool(on) = %v, %v", got, err)
	}
	if got, err := options.ParseInt("8", 10, 64); err != nil || got != 8 {
		t.Fatalf("ParseInt(8) = %v, %v", got, err)
	}
	if got, err := options.ParseUint("9", 10, 64); err != nil || got != 9 {
		t.Fatalf("ParseUint(9) = %v, %v", got, err)
	}
	if got, err := options.ParseFloat("1.25", 64); err != nil || got != 1.25 {
		t.Fatalf("ParseFloat(1.25) = %v, %v", got, err)
	}
}

func TestResultSortedMetadata(t *testing.T) {
	result := Result{
		Matched: []string{"name", "age"},
		Skipped: []string{"zero", "empty"},
		Unused:  []string{"z", "a"},
	}
	result.sort()

	assertEqualStrings(t, []string{"age", "name"}, result.Matched)
	assertEqualStrings(t, []string{"empty", "zero"}, result.Skipped)
	assertEqualStrings(t, []string{"a", "z"}, result.Unused)
}

func TestWithStrictUnusedOption(t *testing.T) {
	cfg := applyOptions(WithStrictUnused(true))
	if !cfg.StrictUnused {
		t.Fatal("WithStrictUnused(true) did not enable strict unused handling")
	}

	cfg = applyOptions(WithStrictUnused(false))
	if cfg.StrictUnused {
		t.Fatal("WithStrictUnused(false) did not disable strict unused handling")
	}
}

func TestBeanCopyDecodeMergeSemanticMatrix(t *testing.T) {
	type profile struct {
		Name string
		Age  int
	}

	var copied profile
	if err := CopyProperties(profile{Name: "alice", Age: 30}, &copied); err != nil {
		t.Fatalf("CopyProperties() error = %v", err)
	}
	if copied != (profile{Name: "alice", Age: 30}) {
		t.Fatalf("CopyProperties() copied = %+v", copied)
	}

	var decoded profile
	result, err := DecodeResult(map[string]any{"name": "bob", "age": "40", "extra": true}, &decoded)
	if err != nil {
		t.Fatalf("DecodeResult() error = %v", err)
	}
	if decoded != (profile{Name: "bob", Age: 40}) {
		t.Fatalf("DecodeResult() decoded = %+v", decoded)
	}
	assertEqualStrings(t, []string{"age", "name"}, result.Matched)
	assertEqualStrings(t, []string{"extra"}, result.Unused)

	merged := profile{Name: "base", Age: 10}
	mergeResult, err := MergeResult(&merged, map[string]any{"name": "first"}, map[string]any{"age": "50"})
	if err != nil {
		t.Fatalf("MergeResult() error = %v", err)
	}
	if merged != (profile{Name: "first", Age: 50}) {
		t.Fatalf("MergeResult() merged = %+v", merged)
	}
	assertEqualStrings(t, []string{"age", "name"}, mergeResult.Matched)
}

func TestMergeStrategyContract(t *testing.T) {
	type profile struct {
		Name  string
		Tags  []string
		Score int
	}

	dst := profile{Name: "base", Tags: []string{"old"}, Score: 10}
	result, err := MergeResultWithOptions(&dst, []any{
		map[string]any{"name": "", "tags": []string{"new"}},
		map[string]any{"score": 0},
	}, WithIgnoreZero(true))
	if err != nil {
		t.Fatalf("MergeResultWithOptions() error = %v", err)
	}
	if dst.Name != "base" || dst.Score != 10 {
		t.Fatalf("MergeResultWithOptions() zero overwrite dst = %+v", dst)
	}
	assertEqualStrings(t, []string{"name", "score"}, result.Skipped)

	result, err = MergeResult(&dst, map[string]any{"tags": []string{"latest"}, "score": "12"})
	if err != nil {
		t.Fatalf("MergeResult() error = %v", err)
	}
	if dst.Score != 12 || len(dst.Tags) != 1 || dst.Tags[0] != "latest" {
		t.Fatalf("MergeResult() replace/convert dst = %+v", dst)
	}
	assertEqualStrings(t, []string{"score", "tags"}, result.Matched)
}

func TestCopyContractFieldTagsAndRequiredGaps(t *testing.T) {
	type source struct {
		Name string `bean:"full_name"`
		Age  int    `bean:"-"`
	}
	type target struct {
		Name string `bean:"full_name"`
		Age  int
	}

	var dst target
	result, err := copyPropertiesResult(source{Name: "alice", Age: 30}, &dst)
	if err != nil {
		t.Fatalf("copyPropertiesResult() error = %v", err)
	}
	if dst != (target{Name: "alice"}) {
		t.Fatalf("copyPropertiesResult() dst = %+v", dst)
	}
	assertEqualStrings(t, []string{"full_name"}, result.Matched)
	assertEqualStrings(t, []string{}, result.Unused)
}

func TestDecodeContractEmbeddedPointerNilAndUnused(t *testing.T) {
	type embedded struct {
		Name string `bean:"name"`
	}
	type target struct {
		embedded
		Age  *int `bean:"age"`
		Note any  `bean:"note"`
	}

	var dst target
	result, err := DecodeResult(map[string]any{
		"name":  "carol",
		"age":   "33",
		"note":  nil,
		"extra": true,
	}, &dst)
	if err != nil {
		t.Fatalf("DecodeResult() embedded/pointer/nil contract error = %v", err)
	}
	if dst.Name != "carol" || dst.Age == nil || *dst.Age != 33 || dst.Note != nil {
		t.Fatalf("DecodeResult() embedded/pointer/nil dst = %+v", dst)
	}
	assertEqualStrings(t, []string{"age", "name", "note"}, result.Matched)
	assertEqualStrings(t, []string{"extra"}, result.Unused)

	var strict target
	if _, err := DecodeResult(map[string]any{"extra": true}, &strict, WithStrictUnused(true)); err == nil {
		t.Fatal("DecodeResult() strict unused error = nil")
	}
}

func TestMergeContractEmptySkipAndMapReplace(t *testing.T) {
	type target struct {
		Name  string
		Attrs map[string]int
	}

	dst := target{Name: "base", Attrs: map[string]int{"old": 1}}
	result, err := MergeResultWithOptions(&dst, []any{
		map[string]any{"name": "", "attrs": map[string]int{}},
	}, WithIgnoreEmpty(true))
	if err != nil {
		t.Fatalf("MergeResultWithOptions() ignore empty contract error = %v", err)
	}
	if dst.Name != "base" || dst.Attrs["old"] != 1 {
		t.Fatalf("MergeResultWithOptions() ignore empty dst = %+v", dst)
	}
	assertEqualStrings(t, []string{"attrs", "name"}, result.Skipped)

	result, err = MergeResult(&dst, map[string]any{"attrs": map[string]int{"new": 2}})
	if err != nil {
		t.Fatalf("MergeResult() map replace contract error = %v", err)
	}
	if len(dst.Attrs) != 1 || dst.Attrs["new"] != 2 {
		t.Fatalf("MergeResult() map replace dst = %+v", dst)
	}
	assertEqualStrings(t, []string{"attrs"}, result.Matched)
}
