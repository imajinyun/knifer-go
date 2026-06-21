package db

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"testing"
	"time"

	knifer "github.com/imajinyun/go-knifer"
)

func TestDBExecAndQueryHappyPath(t *testing.T) {
	ctx := context.Background()
	b := &fakeBehavior{
		queryFunc: func(string) (driver.Rows, error) {
			return mkRows([]string{"id", "name"},
				[]driver.Value{int64(1), []byte("alice")},
				[]driver.Value{int64(2), []byte("bob")},
			), nil
		},
	}
	db := newFakeDB(b)
	defer func() { _ = db.Close() }()

	if _, err := db.Exec(ctx, "DELETE FROM users"); err != nil {
		t.Fatalf("Exec: %v", err)
	}
	rows, err := db.Query(ctx, "SELECT id, name FROM users")
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if len(rows) != 2 || rows[0].Values["name"] != "alice" {
		t.Fatalf("rows = %#v", rows)
	}

	one, ok, err := db.QueryOne(ctx, "SELECT id, name FROM users")
	if err != nil || !ok || one.Values["id"] != int64(1) {
		t.Fatalf("QueryOne ok=%v one=%#v err=%v", ok, one, err)
	}
}

func TestDBExecAndQueryErrors(t *testing.T) {
	ctx := context.Background()
	wantExec := errors.New("exec boom")
	wantQuery := errors.New("query boom")
	db := newFakeDB(&fakeBehavior{execErr: wantExec, queryErr: wantQuery})
	defer func() { _ = db.Close() }()

	if _, err := db.Exec(ctx, "DELETE FROM users"); !errors.Is(err, wantExec) {
		t.Fatalf("Exec err = %v", err)
	}
	if _, err := db.Query(ctx, "SELECT 1"); !errors.Is(err, wantQuery) {
		t.Fatalf("Query err = %v", err)
	}
	if _, _, err := db.QueryOne(ctx, "SELECT 1"); !errors.Is(err, wantQuery) {
		t.Fatalf("QueryOne err = %v", err)
	}
	if _, _, err := db.QueryScalar(ctx, "SELECT 1"); !errors.Is(err, wantQuery) {
		t.Fatalf("QueryScalar err = %v", err)
	}
}

func TestDBQueryScalar(t *testing.T) {
	ctx := context.Background()
	rowsFn := func(string) (driver.Rows, error) {
		return mkRows([]string{"c"}, []driver.Value{int64(42)}), nil
	}
	db := newFakeDB(&fakeBehavior{queryFunc: rowsFn})
	defer func() { _ = db.Close() }()

	val, ok, err := db.QueryScalar(ctx, "SELECT COUNT(*) FROM users")
	if err != nil || !ok || val != int64(42) {
		t.Fatalf("QueryScalar val=%v ok=%v err=%v", val, ok, err)
	}

	// Empty result set: rows.Err() is nil so no error is wrapped, and ok=false.
	empty := newFakeDB(&fakeBehavior{queryFunc: func(string) (driver.Rows, error) {
		return mkRows([]string{"c"}), nil
	}})
	defer func() { _ = empty.Close() }()
	if v, ok, err := empty.QueryScalar(ctx, "SELECT COUNT(*) FROM users"); v != nil || ok || err != nil {
		t.Fatalf("QueryScalar empty = (%v, %v, %v)", v, ok, err)
	}
}

func TestDBExecNamedAndQueryNamed(t *testing.T) {
	ctx := context.Background()
	db := newFakeDB(&fakeBehavior{queryFunc: func(string) (driver.Rows, error) {
		return mkRows([]string{"id"}, []driver.Value{int64(1)}), nil
	}}, WithDialect(DialectPostgres))
	defer func() { _ = db.Close() }()

	if _, err := db.ExecNamed(ctx, "UPDATE users SET name = :name WHERE id = :id", map[string]any{"name": "x", "id": 1}); err != nil {
		t.Fatalf("ExecNamed: %v", err)
	}
	if _, err := db.QueryNamed(ctx, "SELECT id FROM users WHERE id = :id", map[string]any{"id": 1}); err != nil {
		t.Fatalf("QueryNamed: %v", err)
	}

	// Missing named parameter surfaces invalid input from ParseNamed.
	_, err := db.ExecNamed(ctx, "SELECT :missing", map[string]any{})
	assertDBCode(t, err, knifer.ErrCodeInvalidInput)
	_, err = db.QueryNamed(ctx, "SELECT :missing", map[string]any{})
	assertDBCode(t, err, knifer.ErrCodeInvalidInput)
}

func TestDBExecBatch(t *testing.T) {
	ctx := context.Background()
	db := newFakeDB(&fakeBehavior{})
	defer func() { _ = db.Close() }()
	results, err := db.ExecBatch(ctx, "INSERT INTO users(name) VALUES (?)", []any{"a"}, []any{"b"})
	if err != nil || len(results) != 2 {
		t.Fatalf("ExecBatch results=%d err=%v", len(results), err)
	}

	failing := newFakeDB(&fakeBehavior{execErr: errors.New("batch boom")})
	defer func() { _ = failing.Close() }()
	if _, err := failing.ExecBatch(ctx, "INSERT", []any{"a"}); err == nil {
		t.Fatal("ExecBatch should propagate exec error")
	}
}

