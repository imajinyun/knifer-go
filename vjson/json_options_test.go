package vjson_test

import (
	stdjson "encoding/json"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/imajinyun/go-knifer/vjson"
	"github.com/imajinyun/go-knifer/vxml"
)

func TestFacadeJSONOptions(t *testing.T) {
	compact, err := vjson.ToStr(map[string]any{"name": "go", "empty": nil}, vjson.WithIgnoreNullValue(true))
	if err != nil {
		t.Fatalf("ToStr with options: %v", err)
	}
	if strings.Contains(compact, "empty") {
		t.Fatalf("ToStr WithIgnoreNullValue should omit null field: %s", compact)
	}
	formatted := vjson.FormatWithOptions(`{"a":1,"b":{"c":2}}`, vjson.WithFormatIndentWidth(2), vjson.WithFormatSpaceAfterKey(false))
	if !strings.Contains(formatted, "\n  \"a\":1") || strings.Contains(formatted, "\": 1") {
		t.Fatalf("FormatWithOptions = %q", formatted)
	}
	cfg := vjson.NewConfig()
	cfg.IgnoreNullValue = true
	obj, err := vjson.ParseObjWithOptions(map[string]any{"name": "go", "empty": nil}, vjson.WithParseConfig(cfg))
	if err != nil {
		t.Fatalf("ParseObjWithOptions: %v", err)
	}
	if got := obj.ToString(); strings.Contains(got, "empty") {
		t.Fatalf("ParseObjWithOptions should apply config: %s", got)
	}
}

