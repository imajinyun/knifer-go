package csvx

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	knifer "github.com/imajinyun/knifer-go"
)

type csvPerson struct {
	Name   string `csv:"name"`
	Age    int    `csv:"age"`
	hidden string
	Skip   string `csv:"-"`
}

type csvValue string

func (v csvValue) String() string { return "value:" + string(v) }

func TestReadStringWithOptions(t *testing.T) {
	got, err := ReadString("name;age\nalice;30\n", WithComma(';'))
	if err != nil {
		t.Fatalf("ReadString: %v", err)
	}
	if got[1][0] != "alice" || got[1][1] != "30" {
		t.Fatalf("ReadString = %#v", got)
	}
}

func TestReadMapsAndWriteString(t *testing.T) {
	rows, err := ReadMaps(strings.NewReader("name,age\nalice,30\n"))
	if err != nil {
		t.Fatalf("ReadMaps: %v", err)
	}
	if rows[0]["name"] != "alice" || rows[0]["age"] != "30" {
		t.Fatalf("ReadMaps = %#v", rows)
	}
	out, err := WriteString([][]string{{"name", "age"}, {"alice", "30"}})
	if err != nil {
		t.Fatalf("WriteString: %v", err)
	}
	if out != "name,age\nalice,30\n" {
		t.Fatalf("WriteString = %q", out)
	}
}

func TestForEachStopsOnHandlerError(t *testing.T) {
	want := errors.New("stop")
	err := ForEach(strings.NewReader("a\nb\n"), func(record []string) error {
		if record[0] == "b" {
			return want
		}
		return nil
	})
	if !errors.Is(err, want) || !errors.Is(err, knifer.ErrCodeInternal) {
		t.Fatalf("ForEach error = %v", err)
	}
}

func TestMapsAndStructRecords(t *testing.T) {
	records := MapsToRecords([]string{"name", "age"}, []map[string]string{{"name": "alice", "age": "30"}})
	if records[1][0] != "alice" || records[1][1] != "30" {
		t.Fatalf("MapsToRecords = %#v", records)
	}
	rows, err := RecordsToMaps([][]string{{"name", "age"}, {"alice"}})
	if err != nil {
		t.Fatalf("RecordsToMaps: %v", err)
	}
	if rows[0]["age"] != "" {
		t.Fatalf("missing field should become empty string: %#v", rows)
	}
	structRecords, err := StructsToRecords([]csvPerson{{Name: "alice", Age: 30, hidden: "x", Skip: "skip"}})
	if err != nil {
		t.Fatalf("StructsToRecords: %v", err)
	}
	if got := structRecords[0]; len(got) != 2 || got[0] != "name" || got[1] != "age" {
		t.Fatalf("StructsToRecords header = %#v", got)
	}
	if got := structRecords[1]; got[0] != "alice" || got[1] != "30" {
		t.Fatalf("StructsToRecords row = %#v", got)
	}
}

func TestReadOptionsAndStringMaps(t *testing.T) {
	rows, err := ReadStringMaps("# ignored\n name,age\n alice,30,extra\n",
		WithComment('#'),
		WithTrimLeadingSpace(true),
		WithFieldsPerRecord(-1),
	)
	if err != nil {
		t.Fatalf("ReadStringMaps: %v", err)
	}
	if rows[0]["name"] != "alice" || rows[0]["age"] != "30" {
		t.Fatalf("ReadStringMaps = %#v", rows)
	}

	records, err := ReadString("name,quote\nalice,\"hello\n", WithLazyQuotes(true), WithReuseRecord(true))
	if err != nil {
		t.Fatalf("ReadString lazy quotes: %v", err)
	}
	if records[1][1] != "hello\n" {
		t.Fatalf("ReadString lazy quote field = %#v", records)
	}
}

func TestReadStringTrimsUTF8BOMByDefault(t *testing.T) {
	rows, err := ReadStringMaps("\ufeffname,age\nalice,30\n")
	if err != nil {
		t.Fatalf("ReadStringMaps BOM: %v", err)
	}
	if rows[0]["name"] != "alice" || rows[0]["\ufeffname"] != "" {
		t.Fatalf("ReadStringMaps BOM rows = %#v", rows)
	}

	records, err := ReadString("\ufeffname,age\nalice,30\n", WithTrimUTF8BOM(false))
	if err != nil {
		t.Fatalf("ReadString keep BOM: %v", err)
	}
	if records[0][0] != "\ufeffname" {
		t.Fatalf("ReadString keep BOM records = %#v", records)
	}
}

