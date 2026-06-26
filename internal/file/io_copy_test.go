package file

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	knifer "github.com/imajinyun/knifer-go"
)

var errTestCopy = errors.New("copy test error")

type errorReader struct{}

func (errorReader) Read([]byte) (int, error) { return 0, errTestCopy }

type errorWriter struct{}

func (errorWriter) Write([]byte) (int, error) { return 0, errTestCopy }

type shortWriter struct{}

func (shortWriter) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	return len(p) - 1, nil
}

func TestIoCopyWithOptions(t *testing.T) {
	var dst bytes.Buffer
	n, err := IoCopy(&dst, strings.NewReader("io"))
	if err != nil || n != 2 || dst.String() != "io" {
		t.Fatalf("IoCopy n=%d dst=%q err=%v", n, dst.String(), err)
	}
	dst.Reset()
	n, err = IoCopyWithOptions(&dst, strings.NewReader("abc"), WithBufferSize(2), WithMaxBytes(3))
	if err != nil || n != 3 || dst.String() != "abc" {
		t.Fatalf("IoCopyWithOptions n=%d dst=%q err=%v", n, dst.String(), err)
	}
	if n, err := IoCopyWithOptions(nil, strings.NewReader("x")); !errors.Is(err, knifer.ErrCodeInvalidInput) || n != 0 {
		t.Fatalf("IoCopyWithOptions nil writer n=%d err=%v", n, err)
	}
	if n, err := IoCopyWithOptions(&dst, nil); !errors.Is(err, knifer.ErrCodeInvalidInput) || n != 0 {
		t.Fatalf("IoCopyWithOptions nil reader n=%d err=%v", n, err)
	}
	dst.Reset()
	n, err = IoCopyWithOptions(&dst, strings.NewReader("abcdef"), WithBufferSize(2), WithMaxBytes(4))
	if !errors.Is(err, knifer.ErrCodeInvalidInput) || n != 4 || dst.String() != "abcd" {
		t.Fatalf("IoCopyWithOptions limited n=%d dst=%q err=%v", n, dst.String(), err)
	}
}

func TestIoCopyWithOptionsErrors(t *testing.T) {
	if n, err := IoCopyWithOptions(errorWriter{}, strings.NewReader("x")); !errors.Is(err, errTestCopy) || n != 0 {
		t.Fatalf("IoCopyWithOptions write error n=%d err=%v", n, err)
	}
	if n, err := IoCopyWithOptions(shortWriter{}, strings.NewReader("xy")); !errors.Is(err, io.ErrShortWrite) || n != 1 {
		t.Fatalf("IoCopyWithOptions short write n=%d err=%v", n, err)
	}
	if n, err := IoCopyWithOptions(&bytes.Buffer{}, errorReader{}); !errors.Is(err, errTestCopy) || n != 0 {
		t.Fatalf("IoCopyWithOptions read error n=%d err=%v", n, err)
	}

	var dst bytes.Buffer
	n, err := IoCopyWithOptions(&dst, strings.NewReader("unlimited"), WithBufferSize(3), WithUnlimitedRead())
	if err != nil || n != 9 || dst.String() != "unlimited" {
		t.Fatalf("IoCopyWithOptions unlimited n=%d dst=%q err=%v", n, dst.String(), err)
	}
}
