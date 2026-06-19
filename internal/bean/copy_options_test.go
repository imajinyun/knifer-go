package bean

import (
	"strconv"
	"testing"
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
