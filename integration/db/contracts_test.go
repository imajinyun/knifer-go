package db_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/imajinyun/knifer-go/vdb"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "modernc.org/sqlite"
)

type engine struct {
	name, driver, dsn, blob, clock string
	dialect                        vdb.Dialect
}

func TestDatabaseContracts(t *testing.T) {
	names := os.Getenv("KNIFER_DB_ENGINES")
	if names == "" {
		names = "sqlite"
	}
	seen := map[string]bool{}
	for _, name := range strings.Split(names, ",") {
		name = strings.TrimSpace(name)
		if seen[name] {
			t.Fatalf("duplicate engine %s", name)
		}
		seen[name] = true
		t.Run(name, func(t *testing.T) {
			e := engine{name: name, blob: "BLOB", clock: "DATETIME"}
			switch name {
			case "sqlite":
				e.driver = "sqlite"
				e.dialect = vdb.DialectSQLite
				e.dsn = filepath.Join(t.TempDir(), "contract.sqlite")
			case "postgres":
				e.driver = "pgx"
				e.dialect = vdb.DialectPostgres
				e.dsn = os.Getenv("KNIFER_TEST_POSTGRES_DSN")
				e.blob = "BYTEA"
				e.clock = "TIMESTAMP"
			case "mysql":
				e.driver = "mysql"
				e.dialect = vdb.DialectMySQL
				e.dsn = os.Getenv("KNIFER_TEST_MYSQL_DSN")
				e.clock = "DATETIME(6)"
			default:
				t.Fatalf("unknown engine %q", name)
			}
			if e.dsn == "" {
				t.Fatalf("requested engine %s requires its dedicated test DSN", name)
			}
			db, err := vdb.Open(e.driver, e.dsn, vdb.WithDialect(e.dialect), vdb.WithMaxOpenConns(1), vdb.WithMaxIdleConns(1))
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := db.Close(); err != nil {
					t.Error(err)
				}
			})
			ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
			defer cancel()
			if err := db.Ping(ctx); err != nil {
				t.Fatalf("test database unavailable: %v", err)
			}
			versionQuery := "SELECT version()"
			if name == "sqlite" {
				versionQuery = "SELECT sqlite_version()"
			}
			version, _, err := db.QueryScalar(ctx, versionQuery)
			if err != nil {
				t.Fatal(err)
			}
			t.Logf("engine=%s version=%v", name, version)
			for _, contract := range []struct {
				name string
				run  func(*testing.T, context.Context, *vdb.DB, string, engine)
			}{
				{"binding_and_scanning", bindingAndScanning}, {"upsert", upsert}, {"pagination", pagination}, {"transactions", transactions}, {"partial_batch", partialBatch}, {"row_lifecycle", rowLifecycle}, {"cancellation", cancellation},
			} {
				t.Run(contract.name, func(t *testing.T) { table := newTable(t, ctx, db, e); contract.run(t, ctx, db, table, e) })
			}
			if name == "postgres" {
				t.Run("deferred_commit_error", func(t *testing.T) { deferredCommitError(t, ctx, db, e) })
			}
		})
	}
}

func newTable(t *testing.T, ctx context.Context, db *vdb.DB, e engine) string {
	t.Helper()
	name := fmt.Sprintf("knifer_it_%x", time.Now().UnixNano())
	floating := "DOUBLE PRECISION"
	if e.name == "sqlite" {
		floating = "REAL"
	}
	query := fmt.Sprintf("CREATE TABLE %s (id BIGINT PRIMARY KEY, label VARCHAR(128) NOT NULL UNIQUE, n BIGINT NOT NULL, fraction %s, payload %s, optional TEXT, occurred_at %s)", name, floating, e.blob, e.clock)
	if _, err := db.Exec(ctx, query); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		cleanup, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if _, err := db.Exec(cleanup, "DROP TABLE "+name); err != nil {
			t.Error(err)
		}
	})
	return name
}

func insert(t *testing.T, ctx context.Context, db *vdb.DB, table string, id int64, label string) {
	t.Helper()
	if _, err := db.Insert(ctx, vdb.NewEntity(table).Set("id", id).Set("label", label).Set("n", id)); err != nil {
		t.Fatal(err)
	}
}

func count(t *testing.T, ctx context.Context, db *vdb.DB, table string, want int64) {
	t.Helper()
	n, err := db.Count(ctx, table)
	if err != nil || n != want {
		t.Fatalf("count=%d err=%v want=%d", n, err, want)
	}
}

