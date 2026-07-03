package db

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	knifer "github.com/imajinyun/knifer-go"
)

func TestBuildCountSQL(t *testing.T) {
	sqlText, args, err := buildCountSQL(DialectPostgres, WrapperForDialect(DialectPostgres), []string{"users"}, Eq("status", "active"))
	if err != nil {
		t.Fatalf("buildCountSQL() error = %v", err)
	}
	if sqlText != `SELECT COUNT(*) FROM "users" WHERE "status" = $1` {
		t.Fatalf("sql = %q", sqlText)
	}
	if !reflect.DeepEqual(args, []any{"active"}) {
		t.Fatalf("args = %#v", args)
	}

	_, _, err = buildCountSQL(DialectQuestion, Wrapper{}, []string{"users; drop table users"})
	assertDBCode(t, err, knifer.ErrCodeInvalidInput)
}

func TestListColumnsSQLRejectsUnsafeTable(t *testing.T) {
	_, _, _, err := listColumnsSQL(DialectSQLite, "users; drop table users")
	assertDBCode(t, err, knifer.ErrCodeInvalidInput)
}

func TestScanAndMetaHelpersReportInvalidInputAndUnsupported(t *testing.T) {
	_, err := ScanRows(nil)
	assertDBCode(t, err, knifer.ErrCodeInvalidInput)

	err = AssignEntity(NewEntity("users"), nil)
	assertDBCode(t, err, knifer.ErrCodeInvalidInput)

	_, err = listTablesSQL(DialectOracle)
	assertDBCode(t, err, knifer.ErrCodeUnsupported)

	_, _, _, err = listColumnsSQL(DialectOracle, "users")
	assertDBCode(t, err, knifer.ErrCodeUnsupported)
}

func TestScanRowsNormalizesBytesAndReportsIteratorErrors(t *testing.T) {
	rowsDB := newFakeDB(&fakeBehavior{queryFunc: func(string) (driver.Rows, error) {
		return mkRows(
			[]string{"id", "name", "created_at"},
			[]driver.Value{int64(1), []byte("alice"), time.Unix(123, 0).UTC()},
		), nil
	}})
	defer func() { _ = rowsDB.Close() }()
	rows, err := rowsDB.sqlDB.Query("SELECT id, name, created_at FROM users")
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	entities, err := ScanRows(rows)
	if err != nil {
		t.Fatalf("ScanRows: %v", err)
	}
	if len(entities) != 1 || entities[0].Values["name"] != "alice" {
		t.Fatalf("ScanRows entities = %#v", entities)
	}
	if _, ok := entities[0].Values["created_at"].(time.Time); !ok {
		t.Fatalf("created_at type = %T, want time.Time", entities[0].Values["created_at"])
	}

	iterDB := newFakeDB(&fakeBehavior{queryFunc: func(string) (driver.Rows, error) {
		return &fakeRows{cols: []string{"id"}, nextErr: errors.New("iterator boom")}, nil
	}})
	defer func() { _ = iterDB.Close() }()
	iterRows, err := iterDB.sqlDB.Query("SELECT id FROM users")
	if err != nil {
		t.Fatalf("Query iterator rows: %v", err)
	}
	errEntities, err := ScanRows(iterRows)
	if err == nil || len(errEntities) != 0 {
		t.Fatalf("ScanRows iterator error entities=%#v err=%v", errEntities, err)
	}
	if !errors.Is(err, knifer.ErrCodeInternal) || !strings.Contains(err.Error(), "iterate rows") {
		t.Fatalf("ScanRows iterator err = %v, want internal iterate error", err)
	}
}

func TestScanOneOnlyScansFirstRowAndCloses(t *testing.T) {
	closed := false
	rowsDB := newFakeDB(&fakeBehavior{queryFunc: func(string) (driver.Rows, error) {
		return &fakeRows{
			cols:    []string{"id", "name"},
			data:    [][]driver.Value{{int64(1), []byte("alice")}},
			nextErr: errors.New("second row should not be read"),
			closeFn: func() { closed = true },
		}, nil
	}})
	defer func() { _ = rowsDB.Close() }()
	rows, err := rowsDB.sqlDB.Query("SELECT id, name FROM users")
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	entity, ok, err := ScanOne(rows)
	if err != nil {
		t.Fatalf("ScanOne should ignore later iterator error after first row: %v", err)
	}
	if !ok || entity.Values["id"] != int64(1) || entity.Values["name"] != "alice" {
		t.Fatalf("ScanOne entity=%#v ok=%v", entity, ok)
	}
	if !closed {
		t.Fatal("ScanOne should close rows")
	}

	allRowsDB := newFakeDB(&fakeBehavior{queryFunc: func(string) (driver.Rows, error) {
		return &fakeRows{
			cols:    []string{"id"},
			data:    [][]driver.Value{{int64(1)}},
			nextErr: errors.New("iterator boom"),
		}, nil
	}})
	defer func() { _ = allRowsDB.Close() }()
	allRows, err := allRowsDB.sqlDB.Query("SELECT id FROM users")
	if err != nil {
		t.Fatalf("Query all rows: %v", err)
	}
	if _, err := ScanRows(allRows); !errors.Is(err, knifer.ErrCodeInternal) {
		t.Fatalf("ScanRows iterator err = %v, want ErrCodeInternal", err)
	}
}

