package csvx

import (
	"errors"
	"testing"

	knifer "github.com/imajinyun/go-knifer"
)

func TestCSVErrorNilReceiver(t *testing.T) {
	var e *CSVError
	if s := e.Error(); s != "" {
		t.Fatalf("nil CSVError.Error() = %q", s)
	}
	if ec := e.ErrorCode(); ec != "" {
		t.Fatalf("nil CSVError.ErrorCode() = %q", ec)
	}
	if cause := e.Unwrap(); cause != nil {
		t.Fatalf("nil CSVError.Unwrap() = %v", cause)
	}
	if e.Is(nil) {
		t.Fatal("nil CSVError.Is(nil) = true")
	}
}

func TestCSVErrorWithCause(t *testing.T) {
	cause := errors.New("inner")
	e := &CSVError{Code: knifer.ErrCodeInvalidInput, Msg: "bad input", Cause: cause}
	if s := e.Error(); s != "bad input: inner" {
		t.Fatalf("CSVError.Error() = %q", s)
	}
	if !e.Is(knifer.ErrCodeInvalidInput) {
		t.Fatal("CSVError.Is(ErrCodeInvalidInput) = false")
	}
	if !e.Is(&CSVError{Code: knifer.ErrCodeInvalidInput}) {
		t.Fatal("CSVError.Is(same code CSVError) = false")
	}
}

func TestCSVErrorWithoutCause(t *testing.T) {
	e := &CSVError{Code: knifer.ErrCodeNotFound, Msg: "not found"}
	if s := e.Error(); s != "not found" {
		t.Fatalf("CSVError.Error() = %q", s)
	}
}
