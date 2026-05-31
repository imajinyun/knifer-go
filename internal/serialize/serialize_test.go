package serialize

import (
	"errors"
	"reflect"
	"testing"
)

type payload struct {
	Name string
	Tags []string
}

type withFunc struct {
	Fn func()
}

func TestSerializeDeserializeAndClone(t *testing.T) {
	src := payload{Name: "go", Tags: []string{"tool"}}
	data, err := Serialize(src)
	if err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	out, err := DeserializeTo[payload](data)
	if err != nil {
		t.Fatalf("DeserializeTo: %v", err)
	}
	if !reflect.DeepEqual(out, src) {
		t.Fatalf("DeserializeTo mismatch: %#v", out)
	}

	clone, err := Clone(src)
	if err != nil {
		t.Fatalf("Clone: %v", err)
	}
	clone.Tags[0] = "sdk"
	if src.Tags[0] != "tool" {
		t.Fatal("clone changed source")
	}
	if !reflect.DeepEqual(CloneIfPossible(src), src) {
		t.Fatal("CloneIfPossible changed value")
	}
}

func TestDeserializeAcceptedTypes(t *testing.T) {
	src := payload{Name: "go", Tags: []string{"tool"}}
	data, err := Serialize(src)
	if err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	var allowed payload
	if err := Deserialize(data, &allowed, payload{}); err != nil {
		t.Fatalf("Deserialize accepted: %v", err)
	}
	var rejected payload
	if err := Deserialize(data, &rejected, struct{ Other string }{}); err == nil {
		t.Fatal("expected rejected decoded type")
	}
}

func TestSerializeFailures(t *testing.T) {
	if data := SerializeOrNil(withFunc{}); data != nil {
		t.Fatalf("SerializeOrNil should return nil: %v", data)
	}
	if _, err := Serialize(withFunc{}); err == nil {
		t.Fatal("expected serialize error")
	}
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic")
		}
	}()
	_ = MustDeserialize[payload]([]byte("bad"))
}

func TestRegisterInterfaceValue(t *testing.T) {
	type box struct{ Value any }
	type item struct{ Name string }
	Register(item{})
	data, err := Serialize(box{Value: item{Name: "x"}})
	if err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	out, err := DeserializeTo[box](data, box{}, item{})
	if err != nil {
		t.Fatalf("DeserializeTo: %v", err)
	}
	if got, ok := out.Value.(item); !ok || got.Name != "x" {
		t.Fatalf("interface value: %#v", out.Value)
	}
}

func TestValidateAcceptedTypesErrors(t *testing.T) {
	if err := ValidateAcceptedTypes(payload{}, payload{}); err != nil {
		t.Fatalf("accepted payload: %v", err)
	}
	err := ValidateAcceptedTypes(payload{}, errors.New("x"))
	if err == nil {
		t.Fatal("expected validation error")
	}
}
