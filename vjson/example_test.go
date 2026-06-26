package vjson_test

import (
	stdjson "encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/imajinyun/knifer-go"
	"github.com/imajinyun/knifer-go/vjson"
	"github.com/imajinyun/knifer-go/vxml"
)

func Example_cookbookEncodeStruct() {
	type user struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	body, err := vjson.ToStr(user{Name: "knifer-go", Age: 5})
	fmt.Println(body, err)
	// Output: {"age":5,"name":"knifer-go"} <nil>
}

func Example_cookbookDecodeStruct() {
	type user struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	var dst user
	err := vjson.ToBean(`{"name":"knifer-go","age":5}`, &dst)
	fmt.Println(dst.Name, dst.Age, err)
	// Output: knifer-go 5 <nil>
}

func Example_cookbookParseObjectPathDefault() {
	obj, err := vjson.ParseObj(`{"user":{"name":"knifer-go"}}`)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(vjson.GetByPath(obj, "user.name"))
	fmt.Println(vjson.GetByPathOr(obj, "user.email", "missing"))
	// Output:
	// knifer-go
	// missing
}

func Example_cookbookFormatForHumans() {
	fmt.Println(vjson.FormatWithOptions(`{"name":"knifer-go"}`, vjson.WithFormatIndentWidth(2)))
	// Output:
	// {
	//   "name": "knifer-go"
	// }
}

func Example_cookbookConvertXMLAndJSON() {
	obj, err := vjson.XMLToJSON(`<user><name>knifer-go</name></user>`)
	if err != nil {
		fmt.Println(err)
		return
	}
	xmlText, err := vjson.ToXMLWithOptions(obj.GetJSONObject("user"), "user", vxml.WithOmitDeclaration(true))
	fmt.Println(vjson.GetByPath(obj, "user.name"), xmlText, err)
	// Output: knifer-go <user><name>knifer-go</name></user> <nil>
}

func Example_cookbookCustomParsingBehavior() {
	obj, err := vjson.ParseObjWithOptions(`{"n":"ignored"}`, vjson.WithParseDecoderFactory(func(io.Reader) *stdjson.Decoder {
		dec := stdjson.NewDecoder(strings.NewReader(`{"n":7}`))
		dec.UseNumber()
		return dec
	}))
	fmt.Println(obj.GetInt("n"), err)
	// Output: 7 <nil>
}

func Example_cookbookExplicitErrorHandling() {
	_, err := vjson.ParseObj(`["not","an","object"]`)
	fmt.Println(errors.Is(err, knifer.ErrCodeInvalidInput))
	// Output: true
}

func ExampleToStr() {
	s, _ := vjson.ToStr(map[string]any{"name": "go"})
	fmt.Println(s)
	// Output: {"name":"go"}
}

func ExampleIsJSON() {
	fmt.Println(vjson.IsJSON(`{"a":1}`))
	fmt.Println(vjson.IsJSON(`not json`))
	// Output:
	// true
	// false
}

func ExampleGetByPath() {
	root, _ := vjson.Parse(`{"user":{"name":"go"}}`)
	fmt.Println(vjson.GetByPath(root, "user.name"))
	// Output: go
}

func ExampleParseObj_error() {
	_, err := vjson.ParseObj(`[1,2,3]`)
	fmt.Println(errors.Is(err, knifer.ErrCodeInvalidInput))
	// Output: true
}

func ExamplePutByPath() {
	root := vjson.NewObject()
	_ = vjson.PutByPath(root, "user.name", "knifer-go")
	fmt.Println(vjson.GetByPath(root, "user.name"))
	// Output: knifer-go
}

func ExampleToBean() {
	type user struct {
		Name string `json:"name"`
	}
	var u user
	_ = vjson.ToBean(`{"name":"knifer-go"}`, &u)
	fmt.Println(u.Name)
	// Output: knifer-go
}

func ExampleXMLToJSON() {
	obj, _ := vjson.XMLToJSON(`<user><name>knifer-go</name></user>`)
	fmt.Println(vjson.GetByPath(obj, "user.name"))
	// Output: knifer-go
}

func ExampleCreateConfig() {
	cfg := vjson.CreateConfig()
	fmt.Println(cfg.IndentFactor)
	// Output: 4
}

