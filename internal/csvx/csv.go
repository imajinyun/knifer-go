package csvx

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"reflect"
	"slices"
	"strconv"
	"strings"

	knifer "github.com/imajinyun/knifer-go"
)

type config struct {
	comma            rune
	comment          rune
	fieldsPerRecord  int
	lazyQuotes       bool
	trimLeadingSpace bool
	reuseRecord      bool
	useCRLF          bool
	trimUTF8BOM      bool
	writeUTF8BOM     bool
}

var utf8BOM = []byte{0xEF, 0xBB, 0xBF}

// Option customizes CSV reader and writer helpers.
type Option func(*config)

// ReadOption customizes CSV read helpers.
type ReadOption = Option

// WriteOption customizes CSV write helpers.
type WriteOption = Option

func defaultConfig() config {
	return config{
		comma:       ',',
		trimUTF8BOM: true,
	}
}

// WithComma sets the field delimiter used by readers and writers.
func WithComma(comma rune) Option {
	return func(c *config) {
		if comma != 0 {
			c.comma = comma
		}
	}
}

// WithComment sets the comment character used by readers. Zero disables comments.
func WithComment(comment rune) ReadOption {
	return func(c *config) { c.comment = comment }
}

// WithFieldsPerRecord sets the expected number of fields per record.
//
// It follows encoding/csv.Reader semantics: positive values enforce a fixed
// field count, zero infers the count from the first record, and negative values
// allow variable field counts.
func WithFieldsPerRecord(n int) ReadOption {
	return func(c *config) { c.fieldsPerRecord = n }
}

// WithLazyQuotes allows lazy quote handling in read helpers.
func WithLazyQuotes(enabled bool) ReadOption {
	return func(c *config) { c.lazyQuotes = enabled }
}

// WithTrimLeadingSpace trims leading space in unquoted fields when reading.
func WithTrimLeadingSpace(enabled bool) ReadOption {
	return func(c *config) { c.trimLeadingSpace = enabled }
}

// WithReuseRecord allows readers to reuse backing arrays between records.
func WithReuseRecord(enabled bool) ReadOption {
	return func(c *config) { c.reuseRecord = enabled }
}

// WithTrimUTF8BOM controls whether read helpers remove a leading UTF-8 BOM.
func WithTrimUTF8BOM(enabled bool) ReadOption {
	return func(c *config) { c.trimUTF8BOM = enabled }
}

// WithUTF8BOM controls whether write helpers prepend a UTF-8 BOM.
func WithUTF8BOM(enabled bool) WriteOption {
	return func(c *config) { c.writeUTF8BOM = enabled }
}

// WithUseCRLF makes writers terminate records with \r\n.
func WithUseCRLF(enabled bool) WriteOption {
	return func(c *config) { c.useCRLF = enabled }
}

// Read reads all CSV records from r.
func Read(r io.Reader, opts ...ReadOption) ([][]string, error) {
	if r == nil {
		return nil, invalidInput("csv reader is nil")
	}
	reader := newReader(r, opts...)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, wrapCSVError(knifer.ErrCodeInvalidInput, "read csv records", err)
	}
	return records, nil
}

// ReadString reads all CSV records from s.
func ReadString(s string, opts ...ReadOption) ([][]string, error) {
	return Read(strings.NewReader(s), opts...)
}

// ReadMaps reads CSV records into maps keyed by the header row.
func ReadMaps(r io.Reader, opts ...ReadOption) ([]map[string]string, error) {
	records, err := Read(r, opts...)
	if err != nil {
		return nil, err
	}
	return RecordsToMaps(records)
}

// ReadStringMaps reads CSV records from s into maps keyed by the header row.
func ReadStringMaps(s string, opts ...ReadOption) ([]map[string]string, error) {
	return ReadMaps(strings.NewReader(s), opts...)
}

// ForEach reads CSV records from r and invokes handle for each record.
func ForEach(r io.Reader, handle func([]string) error, opts ...ReadOption) error {
	if r == nil {
		return invalidInput("csv reader is nil")
	}
	if handle == nil {
		return invalidInput("csv record handler is nil")
	}
	reader := newReader(r, opts...)
	for {
		record, err := reader.Read()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return wrapCSVError(knifer.ErrCodeInvalidInput, "read csv record", err)
		}
		if err := handle(record); err != nil {
			return wrapCSVError(knifer.ErrCodeInternal, "handle csv record", err)
		}
	}
}

// Write writes CSV records to w.
func Write(w io.Writer, records [][]string, opts ...WriteOption) error {
	if w == nil {
		return invalidInput("csv writer is nil")
	}
	cfg := applyOptions(opts...)
	if cfg.writeUTF8BOM {
		if _, err := w.Write(utf8BOM); err != nil {
			return wrapCSVError(knifer.ErrCodeInternal, "write csv bom", err)
		}
	}
	writer := newWriterWithConfig(w, cfg)
	if err := writer.WriteAll(records); err != nil {
		return wrapCSVError(knifer.ErrCodeInternal, "write csv records", err)
	}
	if err := writer.Error(); err != nil {
		return wrapCSVError(knifer.ErrCodeInternal, "flush csv records", err)
	}
	return nil
}

// WriteString writes CSV records into a string.
func WriteString(records [][]string, opts ...WriteOption) (string, error) {
	var b strings.Builder
	if err := Write(&b, records, opts...); err != nil {
		return "", err
	}
	return b.String(), nil
}

// WriteMaps writes maps using headers as the output column order.
func WriteMaps(w io.Writer, headers []string, rows []map[string]string, opts ...WriteOption) error {
	return Write(w, MapsToRecords(headers, rows), opts...)
}