func bindingAndScanning(t *testing.T, ctx context.Context, db *vdb.DB, table string, e engine) {
	label := "x'); DROP TABLE ignored; --"
	instant := time.Date(2024, 6, 1, 2, 3, 4, 0, time.UTC)
	payload := []byte{0, 1, 255, 'a'}
	query := "INSERT INTO " + table + " (id,label,n,fraction,payload,optional,occurred_at) VALUES (:id,:label,:n,:fraction,:payload,:optional,:at)"
	if _, err := db.ExecNamed(ctx, query, map[string]any{"id": int64(1), "label": label, "n": int64(42), "fraction": 1.5, "payload": payload, "optional": nil, "at": instant}); err != nil {
		t.Fatal(err)
	}
	rows, err := db.QueryNamed(ctx, "SELECT * FROM "+table+" WHERE label=:label OR label=:label", map[string]any{"label": label})
	if err != nil || len(rows) != 1 {
		t.Fatalf("rows=%v err=%v", rows, err)
	}
	var got struct {
		ID         int64
		Label      string
		N          int64
		Fraction   float64
		Payload    []byte
		Optional   sql.NullString
		OccurredAt time.Time `db:"occurred_at"`
	}
	if err := vdb.AssignEntity(rows[0], &got); err != nil {
		t.Fatal(err)
	}
	if got.ID != 1 || got.Label != label || got.N != 42 || got.Fraction != 1.5 || string(got.Payload) != string(payload) || got.Optional.Valid || !got.OccurredAt.Equal(instant) {
		t.Fatalf("scan=%+v", got)
	}
	var narrow struct{ N int8 }
	if err := vdb.AssignEntity(vdb.NewEntity("").Set("n", int64(256)), &narrow); err == nil {
		t.Fatal("numeric overflow accepted")
	}
	count(t, ctx, db, table, 1)
}

func upsert(t *testing.T, ctx context.Context, db *vdb.DB, table string, e engine) {
	entity := vdb.NewEntity(table).Set("id", int64(1)).Set("label", "initial").Set("n", int64(1))
	if _, err := db.Upsert(ctx, entity, []string{"id"}, "label", "n"); err != nil {
		t.Fatal(err)
	}
	entity.Set("label", "updated").Set("n", int64(7))
	result, err := db.Upsert(ctx, entity, []string{"id"}, "label", "n")
	if err != nil {
		t.Fatal(err)
	}
	affected, err := result.RowsAffected()
	want := int64(1)
	if e.name == "mysql" {
		want = 2
	}
	if err != nil || affected != want {
		t.Fatalf("affected=%d err=%v want=%d", affected, err, want)
	}
	row, ok, err := db.Get(ctx, table, "id", 1)
	if err != nil || !ok || row.Values["label"] != "updated" {
		t.Fatalf("upsert row=%+v ok=%v err=%v", row, ok, err)
	}
	count(t, ctx, db, table, 1)
}

func pagination(t *testing.T, ctx context.Context, db *vdb.DB, table string, _ engine) {
	for i := int64(1); i <= 5; i++ {
		insert(t, ctx, db, table, i, fmt.Sprintf("row-%d", i))
	}
	page, err := db.Page(ctx, vdb.NewQuery(table).Where(vdb.Gt("n", 1)), vdb.NewPage(2, 2, vdb.Asc("id")))
	if err != nil {
		t.Fatal(err)
	}
	if page.Total != 4 || len(page.Items) != 2 {
		t.Fatalf("page=%+v", page)
	}
	for i, row := range page.Items {
		var got struct{ ID int64 }
		if err := vdb.AssignEntity(row, &got); err != nil || got.ID != int64(i+4) {
			t.Fatalf("page item=%+v err=%v", row, err)
		}
	}
}

func transactions(t *testing.T, ctx context.Context, db *vdb.DB, table string, _ engine) {
	marker := errors.New("rollback marker")
	err := db.Tx(ctx, nil, func(s *vdb.Session) error {
		if _, err := s.Insert(ctx, vdb.NewEntity(table).Set("id", 1).Set("label", "rolled-back").Set("n", 1)); err != nil {
			return err
		}
		return marker
	})
	if !errors.Is(err, marker) {
		t.Fatalf("rollback error=%v", err)
	}
	count(t, ctx, db, table, 0)
	func() {
		defer func() {
			if recover() != marker {
				t.Error("panic was not preserved")
			}
		}()
		_ = db.Tx(ctx, nil, func(s *vdb.Session) error {
			if _, err := s.Insert(ctx, vdb.NewEntity(table).Set("id", 1).Set("label", "panic").Set("n", 1)); err != nil {
				t.Fatal(err)
			}
			panic(marker)
		})
	}()
	count(t, ctx, db, table, 0)
	err = db.Tx(ctx, nil, func(s *vdb.Session) error {
		if _, err := s.Insert(ctx, vdb.NewEntity(table).Set("id", 1).Set("label", "committed").Set("n", 1)); err != nil {
			return err
		}
		if err := s.Savepoint(ctx, "checkpoint"); err != nil {
			return err
		}
		if _, err := s.Insert(ctx, vdb.NewEntity(table).Set("id", 2).Set("label", "savepoint").Set("n", 2)); err != nil {
			return err
		}
		return s.RollbackTo(ctx, "checkpoint")
	})
	if err != nil {
		t.Fatal(err)
	}
	count(t, ctx, db, table, 1)
}

