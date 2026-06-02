package db

import (
	"reflect"
	"testing"
)

func TestSQLBuilderSelectWherePage(t *testing.T) {
	sqlText, args, err := NewBuilder(WithDialect(DialectPostgres), WithWrapper(WrapperForDialect(DialectPostgres))).
		Select("id", "name").
		From("users").
		Where(Eq("status", "active"), OrWith(In("role", "admin", "owner"))).
		OrderBy(Desc("id")).
		Page(NewPage(2, 10)).
		SQL()
	if err != nil {
		t.Fatalf("SQL() error = %v", err)
	}
	wantSQL := `SELECT "id", "name" FROM "users" WHERE "status" = $1 OR "role" IN ($2, $3) ORDER BY "id" DESC LIMIT 10 OFFSET 10`
	if sqlText != wantSQL {
		t.Fatalf("sql = %q, want %q", sqlText, wantSQL)
	}
	if !reflect.DeepEqual(args, []any{"active", "admin", "owner"}) {
		t.Fatalf("args = %#v", args)
	}
}

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
}

func TestUtilityTypes(t *testing.T) {
	page := NewPage(0, 0, Asc("name"))
	if page.Number != 1 || page.Size != 20 || page.Offset() != 0 || page.Limit() != 20 {
		t.Fatalf("page = %#v", page)
	}
	result := NewPageResult(page, 41, []Entity{NewEntity("users")})
	if result.TotalPage != 3 || !result.IsFirst() || result.IsLast() {
		t.Fatalf("result = %#v", result)
	}
	if got := WrapperForDialect(DialectMySQL).Wrap("users.name"); got != "`users`.`name`" {
		t.Fatalf("wrapped = %q", got)
	}
	if got := BuildLikeValue("go", "contains"); got != "%go%" {
		t.Fatalf("like = %q", got)
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
