package vcsv

import (
	"io"

	csvx "github.com/imajinyun/knifer-go/internal/csvx"
)

type (
	// Option customizes CSV reader and writer helpers.
	Option = csvx.Option
	// ReadOption customizes CSV read helpers.
	ReadOption = csvx.ReadOption
	// WriteOption customizes CSV write helpers.
	WriteOption = csvx.WriteOption
	// Error is the CSV module error type.
	Error = csvx.CSVError
)

func WithComma(comma rune) Option          { return csvx.WithComma(comma) }
func WithComment(comment rune) ReadOption  { return csvx.WithComment(comment) }
func WithFieldsPerRecord(n int) ReadOption { return csvx.WithFieldsPerRecord(n) }

func WithLazyQuotes(enabled bool) ReadOption { return csvx.WithLazyQuotes(enabled) }

func WithTrimLeadingSpace(enabled bool) ReadOption { return csvx.WithTrimLeadingSpace(enabled) }

func WithReuseRecord(enabled bool) ReadOption                  { return csvx.WithReuseRecord(enabled) }
func WithTrimUTF8BOM(enabled bool) ReadOption                  { return csvx.WithTrimUTF8BOM(enabled) }
func WithUTF8BOM(enabled bool) WriteOption                     { return csvx.WithUTF8BOM(enabled) }
func WithUseCRLF(enabled bool) WriteOption                     { return csvx.WithUseCRLF(enabled) }
func Read(r io.Reader, opts ...ReadOption) ([][]string, error) { return csvx.Read(r, opts...) }
func ReadString(s string, opts ...ReadOption) ([][]string, error) {
	return csvx.ReadString(s, opts...)
}

func ReadMaps(r io.Reader, opts ...ReadOption) ([]map[string]string, error) {
	return csvx.ReadMaps(r, opts...)
}

func ReadStringMaps(s string, opts ...ReadOption) ([]map[string]string, error) {
	return csvx.ReadStringMaps(s, opts...)
}

func ForEach(r io.Reader, handle func([]string) error, opts ...ReadOption) error {
	return csvx.ForEach(r, handle, opts...)
}

func Write(w io.Writer, records [][]string, opts ...WriteOption) error {
	return csvx.Write(w, records, opts...)
}

func WriteString(records [][]string, opts ...WriteOption) (string, error) {
	return csvx.WriteString(records, opts...)
}

func WriteMaps(w io.Writer, headers []string, rows []map[string]string, opts ...WriteOption) error {
	return csvx.WriteMaps(w, headers, rows, opts...)
}

func WriteStringMaps(headers []string, rows []map[string]string, opts ...WriteOption) (string, error) {
	return csvx.WriteStringMaps(headers, rows, opts...)
}

func RecordsToMaps(records [][]string) ([]map[string]string, error) {
	return csvx.RecordsToMaps(records)
}

func MapsToRecords(headers []string, rows []map[string]string) [][]string {
	return csvx.MapsToRecords(headers, rows)
}
func StructsToRecords(values any) ([][]string, error) { return csvx.StructsToRecords(values) }
func WriteStructs(w io.Writer, values any, opts ...WriteOption) error {
	return csvx.WriteStructs(w, values, opts...)
}

func WriteStringStructs(values any, opts ...WriteOption) (string, error) {
	return csvx.WriteStringStructs(values, opts...)
}
