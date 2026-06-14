package poi

import (
	"bytes"
	"reflect"
	"testing"
)

func TestRowsBufferRoundTrip(t *testing.T) {
	rows := [][]string{{"a", "b"}, {"1", "2"}}
	buf, err := WriteRowsToBuffer("Data", rows)
	if err != nil {
		t.Fatalf("WriteRowsToBuffer: %v", err)
	}
	if buf.Len() == 0 {
		t.Fatal("buffer is empty")
	}

	got, err := ReadRowsFromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("ReadRowsFromReader: %v", err)
	}
	if !reflect.DeepEqual(got, rows) {
		t.Fatalf("ReadRowsFromReader = %#v, want %#v", got, rows)
	}
}

func TestWriteRowsToBufferOptions(t *testing.T) {
	rows := [][]string{{"x"}}
	buf, err := WriteRowsToBuffer("Data", rows, WithStartCell(2, 2))
	if err != nil {
		t.Fatalf("WriteRowsToBuffer with options: %v", err)
	}
	got, err := ReadRowsFromReader(bytes.NewReader(buf.Bytes()), WithReadSheet("Data"))
	if err != nil {
		t.Fatalf("ReadRowsFromReader: %v", err)
	}
	want := [][]string{nil, {"", "x"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("rows = %#v, want %#v", got, want)
	}
}
