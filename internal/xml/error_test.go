package xml

import (
	"errors"
	"strings"
	"testing"

	knifer "github.com/imajinyun/knifer-go"
)

func TestXMLErrorContract(t *testing.T) {
	_, err := ParseXML("")
	assertXMLInvalidInput(t, err)

	assertXMLInvalidInput(t, WriteTo(nil, CreateXMLWithRoot("root")))
	assertXMLInvalidInput(t, WriteTo(&strings.Builder{}, "unsupported"))

	var dst struct {
		Root struct {
			Value int `json:"value"`
		} `json:"root"`
	}
	assertXMLInvalidInput(t, XMLToBean(`<root><value>not-int</value></root>`, &dst))
}

func TestXMLErrorUnwrap(t *testing.T) {
	cause := errors.New("root cause")
	err := wrapXMLError(knifer.ErrCodeInternal, "wrapped", cause)
	var xmlErr *XMLError
	if !errors.As(err, &xmlErr) {
		t.Fatal("wrapXMLError should return XMLError")
	}
	if got := xmlErr.Unwrap(); got != cause {
		t.Fatalf("Unwrap = %v, want %v", got, cause)
	}
}

func TestXMLErrorUnwrapNil(t *testing.T) {
	var e *XMLError
	if got := e.Unwrap(); got != nil {
		t.Fatalf("nil Unwrap = %v, want nil", got)
	}
}

func TestXMLErrorErrorCodeNil(t *testing.T) {
	var e *XMLError
	if got := e.ErrorCode(); got != "" {
		t.Fatalf("nil ErrorCode = %v, want empty", got)
	}
}

func TestWrapInternal(t *testing.T) {
	cause := errors.New("internal error")
	err := wrapInternal("test wrap", cause)
	if err == nil {
		t.Fatal("wrapInternal should not return nil")
	}
	// nil cause should return nil
	if got := wrapInternal("test", nil); got != nil {
		t.Fatal("wrapInternal with nil cause should return nil")
	}
}
