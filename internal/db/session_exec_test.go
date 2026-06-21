package db

import (
	"context"
	"database/sql/driver"
	"errors"
	"testing"

	knifer "github.com/imajinyun/go-knifer"
)

func TestSessionCRUDAndHelpers(t *testing.T) {
	ctx := context.Background()
	db := newFakeDB(&fakeBehavior{
		queryFunc: func(string) (driver.Rows, error) {
			return mkRows([]string{"id"}, []driver.Value{int64(1)}), nil
		},
	}, WithDialect(DialectPostgres))
	defer func() { _ = db.Close() }()

	entity := EntityFromMap("users", map[string]any{"name": "alice"})
	err := db.Tx(ctx, nil, func(s *Session) error {
		if _, e := s.Exec(ctx, "DELETE FROM users WHERE id = ?", 1); e != nil {
			return e
		}
		if _, e := s.ExecNamed(ctx, "UPDATE users SET name = :name WHERE id = :id", map[string]any{"name": "x", "id": 1}); e != nil {
			return e
		}
		if _, e := s.ExecBatch(ctx, "INSERT INTO users(name) VALUES (?)", []any{"a"}, []any{"b"}); e != nil {
			return e
		}
		if _, e := s.Query(ctx, "SELECT id FROM users"); e != nil {
			return e
		}
		if _, _, e := s.QueryOne(ctx, "SELECT id FROM users"); e != nil {
			return e
		}
		if _, _, e := s.QueryScalar(ctx, "SELECT id FROM users"); e != nil {
			return e
		}
		if _, e := s.Insert(ctx, entity); e != nil {
			return e
		}
		if _, e := s.Upsert(ctx, entity, []string{"name"}); e != nil {
			return e
		}
		if _, e := s.Update(ctx, entity, Eq("id", 1)); e != nil {
			return e
		}
		if _, e := s.UpdateAll(ctx, entity); e != nil {
			return e
		}
		if _, e := s.Delete(ctx, "users", Eq("id", 1)); e != nil {
			return e
		}
		if _, e := s.DeleteAll(ctx, "users"); e != nil {
			return e
		}
		if e := s.Savepoint(ctx, "sp1"); e != nil {
			return e
		}
		return s.RollbackTo(ctx, "sp1")
	})
	if err != nil {
		t.Fatalf("session CRUD: %v", err)
	}
}

func TestSessionGuardsAndInvalidIdentifiers(t *testing.T) {
	ctx := context.Background()
	db := newFakeDB(&fakeBehavior{}, WithDialect(DialectPostgres))
	defer func() { _ = db.Close() }()

	entity := EntityFromMap("users", map[string]any{"name": "x"})
	_ = db.Tx(ctx, nil, func(s *Session) error {
		assertDBCode(t, mustErr2(s.Update(ctx, entity)), knifer.ErrCodeInvalidInput)
		assertDBCode(t, mustErr2(s.Delete(ctx, "users")), knifer.ErrCodeInvalidInput)
		assertDBCode(t, mustErr2(s.Update(ctx, entity, Condition{})), knifer.ErrCodeInvalidInput)
		assertDBCode(t, mustErr2(s.Delete(ctx, "users", AndGroup())), knifer.ErrCodeInvalidInput)
		assertDBCode(t, s.Savepoint(ctx, "bad name"), knifer.ErrCodeInvalidInput)
		assertDBCode(t, s.RollbackTo(ctx, "bad name"), knifer.ErrCodeInvalidInput)
		return errors.New("stop") // force rollback, avoid commit noise
	})
}

func TestDBListTables(t *testing.T) {
	ctx := context.Background()
	db := newFakeDB(&fakeBehavior{queryFunc: func(string) (driver.Rows, error) {
		return mkRows([]string{"name"}, []driver.Value{[]byte("users")}, []driver.Value{[]byte("orders")}), nil
	}}, WithDialect(DialectSQLite))
	defer func() { _ = db.Close() }()

	tables, err := db.ListTables(ctx)
	if err != nil || len(tables) != 2 || tables[0] != "users" {
		t.Fatalf("ListTables = %#v err=%v", tables, err)
	}

	// Unsupported dialect.
	oracle := newFakeDB(&fakeBehavior{}, WithDialect(DialectOracle))
	defer func() { _ = oracle.Close() }()
	assertDBCode(t, mustErr2(oracle.ListTables(ctx)), knifer.ErrCodeUnsupported)
}

