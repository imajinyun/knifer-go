package codec

import (
	"encoding/base64"
	"errors"
	"testing"

	knifer "github.com/imajinyun/go-knifer"
)

func TestCodecErrorContract(t *testing.T) {
	tests := []struct {
		name    string
		decode  func() ([]byte, error)
		wantAs  any
		wantMsg string
	}{
		{
			name: "standard base64 invalid input",
			decode: func() ([]byte, error) {
				return Base64Decode("invalid!")
			},
			wantAs:  new(base64.CorruptInputError),
			wantMsg: "decode base64",
		},
		{
			name: "url base64 invalid input",
			decode: func() ([]byte, error) {
				return Base64URLDecode("invalid!")
			},
			wantMsg: "decode url-safe base64",
		},
		{
			name: "raw url base64 invalid input",
			decode: func() ([]byte, error) {
				return Base64RawURLDecode("invalid!")
			},
			wantMsg: "decode raw url-safe base64",
		},
		{
			name: "hex invalid input",
			decode: func() ([]byte, error) {
				return HexDecode("xyz")
			},
			wantMsg: "decode hex",
		},
		{
			name: "hex string invalid input",
			decode: func() ([]byte, error) {
				_, err := HexDecodeStr("xyz")
				return nil, err
			},
			wantMsg: "decode hex",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.decode()
			assertCodecInvalidInput(t, err)
			if tt.wantMsg != "" && !errors.Is(err, &CodecError{Code: knifer.ErrCodeInvalidInput}) {
				t.Fatalf("errors.Is(%v, *CodecError{Code:%s}) = false", err, knifer.ErrCodeInvalidInput)
			}
			if tt.wantAs != nil {
				assertCodecCauseAs(t, err, tt.wantAs)
			}
		})
	}
}

func TestCodecErrorString(t *testing.T) {
	err := &CodecError{Code: knifer.ErrCodeInternal, Msg: "test error", Cause: nil}
	if got := err.Error(); got != "test error" {
		t.Fatalf("CodecError.Error() = %q, want %q", got, "test error")
	}
	errWithCause := &CodecError{Code: knifer.ErrCodeInternal, Msg: "test", Cause: errors.New("wrapped")}
	if got := errWithCause.Error(); got != "test: wrapped" {
		t.Fatalf("CodecError with cause = %q", got)
	}
}

func assertCodecInvalidInput(t *testing.T, err error) {
	t.Helper()
	const code = knifer.ErrCodeInvalidInput
	if err == nil {
		t.Fatalf("err = nil, want %s", code)
	}
	if !errors.Is(err, code) {
		t.Fatalf("errors.Is(%v, %s) = false", err, code)
	}
	got, ok := knifer.CodeOf(err)
	if !ok || got != code {
		t.Fatalf("CodeOf(%v) = %q, %v; want %q, true", err, got, ok, code)
	}
	var codecErr *CodecError
	if !errors.As(err, &codecErr) {
		t.Fatalf("errors.As(err, *CodecError) = false: %v", err)
	}
}

func assertCodecCauseAs(t *testing.T, err error, target any) {
	t.Helper()
	if !errors.As(err, target) {
		t.Fatalf("errors.As(%v, %T) = false", err, target)
	}
}
