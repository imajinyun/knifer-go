package db

import (
	"reflect"
	"testing"

	knifer "github.com/imajinyun/knifer-go"
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

func TestSQLBuilderPageOrders(t *testing.T) {
	sqlText, args, err := NewBuilder(WithDialect(DialectMySQL)).
		Select("id", "created_at").
		From("orders").
		Page(NewPage(2, 10, Desc("created_at"))).
		SQL()
	if err != nil {
		t.Fatalf("SQL() error = %v", err)
	}
	wantSQL := "SELECT `id`, `created_at` FROM `orders` ORDER BY `created_at` DESC LIMIT 10 OFFSET 10"
	if sqlText != wantSQL {
		t.Fatalf("sql = %q, want %q", sqlText, wantSQL)
	}
	if len(args) != 0 {
		t.Fatalf("args = %#v", args)
	}
}

func BenchmarkSQLBuilderPageOrders(b *testing.B) {
	page := NewPage(2, 10, Desc("created_at"))
	for b.Loop() {
		_, _, err := NewBuilder(WithDialect(DialectMySQL)).
			Select("id", "created_at").
			From("orders").
			Where(Eq("status", "paid")).
			Page(page).
			SQL()
		if err != nil {
			b.Fatal(err)
		}
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
		{name: "page order field", err: sqlErr(Select("id").From("users").Page(NewPage(1, 10, Asc("id desc; drop table users"))).SQL())},
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

func TestSQLBuilderRawAppendJoinGroupHaving(t *testing.T) {
	sqlText, args, err := Raw("SELECT * FROM users WHERE id = ?", 7).
		Append("AND status = ?", "active").
		SQL()
	if err != nil {
		t.Fatalf("Raw SQL error = %v", err)
	}
	if sqlText != "SELECT * FROM users WHERE id = ? AND status = ?" || !reflect.DeepEqual(args, []any{7, "active"}) {
		t.Fatalf("raw sql=%q args=%#v", sqlText, args)
	}

	sqlText, args, err = NewBuilder(WithDialect(DialectQuestion), WithWrapper(WrapperForDialect(DialectMySQL))).
		Select("users.id", "orders.total").
		From("users").
		Join("LEFT JOIN orders ON orders.user_id = users.id").
		Where(Eq("users.status", "active")).
		GroupBy("users.id", "orders.total").
		Having("COUNT(*) > 1").
		SQL()
	if err != nil {
		t.Fatalf("join/group SQL error = %v", err)
	}
	wantSQL := "SELECT `users`.`id`, `orders`.`total` FROM `users` LEFT JOIN orders ON orders.user_id = users.id WHERE `users`.`status` = ? GROUP BY `users`.`id`, `orders`.`total` HAVING COUNT(*) > 1"
	if sqlText != wantSQL || !reflect.DeepEqual(args, []any{"active"}) {
		t.Fatalf("join/group sql=%q args=%#v", sqlText, args)
	}
}

func TestSQLBuilderQueryAndOrderHelpers(t *testing.T) {
	q := NewQuery("users").Select("id", "name").Where(Eq("status", "active")).WithPage(NewPage(2, 3)).OrderBy(Asc(""), Desc("id"))
	sqlText, args, err := NewBuilder(WithDialect(DialectSQLServer), WithWrapper(WrapperForDialect(DialectSQLServer))).Query(q).SQL()
	if err != nil {
		t.Fatalf("Query SQL error = %v", err)
	}
	want := "SELECT [id], [name] FROM [users] WHERE [status] = @p1 ORDER BY [id] DESC OFFSET 3 ROWS FETCH NEXT 3 ROWS ONLY"
	if sqlText != want || !reflect.DeepEqual(args, []any{"active"}) {
		t.Fatalf("query sql=%q args=%#v", sqlText, args)
	}

	if got := RemoveOuterOrderBy("SELECT * FROM (SELECT * FROM users ORDER BY id) t ORDER BY name"); got != "SELECT * FROM (SELECT * FROM users ORDER BY id) t" {
		t.Fatalf("RemoveOuterOrderBy nested = %q", got)
	}
	if got := RemoveOuterOrderBy("SELECT 'ORDER BY literal' AS x"); got != "SELECT 'ORDER BY literal' AS x" {
		t.Fatalf("RemoveOuterOrderBy literal = %q", got)
	}
	if !IsInClause("WHERE id IN (?, ?)") || IsInClause("WHERE name = ?") {
		t.Fatal("IsInClause returned unexpected result")
	}
}