func TestWriteMapsStructsAndOptions(t *testing.T) {
	out, err := WriteStringMaps([]string{"name", "age"}, []map[string]string{{"name": "alice", "age": "30"}},
		WithComma(';'),
		WithUseCRLF(true),
	)
	if err != nil {
		t.Fatalf("WriteStringMaps: %v", err)
	}
	if out != "name;age\r\nalice;30\r\n" {
		t.Fatalf("WriteStringMaps = %q", out)
	}

	var b strings.Builder
	if err := WriteStructs(&b, []csvPerson{{Name: "bob", Age: 40}}); err != nil {
		t.Fatalf("WriteStructs: %v", err)
	}
	if b.String() != "name,age\nbob,40\n" {
		t.Fatalf("WriteStructs = %q", b.String())
	}

	bomOut, err := WriteString([][]string{{"name"}, {"alice"}}, WithUTF8BOM(true))
	if err != nil {
		t.Fatalf("WriteString BOM: %v", err)
	}
	if bomOut != "\ufeffname\nalice\n" {
		t.Fatalf("WriteString BOM = %q", bomOut)
	}
}

func TestStructsToRecordsPointersAndScalarKinds(t *testing.T) {
	type scalarRow struct {
		Name  *string  `csv:"name"`
		OK    bool     `csv:"ok"`
		Count uint     `csv:"count"`
		Rate  float64  `csv:"rate"`
		Value csvValue `csv:"value"`
	}
	name := "alice"
	records, err := StructsToRecords([]*scalarRow{
		{Name: &name, OK: true, Count: 2, Rate: 1.5, Value: "x"},
		nil,
	})
	if err != nil {
		t.Fatalf("StructsToRecords: %v", err)
	}
	if got := records[1]; got[0] != "alice" || got[1] != "true" || got[2] != "2" || got[3] != "1.5" || got[4] != "value:x" {
		t.Fatalf("StructsToRecords scalar row = %#v", got)
	}
	if got := records[2]; strings.Join(got, ",") != ",,,," {
		t.Fatalf("StructsToRecords nil pointer row = %#v", got)
	}
}

func TestWriteMaps(t *testing.T) {
	var b strings.Builder
	err := WriteMaps(&b, []string{"name", "age"}, []map[string]string{{"name": "alice", "age": "30"}})
	if err != nil {
		t.Fatalf("WriteMaps() error = %v", err)
	}
	if b.String() != "name,age\nalice,30\n" {
		t.Fatalf("WriteMaps() = %q", b.String())
	}
}

func TestWriteStringStructs(t *testing.T) {
	out, err := WriteStringStructs([]csvPerson{{Name: "bob", Age: 40}})
	if err != nil {
		t.Fatalf("WriteStringStructs() error = %v", err)
	}
	if out != "name,age\nbob,40\n" {
		t.Fatalf("WriteStringStructs() = %q", out)
	}
}

func TestErrorWrappingAndParseErrors(t *testing.T) {
	_, err := ReadString("\"unterminated\n")
	if !errors.Is(err, knifer.ErrCodeInvalidInput) {
		t.Fatalf("parse error = %v", err)
	}
	var csvErr *CSVError
	if !errors.As(err, &csvErr) || csvErr.ErrorCode() != knifer.ErrCodeInvalidInput || csvErr.Unwrap() == nil {
		t.Fatalf("wrapped parse error = %#v", err)
	}
	if !csvErr.Is(&CSVError{Code: knifer.ErrCodeInvalidInput}) {
		t.Fatalf("CSVError.Is did not match same code")
	}

	if err := ForEach(strings.NewReader("a\n"), nil); !errors.Is(err, knifer.ErrCodeInvalidInput) {
		t.Fatalf("ForEach nil handler error = %v", err)
	}
}

func TestWritePropagatesWriterError(t *testing.T) {
	err := Write(errorWriter{}, [][]string{{"a"}})
	if !errors.Is(err, knifer.ErrCodeInternal) {
		t.Fatalf("Write error = %v", err)
	}
}

type errorWriter struct{}

func (errorWriter) Write([]byte) (int, error) { return 0, fmt.Errorf("write failed") }

func TestInvalidInputsReturnCodeErrors(t *testing.T) {
	if _, err := Read(nil); !errors.Is(err, knifer.ErrCodeInvalidInput) {
		t.Fatalf("Read(nil) error = %v", err)
	}
	if err := Write(nil, nil); !errors.Is(err, knifer.ErrCodeInvalidInput) {
		t.Fatalf("Write(nil) error = %v", err)
	}
	if _, err := StructsToRecords(csvPerson{}); !errors.Is(err, knifer.ErrCodeInvalidInput) {
		t.Fatalf("StructsToRecords(non-slice) error = %v", err)
	}
}
