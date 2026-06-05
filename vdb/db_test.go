package vdb

import (
	"errors"
	"testing"

	knifer "github.com/imajinyun/go-knifer"
)

func TestFacadeBuilder(t *testing.T) {
	sqlText, args, err := NewBuilder(WithDialect(DialectPostgres), WithWrapper(WrapperForDialect(DialectPostgres))).
		Select("id").
		From("users").
		Where(Eq("name", "alice")).
		SQL()
	if err != nil {
		t.Fatalf("SQL() error = %v", err)
	}
	if sqlText != `SELECT "id" FROM "users" WHERE "name" = $1` {
		t.Fatalf("sql = %q", sqlText)
	}
	if len(args) != 1 || args[0] != "alice" {
		t.Fatalf("args = %#v", args)
	}
}

func TestFacadeBuilderOptionsWrapperPrecedence(t *testing.T) {
	sqlText, _, err := NewBuilder(WithDialect(DialectMySQL)).Select("id").From("users").SQL()
	if err != nil {
		t.Fatalf("SQL() with dialect option error = %v", err)
	}
	if sqlText != "SELECT `id` FROM `users`" {
		t.Fatalf("SQL() with dialect default wrapper = %q", sqlText)
	}

	sqlText, _, err = NewBuilder(WithDialect(DialectMySQL), WithWrapper(NewWrapper("\"", "\""))).Select("id").From("users").SQL()
	if err != nil {
		t.Fatalf("SQL() with wrapper option error = %v", err)
	}
	if sqlText != `SELECT "id" FROM "users"` {
		t.Fatalf("SQL() with explicit wrapper = %q", sqlText)
	}
}

func TestFacadeNamedSQL(t *testing.T) {
	named, err := ParseNamed("select * from users where id=:id", map[string]any{"id": 1}, DialectQuestion)
	if err != nil {
		t.Fatalf("ParseNamed() error = %v", err)
	}
	if named.SQL != "select * from users where id=?" || named.Params[0] != 1 {
		t.Fatalf("named = %#v", named)
	}
}

func TestFacadeDBErrorContract(t *testing.T) {
	_, _, err := NewBuilder().SQL()
	if err == nil {
		t.Fatal("SQL() error = nil, want invalid input error")
	}
	if !errors.Is(err, knifer.ErrCodeInvalidInput) {
		t.Fatalf("errors.Is(err, ErrCodeInvalidInput) = false: %v", err)
	}
	code, ok := knifer.CodeOf(err)
	if !ok || code != knifer.ErrCodeInvalidInput {
		t.Fatalf("CodeOf(err) = %q, %v; want invalid input", code, ok)
	}
	var dbErr *DBError
	if !errors.As(err, &dbErr) {
		t.Fatalf("errors.As(err, *DBError) = false: %v", err)
	}
}