func ExampleNewConfig() {
	cfg := vjson.NewConfig()
	cfg.IgnoreNullValue = true
	fmt.Println(cfg.IgnoreNullValue)
	// Output: true
}

func ExampleNewObject() {
	obj := vjson.NewObject().Set("name", "go").Set("count", 2)
	fmt.Println(obj.ToString())
	// Output: {"name":"go","count":2}
}

func ExampleNewObjectWithConfig() {
	cfg := vjson.NewConfig()
	cfg.IgnoreNullValue = true
	obj := vjson.NewObjectWithConfig(cfg).Set("name", "go").Set("empty", nil)
	fmt.Println(obj.ToString())
	// Output: {"name":"go"}
}

func ExampleNewArray() {
	arr := vjson.NewArray().Add("go").Add(2).Add(true)
	fmt.Println(arr.ToString())
	// Output: ["go",2,true]
}

func ExampleNewArrayWithConfig() {
	cfg := vjson.NewConfig()
	cfg.IgnoreNullValue = true
	arr := vjson.NewArrayWithConfig(cfg).Add("go").Add(nil).Add("knifer")
	fmt.Println(arr.ToString())
	// Output: ["go","knifer"]
}

func ExampleNewJSONError() {
	err := vjson.NewJSONError("bad value %q", "name")
	fmt.Println(errors.Is(err, knifer.ErrCodeInvalidInput))
	fmt.Println(err.Error())
	// Output:
	// true
	// bad value "name"
}

func ExampleWrapJSONError() {
	cause := errors.New("decode failed")
	err := vjson.WrapJSONError(cause, "json parse")
	fmt.Println(errors.Is(err, knifer.ErrCodeInvalidInput))
	fmt.Println(errors.Is(err, cause))
	// Output:
	// true
	// true
}

func ExampleWithConfig() {
	cfg := vjson.NewConfig()
	cfg.IgnoreNullValue = true
	s, _ := vjson.ToStr(map[string]any{"name": "go", "empty": nil}, vjson.WithConfig(cfg))
	fmt.Println(s)
	// Output: {"name":"go"}
}

func ExampleWithIndent() {
	s, _ := vjson.ToPrettyStr(vjson.NewObject().Set("name", "go"), vjson.WithIndent(2))
	fmt.Println(s)
	// Output:
	// {
	//   "name": "go"
	// }
}

func ExampleWithIgnoreNullValue() {
	s, _ := vjson.ToStr(map[string]any{"name": "go", "empty": nil}, vjson.WithIgnoreNullValue(true))
	fmt.Println(s)
	// Output: {"name":"go"}
}

func ExampleWithDateFormat() {
	when := time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)
	s, _ := vjson.ToStr(when, vjson.WithDateFormat("2006-01-02"))
	fmt.Println(s)
	// Output: "2024-01-02"
}

func ExampleWithMarshalFunc() {
	s, _ := vjson.ToStr(struct{ Name string }{}, vjson.WithMarshalFunc(func(any) ([]byte, error) {
		return []byte(`{"name":"provided"}`), nil
	}))
	fmt.Println(s)
	// Output: {"name":"provided"}
}

func ExampleWithUnmarshalFunc() {
	s, _ := vjson.ToStr(struct{ Name string }{},
		vjson.WithMarshalFunc(func(any) ([]byte, error) { return []byte(`{"ignored":true}`), nil }),
		vjson.WithUnmarshalFunc(func(_ []byte, dst any) error {
			*(dst.(*any)) = map[string]any{"name": "provided"}
			return nil
		}),
	)
	fmt.Println(s)
	// Output: {"name":"provided"}
}

func ExampleWithDecoderFactory() {
	s, _ := vjson.ToStr(struct{ Name string }{},
		vjson.WithMarshalFunc(func(any) ([]byte, error) { return []byte(`{"ignored":true}`), nil }),
		vjson.WithDecoderFactory(func(io.Reader) *stdjson.Decoder {
			return stdjson.NewDecoder(strings.NewReader(`{"name":"decoder"}`))
		}),
	)
	fmt.Println(s)
	// Output: {"name":"decoder"}
}

func ExampleWithSprintFunc() {
	s, _ := vjson.ToStr(map[int]string{7: "seven"}, vjson.WithSprintFunc(func(v any) string {
		return "key-" + fmt.Sprint(v)
	}))
	fmt.Println(s)
	// Output: {"key-7":"seven"}
}

