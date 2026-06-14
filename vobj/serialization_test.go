package vobj_test

import (
	"errors"
	"io"
	"reflect"
	"testing"

	"github.com/imajinyun/go-knifer/vobj"
)

func TestFacadeCloneAndSerialize(t *testing.T) {
	src := record{Name: "go", Tags: []string{"tool"}}
	clone, err := vobj.Clone(src)
	if err != nil {
		t.Fatalf("Clone: %v", err)
	}
	clone.Tags[0] = "sdk"
	if src.Tags[0] != "tool" {
		t.Fatal("clone changed source")
	}
	data, err := vobj.Serialize(src)
	if err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	var out record
	if err := vobj.Deserialize(data, &out); err != nil || out.Name != src.Name {
		t.Fatalf("Deserialize: %#v %v", out, err)
	}
}

func TestFacadeSerializeExtended(t *testing.T) {
	src := record{Name: "go", Tags: []string{"tool"}}

	data, err := vobj.Serialize(src)
	if err != nil {
		t.Fatalf("Serialize: %v", err)
	}

	out, err := vobj.DeserializeTo[record](data)
	if err != nil {
		t.Fatalf("DeserializeTo: %v", err)
	}
	if out.Name != src.Name || len(out.Tags) != 1 || out.Tags[0] != "tool" {
		t.Fatalf("DeserializeTo mismatch: %#v", out)
	}

	nilData := vobj.SerializeOrNil(src)
	if nilData == nil {
		t.Fatal("SerializeOrNil should not return nil for valid input")
	}
}

func TestFacadeSerializationOptionsAndValidation(t *testing.T) {
	src := record{Name: "go", Tags: []string{"tool"}}
	clone, err := vobj.CloneByStream(src)
	if err != nil || clone.Name != src.Name {
		t.Fatalf("CloneByStream = %#v, %v", clone, err)
	}
	clone = vobj.CloneIfPossible(src)
	clone.Tags[0] = "copy"
	if src.Tags[0] != "tool" {
		t.Fatal("CloneIfPossible changed source")
	}

	failingOpt := vobj.WithEncoderFactory(func(io.Writer) vobj.Encoder {
		return encoderFunc(func(any) error { return errors.New("encode failed") })
	})
	if got := vobj.SerializeOrNilWithOptions(src, failingOpt); got != nil {
		t.Fatalf("SerializeOrNilWithOptions failing = %v", got)
	}
	if got := vobj.CloneIfPossibleWithOptions(src, failingOpt); !reflect.DeepEqual(got, src) {
		t.Fatalf("CloneIfPossibleWithOptions fallback = %#v", got)
	}

	var decoded record
	err = vobj.DeserializeWithOptions([]byte("ignored"), &decoded, nil,
		vobj.WithDecoderFactory(func(io.Reader) vobj.Decoder {
			return decoderFunc(func(out any) error {
				*out.(*record) = record{Name: "decoded", Tags: []string{"via-option"}}
				return nil
			})
		}),
	)
	if err != nil || decoded.Name != "decoded" || decoded.Tags[0] != "via-option" {
		t.Fatalf("DeserializeWithOptions = %#v, %v", decoded, err)
	}

	data, err := vobj.Serialize(src)
	if err != nil {
		t.Fatal(err)
	}
	out := vobj.MustDeserialize[record](data)
	if !reflect.DeepEqual(out, src) {
		t.Fatalf("MustDeserialize = %#v", out)
	}
	if _, err := vobj.DeserializeToWithOptions[record]([]byte("bad"), nil); err == nil {
		t.Fatal("DeserializeToWithOptions invalid data error = nil")
	}
	if err := vobj.ValidateAcceptedTypes(src, record{}); err != nil {
		t.Fatalf("ValidateAcceptedTypes accepted record: %v", err)
	}
	if err := vobj.ValidateAcceptedTypes(src, "not-record"); err == nil {
		t.Fatal("ValidateAcceptedTypes rejected type error = nil")
	}
}
