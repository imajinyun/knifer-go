package vxml

import (
	"errors"
	"testing"

	knifer "github.com/imajinyun/knifer-go"
)

func TestFacadeXMLErrorContract(t *testing.T) {
	_, err := ParseXML(`<root><unclosed></root>`)
	if err == nil {
		t.Fatal("ParseXML() error = nil, want invalid input")
	}
	if !errors.Is(err, knifer.ErrCodeInvalidInput) {
		t.Fatalf("errors.Is(err, ErrCodeInvalidInput) = false: %v", err)
	}
	code, ok := knifer.CodeOf(err)
	if !ok || code != knifer.ErrCodeInvalidInput {
		t.Fatalf("CodeOf(err) = %q, %v; want invalid input", code, ok)
	}
	var xmlErr *Error
	if !errors.As(err, &xmlErr) {
		t.Fatalf("errors.As(err, *vxml.Error) = false: %v", err)
	}
}
