package db

import (
	"reflect"
	"strings"
	"testing"
	"unicode"
)

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
}

func TestIdentifierSafety(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want bool
	}{
		{name: "plain identifier", in: "users", want: true},
		{name: "qualified identifier", in: "users.name", want: true},
		{name: "qualified wildcard", in: "users.*", want: true},
		{name: "wrapped identifier", in: "`users`.`name`", want: true},
		{name: "quoted identifier", in: `"users"."name"`, want: true},
		{name: "bracket identifier", in: "[users].[name]", want: true},
		{name: "unicode identifier", in: "订单.用户", want: true},
		{name: "leading underscore", in: "_system.table_1", want: true},
		{name: "statement injection", in: "users; drop table users", want: false},
		{name: "function expression", in: "COUNT(*)", want: false},
		{name: "whitespace separated", in: "users name", want: false},
		{name: "comment injection", in: "users--", want: false},
		{name: "block comment injection", in: "users/*comment*/", want: false},
		{name: "line comment injection", in: "users--comment", want: false},
		{name: "slash comment", in: "users/name", want: false},
		{name: "backslash path", in: `users\name`, want: false},
		{name: "empty path part", in: "users..name", want: false},
		{name: "leading digit", in: "1users", want: false},
		{name: "quoted expression", in: "`COUNT(*)`", want: false},
		{name: "mismatched quote", in: "`users\"", want: false},
		{name: "wildcard in middle", in: "users.*.name", want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsSafeIdentifier(tt.in); got != tt.want {
				t.Fatalf("IsSafeIdentifier(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func FuzzIsSafeIdentifier(f *testing.F) {
	seeds := []string{
		"users",
		"users.name",
		"users.*",
		"`users`.`name`",
		`"users"."name"`,
		"[users].[name]",
		"订单.用户",
		"users; drop table users",
		"COUNT(*)",
		"users--comment",
		"users/*comment*/",
		"users name",
		"users..name",
		"1users",
		"",
	}
	for _, seed := range seeds {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, input string) {
		if !IsSafeIdentifier(input) {
			return
		}
		trimmed := strings.TrimSpace(input)
		if trimmed == "" || strings.Contains(trimmed, "..") {
			t.Fatalf("IsSafeIdentifier(%q)=true for empty or double-dot path", input)
		}
		dangerous := []string{";", "--", "/*", "*/", "(", ")", "/", `\`, " "}
		for _, token := range dangerous {
			if strings.Contains(trimmed, token) {
				t.Fatalf("IsSafeIdentifier(%q)=true with dangerous token %q", input, token)
			}
		}
		for _, part := range strings.Split(trimmed, ".") {
			part = strings.TrimSpace(part)
			if part == "" {
				t.Fatalf("IsSafeIdentifier(%q)=true with empty part", input)
			}
			if part == "*" {
				continue
			}
			if isWrappedIdentifierPart(part) {
				part = part[1 : len(part)-1]
			}
			if part == "" {
				t.Fatalf("IsSafeIdentifier(%q)=true with empty wrapped part", input)
			}
			for i, r := range part {
				if i == 0 {
					if r != '_' && !unicode.IsLetter(r) {
						t.Fatalf("IsSafeIdentifier(%q)=true with invalid leading rune %q", input, r)
					}
					continue
				}
				if r != '_' && !unicode.IsLetter(r) && !unicode.IsDigit(r) {
					t.Fatalf("IsSafeIdentifier(%q)=true with invalid rune %q", input, r)
				}
			}
		}
	})
}

func TestNormalizeDialectAndWrappers(t *testing.T) {
	tests := []struct {
		input string
		want  Dialect
	}{
		{input: " mysql ", want: DialectMySQL},
		{input: "mariadb", want: DialectMySQL},
		{input: "sqlite3", want: DialectSQLite},
		{input: "moderncsqlite", want: DialectSQLite},
		{input: "postgresql", want: DialectPostgres},
		{input: "pgx", want: DialectPostgres},
		{input: "mssql", want: DialectSQLServer},
		{input: "godror", want: DialectOracle},
		{input: "clickhouse", want: DialectClickHouse},
		{input: "unknown", want: DialectQuestion},
	}
	for _, tt := range tests {
		if got := NormalizeDialect(tt.input); got != tt.want {
			t.Fatalf("NormalizeDialect(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}

	if got := NewWrapper("`", "").Unwrap("`users`.`name`"); got != "users.name" {
		t.Fatalf("backtick unwrap = %q", got)
	}
	if got := WrapperForDialect(DialectSQLServer).Wrap("dbo.users"); got != "[dbo].[users]" {
		t.Fatalf("sqlserver wrap = %q", got)
	}
	if got := WrapperForDialect(DialectSQLServer).Unwrap("[dbo].[users]"); got != "dbo.users" {
		t.Fatalf("sqlserver unwrap = %q", got)
	}
	if got := (Wrapper{}).Unwrap(" users.name "); got != "users.name" {
		t.Fatalf("empty wrapper unwrap = %q", got)
	}
}

func TestEntityHelpers(t *testing.T) {
	values := map[string]any{"id": 1, "name": "alice", "skip": nil}
	entity := EntityFromMap("users", values)
	values["id"] = 2
	if entity.Table != "users" || entity.Values["id"] != 1 {
		t.Fatalf("EntityFromMap should copy values, got %#v", entity)
	}

	entity = entity.SetIfNotNil("email", "a@example.com").SetIfNotNil("nil", nil).Select("id", "email")
	if _, ok := entity.Values["nil"]; ok {
		t.Fatal("SetIfNotNil should skip nil values")
	}
	if !reflect.DeepEqual(entity.Fields, []string{"id", "email"}) {
		t.Fatalf("Select fields = %#v", entity.Fields)
	}
	filtered := EntityFromMap("users", map[string]any{"id": 1, "name": "alice", "email": "a@example.com"}).Filter("id", "email")
	if _, ok := filtered.Values["name"]; ok || filtered.Values["id"] != 1 || filtered.Values["email"] != "a@example.com" {
		t.Fatalf("Filter = %#v", filtered.Values)
	}
	removed := filtered.Remove("email", "missing")
	if _, ok := removed.Values["email"]; ok || removed.Values["id"] != 1 {
		t.Fatalf("Remove = %#v", removed.Values)
	}
}

func TestQueryHelpers(t *testing.T) {
	q := NewQuery("users", "profiles").
		Select("users.id", "profiles.name").
		Where(Eq("users.status", "active")).
		WithPage(NewPage(2, 5)).
		OrderBy(Desc("users.id"))

	if q.FirstTable() != "users" {
		t.Fatalf("FirstTable = %q", q.FirstTable())
	}
	if len(q.Fields) != 2 || len(q.Conditions) != 1 || q.Page == nil || len(q.Orders) != 1 {
		t.Fatalf("query = %#v", q)
	}
	if got := NewQuery().FirstTable(); got != "" {
		t.Fatalf("empty FirstTable = %q", got)
	}
}
