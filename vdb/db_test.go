package vdb

import "testing"

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

func TestFacadeNamedSQL(t *testing.T) {
	named, err := ParseNamed("select * from users where id=:id", map[string]any{"id": 1}, DialectQuestion)
	if err != nil {
		t.Fatalf("ParseNamed() error = %v", err)
	}
	if named.SQL != "select * from users where id=?" || named.Params[0] != 1 {
		t.Fatalf("named = %#v", named)
	}
}