// WriteStringMaps writes maps into a CSV string using headers as column order.
func WriteStringMaps(headers []string, rows []map[string]string, opts ...WriteOption) (string, error) {
	return WriteString(MapsToRecords(headers, rows), opts...)
}

// RecordsToMaps converts records into maps keyed by the first row.
func RecordsToMaps(records [][]string) ([]map[string]string, error) {
	if len(records) == 0 {
		return nil, nil
	}
	headers := records[0]
	rows := make([]map[string]string, 0, len(records)-1)
	for _, record := range records[1:] {
		row := make(map[string]string, len(headers))
		for i, header := range headers {
			if header == "" {
				continue
			}
			if i < len(record) {
				row[header] = record[i]
			} else {
				row[header] = ""
			}
		}
		rows = append(rows, row)
	}
	return rows, nil
}

// MapsToRecords converts maps into records using headers as column order.
func MapsToRecords(headers []string, rows []map[string]string) [][]string {
	records := make([][]string, 0, len(rows)+1)
	records = append(records, slices.Clone(headers))
	for _, row := range rows {
		record := make([]string, len(headers))
		for i, header := range headers {
			record[i] = row[header]
		}
		records = append(records, record)
	}
	return records
}

// StructsToRecords converts a slice of structs into CSV records.
//
// Exported fields are included in declaration order. The csv tag overrides the
// header name, and `csv:"-"` skips a field.
func StructsToRecords(values any) ([][]string, error) {
	rv := reflect.ValueOf(values)
	if !rv.IsValid() {
		return nil, invalidInput("csv structs value is nil")
	}
	if rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return nil, invalidInput("csv structs pointer is nil")
		}
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
		return nil, invalidInput("csv structs value must be a slice or array")
	}
	if rv.Len() == 0 {
		return nil, nil
	}
	elemType := rv.Type().Elem()
	if elemType.Kind() == reflect.Pointer {
		elemType = elemType.Elem()
	}
	if elemType.Kind() != reflect.Struct {
		return nil, invalidInput("csv structs element must be a struct")
	}
	fields := csvFields(elemType)
	headers := make([]string, len(fields))
	for i, field := range fields {
		headers[i] = field.header
	}
	records := make([][]string, 0, rv.Len()+1)
	records = append(records, headers)
	for i := 0; i < rv.Len(); i++ {
		elem := rv.Index(i)
		if elem.Kind() == reflect.Pointer {
			if elem.IsNil() {
				records = append(records, make([]string, len(fields)))
				continue
			}
			elem = elem.Elem()
		}
		record := make([]string, len(fields))
		for j, field := range fields {
			record[j] = formatValue(elem.Field(field.index))
		}
		records = append(records, record)
	}
	return records, nil
}

// WriteStructs writes a slice of structs as CSV records.
func WriteStructs(w io.Writer, values any, opts ...WriteOption) error {
	records, err := StructsToRecords(values)
	if err != nil {
		return err
	}
	return Write(w, records, opts...)
}

// WriteStringStructs writes a slice of structs into a CSV string.
func WriteStringStructs(values any, opts ...WriteOption) (string, error) {
	records, err := StructsToRecords(values)
	if err != nil {
		return "", err
	}
	return WriteString(records, opts...)
}

func newReader(r io.Reader, opts ...ReadOption) *csv.Reader {
	cfg := applyOptions(opts...)
	if cfg.trimUTF8BOM {
		r = trimUTF8BOM(r)
	}
	reader := csv.NewReader(r)
	reader.Comma = cfg.comma
	reader.Comment = cfg.comment
	reader.FieldsPerRecord = cfg.fieldsPerRecord
	reader.LazyQuotes = cfg.lazyQuotes
	reader.TrimLeadingSpace = cfg.trimLeadingSpace
	reader.ReuseRecord = cfg.reuseRecord
	return reader
}

func applyOptions(opts ...Option) config {
	cfg := defaultConfig()
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	return cfg
}

func newWriterWithConfig(w io.Writer, cfg config) *csv.Writer {
	writer := csv.NewWriter(w)
	writer.Comma = cfg.comma
	writer.UseCRLF = cfg.useCRLF
	return writer
}

func trimUTF8BOM(r io.Reader) io.Reader {
	br := bufio.NewReader(r)
	prefix, err := br.Peek(len(utf8BOM))
	if err == nil && bytes.Equal(prefix, utf8BOM) {
		_, _ = br.Discard(len(utf8BOM))
	}
	return br
}

type csvField struct {
	index  int
	header string
}

func csvFields(typ reflect.Type) []csvField {
	fields := make([]csvField, 0, typ.NumField())
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if field.PkgPath != "" {
			continue
		}
		name := field.Tag.Get("csv")
		if name == "-" {
			continue
		}
		if idx := strings.IndexByte(name, ','); idx >= 0 {
			name = name[:idx]
		}
		if name == "" {
			name = field.Name
		}
		fields = append(fields, csvField{index: i, header: name})
	}
	return fields
}

func formatValue(value reflect.Value) string {
	if !value.IsValid() {
		return ""
	}
	if value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return ""
		}
		value = value.Elem()
	}
	if value.CanInterface() {
		if s, ok := value.Interface().(fmt.Stringer); ok {
			return s.String()
		}
	}
	switch value.Kind() {
	case reflect.String:
		return value.String()
	case reflect.Bool:
		return strconv.FormatBool(value.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(value.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return strconv.FormatUint(value.Uint(), 10)
	case reflect.Float32:
		return strconv.FormatFloat(value.Float(), 'f', -1, 32)
	case reflect.Float64:
		return strconv.FormatFloat(value.Float(), 'f', -1, 64)
	default:
		if !value.CanInterface() {
			return ""
		}
		return fmt.Sprint(value.Interface())
	}
}
