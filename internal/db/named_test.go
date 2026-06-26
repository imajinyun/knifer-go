package db

import (
	"reflect"
	"testing"

	knifer "github.com/imajinyun/knifer-go"
)

func TestParseNamed(t *testing.T) {
	named, err := ParseNamed("select * from users where id = :id and name = ':skip' and role = :role", map[string]any{"id": 1, "role": "admin"}, DialectPostgres)
	if err != nil {
		t.Fatalf("ParseNamed() error = %v", err)
	}
	if named.SQL != "select * from users where id = $1 and name = ':skip' and role = $2" {
		t.Fatalf("sql = %q", named.SQL)
	}
	if !reflect.DeepEqual(named.Params, []any{1, "admin"}) || !reflect.DeepEqual(named.Names, []string{"id", "role"}) {
		t.Fatalf("named = %#v", named)
	}

	named, err = ParseNamed("select * from users where id = :id::int and role = :role", map[string]any{"id": 7, "role": "owner"}, DialectPostgres)
	if err != nil {
		t.Fatalf("ParseNamed() with PostgreSQL cast error = %v", err)
	}
	if named.SQL != "select * from users where id = $1::int and role = $2" {
		t.Fatalf("sql with PostgreSQL cast = %q", named.SQL)
	}
	if !reflect.DeepEqual(named.Params, []any{7, "owner"}) || !reflect.DeepEqual(named.Names, []string{"id", "role"}) {
		t.Fatalf("named with PostgreSQL cast = %#v", named)
	}
}

func TestParseNamedReportsMissingParameter(t *testing.T) {
	_, err := ParseNamed("select * from users where id=:id", nil, DialectQuestion)
	assertDBCode(t, err, knifer.ErrCodeInvalidInput)
}
