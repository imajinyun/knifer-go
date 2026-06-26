package vcsv_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/imajinyun/knifer-go/vcsv"
)

type facadePerson struct {
	Name string `csv:"name"`
	Age  int    `csv:"age"`
}

func TestFacadeReadMapsAndWriteString(t *testing.T) {
	rows, err := vcsv.ReadMaps(strings.NewReader("name,age\nalice,30\n"))
	if err != nil {
		t.Fatalf("ReadMaps: %v", err)
	}
	if rows[0]["name"] != "alice" || rows[0]["age"] != "30" {
		t.Fatalf("ReadMaps = %#v", rows)
	}
	out, err := vcsv.WriteString([][]string{{"name"}, {"alice"}})
	if err != nil {
		t.Fatalf("WriteString: %v", err)
	}
	if out != "name\nalice\n" {
		t.Fatalf("WriteString = %q", out)
	}
}

func TestFacadeOptionsAndStructs(t *testing.T) {
	records, err := vcsv.ReadString("\ufeffname;age\nalice;30\n", vcsv.WithComma(';'), vcsv.WithTrimUTF8BOM(true))
	if err != nil {
		t.Fatalf("ReadString: %v", err)
	}
	if records[1][0] != "alice" {
		t.Fatalf("ReadString = %#v", records)
	}
	out, err := vcsv.WriteStringStructs([]facadePerson{{Name: "alice", Age: 30}}, vcsv.WithUTF8BOM(true))
	if err != nil {
		t.Fatalf("WriteStringStructs: %v", err)
	}
	if out != "\ufeffname,age\nalice,30\n" {
		t.Fatalf("WriteStringStructs = %q", out)
	}
}

func TestFacadeAdditionalWrappers(t *testing.T) {
	records, err := vcsv.Read(strings.NewReader("#skip\n name,age\n alice,30\n"),
		vcsv.WithComment('#'),
		vcsv.WithFieldsPerRecord(2),
		vcsv.WithTrimLeadingSpace(true),
		vcsv.WithLazyQuotes(true),
		vcsv.WithReuseRecord(false),
	)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if records[1][0] != "alice" {
		t.Fatalf("Read = %#v", records)
	}

	var seen []string
	if err := vcsv.ForEach(strings.NewReader("a\nb\n"), func(record []string) error {
		seen = append(seen, record[0])
		return nil
	}); err != nil {
		t.Fatalf("ForEach: %v", err)
	}
	if strings.Join(seen, "") != "ab" {
		t.Fatalf("ForEach seen = %#v", seen)
	}

	var b bytes.Buffer
	if err := vcsv.Write(&b, [][]string{{"name"}, {"alice"}}, vcsv.WithUseCRLF(true)); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if b.String() != "name\r\nalice\r\n" {
		t.Fatalf("Write = %q", b.String())
	}

	b.Reset()
	if err := vcsv.WriteMaps(&b, []string{"name"}, []map[string]string{{"name": "alice"}}); err != nil {
		t.Fatalf("WriteMaps: %v", err)
	}
	if b.String() != "name\nalice\n" {
		t.Fatalf("WriteMaps = %q", b.String())
	}

	out, err := vcsv.WriteStringMaps([]string{"name"}, []map[string]string{{"name": "alice"}})
	if err != nil || out != "name\nalice\n" {
		t.Fatalf("WriteStringMaps = %q, %v", out, err)
	}
	rows, err := vcsv.ReadStringMaps(out)
	if err != nil || rows[0]["name"] != "alice" {
		t.Fatalf("ReadStringMaps = %#v, %v", rows, err)
	}
	maps, err := vcsv.RecordsToMaps([][]string{{"name"}, {"alice"}})
	if err != nil || maps[0]["name"] != "alice" {
		t.Fatalf("RecordsToMaps = %#v, %v", maps, err)
	}
	if got := vcsv.MapsToRecords([]string{"name"}, maps); got[1][0] != "alice" {
		t.Fatalf("MapsToRecords = %#v", got)
	}
	structRecords, err := vcsv.StructsToRecords([]facadePerson{{Name: "alice", Age: 30}})
	if err != nil || structRecords[1][0] != "alice" {
		t.Fatalf("StructsToRecords = %#v, %v", structRecords, err)
	}
	b.Reset()
	if err := vcsv.WriteStructs(&b, []facadePerson{{Name: "alice", Age: 30}}); err != nil {
		t.Fatalf("WriteStructs: %v", err)
	}
	if b.String() != "name,age\nalice,30\n" {
		t.Fatalf("WriteStructs = %q", b.String())
	}
}