func partialBatch(t *testing.T, ctx context.Context, db *vdb.DB, table string, e engine) {
	named, err := vdb.ParseNamed("INSERT INTO "+table+" (id,label,n) VALUES (:id,:label,:n)", map[string]any{"id": 1, "label": "x", "n": 1}, e.dialect)
	if err != nil {
		t.Fatal(err)
	}
	batches := [][]any{{int64(1), "first", int64(1)}, {int64(1), "duplicate", int64(2)}, {int64(3), "not-executed", int64(3)}}
	results, err := db.ExecBatch(ctx, named.SQL, batches...)
	if err == nil || len(results) != 1 {
		t.Fatalf("batch results=%d error=%v", len(results), err)
	}
	count(t, ctx, db, table, 1)
	if _, err := db.DeleteAll(ctx, table); err != nil {
		t.Fatal(err)
	}
	err = db.Tx(ctx, nil, func(s *vdb.Session) error {
		results, err := s.ExecBatch(ctx, named.SQL, batches...)
		if len(results) != 1 {
			t.Errorf("transaction partial result count=%d", len(results))
		}
		return err
	})
	if err == nil {
		t.Fatal("transaction batch failure was swallowed")
	}
	count(t, ctx, db, table, 0)
}

func rowLifecycle(t *testing.T, ctx context.Context, db *vdb.DB, table string, _ engine) {
	insert(t, ctx, db, table, 1, "first")
	insert(t, ctx, db, table, 2, "second")
	rows, err := db.SQLDB().QueryContext(ctx, "SELECT * FROM "+table+" ORDER BY id")
	if err != nil {
		t.Fatal(err)
	}
	_, ok, err := vdb.ScanOne(rows)
	if err != nil || !ok {
		t.Fatalf("ScanOne=%v %v", ok, err)
	}
	if db.SQLDB().Stats().InUse != 0 {
		t.Fatal("ScanOne did not release its connection")
	}
	if _, ok, err := db.QueryOne(ctx, "SELECT * FROM "+table+" WHERE id=-1"); err != nil || ok {
		t.Fatalf("missing row=%v %v", ok, err)
	}
	value, ok, err := db.QueryScalar(ctx, "SELECT optional FROM "+table+" WHERE id=1")
	if err != nil || !ok || value != nil {
		t.Fatalf("NULL scalar=%v %v %v", value, ok, err)
	}
}

func cancellation(t *testing.T, ctx context.Context, db *vdb.DB, _ string, e engine) {
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := db.Query(cancelled, "SELECT 1"); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled query=%v", err)
	}
	query := "SELECT pg_sleep(5)"
	if e.name == "mysql" {
		query = "SELECT SLEEP(5)"
	}
	if e.name == "sqlite" {
		query = "WITH RECURSIVE c(n) AS (SELECT 0 UNION ALL SELECT n+1 FROM c WHERE n<100000000) SELECT SUM(n) FROM c"
	}
	deadline, stop := context.WithTimeout(ctx, 100*time.Millisecond)
	defer stop()
	start := time.Now()
	_, err := db.Query(deadline, query)
	if err == nil || deadline.Err() == nil || time.Since(start) > 3*time.Second {
		t.Fatalf("in-flight cancellation err=%v ctx=%v elapsed=%v", err, deadline.Err(), time.Since(start))
	}
	if err := db.Ping(ctx); err != nil {
		t.Fatalf("connection pool did not recover: %v", err)
	}
}

func deferredCommitError(t *testing.T, ctx context.Context, db *vdb.DB, e engine) {
	parent := newTable(t, ctx, db, e)
	child := parent + "_child"
	if _, err := db.Exec(ctx, "CREATE TABLE "+child+" (id BIGINT REFERENCES "+parent+"(id) DEFERRABLE INITIALLY DEFERRED)"); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if _, err := db.Exec(ctx, "DROP TABLE "+child); err != nil {
			t.Error(err)
		}
	}()
	inserted := false
	err := db.Tx(ctx, nil, func(s *vdb.Session) error {
		_, err := s.Exec(ctx, "INSERT INTO "+child+" (id) VALUES (999)")
		inserted = err == nil
		return err
	})
	if !inserted || err == nil {
		t.Fatalf("expected real commit-time failure: inserted=%v err=%v", inserted, err)
	}
	count(t, ctx, db, child, 0)
}
