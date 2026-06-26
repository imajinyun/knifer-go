package vjson_test

import (
	"errors"
	"testing"

	knifer "github.com/imajinyun/knifer-go"
	"github.com/imajinyun/knifer-go/vjson"
)

func TestFacadeErrorNameWithoutJSONPrefix(t *testing.T) {
	_, err := vjson.ParseObj(`[1,2]`)
	var jsonErr *vjson.Error
	if !errors.As(err, &jsonErr) {
		t.Fatalf("ParseObj() error type = %T, want *vjson.Error", err)
	}
	if !errors.Is(err, knifer.ErrCodeInvalidInput) {
		t.Fatalf("errors.Is(err, ErrCodeInvalidInput) = false: %v", err)
	}
	code, ok := knifer.CodeOf(err)
	if !ok || code != knifer.ErrCodeInvalidInput {
		t.Fatalf("CodeOf(err) = %q, %v; want invalid input", code, ok)
	}

	_, err = vjson.XMLToJSON(`<root><unclosed></root>`)
	if !errors.Is(err, knifer.ErrCodeInvalidInput) {
		t.Fatalf("XMLToJSON malformed XML code = %v, want invalid input", err)
	}
}

func TestFacadeNewJSONError(t *testing.T) {
	err := vjson.NewJSONError("code %d", 42)
	if err == nil || err.Error() != "code 42" {
		t.Fatalf("NewJSONError = %v", err)
	}
	var jsonErr *vjson.Error
	if !errors.As(err, &jsonErr) {
		t.Fatal("NewJSONError should produce *vjson.Error")
	}
}

func TestFacadeWrapJSONError(t *testing.T) {
	cause := errors.New("root cause")
	err := vjson.WrapJSONError(cause, "wrapped")
	if err == nil || !errors.Is(err, cause) {
		t.Fatalf("WrapJSONError = %v", err)
	}
	if err.Error() != "wrapped: root cause" {
		t.Fatalf("WrapJSONError message = %q", err.Error())
	}
}
