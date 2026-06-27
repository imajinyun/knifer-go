package str

import (
	"bytes"
	"errors"
	"testing"

	knifer "github.com/imajinyun/knifer-go"
)

func TestBOM(t *testing.T) {
	data := []byte{0xEF, 0xBB, 0xBF, 'g', 'o'}
	if got := HasBOM(data); got != BOMUTF8 {
		t.Fatalf("HasBOM = %q", got)
	}
	if got := StripBOM(data); !bytes.Equal(got, []byte("go")) {
		t.Fatalf("StripBOM = %v", got)
	}
	if got := StripBOM([]byte("go")); !bytes.Equal(got, []byte("go")) {
		t.Fatalf("StripBOM without bom = %v", got)
	}
}

func TestCharsetRoundTrip(t *testing.T) {
	gbk, err := FromUTF8([]byte("中文"), "gbk")
	if err != nil {
		t.Fatalf("FromUTF8 error = %v", err)
	}
	utf8, err := ToUTF8(append([]byte{0xEF, 0xBB, 0xBF}, gbk...), "gbk")
	if err != nil {
		t.Fatalf("ToUTF8 error = %v", err)
	}
	if string(utf8) != "中文" {
		t.Fatalf("ToUTF8 = %q", utf8)
	}
}

func TestCharsetUnsupported(t *testing.T) {
	_, err := ToUTF8([]byte("x"), "unknown-charset")
	if !errors.Is(err, knifer.ErrCodeUnsupported) {
		t.Fatalf("ToUTF8 unsupported err = %v", err)
	}
}