func TestDBInsertUpdateDelete(t *testing.T) {
	ctx := context.Background()
	db := newFakeDB(&fakeBehavior{execResult: fakeResult{lastID: 99}})
	defer func() { _ = db.Close() }()

	entity := EntityFromMap("users", map[string]any{"name": "alice"})
	if _, err := db.Insert(ctx, entity); err != nil {
		t.Fatalf("Insert: %v", err)
	}
	id, err := db.InsertGetID(ctx, entity)
	if err != nil || id != 99 {
		t.Fatalf("InsertGetID id=%d err=%v", id, err)
	}

	if _, err := db.Update(ctx, entity, Eq("id", 1)); err != nil {
		t.Fatalf("Update: %v", err)
	}
	if _, err := db.UpdateAll(ctx, entity); err != nil {
		t.Fatalf("UpdateAll: %v", err)
	}
	if _, err := db.Delete(ctx, "users", Eq("id", 1)); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := db.DeleteAll(ctx, "users"); err != nil {
		t.Fatalf("DeleteAll: %v", err)
	}
	if _, err := db.DeleteEntity(ctx, EntityFromMap("users", map[string]any{"id": 1})); err != nil {
		t.Fatalf("DeleteEntity: %v", err)
	}

	// Guard rails: condition-less Update/Delete are rejected.
	assertDBCode(t, mustErr2(db.Update(ctx, entity)), knifer.ErrCodeInvalidInput)
	assertDBCode(t, mustErr2(db.Delete(ctx, "users")), knifer.ErrCodeInvalidInput)
	assertDBCode(t, mustErr2(db.Update(ctx, entity, Condition{})), knifer.ErrCodeInvalidInput)
	assertDBCode(t, mustErr2(db.Delete(ctx, "users", Condition{Field: " "})), knifer.ErrCodeInvalidInput)
}

func TestDBInsertGetIDErrors(t *testing.T) {
	ctx := context.Background()
	entity := EntityFromMap("users", map[string]any{"name": "x"})

	failInsert := newFakeDB(&fakeBehavior{execErr: errors.New("insert boom")})
	defer func() { _ = failInsert.Close() }()
	if _, err := failInsert.InsertGetID(ctx, entity); err == nil {
		t.Fatal("InsertGetID should propagate insert error")
	}

	failID := newFakeDB(&fakeBehavior{execResult: fakeResult{lastIDErr: errors.New("no id")}})
	defer func() { _ = failID.Close() }()
	if _, err := failID.InsertGetID(ctx, entity); !errors.Is(err, knifer.ErrCodeInternal) {
		t.Fatalf("InsertGetID id error = %v", err)
	}
}

func TestDBFindGetCountPage(t *testing.T) {
	ctx := context.Background()
	rowsFn := func(query string) (driver.Rows, error) {
		if containsCount(query) {
			return mkRows([]string{"c"}, []driver.Value{int64(3)}), nil
		}
		return mkRows([]string{"id"}, []driver.Value{int64(1)}, []driver.Value{int64(2)}), nil
	}
	db := newFakeDB(&fakeBehavior{queryFunc: rowsFn}, WithDialect(DialectPostgres))
	defer func() { _ = db.Close() }()

	if _, err := db.FindAll(ctx, "users"); err != nil {
		t.Fatalf("FindAll: %v", err)
	}
	if _, err := db.FindBy(ctx, "users", "status", "active"); err != nil {
		t.Fatalf("FindBy: %v", err)
	}
	if _, err := db.FindLike(ctx, "users", "name", "ali", "prefix"); err != nil {
		t.Fatalf("FindLike: %v", err)
	}
	if _, _, err := db.Get(ctx, "users", "id", 1); err != nil {
		t.Fatalf("Get: %v", err)
	}
	n, err := db.Count(ctx, "users", Eq("status", "active"))
	if err != nil || n != 3 {
		t.Fatalf("Count n=%d err=%v", n, err)
	}
	page, err := db.Page(ctx, NewQuery("users"), NewPage(1, 10))
	if err != nil || page.Total != 3 || len(page.Items) != 2 {
		t.Fatalf("Page = %#v err=%v", page, err)
	}
}

func TestDBTxCommitAndRollback(t *testing.T) {
	ctx := context.Background()
	db := newFakeDB(&fakeBehavior{})
	defer func() { _ = db.Close() }()

	if err := db.Tx(ctx, nil, func(s *Session) error {
		_, err := s.Exec(ctx, "INSERT INTO users(name) VALUES (?)", "alice")
		return err
	}); err != nil {
		t.Fatalf("Tx commit: %v", err)
	}

	// fn returns error -> rollback and error propagation.
	wantErr := errors.New("fn boom")
	if err := db.Tx(ctx, nil, func(*Session) error { return wantErr }); !errors.Is(err, wantErr) {
		t.Fatalf("Tx rollback err = %v", err)
	}
}

