package vdb

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"reflect"
	"testing"

	knifer "github.com/imajinyun/knifer-go"
)

func TestFacadeTopLevelBuildersAndConditions(t *testing.T) {
	rawSQL, rawArgs, err := Raw("SELECT ? AS id", 7).SQL()
	if err != nil || rawSQL != "SELECT ? AS id" || !reflect.DeepEqual(rawArgs, []any{7}) {
		t.Fatalf("Raw SQL=%q args=%#v err=%v", rawSQL, rawArgs, err)
	}

	insertSQL, insertArgs, err := Insert(EntityFromMap("users", map[string]any{"name": "alice", "age": 18})).SQL()
	if err != nil {
		t.Fatalf("Insert SQL: %v", err)
	}
	if insertSQL != "INSERT INTO users (age, name) VALUES (?, ?)" || !reflect.DeepEqual(insertArgs, []any{18, "alice"}) {
		t.Fatalf("Insert SQL=%q args=%#v", insertSQL, insertArgs)
	}

	updateSQL, updateArgs, err := Update(NewEntity("users").Set("name", "bob")).
		Where(AndGroup(Gt("id", 10), Lte("id", 20)), OrWith(IsNull("deleted_at"))).
		SQL()
	if err != nil {
		t.Fatalf("Update SQL: %v", err)
	}
	if updateSQL != "UPDATE users SET name = ? WHERE (id > ? AND id <= ?) OR deleted_at IS NULL" {
		t.Fatalf("Update SQL = %q", updateSQL)
	}
	if !reflect.DeepEqual(updateArgs, []any{"bob", 10, 20}) {
		t.Fatalf("Update args = %#v", updateArgs)
	}

	deleteSQL, deleteArgs, err := Delete("users").
		Where(OrGroup(Ne("status", "active"), Between("created_at", 1, 9), IsNotNull("blocked_at"))).
		SQL()
	if err != nil {
		t.Fatalf("Delete SQL: %v", err)
	}
	if deleteSQL != "DELETE FROM users WHERE (status <> ? AND created_at BETWEEN ? AND ? AND blocked_at IS NOT NULL)" {
		t.Fatalf("Delete SQL = %q", deleteSQL)
	}
	if !reflect.DeepEqual(deleteArgs, []any{"active", 1, 9}) {
		t.Fatalf("Delete args = %#v", deleteArgs)
	}

	conditionSQL, conditionArgs, err := BuildConditions(Like("name", BuildLikeValue("go", "prefix")), In("role", "admin", "owner"))
	if err != nil {
		t.Fatalf("BuildConditions: %v", err)
	}
	if conditionSQL != "name LIKE ? AND role IN (?, ?)" || !reflect.DeepEqual(conditionArgs, []any{"go%", "admin", "owner"}) {
		t.Fatalf("BuildConditions SQL=%q args=%#v", conditionSQL, conditionArgs)
	}

	conds := ConditionsFromEntity(NewEntity("users").Set("id", 1).Set("name", "alice"))
	if len(conds) != 2 || conds[0].Field != "id" || conds[1].Field != "name" {
		t.Fatalf("ConditionsFromEntity = %#v", conds)
	}
}

func TestFacadeNewQueryAndConditions(t *testing.T) {
	q := NewQuery("users")
	if len(q.Tables) != 1 || q.Tables[0] != "users" {
		t.Fatalf("NewQuery Tables = %#v", q.Tables)
	}

	// Gte and Lt condition builders
	rawSQL, rawArgs, err := Delete("users").Where(Gte("id", 5), Lt("age", 30)).SQL()
	if err != nil {
		t.Fatalf("Delete SQL: %v", err)
	}
	if rawSQL != "DELETE FROM users WHERE id >= ? AND age < ?" {
		t.Fatalf("Delete SQL = %q", rawSQL)
	}
	if !reflect.DeepEqual(rawArgs, []any{5, 30}) {
		t.Fatalf("Delete args = %#v", rawArgs)
	}
}

func TestFacadeEntityScanners(t *testing.T) {
	// All public functions should be callable - we just test they don't panic and return correctly
	e := NewEntity("users")
	e.Set("id", 1)
	e.Set("name", "alice")

	// AssignEntity assigns to a map
	var m map[string]any
	err := AssignEntity(e, &m)
	if err != nil {
		t.Fatalf("AssignEntity error = %v", err)
	}
	if m["id"] != 1 || m["name"] != "alice" {
		t.Fatalf("AssignEntity map = %#v", m)
	}
}