func TestDBErrorsAndOptions(t *testing.T) {
	cause := errors.New("driver down")
	if got := (*DBError)(nil).Error(); got != "" {
		t.Fatalf("nil DBError Error = %q", got)
	}
	if got := (*DBError)(nil).ErrorCode(); got != "" {
		t.Fatalf("nil DBError code = %q", got)
	}
	if got := (*DBError)(nil).Unwrap(); got != nil {
		t.Fatalf("nil DBError unwrap = %v", got)
	}
	if wrapDBError(knifer.ErrCodeInternal, "ignored", nil) != nil {
		t.Fatal("wrapDBError nil cause should return nil")
	}
	err := wrapInternal("db failed", cause)
	if !errors.Is(err, cause) {
		t.Fatalf("wrapInternal should unwrap cause, got %v", err)
	}
	assertDBCode(t, err, knifer.ErrCodeInternal)

	opened := false
	cfg := applyOptions(
		WithMaxOpenConns(5),
		WithMaxIdleConns(2),
		WithConnMaxLifetime(time.Minute),
		WithConnMaxIdleTime(time.Second),
		WithSQLOpenFunc(func(driverName, dataSourceName string) (*sql.DB, error) {
			opened = driverName == "test" && dataSourceName == "dsn"
			return nil, cause
		}),
	)
	if cfg.MaxOpenConns != 5 || cfg.MaxIdleConns != 2 || cfg.ConnMaxLifetime != time.Minute || cfg.ConnMaxIdleTime != time.Second {
		t.Fatalf("options = %#v", cfg)
	}
	_, openErr := cfg.SQLOpen("test", "dsn")
	if !opened || !errors.Is(openErr, cause) {
		t.Fatalf("custom SQLOpen opened=%v err=%v", opened, openErr)
	}
	defaultCfg := applyOptions(WithSQLOpenFunc(nil))
	if defaultCfg.SQLOpen == nil {
		t.Fatal("default SQLOpen should be populated")
	}
}

func TestAssignEntityStructAndMap(t *testing.T) {
	type userDTO struct {
		ID       int64  `db:"id"`
		FullName string `json:"full_name"`
		Email    string `bean:"email"`
		Score    int64
		Active   bool
		ignored  string
	}

	entity := EntityFromMap("users", map[string]any{
		"id":        int32(7),
		"full_name": "Alice",
		"email":     true,
		"score":     nil,
		"active":    true,
		"ignored":   "hidden",
	})
	var dto userDTO
	if err := AssignEntity(entity, &dto); err != nil {
		t.Fatalf("AssignEntity struct: %v", err)
	}
	if dto.ID != 7 || dto.FullName != "Alice" || dto.Email != "true" || dto.Score != 0 || !dto.Active || dto.ignored != "" {
		t.Fatalf("assigned dto = %#v", dto)
	}

	var out map[string]any
	if err := AssignEntity(entity, &out); err != nil {
		t.Fatalf("AssignEntity map: %v", err)
	}
	if out["id"] != int32(7) || out["full_name"] != "Alice" {
		t.Fatalf("assigned map = %#v", out)
	}
}

func TestAssignEntityInvalidTargetsAndTypeMismatch(t *testing.T) {
	entity := EntityFromMap("users", map[string]any{"age": "not-int"})
	var nonStruct int
	for name, target := range map[string]any{
		"non-pointer": struct{}{},
		"nil-pointer": (*struct{})(nil),
		"non-struct":  &nonStruct,
	} {
		t.Run(name, func(t *testing.T) {
			err := AssignEntity(entity, target)
			assertDBCode(t, err, knifer.ErrCodeInvalidInput)
		})
	}

	var dst struct{ Age int }
	err := AssignEntity(entity, &dst)
	assertDBCode(t, err, knifer.ErrCodeInternal)
}

func TestAssignEntityRejectsNumericOverflow(t *testing.T) {
	t.Run("signed integer overflow", func(t *testing.T) {
		entity := EntityFromMap("users", map[string]any{"age": int64(128)})
		var dst struct{ Age int8 }
		err := AssignEntity(entity, &dst)
		assertDBCode(t, err, knifer.ErrCodeInternal)
		if dst.Age != 0 {
			t.Fatalf("Age = %d, want unchanged zero value", dst.Age)
		}
	})

	t.Run("negative to unsigned", func(t *testing.T) {
		entity := EntityFromMap("users", map[string]any{"age": int64(-1)})
		var dst struct{ Age uint8 }
		err := AssignEntity(entity, &dst)
		assertDBCode(t, err, knifer.ErrCodeInternal)
		if dst.Age != 0 {
			t.Fatalf("Age = %d, want unchanged zero value", dst.Age)
		}
	})

	t.Run("fractional float to integer", func(t *testing.T) {
		entity := EntityFromMap("users", map[string]any{"age": 1.5})
		var dst struct{ Age int }
		err := AssignEntity(entity, &dst)
		assertDBCode(t, err, knifer.ErrCodeInternal)
		if dst.Age != 0 {
			t.Fatalf("Age = %d, want unchanged zero value", dst.Age)
		}
	})
}