func TestDBListColumnsSQLite(t *testing.T) {
	ctx := context.Background()
	// PRAGMA table_info columns: cid, name, type, notnull, dflt_value, pk
	db := newFakeDB(&fakeBehavior{queryFunc: func(string) (driver.Rows, error) {
		return mkRows(
			[]string{"cid", "name", "type", "notnull", "dflt_value", "pk"},
			[]driver.Value{int64(0), []byte("id"), []byte("INTEGER"), int64(1), nil, int64(1)},
			[]driver.Value{int64(1), []byte("name"), []byte("TEXT"), int64(0), nil, int64(0)},
		), nil
	}}, WithDialect(DialectSQLite))
	defer func() { _ = db.Close() }()

	cols, err := db.ListColumns(ctx, "users")
	if err != nil || len(cols) != 2 {
		t.Fatalf("ListColumns = %#v err=%v", cols, err)
	}
	if !cols[0].PrimaryKey || cols[0].Nullable {
		t.Fatalf("id column = %#v", cols[0])
	}
	if cols[1].Name != "name" || !cols[1].Nullable {
		t.Fatalf("name column = %#v", cols[1])
	}

	names, err := db.ColumnNames(ctx, "users")
	if err != nil || len(names) != 2 || names[0] != "id" {
		t.Fatalf("ColumnNames = %#v err=%v", names, err)
	}
	meta, err := db.TableMeta(ctx, "users")
	if err != nil || len(meta.PrimaryKeys) != 1 || meta.PrimaryKeys[0] != "id" {
		t.Fatalf("TableMeta = %#v err=%v", meta, err)
	}
	pks, err := db.PrimaryKeys(ctx, "users")
	if err != nil || len(pks) != 1 {
		t.Fatalf("PrimaryKeys = %#v err=%v", pks, err)
	}
}

func TestDBListColumnsInformationSchema(t *testing.T) {
	ctx := context.Background()
	// column_name, data_type, is_nullable, column_default, column_key, extra
	db := newFakeDB(&fakeBehavior{queryFunc: func(string) (driver.Rows, error) {
		return mkRows(
			[]string{"column_name", "data_type", "is_nullable", "column_default", "column_key", "extra"},
			[]driver.Value{[]byte("id"), []byte("bigint"), []byte("NO"), nil, []byte("PRI"), []byte("auto_increment")},
		), nil
	}}, WithDialect(DialectMySQL))
	defer func() { _ = db.Close() }()

	cols, err := db.ListColumns(ctx, "users")
	if err != nil || len(cols) != 1 {
		t.Fatalf("ListColumns = %#v err=%v", cols, err)
	}
	if !cols[0].PrimaryKey || !cols[0].AutoIncrement || cols[0].Nullable {
		t.Fatalf("column = %#v", cols[0])
	}
}

func TestDBMetaErrors(t *testing.T) {
	ctx := context.Background()
	queryFail := newFakeDB(&fakeBehavior{queryErr: errors.New("query boom")}, WithDialect(DialectSQLite))
	defer func() { _ = queryFail.Close() }()
	if _, err := queryFail.ListTables(ctx); !errors.Is(err, knifer.ErrCodeInternal) {
		t.Fatalf("ListTables query error = %v", err)
	}
	if _, err := queryFail.ListColumns(ctx, "users"); !errors.Is(err, knifer.ErrCodeInternal) {
		t.Fatalf("ListColumns query error = %v", err)
	}

	// Invalid identifier rejected before query.
	ok := newFakeDB(&fakeBehavior{}, WithDialect(DialectSQLite))
	defer func() { _ = ok.Close() }()
	assertDBCode(t, mustErr2(ok.ListColumns(ctx, "bad name")), knifer.ErrCodeInvalidInput)
}
