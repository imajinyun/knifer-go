package db

import (
	"database/sql"
	"errors"
	"reflect"
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
