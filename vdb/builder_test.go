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

func TestFacadePageOrdersAndIdentifierGuard(t *testing.T) {
	sqlText, args, err := NewBuilder(WithDialect(DialectMySQL)).
		Select("id", "created_at").
		From("orders").
		Page(NewPage(2, 10, Desc("created_at"))).
		SQL()
	if err != nil {
		t.Fatalf("SQL() with page orders error = %v", err)
	}
	if sqlText != "SELECT `id`, `created_at` FROM `orders` ORDER BY `created_at` DESC LIMIT 10 OFFSET 10" {
		t.Fatalf("SQL() with page orders = %q", sqlText)
	}
	if len(args) != 0 {
		t.Fatalf("args = %#v", args)
	}

	if !IsSafeIdentifier("orders.created_at") {
		t.Fatal("IsSafeIdentifier rejected a dotted identifier")
	}
	if IsSafeIdentifier("orders; drop table orders") {
		t.Fatal("IsSafeIdentifier accepted unsafe SQL")
	}
}

func BenchmarkFacadePageOrders(b *testing.B) {
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
