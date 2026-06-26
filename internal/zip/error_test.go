package zip

import (
	"bytes"
	"io"
	"path/filepath"
	"testing"

	knifer "github.com/imajinyun/knifer-go"
)

func TestZipErrorContract(t *testing.T) {
	_, err := GetStream(nil)
	assertZipCode(t, err, knifer.ErrCodeInvalidInput)
	assertZipCode(t, UnzipReaderToLimit(nil, t.TempDir(), -1), knifer.ErrCodeInvalidInput)

	var buf bytes.Buffer
	err = ZipEntriesToWriter(&buf, EntryData{Name: "../evil.txt", Data: []byte("bad")})
	assertZipCode(t, err, knifer.ErrCodeInvalidInput)
}

func TestZipNilBoundaryErrors(t *testing.T) {
	assertZipCode(t, ZipEntriesToWriter(nil, EntryData{Name: "a.txt", Data: []byte("a")}), knifer.ErrCodeInvalidInput)

	var buf bytes.Buffer
	err := ZipStreamsToWriterWithOptions(nil, []StreamEntry{{Name: "a.txt", Reader: bytes.NewReader([]byte("a"))}})
	assertZipCode(t, err, knifer.ErrCodeInvalidInput)
	err = ZipStreamsToWriterWithOptions(&buf, []StreamEntry{{Name: "a.txt"}})
	assertZipCode(t, err, knifer.ErrCodeInvalidInput)

	_, err = readAllLimit(nil, 1)
	assertZipCode(t, err, knifer.ErrCodeInvalidInput)
	_, err = copyLimit(io.Discard, nil, 1)
	assertZipCode(t, err, knifer.ErrCodeInvalidInput)
	_, err = copyLimit(nil, bytes.NewReader([]byte("a")), 1)
	assertZipCode(t, err, knifer.ErrCodeInvalidInput)

	archive := filepath.Join(t.TempDir(), "entries.zip")
	if err := ZipEntries(archive, EntryData{Name: "a.txt", Data: []byte("a")}); err != nil {
		t.Fatalf("ZipEntries() error = %v", err)
	}
	assertZipCode(t, ReadWithOptions(archive, nil), knifer.ErrCodeInvalidInput)
}
