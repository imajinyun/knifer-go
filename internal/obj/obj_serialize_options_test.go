package obj

import (
	"encoding/gob"
	"io"
	"reflect"
	"testing"
)

type recordingEncoder struct {
	inner  Encoder
	called *bool
}

func (e recordingEncoder) Encode(v any) error {
	*e.called = true
	return e.inner.Encode(v)
}

type recordingDecoder struct {
	inner  Decoder
	called *bool
}

func (d recordingDecoder) Decode(v any) error {
	*d.called = true
	return d.inner.Decode(v)
}

func TestSerializeWithOptionsUsesCodecFactories(t *testing.T) {
	src := sample{Name: "n", Tags: []string{"a"}}
	encoderCalled := false
	data, err := SerializeWithOptions(src, WithEncoderFactory(func(w io.Writer) Encoder {
		return recordingEncoder{inner: gob.NewEncoder(w), called: &encoderCalled}
	}))
	if err != nil {
		t.Fatalf("SerializeWithOptions: %v", err)
	}
	if !encoderCalled {
		t.Fatal("custom encoder factory was not used")
	}

	decoderCalled := false
	var out sample
	err = DeserializeWithOptions(data, &out, nil, WithDecoderFactory(func(r io.Reader) Decoder {
		return recordingDecoder{inner: gob.NewDecoder(r), called: &decoderCalled}
	}))
	if err != nil {
		t.Fatalf("DeserializeWithOptions: %v", err)
	}
	if !decoderCalled || !reflect.DeepEqual(out, src) {
		t.Fatalf("decoderCalled=%v out=%#v", decoderCalled, out)
	}

	clone, err := CloneWithOptions(src,
		WithEncoderFactory(func(w io.Writer) Encoder { return gob.NewEncoder(w) }),
		WithDecoderFactory(func(r io.Reader) Decoder { return gob.NewDecoder(r) }),
	)
	if err != nil || !reflect.DeepEqual(clone, src) {
		t.Fatalf("CloneWithOptions = %#v, %v", clone, err)
	}
}