func ExampleWithFormatIntFunc() {
	s, _ := vjson.ToStr(map[string]any{"n": 7}, vjson.WithFormatIntFunc(func(int64, int) string { return `"int!"` }))
	fmt.Println(s)
	// Output: {"n":"int!"}
}

func ExampleWithFormatFloatFunc() {
	s, _ := vjson.ToStr(map[string]any{"n": 1.5}, vjson.WithFormatFloatFunc(func(float64, byte, int, int) string { return `"float!"` }))
	fmt.Println(s)
	// Output: {"n":"float!"}
}

func ExampleWithFormatIndent() {
	formatted := vjson.FormatWithOptions(`{"a":1}`, vjson.WithFormatIndent("--"))
	fmt.Println(formatted)
	// Output:
	// {
	// --"a": 1
	// }
}

func ExampleWithFormatIndentWidth() {
	formatted := vjson.FormatWithOptions(`{"a":1}`, vjson.WithFormatIndentWidth(2))
	fmt.Println(formatted)
	// Output:
	// {
	//   "a": 1
	// }
}

func ExampleWithFormatSpaceAfterKey() {
	formatted := vjson.FormatWithOptions(`{"a":1}`, vjson.WithFormatSpaceAfterKey(false))
	fmt.Println(formatted)
	// Output:
	// {
	//     "a":1
	// }
}

func ExampleWithParseConfig() {
	cfg := vjson.NewConfig()
	cfg.IgnoreNullValue = true
	obj, _ := vjson.ParseObjWithOptions(map[string]any{"name": "go", "empty": nil}, vjson.WithParseConfig(cfg))
	fmt.Println(obj.ToString())
	// Output: {"name":"go"}
}

func ExampleWithParseUnmarshalFunc() {
	obj, _ := vjson.ParseObjWithOptions(`{}`, vjson.WithParseUnmarshalFunc(func(_ []byte, dst any) error {
		*(dst.(*any)) = map[string]any{"name": "provided"}
		return nil
	}))
	fmt.Println(obj.GetString("name"))
	// Output: provided
}

func ExampleWithParseDecoderFactory() {
	obj, _ := vjson.ParseObjWithOptions(`{"ignored":true}`, vjson.WithParseDecoderFactory(func(io.Reader) *stdjson.Decoder {
		return stdjson.NewDecoder(strings.NewReader(`{"name":"decoder"}`))
	}))
	fmt.Println(obj.GetString("name"))
	// Output: decoder
}

func ExampleWithJSONValidFunc() {
	ok := vjson.IsJSONWithOptions("not-json", vjson.WithJSONValidFunc(func(b []byte) bool {
		return string(b) == "not-json"
	}))
	fmt.Println(ok)
	// Output: true
}

func ExampleWithBeanConfig() {
	type user struct {
		Name string `json:"name"`
	}
	cfg := vjson.NewConfig()
	cfg.UnmarshalFunc = func(_ []byte, dst any) error {
		dst.(*user).Name = "configured"
		return nil
	}
	var u user
	_ = vjson.ToBeanWithOptions(`{"name":"ignored"}`, &u, vjson.WithBeanConfig(cfg))
	fmt.Println(u.Name)
	// Output: configured
}

func ExampleWithBeanUnmarshalFunc() {
	type user struct {
		Name string `json:"name"`
	}
	var u user
	_ = vjson.ToBeanWithOptions(`{"name":"ignored"}`, &u, vjson.WithBeanUnmarshalFunc(func(_ []byte, dst any) error {
		dst.(*user).Name = "provided"
		return nil
	}))
	fmt.Println(u.Name)
	// Output: provided
}

func ExampleIsNull() {
	fmt.Println(vjson.IsNull(nil))
	fmt.Println(vjson.IsNull(""))
	// Output:
	// true
	// false
}

func ExampleParse() {
	parsed, _ := vjson.Parse(`{"name":"go"}`)
	obj := parsed.(*vjson.Object)
	fmt.Println(obj.GetString("name"))
	// Output: go
}

