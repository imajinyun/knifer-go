package vbean_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/imajinyun/knifer-go/vbean"
)

func TestFacadeBeanOptions(t *testing.T) {
	type customTagged struct {
		Name string `db:"user_name"`
		Age  int    `db:"age"`
	}
	got, err := vbean.ToMap(customTagged{Name: "casey", Age: 0},
		vbean.WithTagNames("db"),
		vbean.WithIgnoreZero(true),
	)
	if err != nil {
		t.Fatalf("ToMap() with options error = %v", err)
	}
	if got["user_name"] != "casey" {
		t.Fatalf("ToMap() user_name = %#v", got["user_name"])
	}
	if _, ok := got["age"]; ok {
		t.Fatalf("ToMap() should skip zero age with WithIgnoreZero: %#v", got)
	}

	var dst userModel
	if err := vbean.ToStruct(map[string]any{"FULL_NAME": "drew", "age": "21"}, &dst,
		vbean.WithCaseInsensitive(true),
		vbean.WithWeaklyTyped(true),
	); err != nil {
		t.Fatalf("ToStruct() with options error = %v", err)
	}
	if dst.Name != "drew" || dst.Age != 21 {
		t.Fatalf("ToStruct() with options dst = %+v", dst)
	}

	dst = userModel{Name: "existing", Age: 30}
	if err := vbean.Copy(map[string]any{"full_name": "", "age": "22"}, &dst, vbean.WithIgnoreEmpty(true)); err != nil {
		t.Fatalf("Copy() with WithIgnoreEmpty error = %v", err)
	}
	if dst.Name != "existing" || dst.Age != 22 {
		t.Fatalf("Copy() WithIgnoreEmpty dst = %+v", dst)
	}

	var strict userModel
	if err := vbean.CopyProperties(map[string]any{"age": "23"}, &strict, vbean.WithWeaklyTyped(false)); err == nil {
		t.Fatal("CopyProperties() WithWeaklyTyped(false) error = nil, want strict assignment error")
	}
}

func TestFacadeBeanNewOptions(t *testing.T) {
	opts := vbean.NewOptions()
	_ = opts
}

func TestFacadeBeanParserOptions(t *testing.T) {
	parser1 := func(s string) (bool, error) { return s == "yes", nil }
	opt1 := vbean.WithBoolParser(parser1)
	if opt1 == nil {
		t.Fatal("WithBoolParser returned nil")
	}

	parser2 := func(s string, base, bits int) (int64, error) { return 0, nil }
	opt2 := vbean.WithIntParser(parser2)
	if opt2 == nil {
		t.Fatal("WithIntParser returned nil")
	}

	parser3 := func(s string, base, bits int) (uint64, error) { return 0, nil }
	opt3 := vbean.WithUintParser(parser3)
	if opt3 == nil {
		t.Fatal("WithUintParser returned nil")
	}

	parser4 := func(s string, bits int) (float64, error) { return 0, nil }
	opt4 := vbean.WithFloatParser(parser4)
	if opt4 == nil {
		t.Fatal("WithFloatParser returned nil")
	}

	opt5 := vbean.WithDecodeHook(func(from, to reflect.Type, value any) (any, error) {
		return value, nil
	})
	if opt5 == nil {
		t.Fatal("WithDecodeHook returned nil")
	}
}

func TestFacadeBeanDecodeHookApplied(t *testing.T) {
	type target struct {
		Created time.Time
	}
	var dst target
	if err := vbean.Decode(map[string]any{"created": "2026-06-22"}, &dst,
		vbean.WithDecodeHook(func(from, to reflect.Type, value any) (any, error) {
			if from.Kind() == reflect.String && to == reflect.TypeOf(time.Time{}) {
				return time.Parse(time.DateOnly, value.(string))
			}
			return value, nil
		}),
	); err != nil {
		t.Fatalf("Decode() with hook error = %v", err)
	}
	if got := dst.Created.Format(time.DateOnly); got != "2026-06-22" {
		t.Fatalf("Created = %q", got)
	}
}

func TestFacadeBeanComposedDecodeHooksApplied(t *testing.T) {
	type target struct {
		Created time.Time
		Delay   time.Duration
	}
	var dst target
	if err := vbean.Decode(map[string]any{"created": "2026-06-22", "delay": "250ms"}, &dst,
		vbean.WithDecodeHook(vbean.ComposeDecodeHook(
			vbean.StringToTimeHook(time.DateOnly),
			vbean.StringToDurationHook(),
		)),
	); err != nil {
		t.Fatalf("Decode() with composed hooks error = %v", err)
	}
	if got := dst.Created.Format(time.DateOnly); got != "2026-06-22" {
		t.Fatalf("Created = %q", got)
	}
	if dst.Delay != 250*time.Millisecond {
		t.Fatalf("Delay = %v, want 250ms", dst.Delay)
	}
}

func TestFacadeBeanParseParserApplied(t *testing.T) {
	type P struct {
		Active bool
		Number int64
		UVal   uint64
		FVal   float64
	}

	var dst P
	src := map[string]any{
		"active": "true",
		"number": "42",
		"uval":   "100",
		"fval":   "3.14",
	}
	err := vbean.CopyProperties(src, &dst,
		vbean.WithWeaklyTyped(true),
	)
	if err != nil {
		t.Fatalf("CopyProperties with weak types: %v", err)
	}
	if dst.Active != true || dst.Number != 42 || dst.UVal != 100 {
		t.Fatalf("parsed values: %+v", dst)
	}
	if dst.FVal < 3.13 || dst.FVal > 3.15 {
		t.Fatalf("float parsed = %v", dst.FVal)
	}
}

func TestFacadeBeanFillMap(t *testing.T) {
	type S struct {
		Name string
		Age  int
	}
	src := S{Name: "test", Age: 25}
	dst := make(map[string]any)
	if err := vbean.FillMap(src, dst); err != nil {
		t.Fatalf("FillMap error = %v", err)
	}
	if dst["Name"] != "test" || dst["Age"] != 25 {
		t.Fatalf("FillMap dst = %#v", dst)
	}
}

func TestFacadeBeanWithStrictUnused(t *testing.T) {
	cfg := vbean.NewOptions()
	vbean.WithStrictUnused(true)(&cfg)
	if !cfg.StrictUnused {
		t.Fatal("WithStrictUnused(true) did not enable strict unused handling")
	}
}