func TestFacadeJSONEncodeProviderOptions(t *testing.T) {
	t.Run("array config skips null values", func(t *testing.T) {
		cfg := vjson.CreateConfig()
		cfg.IgnoreNullValue = true
		arr := vjson.NewArrayWithConfig(cfg).Add("go").Add(nil).AddAll("knifer")
		if arr.Len() != 2 || arr.GetString(1) != "knifer" {
			t.Fatalf("NewArrayWithConfig/AddAll = %s len=%d", arr.ToString(), arr.Len())
		}
		if empty := vjson.NewArray(); empty.Len() != 0 || empty.ToString() != "[]" {
			t.Fatalf("NewArray() = %s len=%d", empty.ToString(), empty.Len())
		}
	})

	t.Run("config and scalar providers flow through serialization", func(t *testing.T) {
		cfg := vjson.NewConfig()
		cfg.IgnoreNullValue = true
		compact, err := vjson.ToStrWithConfig(map[string]any{"name": "go", "empty": nil}, cfg)
		if err != nil {
			t.Fatalf("ToStrWithConfig: %v", err)
		}
		if strings.Contains(compact, "empty") || !strings.Contains(compact, `"name":"go"`) {
			t.Fatalf("ToStrWithConfig = %s", compact)
		}

		pretty, err := vjson.ToPrettyStrWithConfig(map[string]any{"name": "go"}, cfg)
		if err != nil {
			t.Fatalf("ToPrettyStrWithConfig: %v", err)
		}
		if !strings.Contains(pretty, "\n    \"name\":") {
			t.Fatalf("ToPrettyStrWithConfig = %q", pretty)
		}

		indented, err := vjson.ToStrIndent(map[string]any{"name": "go"}, 2, vjson.WithConfig(cfg))
		if err != nil {
			t.Fatalf("ToStrIndent: %v", err)
		}
		if !strings.Contains(indented, "\n  \"name\":") {
			t.Fatalf("ToStrIndent = %q", indented)
		}

		formattedTime, err := vjson.ToStr(time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC), vjson.WithDateFormat("2006-01-02"))
		if err != nil {
			t.Fatalf("ToStr WithDateFormat: %v", err)
		}
		if formattedTime != `"2024-01-02"` {
			t.Fatalf("formatted time = %s", formattedTime)
		}

		custom, err := vjson.ToStr(
			struct{ Name string }{Name: "ignored"},
			vjson.WithMarshalFunc(func(any) ([]byte, error) { return []byte(`{"name":"marshal"}`), nil }),
			vjson.WithUnmarshalFunc(func(_ []byte, dst any) error {
				*(dst.(*any)) = map[string]any{"name": "unmarshal"}
				return nil
			}),
		)
		if err != nil {
			t.Fatalf("ToStr with marshal/unmarshal providers: %v", err)
		}
		if custom != `{"name":"unmarshal"}` {
			t.Fatalf("custom providers = %s", custom)
		}

		_, err = vjson.ToStr("fallback",
			vjson.WithSprintFunc(func(any) string { return "sprint" }),
			vjson.WithParseIntFunc(func(string, int, int) (int64, error) { return 1, nil }),
			vjson.WithParseFloatFunc(func(string, int) (float64, error) { return 1.5, nil }),
			vjson.WithParseBoolFunc(func(string) (bool, error) { return true, nil }),
			vjson.WithFormatIntFunc(func(int64, int) string { return "1" }),
			vjson.WithFormatFloatFunc(func(float64, byte, int, int) string { return "1.5" }),
			vjson.WithDecoderFactory(stdjson.NewDecoder),
		)
		if err != nil {
			t.Fatalf("provider option wrapper smoke: %v", err)
		}
	})

	t.Run("decoder and scalar conversion providers are stored in config", func(t *testing.T) {
		calledDecoder := false
		parsed, err := vjson.ParseWithConfig(`{"ignored":true}`, (&vjson.Config{}))
		if err != nil || parsed == nil {
			t.Fatalf("ParseWithConfig smoke = %#v, %v", parsed, err)
		}

		obj, err := vjson.ParseObjWithConfig(`{"ignored":true}`, &vjson.Config{
			DecoderFactory: func(io.Reader) *stdjson.Decoder {
				calledDecoder = true
				return stdjson.NewDecoder(strings.NewReader(`{"provided":true}`))
			},
		})
		if err != nil {
			t.Fatalf("ParseObjWithConfig: %v", err)
		}
		if !calledDecoder || !obj.GetBool("provided") {
			t.Fatalf("decoder provider called=%v obj=%s", calledDecoder, obj.ToString())
		}

		converted, err := vjson.ParseObjWithConfig(map[string]any{"n": "10", "f": "1.5", "b": "enabled", "raw": int64(7), "num": 2.25}, &vjson.Config{
			ParseIntFunc:    func(string, int, int) (int64, error) { return 42, nil },
			ParseFloatFunc:  func(string, int) (float64, error) { return 3.5, nil },
			ParseBoolFunc:   func(string) (bool, error) { return true, nil },
			FormatIntFunc:   func(int64, int) string { return "int!" },
			FormatFloatFunc: func(float64, byte, int, int) string { return "float!" },
		})
		if err != nil {
			t.Fatalf("ParseObjWithConfig conversions: %v", err)
		}
		if converted.GetInt64("n") != 42 || converted.GetFloat64("f") != 3.5 || !converted.GetBool("b") {
			t.Fatalf("custom parsers not applied: %s", converted.ToString())
		}
		if converted.GetString("raw") != "int!" || converted.GetString("num") != "float!" {
			t.Fatalf("custom formatters not applied: raw=%q num=%q", converted.GetString("raw"), converted.GetString("num"))
		}
	})
}

