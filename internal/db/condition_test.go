package db

import (
	"reflect"
	"testing"

	knifer "github.com/imajinyun/knifer-go"
)

func TestSQLBuilderRejectsUnsafeConditionOperator(t *testing.T) {
	_, _, err := NewBuilder(WithDialect(DialectQuestion)).
		Select("id").
		From("users").
		Where(Condition{Field: "name", Op: "= ? OR 1=1 --", Value: "alice"}).
		SQL()
	assertDBCode(t, err, knifer.ErrCodeInvalidInput)

	sqlText, args, err := NewBuilder(WithDialect(DialectQuestion)).
		Select("id").
		From("users").
		Where(Condition{Field: "name", Op: "NOT LIKE", Value: "%bot%"}).
		SQL()
	if err != nil {
		t.Fatalf("SQL() NOT LIKE error = %v", err)
	}
	if sqlText != "SELECT id FROM users WHERE name NOT LIKE ?" || !reflect.DeepEqual(args, []any{"%bot%"}) {
		t.Fatalf("NOT LIKE sql=%q args=%#v", sqlText, args)
	}
}

func TestBuildLikeValue(t *testing.T) {
	if got := BuildLikeValue("go", "contains"); got != "%go%" {
		t.Fatalf("contains like = %q", got)
	}
}

func TestConditionConstructorsAndBuildConditions(t *testing.T) {
	part, args, err := BuildConditions(
		Ne("status", "deleted"),
		Gt("age", 18),
		Gte("score", 90),
		Lt("rank", 10),
		Lte("quota", 100),
		Like("name", "A%"),
		Between("created_at", 1, 9),
		IsNull("archived_at"),
		IsNotNull("email"),
	)
	if err != nil {
		t.Fatalf("BuildConditions error = %v", err)
	}
	wantPart := "status <> ? AND age > ? AND score >= ? AND rank < ? AND quota <= ? AND name LIKE ? AND created_at BETWEEN ? AND ? AND archived_at IS NULL AND email IS NOT NULL"
	if part != wantPart {
		t.Fatalf("part = %q, want %q", part, wantPart)
	}
	if !reflect.DeepEqual(args, []any{"deleted", 18, 90, 10, 100, "A%", 1, 9}) {
		t.Fatalf("args = %#v", args)
	}
}

func TestConditionGroupsEntityAndInBoundaries(t *testing.T) {
	conds := ConditionsFromEntity(EntityFromMap("users", map[string]any{"name": "alice", "id": 7}))
	if len(conds) != 2 || conds[0].Field != "id" || conds[1].Field != "name" {
		t.Fatalf("ConditionsFromEntity sorted conds = %#v", conds)
	}

	part, args, err := BuildConditions(
		AndGroup(Eq("tenant_id", 1), OrWith(Eq("tenant_id", 2))),
		OrGroup(Condition{Field: "empty", Op: "IN", Value: []int{}}, Condition{Field: "non_empty", Op: "NOT IN", Value: []string{"x", "y"}}),
	)
	if err != nil {
		t.Fatalf("BuildConditions groups error = %v", err)
	}
	want := "(tenant_id = ? OR tenant_id = ?) OR (1 = 0 AND non_empty NOT IN (?, ?))"
	if part != want {
		t.Fatalf("group part = %q, want %q", part, want)
	}
	if !reflect.DeepEqual(args, []any{1, 2, "x", "y"}) {
		t.Fatalf("group args = %#v", args)
	}

	part, args, err = BuildConditions(Condition{Field: "ignored"})
	if err != nil || part != "ignored = ?" || !reflect.DeepEqual(args, []any{nil}) {
		t.Fatalf("default op part=%q args=%#v err=%v", part, args, err)
	}
}

func TestBuildLikeValueModes(t *testing.T) {
	cases := map[string]string{
		"prefix": "go%",
		"start":  "go%",
		"suffix": "%go",
		"end":    "%go",
		"exact":  "go",
		"none":   "go",
		"":       "%go%",
	}
	for mode, want := range cases {
		if got := BuildLikeValue("go", mode); got != want {
			t.Fatalf("BuildLikeValue mode %q = %q, want %q", mode, got, want)
		}
	}
}
