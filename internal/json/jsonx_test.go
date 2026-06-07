package json

import (
	stdjson "encoding/json"
	"io"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestObjectOrderPreserved(t *testing.T) {
	obj := NewJSONObject()
	obj.Set("c", 3).Set("a", 1).Set("b", 2)
	if got := strings.Join(obj.Keys(), ","); got != "c,a,b" {
		t.Fatalf("expect insertion order, got %s", got)
	}
	s := obj.String()
	if s != `{"c":3,"a":1,"b":2}` {
		t.Fatalf("compact: %s", s)
	}
}

func TestParseAndStringify(t *testing.T) {
	src := `{"name":"alice","age":30,"tags":["a","b"],"meta":{"x":1}}`
	v, err := Parse(src)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	obj, ok := v.(*JSONObject)
	if !ok {
		t.Fatalf("expect *JSONObject, got %T", v)
	}
	if obj.GetString("name") != "alice" {
		t.Fatalf("name: %v", obj.GetString("name"))
	}
	if obj.GetInt("age") != 30 {
		t.Fatalf("age")
	}
	arr := obj.GetJSONArray("tags")
	if arr == nil || arr.Len() != 2 || arr.GetString(0) != "a" {
		t.Fatalf("tags %v", arr)
	}
	if obj.GetJSONObject("meta").GetInt("x") != 1 {
		t.Fatalf("meta.x")
	}
	out := obj.String()
	if out != src {
		t.Fatalf("round-trip mismatch:\n  in : %s\n  out: %s", src, out)
	}
}

func TestParseRejectsTrailingContent(t *testing.T) {
	if _, err := Parse(`{"a":1} {"b":2}`); err == nil {
		t.Fatal("expected trailing content error")
	}
}

func TestPretty(t *testing.T) {
	obj := NewJSONObject().Set("a", 1).Set("b", NewJSONArray().Add(1).Add(2))
	out := obj.ToStringPretty()
	expect := "{\n    \"a\": 1,\n    \"b\": [\n        1,\n        2\n    ]\n}"
	if out != expect {
		t.Fatalf("pretty mismatch:\n%s\n--\n%s", out, expect)
	}
}

func TestFormatJSONStr(t *testing.T) {
	in := `{"a":1,"b":[1,2],"c":"x"}`
	out := FormatJSONStr(in)
	if !strings.Contains(out, "\n") {
		t.Fatalf("expect formatted: %q", out)
	}
	custom := FormatJSONStrWithOptions(in, WithFormatIndentWidth(2), WithFormatSpaceAfterKey(false))
	if !strings.Contains(custom, "\n  \"a\":1") {
		t.Fatalf("custom format = %q", custom)
	}
}

func TestIsJSONWithOptions(t *testing.T) {
	called := false
	valid := func(data []byte) bool {
		called = true
		return string(data) == "custom"
	}
	if !IsJSONWithOptions("custom", WithJSONValidFunc(valid)) || !called {
		t.Fatalf("IsJSONWithOptions called=%v", called)
	}
	if !IsJSONObjWithOptions("{custom}", WithJSONValidFunc(func([]byte) bool { return true })) {
		t.Fatal("IsJSONObjWithOptions should use custom validator")
	}
	if !IsJSONArrayWithOptions("[custom]", WithJSONValidFunc(func([]byte) bool { return true })) {
		t.Fatal("IsJSONArrayWithOptions should use custom validator")
	}
}

func TestNullHandling(t *testing.T) {
	obj := NewJSONObject().Set("a", nil)
	if !obj.IsNull("a") {
		t.Fatalf("expect a is null")
	}
	if obj.String() != `{"a":null}` {
		t.Fatalf("got %s", obj.String())
	}
}

func TestEncodeOptions(t *testing.T) {
	obj := map[string]any{"a": nil, "b": 1}
	out, err := ToJSONStr(obj, WithIgnoreNullValue(true))
	if err != nil {
		t.Fatalf("ToJSONStr: %v", err)
	}
	if out != `{"b":1}` {
		t.Fatalf("ignore null output = %s", out)
	}
	out, err = ToJSONPrettyStr(map[string]any{"a": 1}, WithIndent(2))
	if err != nil {
		t.Fatalf("ToJSONPrettyStr: %v", err)
	}
	if !strings.Contains(out, "\n  \"a\": 1") {
		t.Fatalf("pretty indent output = %s", out)
	}
	out, err = ToJSONStr(map[string]any{"t": time.Date(2026, 6, 2, 3, 4, 5, 0, time.UTC)}, WithDateFormat("2006-01-02"))
	if err != nil {
		t.Fatalf("ToJSONStr date: %v", err)
	}
	if !strings.Contains(out, "2026-06-02") {
		t.Fatalf("date output = %s", out)
	}
}

func TestEncodeOptionsUseMarshalFunc(t *testing.T) {
	type tagged struct {
		Name string `json:"name"`
	}
	called := false
	out, err := ToJSONStr(tagged{Name: "ignored"}, WithMarshalFunc(func(any) ([]byte, error) {
		called = true
		return []byte(`{"name":"provided"}`), nil
	}))
	if err != nil {
		t.Fatalf("ToJSONStr: %v", err)
	}
	if !called || out != `{"name":"provided"}` {
		t.Fatalf("marshal provider called=%v out=%s", called, out)
	}
}

func TestPathGetPut(t *testing.T) {
	src := `{"a":{"b":[10,20,{"c":"hit"}]}}`
	v, _ := Parse(src)
	if got := GetByPath(v, "a.b[2].c"); got != "hit" {
		t.Fatalf("path get: %v", got)
	}
	if got := GetByPath(v, "$.a.b[0]"); got != int64(10) {
		t.Fatalf("path get with $: %v", got)
	}
	obj := v.(*JSONObject)
	if err := obj.PutByPath("a.b[1]", "X"); err != nil {
		t.Fatalf("put: %v", err)
	}
	if got := obj.GetByPath("a.b[1]"); got != "X" {
		t.Fatalf("after put: %v", got)
	}
}

func TestPathCreatesIntermediate(t *testing.T) {
	obj := NewJSONObject()
	if err := obj.PutByPath("a.b.c", 42); err != nil {
		t.Fatalf("put: %v", err)
	}
	if obj.GetByPath("a.b.c") != int64(42) {
		t.Fatalf("nested put")
	}
}

func TestParseObjAndArrayErrors(t *testing.T) {
	if _, err := ParseObj(`[1,2]`); err == nil {
		t.Fatalf("expect error parsing array as obj")
	}
	if _, err := ParseArray(`{}`); err == nil {
		t.Fatalf("expect error parsing object as array")
	}
}

func TestParseObjAndArrayWithOptionsUseUnmarshalFunc(t *testing.T) {
	objCalled := false
	obj, err := ParseObjWithOptions(`{"ignored":true}`, WithParseUnmarshalFunc(func(_ []byte, dst any) error {
		objCalled = true
		*(dst.(*any)) = map[string]any{"provided": "yes"}
		return nil
	}))
	if err != nil {
		t.Fatalf("ParseObjWithOptions: %v", err)
	}
	if !objCalled || obj.GetString("provided") != "yes" {
		t.Fatalf("object unmarshal provider called=%v obj=%s", objCalled, obj.String())
	}

	arrCalled := false
	arr, err := ParseArrayWithOptions(`["ignored"]`, WithParseUnmarshalFunc(func(_ []byte, dst any) error {
		arrCalled = true
		*(dst.(*any)) = []any{"provided"}
		return nil
	}))
	if err != nil {
		t.Fatalf("ParseArrayWithOptions: %v", err)
	}
	if !arrCalled || arr.GetString(0) != "provided" {
		t.Fatalf("array unmarshal provider called=%v arr=%s", arrCalled, arr.String())
	}
}

func TestParseWithOptionsUsesDecoderFactory(t *testing.T) {
	called := false
	v, err := ParseWithOptions(`{"ignored":true}`, WithParseDecoderFactory(func(io.Reader) *stdjson.Decoder {
		called = true
		dec := stdjson.NewDecoder(strings.NewReader(`{"provided":"yes"}`))
		dec.UseNumber()
		return dec
	}))
	if err != nil {
		t.Fatalf("ParseWithOptions decoder factory: %v", err)
	}
	obj, ok := v.(*JSONObject)
	if !called || !ok || obj.GetString("provided") != "yes" {
		t.Fatalf("decoder factory called=%v value=%#v", called, v)
	}
	if _, err := ParseWithOptions(`{"ignored":true}`, WithParseDecoderFactory(func(io.Reader) *stdjson.Decoder { return nil })); err == nil {
		t.Fatal("nil decoder factory should fail")
	}
}

func TestWrapUsesConfigDecoderFactory(t *testing.T) {
	type tagged struct {
		Name string `json:"name"`
	}
	called := false
	out, err := ToJSONStr(tagged{Name: "ignored"}, WithDecoderFactory(func(io.Reader) *stdjson.Decoder {
		called = true
		dec := stdjson.NewDecoder(strings.NewReader(`{"name":"provided"}`))
		dec.UseNumber()
		return dec
	}))
	if err != nil {
		t.Fatalf("ToJSONStr: %v", err)
	}
	if !called || out != `{"name":"provided"}` {
		t.Fatalf("decoder factory called=%v out=%s", called, out)
	}
}

func TestJSONScalarProviders(t *testing.T) {
	cfg := NewConfig()
	cfg.ParseIntFunc = func(s string, base, bitSize int) (int64, error) {
		if s == "custom-int" {
			return 77, nil
		}
		return strconv.ParseInt(s, base, bitSize)
	}
	cfg.ParseFloatFunc = func(s string, bitSize int) (float64, error) {
		if s == "custom-float" {
			return 8.5, nil
		}
		return strconv.ParseFloat(s, bitSize)
	}
	cfg.ParseBoolFunc = func(s string) (bool, error) {
		if s == "yep" {
			return true, nil
		}
		return false, strconv.ErrSyntax
	}
	obj := NewJSONObjectWithConfig(cfg)
	obj.Set("int", "custom-int").Set("float", "custom-float").Set("bool", "yep")
	if got := obj.GetInt64("int"); got != 77 {
		t.Fatalf("custom int = %d", got)
	}
	if got := obj.GetFloat64("float"); got != 8.5 {
		t.Fatalf("custom float = %v", got)
	}
	if !obj.GetBool("bool") {
		t.Fatal("custom bool parser not used")
	}

	out, err := ToJSONStr(map[string]any{"n": int64(7), "f": 1.25},
		WithFormatIntFunc(func(v int64, base int) string { return strconv.FormatInt(v*10, base) }),
		WithFormatFloatFunc(func(v float64, fmtByte byte, prec, bitSize int) string {
			return strconv.FormatFloat(v*2, fmtByte, prec, bitSize)
		}),
	)
	if err != nil {
		t.Fatalf("ToJSONStr with scalar providers: %v", err)
	}
	if out != `{"f":2.5,"n":70}` && out != `{"n":70,"f":2.5}` {
		t.Fatalf("formatted json = %s", out)
	}

	out, err = ToJSONStr(map[customKey]string{{name: "k"}: "v"}, WithSprintFunc(func(any) string { return "custom-key" }))
	if err != nil {
		t.Fatalf("ToJSONStr with sprint provider: %v", err)
	}
	if out != `{"custom-key":"v"}` {
		t.Fatalf("sprint json = %s", out)
	}
}

type customKey struct{ name string }

func TestToBeanWithOptionsUsesDecoderFactory(t *testing.T) {
	type tagged struct {
		Name string `json:"name"`
	}
	called := false
	var out tagged
	if err := ToBeanWithOptions([]byte(`{"ignored":true}`), &out, WithBeanConfig(&Config{DecoderFactory: func(io.Reader) *stdjson.Decoder {
		called = true
		dec := stdjson.NewDecoder(strings.NewReader(`{"name":"provided"}`))
		dec.UseNumber()
		return dec
	}})); err != nil {
		t.Fatalf("ToBeanWithOptions decoder factory: %v", err)
	}
	if !called || out.Name != "provided" {
		t.Fatalf("decoder factory called=%v out=%#v", called, out)
	}
	if err := ToBeanWithOptions([]byte(`{"ignored":true}`), &out, WithBeanConfig(&Config{DecoderFactory: func(io.Reader) *stdjson.Decoder { return nil }})); err == nil {
		t.Fatal("nil decoder factory should fail")
	}
}

func TestIsJSONHelpers(t *testing.T) {
	if !IsJSON(`{"a":1}`) || !IsJSONObj(`{"a":1}`) {
		t.Fatalf("obj")
	}
	if !IsJSONArray(`[1,2]`) {
		t.Fatalf("array")
	}
	if IsJSON("not json") {
		t.Fatalf("invalid")
	}
}

func TestXMLToJSON(t *testing.T) {
	xmlStr := `<root><name>alice</name><age>30</age><tags>a</tags><tags>b</tags></root>`
	obj, err := XMLToJSON(xmlStr)
	if err != nil {
		t.Fatalf("xml->json: %v", err)
	}
	root := obj.GetJSONObject("root")
	if root == nil {
		t.Fatalf("missing root: %s", obj.String())
	}
	if root.GetString("name") != "alice" {
		t.Fatalf("name: %v", root.GetString("name"))
	}
	if root.GetInt("age") != 30 {
		t.Fatalf("age: %v", root.GetInt("age"))
	}
	tags := root.GetJSONArray("tags")
	if tags == nil || tags.Len() != 2 {
		t.Fatalf("tags: %v", tags)
	}
}

func TestJSONToXML(t *testing.T) {
	root := NewJSONObject().
		Set("name", "alice").
		Set("tags", NewJSONArray().Add("a").Add("b"))
	x, err := JSONToXML(root, "user")
	if err != nil {
		t.Fatalf("json->xml: %v", err)
	}
	want := "<user><name>alice</name><tags>a</tags><tags>b</tags></user>"
	if x != want {
		t.Fatalf("xml mismatch:\n got: %s\nwant: %s", x, want)
	}
}

func TestToBean(t *testing.T) {
	src := `{"name":"alice","age":30}`
	type user struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	var u user
	if err := ToBean(src, &u); err != nil {
		t.Fatalf("to bean: %v", err)
	}
	if u.Name != "alice" || u.Age != 30 {
		t.Fatalf("got %+v", u)
	}
}

func TestToBeanWithOptionsUsesUnmarshalFunc(t *testing.T) {
	type user struct {
		Name string `json:"name"`
	}
	called := false
	var u user
	err := ToBeanWithOptions(`{"name":"ignored"}`, &u, WithBeanUnmarshalFunc(func(_ []byte, dst any) error {
		called = true
		dst.(*user).Name = "provided"
		return nil
	}))
	if err != nil {
		t.Fatalf("ToBeanWithOptions: %v", err)
	}
	if !called || u.Name != "provided" {
		t.Fatalf("unmarshal provider called=%v user=%+v", called, u)
	}
}

func TestArrayOps(t *testing.T) {
	arr := NewJSONArray()
	arr.Add(1).Add("x").Add(true).Add(nil)
	if arr.Len() != 4 || arr.GetInt(0) != 1 || arr.GetString(1) != "x" || !arr.GetBool(2) || !arr.IsNull(3) {
		t.Fatalf("array basic: %s", arr.String())
	}
	arr.Insert(1, "y")
	if arr.GetString(1) != "y" {
		t.Fatalf("insert")
	}
	arr.Remove(0)
	if arr.GetString(0) != "y" {
		t.Fatalf("remove")
	}
}