func TestFacadeJSONParseValidPathAndBeanOptions(t *testing.T) {
	t.Run("parse and valid providers", func(t *testing.T) {
		calledUnmarshal := false
		arr, err := vjson.ParseArrayWithOptions(`[]`, vjson.WithParseUnmarshalFunc(func(_ []byte, dst any) error {
			calledUnmarshal = true
			*(dst.(*any)) = []any{"provided"}
			return nil
		}))
		if err != nil {
			t.Fatalf("ParseArrayWithOptions: %v", err)
		}
		if !calledUnmarshal || arr.GetString(0) != "provided" {
			t.Fatalf("unmarshal provider called=%v arr=%s", calledUnmarshal, arr.ToString())
		}

		cfg := vjson.NewConfig()
		cfg.IgnoreNullValue = true
		parsed, err := vjson.ParseWithConfig(map[string]any{"name": "go", "empty": nil}, cfg)
		if err != nil {
			t.Fatalf("ParseWithConfig: %v", err)
		}
		if got := parsed.(*vjson.Object).ToString(); strings.Contains(got, "empty") {
			t.Fatalf("ParseWithConfig should apply config: %s", got)
		}

		arrWithConfig, err := vjson.ParseArrayWithConfig([]any{"go", nil}, cfg)
		if err != nil {
			t.Fatalf("ParseArrayWithConfig: %v", err)
		}
		if arrWithConfig.Len() != 1 || arrWithConfig.GetString(0) != "go" {
			t.Fatalf("ParseArrayWithConfig = %s len=%d", arrWithConfig.ToString(), arrWithConfig.Len())
		}

		validCalled := false
		if !vjson.IsJSONWithOptions("not-json", vjson.WithJSONValidFunc(func(b []byte) bool {
			validCalled = string(b) == "not-json"
			return true
		})) || !validCalled {
			t.Fatal("IsJSONWithOptions should use custom validator")
		}
		if !vjson.IsObjWithOptions(`{"x":1}`, vjson.WithJSONValidFunc(func([]byte) bool { return true })) {
			t.Fatal("IsObjWithOptions should accept valid object with custom validator")
		}
		if !vjson.IsArray(`[1]`) || !vjson.IsArrayWithOptions(`[1]`, vjson.WithJSONValidFunc(func([]byte) bool { return true })) {
			t.Fatal("array validators should accept JSON arrays")
		}
	})

	t.Run("path and bean options", func(t *testing.T) {
		root := vjson.NewObject()
		if err := vjson.PutByPath(root, "user.name", "go"); err != nil {
			t.Fatalf("PutByPath: %v", err)
		}
		if got := vjson.GetByPath(root, "user.name"); got != "go" {
			t.Fatalf("GetByPath after PutByPath = %v", got)
		}

		type user struct{ Name string }
		var bean user
		beanCalled := false
		if err := vjson.ToBeanWithOptions(`{"name":"ignored"}`, &bean, vjson.WithBeanUnmarshalFunc(func(_ []byte, dst any) error {
			beanCalled = true
			dst.(*user).Name = "provided"
			return nil
		})); err != nil {
			t.Fatalf("ToBeanWithOptions: %v", err)
		}
		if !beanCalled || bean.Name != "provided" {
			t.Fatalf("ToBeanWithOptions provider called=%v bean=%+v", beanCalled, bean)
		}

		var list []user
		listCalled := false
		if err := vjson.ToListWithOptions(`[]`, &list, vjson.WithBeanUnmarshalFunc(func(_ []byte, dst any) error {
			listCalled = true
			*(dst.(*[]user)) = []user{{Name: "provided"}}
			return nil
		})); err != nil {
			t.Fatalf("ToListWithOptions: %v", err)
		}
		if !listCalled || len(list) != 1 || list[0].Name != "provided" {
			t.Fatalf("ToListWithOptions provider called=%v list=%+v", listCalled, list)
		}
	})
}

func TestFacadeJSONXMLConversionOptions(t *testing.T) {
	xmlObj, err := vjson.XMLToJSONWithOptions(`<root><n>42</n></root>`, vxml.WithScalarIntParser(func(string, int, int) (int64, error) { return 7, nil }))
	if err != nil {
		t.Fatalf("XMLToJSONWithOptions: %v", err)
	}
	if got := xmlObj.GetJSONObject("root").GetInt64("n"); got != 7 {
		t.Fatalf("XMLToJSONWithOptions scalar parser = %d", got)
	}

	xmlStr, err := vjson.ToXMLWithOptions(map[string]any{"name": "go"}, "user", vxml.WithOmitDeclaration(true), vxml.WithNamespace("urn:test"))
	if err != nil {
		t.Fatalf("ToXMLWithOptions: %v", err)
	}
	if !strings.Contains(xmlStr, `<user xmlns="urn:test">`) || !strings.Contains(xmlStr, `<name>go</name>`) {
		t.Fatalf("ToXMLWithOptions = %q", xmlStr)
	}
}