func TestAssignEntityAllowsSafeNumericConversion(t *testing.T) {
	entity := EntityFromMap("users", map[string]any{
		"small_int":  int64(127),
		"small_uint": uint16(255),
		"whole":      42.0,
	})
	var dst struct {
		SmallInt  int8
		SmallUint uint8
		Whole     int
	}
	if err := AssignEntity(entity, &dst); err != nil {
		t.Fatalf("AssignEntity safe numeric conversion: %v", err)
	}
	if dst.SmallInt != 127 || dst.SmallUint != 255 || dst.Whole != 42 {
		t.Fatalf("assigned numeric dst = %#v", dst)
	}
}

func TestAssignEntityPointerAndSQLScannerFields(t *testing.T) {
	type nullableDTO struct {
		Name       *string        `db:"name"`
		Nickname   *string        `db:"nickname"`
		Age        *int           `db:"age"`
		CreatedAt  *time.Time     `db:"created_at"`
		NullName   sql.NullString `db:"null_name"`
		NullAge    sql.NullInt64  `db:"null_age"`
		MissingPtr *int           `db:"missing_ptr"`
	}
	createdAt := time.Unix(1700000000, 0).UTC()
	entity := EntityFromMap("users", map[string]any{
		"name":        "alice",
		"nickname":    nil,
		"age":         int64(42),
		"created_at":  createdAt,
		"null_name":   "Alice",
		"null_age":    int64(43),
		"missing_ptr": nil,
	})
	var dst nullableDTO
	if err := AssignEntity(entity, &dst); err != nil {
		t.Fatalf("AssignEntity pointer/scanner fields: %v", err)
	}
	if dst.Name == nil || *dst.Name != "alice" {
		t.Fatalf("Name = %#v, want pointer to alice", dst.Name)
	}
	if dst.Nickname != nil {
		t.Fatalf("Nickname = %#v, want nil", dst.Nickname)
	}
	if dst.Age == nil || *dst.Age != 42 {
		t.Fatalf("Age = %#v, want pointer to 42", dst.Age)
	}
	if dst.CreatedAt == nil || !dst.CreatedAt.Equal(createdAt) {
		t.Fatalf("CreatedAt = %#v, want %v", dst.CreatedAt, createdAt)
	}
	if !dst.NullName.Valid || dst.NullName.String != "Alice" {
		t.Fatalf("NullName = %#v, want valid Alice", dst.NullName)
	}
	if !dst.NullAge.Valid || dst.NullAge.Int64 != 43 {
		t.Fatalf("NullAge = %#v, want valid 43", dst.NullAge)
	}
	if dst.MissingPtr != nil {
		t.Fatalf("MissingPtr = %#v, want nil", dst.MissingPtr)
	}
}

type failingScanner struct{}

func (f *failingScanner) Scan(any) error {
	return fmt.Errorf("scanner rejected value")
}

func TestAssignEntityScannerError(t *testing.T) {
	entity := EntityFromMap("users", map[string]any{"value": "bad"})
	var dst struct {
		Value failingScanner `db:"value"`
	}
	err := AssignEntity(entity, &dst)
	assertDBCode(t, err, knifer.ErrCodeInternal)
	if !strings.Contains(err.Error(), "scanner rejected value") {
		t.Fatalf("AssignEntity scanner error = %v, want scanner cause", err)
	}
}

func TestAssignEntityParsesTextScalars(t *testing.T) {
	entity := EntityFromMap("users", map[string]any{
		"active":     []byte("true"),
		"age":        "42",
		"ratio":      []byte("3.5"),
		"unsigned":   "7",
		"name_bytes": []byte("alice"),
	})
	var dst struct {
		Active    bool
		Age       int8
		Ratio     float32
		Unsigned  uint8
		NameBytes *string `db:"name_bytes"`
	}
	if err := AssignEntity(entity, &dst); err != nil {
		t.Fatalf("AssignEntity text scalars: %v", err)
	}
	if !dst.Active || dst.Age != 42 || dst.Ratio != 3.5 || dst.Unsigned != 7 || dst.NameBytes == nil || *dst.NameBytes != "alice" {
		t.Fatalf("assigned text scalar dst = %#v", dst)
	}
}

func TestAssignEntityRejectsInvalidTextScalars(t *testing.T) {
	tests := []struct {
		name   string
		entity Entity
		dst    any
	}{
		{name: "invalid bool", entity: EntityFromMap("users", map[string]any{"active": "not-bool"}), dst: &struct{ Active bool }{}},
		{name: "int overflow", entity: EntityFromMap("users", map[string]any{"age": "128"}), dst: &struct{ Age int8 }{}},
		{name: "invalid float", entity: EntityFromMap("users", map[string]any{"ratio": "nan?"}), dst: &struct{ Ratio float64 }{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := AssignEntity(tt.entity, tt.dst)
			assertDBCode(t, err, knifer.ErrCodeInternal)
		})
	}
}
