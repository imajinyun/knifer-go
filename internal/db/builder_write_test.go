package db

import (
	"reflect"
	"testing"

	knifer "github.com/imajinyun/go-knifer"
)

func TestSQLBuilderInsertUpdateDelete(t *testing.T) {
	entity := NewEntity("users").Set("name", "alice").Set("age", 18)
	insertSQL, insertArgs, err := NewBuilder(WithDialect(DialectQuestion)).Insert(entity).SQL()
	if err != nil {
		t.Fatalf("Insert SQL() error = %v", err)
	}
	if insertSQL != "INSERT INTO users (age, name) VALUES (?, ?)" {
		t.Fatalf("insert sql = %q", insertSQL)
	}
	if !reflect.DeepEqual(insertArgs, []any{18, "alice"}) {
		t.Fatalf("insert args = %#v", insertArgs)
	}

	updateSQL, updateArgs, err := NewBuilder(WithDialect(DialectQuestion)).Update(entity).Where(Eq("id", 7)).SQL()
	if err != nil {
		t.Fatalf("Update SQL() error = %v", err)
	}
	if updateSQL != "UPDATE users SET age = ?, name = ? WHERE id = ?" {
		t.Fatalf("update sql = %q", updateSQL)
	}
	if !reflect.DeepEqual(updateArgs, []any{18, "alice", 7}) {
		t.Fatalf("update args = %#v", updateArgs)
	}

	deleteSQL, deleteArgs, err := NewBuilder(WithDialect(DialectQuestion)).Delete("users").Where(Eq("id", 7)).SQL()
	if err != nil {
		t.Fatalf("Delete SQL() error = %v", err)
	}
	if deleteSQL != "DELETE FROM users WHERE id = ?" || !reflect.DeepEqual(deleteArgs, []any{7}) {
		t.Fatalf("delete = %q %#v", deleteSQL, deleteArgs)
	}
}

func TestUpsertSQL(t *testing.T) {
	entity := NewEntity("users").Set("id", 1).Set("name", "alice")
	sqlText, args, err := buildUpsertSQL(DialectSQLite, WrapperForDialect(DialectSQLite), entity, []string{"id"})
	if err != nil {
		t.Fatalf("buildUpsertSQL() error = %v", err)
	}
	want := "INSERT INTO `users` (`id`, `name`) VALUES (?, ?) ON CONFLICT (`id`) DO UPDATE SET `name` = excluded.`name`"
	if sqlText != want {
		t.Fatalf("sql = %q, want %q", sqlText, want)
	}
	if !reflect.DeepEqual(args, []any{1, "alice"}) {
		t.Fatalf("args = %#v", args)
	}
}

func TestUpsertSQLReportsUnsupportedAndInvalidInput(t *testing.T) {
	entity := NewEntity("users").Set("id", 1).Set("name", "alice")
	_, _, err := buildUpsertSQL(DialectOracle, WrapperForDialect(DialectOracle), entity, []string{"id"})
	assertDBCode(t, err, knifer.ErrCodeUnsupported)

	_, _, err = buildUpsertSQL(DialectSQLite, WrapperForDialect(DialectSQLite), entity, nil)
	assertDBCode(t, err, knifer.ErrCodeInvalidInput)

	doNothingEntity := NewEntity("users").Set("id", 1)
	_, _, err = buildUpsertSQL(DialectSQLite, WrapperForDialect(DialectSQLite), doNothingEntity, []string{"id", "bad;drop"})
	assertDBCode(t, err, knifer.ErrCodeInvalidInput)
}