func ExampleParseWithOptions() {
	parsed, _ := vjson.ParseWithOptions(`{}`, vjson.WithParseUnmarshalFunc(func(_ []byte, dst any) error {
		*(dst.(*any)) = []any{"provided"}
		return nil
	}))
	arr := parsed.(*vjson.Array)
	fmt.Println(arr.GetString(0))
	// Output: provided
}

func ExampleParseWithConfig() {
	cfg := vjson.NewConfig()
	cfg.IgnoreCase = true
	parsed, _ := vjson.ParseWithConfig(`{"Name":"go"}`, cfg)
	obj := parsed.(*vjson.Object)
	fmt.Println(obj.GetString("name"))
	// Output: go
}

func ExampleParseObjWithOptions() {
	obj, _ := vjson.ParseObjWithOptions(`{}`, vjson.WithParseUnmarshalFunc(func(_ []byte, dst any) error {
		*(dst.(*any)) = map[string]any{"name": "provided"}
		return nil
	}))
	fmt.Println(obj.GetString("name"))
	// Output: provided
}

func ExampleParseObjWithConfig() {
	cfg := vjson.NewConfig()
	cfg.IgnoreCase = true
	obj, _ := vjson.ParseObjWithConfig(`{"Name":"go"}`, cfg)
	fmt.Println(obj.GetString("name"))
	// Output: go
}

func ExampleParseArray() {
	arr, _ := vjson.ParseArray(`["go",2,true]`)
	fmt.Println(arr.GetString(0), arr.GetInt(1), arr.GetBool(2))
	// Output: go 2 true
}

func ExampleParseArrayWithOptions() {
	arr, _ := vjson.ParseArrayWithOptions(`[]`, vjson.WithParseUnmarshalFunc(func(_ []byte, dst any) error {
		*(dst.(*any)) = []any{"provided"}
		return nil
	}))
	fmt.Println(arr.GetString(0))
	// Output: provided
}

func ExampleParseArrayWithConfig() {
	cfg := vjson.NewConfig()
	cfg.IgnoreNullValue = true
	arr, _ := vjson.ParseArrayWithConfig([]any{"go", nil, "knifer"}, cfg)
	fmt.Println(arr.ToString())
	// Output: ["go","knifer"]
}

func ExampleToPrettyStr() {
	s, _ := vjson.ToPrettyStr(vjson.NewObject().Set("name", "go"))
	fmt.Println(s)
	// Output:
	// {
	//     "name": "go"
	// }
}

func ExampleToStrIndent() {
	s, _ := vjson.ToStrIndent(vjson.NewObject().Set("name", "go"), 2)
	fmt.Println(s)
	// Output:
	// {
	//   "name": "go"
	// }
}

func ExampleToStrWithConfig() {
	cfg := vjson.NewConfig()
	cfg.IgnoreNullValue = true
	s, _ := vjson.ToStrWithConfig(map[string]any{"name": "go", "empty": nil}, cfg)
	fmt.Println(s)
	// Output: {"name":"go"}
}

func ExampleToPrettyStrWithConfig() {
	cfg := vjson.NewConfig()
	cfg.IndentFactor = 2
	s, _ := vjson.ToPrettyStrWithConfig(vjson.NewObject().Set("name", "go"), cfg)
	fmt.Println(s)
	// Output:
	// {
	//   "name": "go"
	// }
}

func ExampleFormat() {
	fmt.Println(vjson.Format(`{"a":1}`))
	// Output:
	// {
	//     "a": 1
	// }
}

func ExampleFormatWithOptions() {
	fmt.Println(vjson.FormatWithOptions(`{"a":1}`, vjson.WithFormatIndentWidth(2), vjson.WithFormatSpaceAfterKey(false)))
	// Output:
	// {
	//   "a":1
	// }
}

func ExampleIsJSONWithOptions() {
	fmt.Println(vjson.IsJSONWithOptions("custom", vjson.WithJSONValidFunc(func(b []byte) bool {
		return string(b) == "custom"
	})))
	// Output: true
}

func ExampleIsObj() {
	fmt.Println(vjson.IsObj(`{"a":1}`))
	fmt.Println(vjson.IsObj(`[1]`))
	// Output:
	// true
	// false
}

func ExampleIsObjWithOptions() {
	fmt.Println(vjson.IsObjWithOptions(`{"custom":true}`, vjson.WithJSONValidFunc(func([]byte) bool { return true })))
	// Output: true
}

