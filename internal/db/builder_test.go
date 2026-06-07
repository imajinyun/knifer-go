package db

import (
	"errors"
	"reflect"
	"testing"

	knifer "github.com/imajinyun/go-knifer"
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
	if !IsSafeIdentifier("users.name") || !IsSafeIdentifier("users.*") || !IsSafeIdentifier("`users`.`name`") {
		t.Fatal("expected safe identifiers to be accepted")
	}
	if IsSafeIdentifier("users; drop table users") || IsSafeIdentifier("COUNT(*)") || IsSafeIdentifier("users name") {
		t.Fatal("expected unsafe identifiers to be rejected")
	}
}

func TestSQLBuilderRejectsUnsafeIdentifiers(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{name: "select field", err: sqlErr(Select("id; drop table users").From("users").SQL())},
		{name: "from table", err: sqlErr(Select("id").From("users; drop table users").SQL())},
		{name: "where field", err: sqlErr(Select("id").From("users").Where(Eq("id OR 1=1", 1)).SQL())},
		{name: "order field", err: sqlErr(Select("id").From("users").OrderBy(Asc("id desc; drop table users")).SQL())},
		{name: "insert table", err: sqlErr(Insert(NewEntity("users; drop table users").Set("name", "alice")).SQL())},
		{name: "insert field", err: sqlErr(Insert(NewEntity("users").Set("name; drop", "alice")).SQL())},
		{name: "update field", err: sqlErr(Update(NewEntity("users").Set("name = hacked", "alice")).Where(Eq("id", 1)).SQL())},
		{name: "delete table", err: sqlErr(Delete("users; drop table users").Where(Eq("id", 1)).SQL())},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertDBCode(t, tt.err, knifer.ErrCodeInvalidInput)
		})
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

func TestDBErrorContract(t *testing.T) {
	tests := []struct {
		name string
		err  error
		code knifer.ErrCode
	}{
		{
			name: "empty builder",
			err:  sqlErr(NewBuilder().SQL()),
			code: knifer.ErrCodeInvalidInput,
		},
		{
			name: "select without table",
			err:  sqlErr(Select("id").SQL()),
			code: knifer.ErrCodeInvalidInput,
		},
		{
			name: "insert without values",
			err:  sqlErr(Insert(NewEntity("users")).SQL()),
			code: knifer.ErrCodeInvalidInput,
		},
		{
			name: "update without values",
			err:  sqlErr(Update(NewEntity("users")).SQL()),
			code: knifer.ErrCodeInvalidInput,
		},
		{
			name: "delete without table",
			err:  sqlErr(Delete("").SQL()),
			code: knifer.ErrCodeInvalidInput,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertDBCode(t, tt.err, tt.code)
		})
	}
}

func TestNamedScanMetaAndUpsertErrorContract(t *testing.T) {
	_, err := ParseNamed("select * from users where id=:id", nil, DialectQuestion)
	assertDBCode(t, err, knifer.ErrCodeInvalidInput)

	_, err = ScanRows(nil)
	assertDBCode(t, err, knifer.ErrCodeInvalidInput)

	err = AssignEntity(NewEntity("users"), nil)
	assertDBCode(t, err, knifer.ErrCodeInvalidInput)

	_, err = listTablesSQL(DialectOracle)
	assertDBCode(t, err, knifer.ErrCodeUnsupported)

	_, _, _, err = listColumnsSQL(DialectOracle, "users")
	assertDBCode(t, err, knifer.ErrCodeUnsupported)

	entity := NewEntity("users").Set("id", 1).Set("name", "alice")
	_, _, err = buildUpsertSQL(DialectOracle, WrapperForDialect(DialectOracle), entity, []string{"id"})
	assertDBCode(t, err, knifer.ErrCodeUnsupported)

	_, _, err = buildUpsertSQL(DialectSQLite, WrapperForDialect(DialectSQLite), entity, nil)
	assertDBCode(t, err, knifer.ErrCodeInvalidInput)
}

func sqlErr(_ string, _ []any, err error) error { return err }

func assertDBCode(t *testing.T, err error, code knifer.ErrCode) {
	t.Helper()
	if err == nil {
		t.Fatalf("err = nil, want %s", code)
	}
	if !errors.Is(err, code) {
		t.Fatalf("errors.Is(%v, %s) = false", err, code)
	}
	got, ok := knifer.CodeOf(err)
	if !ok || got != code {
		t.Fatalf("CodeOf(%v) = %q, %v; want %q, true", err, got, ok, code)
	}
}
