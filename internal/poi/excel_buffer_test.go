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

func BenchmarkWriteRowsToBuffer(b *testing.B) {
	rows := [][]string{{"id", "name", "score"}, {"1", "alice", "100"}, {"2", "bob", "95"}}
	b.ReportAllocs()
	var sink int
	for b.Loop() {
		buf, err := WriteRowsToBuffer("Scores", rows)
		if err != nil {
			b.Fatal(err)
		}
		sink = buf.Len()
	}
	_ = sink
}