func ExampleIsArray() {
	fmt.Println(vjson.IsArray(`[1]`))
	fmt.Println(vjson.IsArray(`{"a":1}`))
	// Output:
	// true
	// false
}

func ExampleIsArrayWithOptions() {
	fmt.Println(vjson.IsArrayWithOptions(`[custom]`, vjson.WithJSONValidFunc(func([]byte) bool { return true })))
	// Output: true
}

func ExampleGetByPathOr() {
	root, _ := vjson.Parse(`{"user":{"name":"go"}}`)
	fmt.Println(vjson.GetByPathOr(root, "user.name", "unknown"))
	fmt.Println(vjson.GetByPathOr(root, "user.email", "missing"))
	// Output:
	// go
	// missing
}

func ExampleQuote() {
	fmt.Println(vjson.Quote("go\nknifer"))
	// Output: "go\nknifer"
}

func ExampleToBeanWithOptions() {
	type user struct {
		Name string `json:"name"`
	}
	var u user
	_ = vjson.ToBeanWithOptions(`{"name":"ignored"}`, &u, vjson.WithBeanUnmarshalFunc(func(_ []byte, dst any) error {
		dst.(*user).Name = "provided"
		return nil
	}))
	fmt.Println(u.Name)
	// Output: provided
}

func ExampleToList() {
	type user struct {
		Name string `json:"name"`
	}
	var users []user
	_ = vjson.ToList(`[{"name":"go"},{"name":"knifer"}]`, &users)
	fmt.Println(users[0].Name, users[1].Name)
	// Output: go knifer
}

func ExampleToListWithOptions() {
	type user struct {
		Name string `json:"name"`
	}
	var users []user
	_ = vjson.ToListWithOptions(`[]`, &users, vjson.WithBeanUnmarshalFunc(func(_ []byte, dst any) error {
		*(dst.(*[]user)) = []user{{Name: "provided"}}
		return nil
	}))
	fmt.Println(users[0].Name)
	// Output: provided
}

func ExampleXMLToJSONWithOptions() {
	obj, _ := vjson.XMLToJSONWithOptions(`<root><n>42</n></root>`, vxml.WithScalarIntParser(func(string, int, int) (int64, error) {
		return 7, nil
	}))
	fmt.Println(obj.GetJSONObject("root").GetInt64("n"))
	// Output: 7
}

func ExampleToXML() {
	xmlText, _ := vjson.ToXML(vjson.NewObject().Set("name", "go"), "user")
	fmt.Println(strings.Contains(xmlText, "<name>go</name>"))
	// Output: true
}

func ExampleToXMLWithOptions() {
	xmlText, _ := vjson.ToXMLWithOptions(vjson.NewObject().Set("name", "go"), "user", vxml.WithOmitDeclaration(true))
	fmt.Println(xmlText)
	// Output: <user><name>go</name></user>
}

func ExampleWithParseIntFunc() {
	cfg := vjson.NewConfig()
	cfg.ParseIntFunc = func(string, int, int) (int64, error) { return 42, nil }
	obj := vjson.NewObjectWithConfig(cfg).Set("n", "ignored")
	option := vjson.WithParseIntFunc(strconv.ParseInt)
	fmt.Println(obj.GetInt64("n"))
	fmt.Println(option != nil)
	// Output:
	// 42
	// true
}

func ExampleWithParseFloatFunc() {
	cfg := vjson.NewConfig()
	cfg.ParseFloatFunc = func(string, int) (float64, error) { return 3.5, nil }
	obj := vjson.NewObjectWithConfig(cfg).Set("n", "ignored")
	option := vjson.WithParseFloatFunc(strconv.ParseFloat)
	fmt.Println(obj.GetFloat64("n"))
	fmt.Println(option != nil)
	// Output:
	// 3.5
	// true
}

func ExampleWithParseBoolFunc() {
	cfg := vjson.NewConfig()
	cfg.ParseBoolFunc = func(string) (bool, error) { return true, nil }
	obj := vjson.NewObjectWithConfig(cfg).Set("enabled", "custom")
	option := vjson.WithParseBoolFunc(strconv.ParseBool)
	fmt.Println(obj.GetBool("enabled"))
	fmt.Println(option != nil)
	// Output:
	// true
	// true
}
