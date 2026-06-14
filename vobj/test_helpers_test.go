package vobj_test

type record struct {
	Name string
	Tags []string
}

type encoderFunc func(any) error

func (f encoderFunc) Encode(v any) error { return f(v) }

type decoderFunc func(any) error

func (f decoderFunc) Decode(v any) error { return f(v) }