func TestDBTxBeginAndCommitErrors(t *testing.T) {
	ctx := context.Background()
	beginFail := newFakeDB(&fakeBehavior{beginErr: errors.New("begin boom")})
	defer func() { _ = beginFail.Close() }()
	if err := beginFail.Tx(ctx, nil, func(*Session) error { return nil }); !errors.Is(err, knifer.ErrCodeInternal) {
		t.Fatalf("Tx begin error = %v", err)
	}

	commitFail := newFakeDB(&fakeBehavior{commitErr: errors.New("commit boom")})
	defer func() { _ = commitFail.Close() }()
	if err := commitFail.Tx(ctx, nil, func(*Session) error { return nil }); !errors.Is(err, knifer.ErrCodeInternal) {
		t.Fatalf("Tx commit error = %v", err)
	}
}

func TestOpenUseAccessorsAndPing(t *testing.T) {
	ctx := context.Background()

	// Open uses the injected SQLOpenFunc and applies pool options.
	opened := false
	db, err := Open("sqlite3", "file::memory:",
		WithSQLOpenFunc(func(string, string) (*sql.DB, error) {
			opened = true
			return sql.OpenDB(&fakeConnector{b: &fakeBehavior{}}), nil
		}),
		WithMaxOpenConns(3),
		WithMaxIdleConns(1),
		WithConnMaxLifetime(time.Minute),
		WithConnMaxIdleTime(time.Second),
	)
	if err != nil || !opened {
		t.Fatalf("Open opened=%v err=%v", opened, err)
	}
	defer func() { _ = db.Close() }()

	if db.SQLDB() == nil {
		t.Fatal("SQLDB() = nil")
	}
	if db.Dialect() != DialectSQLite {
		t.Fatalf("Dialect() = %v", db.Dialect())
	}
	if db.Wrapper() != WrapperForDialect(DialectSQLite) {
		t.Fatalf("Wrapper() = %#v", db.Wrapper())
	}
	if err := db.Ping(ctx); err != nil {
		t.Fatalf("Ping: %v", err)
	}
}

func TestOpenError(t *testing.T) {
	want := errors.New("open boom")
	_, err := Open("bad", "dsn", WithSQLOpenFunc(func(string, string) (*sql.DB, error) {
		return nil, want
	}))
	if !errors.Is(err, knifer.ErrCodeInternal) || !errors.Is(err, want) {
		t.Fatalf("Open error = %v", err)
	}
}

func TestDBUpsert(t *testing.T) {
	ctx := context.Background()
	db := newFakeDB(&fakeBehavior{}, WithDialect(DialectPostgres))
	defer func() { _ = db.Close() }()
	entity := EntityFromMap("users", map[string]any{"id": 1, "name": "alice"})
	if _, err := db.Upsert(ctx, entity, []string{"id"}); err != nil {
		t.Fatalf("Upsert: %v", err)
	}

	// Exec error propagation.
	failing := newFakeDB(&fakeBehavior{execErr: errors.New("upsert boom")}, WithDialect(DialectPostgres))
	defer func() { _ = failing.Close() }()
	if _, err := failing.Upsert(ctx, entity, []string{"id"}); !errors.Is(err, knifer.ErrCodeInternal) {
		t.Fatalf("Upsert error = %v", err)
	}

	// Invalid upsert (missing conflict fields for Postgres) surfaces invalid input.
	assertDBCode(t, mustErr2(db.Upsert(ctx, entity, nil)), knifer.ErrCodeInvalidInput)
}

func TestDBErrorErrorAndIs(t *testing.T) {
	// Error() without cause returns Msg only.
	e := dbErrorf(knifer.ErrCodeInvalidInput, "bad %s", "input")
	if e.Error() != "bad input" {
		t.Fatalf("Error() = %q", e.Error())
	}
	// Is matches by code and against another *DBError, but not unrelated codes/targets.
	if !e.Is(knifer.ErrCodeInvalidInput) {
		t.Fatal("Is(code) should match")
	}
	if e.Is(knifer.ErrCodeInternal) {
		t.Fatal("Is(other code) should not match")
	}
	if !e.Is(&DBError{Code: knifer.ErrCodeInvalidInput}) {
		t.Fatal("Is(*DBError same code) should match")
	}
	if e.Is(errors.New("plain")) {
		t.Fatal("Is(plain error) should not match")
	}
	if e.Is(nil) {
		t.Fatal("Is(nil) should not match")
	}
	if (*DBError)(nil).Is(knifer.ErrCodeInvalidInput) {
		t.Fatal("nil receiver Is should be false")
	}
}

func mustErr2[T any](_ T, err error) error { return err }

func containsCount(q string) bool {
	for i := 0; i+5 <= len(q); i++ {
		if q[i:i+5] == "COUNT" {
			return true
		}
	}
	return false
}