func TestFacadeAssignEntityRejectsNumericOverflow(t *testing.T) {
	entity := EntityFromMap("users", map[string]any{"age": int64(128)})
	var dst struct{ Age int8 }
	err := AssignEntity(entity, &dst)
	if !errors.Is(err, knifer.ErrCodeInternal) {
		t.Fatalf("AssignEntity overflow error = %v, want ErrCodeInternal", err)
	}
	if dst.Age != 0 {
		t.Fatalf("Age = %d, want unchanged zero value", dst.Age)
	}
}

func TestFacadeAssignEntityPointerAndNullFields(t *testing.T) {
	entity := EntityFromMap("users", map[string]any{
		"name":     "alice",
		"nickname": nil,
		"age":      int64(42),
		"label":    "A",
	})
	var dst struct {
		Name     *string        `db:"name"`
		Nickname *string        `db:"nickname"`
		Age      *int           `db:"age"`
		Label    sql.NullString `db:"label"`
	}
	if err := AssignEntity(entity, &dst); err != nil {
		t.Fatalf("AssignEntity pointer/null fields: %v", err)
	}
	if dst.Name == nil || *dst.Name != "alice" {
		t.Fatalf("Name = %#v, want pointer to alice", dst.Name)
	}
	if dst.Nickname != nil {
		t.Fatalf("Nickname = %#v, want nil", dst.Nickname)
	}
	if dst.Age == nil || *dst.Age != 42 {
		t.Fatalf("Age = %#v, want pointer to 42", dst.Age)
	}
	if !dst.Label.Valid || dst.Label.String != "A" {
		t.Fatalf("Label = %#v, want valid A", dst.Label)
	}
}

func TestFacadeAssignEntityParsesTextScalars(t *testing.T) {
	entity := EntityFromMap("users", map[string]any{
		"active": []byte("true"),
		"age":    "42",
		"ratio":  []byte("3.5"),
	})
	var dst struct {
		Active bool
		Age    int8
		Ratio  float32
	}
	if err := AssignEntity(entity, &dst); err != nil {
		t.Fatalf("AssignEntity text scalars: %v", err)
	}
	if !dst.Active || dst.Age != 42 || dst.Ratio != 3.5 {
		t.Fatalf("assigned text scalar dst = %#v", dst)
	}
}

func TestFacadeScanRowsAndScanOne(t *testing.T) {
	db, err := sql.Open("vdb_pool_test", "")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// ScanRows on empty result set
	rows, err := db.Query("SELECT id")
	if err != nil {
		t.Fatal(err)
	}
	got, err := ScanRows(rows)
	if err != nil {
		t.Fatalf("ScanRows error = %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("ScanRows = %#v, want empty", got)
	}

	// ScanOne on empty result set
	rows2, err := db.Query("SELECT id")
	if err != nil {
		t.Fatal(err)
	}
	entity, ok, err := ScanOne(rows2)
	if err != nil {
		t.Fatalf("ScanOne error = %v", err)
	}
	if ok {
		t.Fatal("ScanOne ok = true, want false for empty result")
	}
	if entity.Table != "" {
		t.Fatalf("ScanOne entity = %#v, want empty", entity)
	}
}

func TestFacadeScanOneOnlyReadsFirstRow(t *testing.T) {
	closed := false
	scriptedRowsForTest = &scriptedRows{
		cols:    []string{"id", "name"},
		data:    [][]driver.Value{{int64(1), []byte("alice")}},
		nextErr: errors.New("second row should not be read"),
		closed:  &closed,
	}
	defer func() { scriptedRowsForTest = nil }()

	db, err := sql.Open("vdb_scripted_rows_test", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	rows, err := db.Query("SELECT id, name FROM users")
	if err != nil {
		t.Fatal(err)
	}
	entity, ok, err := ScanOne(rows)
	if err != nil {
		t.Fatalf("ScanOne should ignore later iterator error after first row: %v", err)
	}
	if !ok || entity.Values["id"] != int64(1) || entity.Values["name"] != "alice" {
		t.Fatalf("ScanOne entity=%#v ok=%v", entity, ok)
	}
	if !closed {
		t.Fatal("ScanOne should close rows")
	}
}
