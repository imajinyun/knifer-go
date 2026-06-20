package vpoi_test

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/imajinyun/go-knifer/vpoi"
)

func TestExcelFacadeBufferRoundTrip(t *testing.T) {
	rows := [][]string{{"id", "name"}, {"1", "alice"}}
	buf, err := vpoi.WriteRowsToBuffer("Users", rows)
	if err != nil {
		t.Fatalf("WriteRowsToBuffer: %v", err)
	}

	got, err := vpoi.ReadRowsFromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("ReadRowsFromReader: %v", err)
	}
	if !reflect.DeepEqual(got, rows) {
		t.Fatalf("ReadRowsFromReader = %#v, want %#v", got, rows)
	}
}

func BenchmarkExcelFacadeWriteRowsToBuffer(b *testing.B) {
	rows := [][]string{{"id", "name", "score"}, {"1", "alice", "100"}, {"2", "bob", "95"}}
	b.ReportAllocs()
	var sink int
	for b.Loop() {
		buf, err := vpoi.WriteRowsToBuffer("Scores", rows)
		if err != nil {
			b.Fatal(err)
		}
		sink = buf.Len()
	}
	_ = sink
}
