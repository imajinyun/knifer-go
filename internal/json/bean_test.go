package json

import (
	stdjson "encoding/json"
	"io"
	"strings"
	"testing"
)

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

func TestToList(t *testing.T) {
	var out []int
	if err := ToList(`[1, 2, 3]`, &out); err != nil {
		t.Fatalf("ToList error = %v", err)
	}
	if len(out) != 3 || out[0] != 1 {
		t.Fatalf("ToList = %v", out)
	}
}

func TestToListWithOptions(t *testing.T) {
	var out []int
	if err := ToListWithOptions(`[1, 2, 3]`, &out); err != nil {
		t.Fatalf("ToListWithOptions error = %v", err)
	}
	if len(out) != 3 {
		t.Fatalf("ToListWithOptions = %v", out)
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
