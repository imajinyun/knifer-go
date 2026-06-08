package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

type sqlExecutor interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

// DB wraps database/sql with SQL helper methods.
type DB struct {
	sqlDB   *sql.DB
	dialect Dialect
	wrapper Wrapper
}

// Open opens a database using database/sql and applies pool options.
func Open(driverName, dataSourceName string, opts ...Option) (*DB, error) {
	cfg := applyOptions(append([]Option{WithDialect(NormalizeDialect(driverName))}, opts...)...)
	sqlDB, err := cfg.SQLOpen(driverName, dataSourceName)
	if err != nil {
		return nil, wrapInternal("db: open database", err)
	}
	applyPoolOptions(sqlDB, cfg)
	return &DB{sqlDB: sqlDB, dialect: cfg.Dialect, wrapper: cfg.Wrapper}, nil
}

// Use wraps an existing *sql.DB.
func Use(sqlDB *sql.DB, opts ...Option) *DB {
	cfg := applyOptions(opts...)
	applyPoolOptions(sqlDB, cfg)
	return &DB{sqlDB: sqlDB, dialect: cfg.Dialect, wrapper: cfg.Wrapper}
}

// SQLDB returns the underlying *sql.DB.
func (db *DB) SQLDB() *sql.DB { return db.sqlDB }

// Dialect returns the configured dialect.
func (db *DB) Dialect() Dialect { return db.dialect }

// Wrapper returns the configured wrapper.
func (db *DB) Wrapper() Wrapper { return db.wrapper }

// Close closes the underlying database.
func (db *DB) Close() error { return db.sqlDB.Close() }

// Ping pings the database.
func (db *DB) Ping(ctx context.Context) error { return db.sqlDB.PingContext(ctx) }

// Exec executes SQL.
func (db *DB) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	result, err := db.sqlDB.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, wrapInternal("db: execute SQL", err)
	}
	return result, nil
}

// ExecNamed executes named-parameter SQL.
func (db *DB) ExecNamed(ctx context.Context, query string, args map[string]any) (sql.Result, error) {
	named, err := ParseNamed(query, args, db.dialect)
	if err != nil {
		return nil, err
	}
	return db.Exec(ctx, named.SQL, named.Params...)
}

// ExecBatch executes the same SQL with multiple argument sets.
func (db *DB) ExecBatch(ctx context.Context, query string, batches ...[]any) ([]sql.Result, error) {
	return execBatch(ctx, db.sqlDB, query, batches...)
}

// Query executes SQL and scans rows into Entity values.
func (db *DB) Query(ctx context.Context, query string, args ...any) ([]Entity, error) {
	return queryEntities(ctx, db.sqlDB, query, args...)
}

// QueryNamed executes named-parameter SQL and scans rows into Entity values.
func (db *DB) QueryNamed(ctx context.Context, query string, args map[string]any) ([]Entity, error) {
	named, err := ParseNamed(query, args, db.dialect)
	if err != nil {
		return nil, err
	}
	return db.Query(ctx, named.SQL, named.Params...)
}

// QueryOne returns the first row.
func (db *DB) QueryOne(ctx context.Context, query string, args ...any) (Entity, bool, error) {
	return queryOne(ctx, db.sqlDB, query, args...)
}

// QueryScalar returns the first column from the first row.
func (db *DB) QueryScalar(ctx context.Context, query string, args ...any) (any, bool, error) {
	return queryScalar(ctx, db.sqlDB, query, args...)
}

// Insert inserts entity values.
func (db *DB) Insert(ctx context.Context, entity Entity) (sql.Result, error) {
	return insertEntity(ctx, db.sqlDB, db.dialect, db.wrapper, entity)
}

// InsertGetID inserts entity values and returns LastInsertId.
func (db *DB) InsertGetID(ctx context.Context, entity Entity) (int64, error) {
	result, err := db.Insert(ctx, entity)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, wrapInternal("db: read last insert id", err)
	}
	return id, nil
}

// Upsert inserts entity values or updates existing rows when supported by the dialect.
func (db *DB) Upsert(ctx context.Context, entity Entity, conflictFields []string, updateFields ...string) (sql.Result, error) {
	sqlText, args, err := buildUpsertSQL(db.dialect, db.wrapper, entity, conflictFields, updateFields...)
	if err != nil {
		return nil, err
	}
	result, err := db.sqlDB.ExecContext(ctx, sqlText, args...)
	if err != nil {
		return nil, wrapInternal("db: execute upsert", err)
	}
	return result, nil
}

// Update updates rows matching conditions.
func (db *DB) Update(ctx context.Context, entity Entity, conds ...Condition) (sql.Result, error) {
	if len(conds) == 0 {
		return nil, invalidInputf("db: UPDATE without conditions is unsafe; use UpdateAll to update every row explicitly")
	}
	return updateEntity(ctx, db.sqlDB, db.dialect, db.wrapper, entity, conds...)
}

// UpdateAll updates all rows in entity.Table. Use deliberately; no WHERE clause is generated.
func (db *DB) UpdateAll(ctx context.Context, entity Entity) (sql.Result, error) {
	return updateEntity(ctx, db.sqlDB, db.dialect, db.wrapper, entity)
}

// Delete deletes rows matching conditions.
func (db *DB) Delete(ctx context.Context, table string, conds ...Condition) (sql.Result, error) {
	if len(conds) == 0 {
		return nil, invalidInputf("db: DELETE without conditions is unsafe; use DeleteAll to delete every row explicitly")
	}
	return deleteRows(ctx, db.sqlDB, db.dialect, db.wrapper, table, conds...)
}

// DeleteAll deletes all rows from table. Use deliberately; no WHERE clause is generated.
func (db *DB) DeleteAll(ctx context.Context, table string) (sql.Result, error) {
	return deleteRows(ctx, db.sqlDB, db.dialect, db.wrapper, table)
}

// DeleteEntity deletes rows using entity values as equality conditions.
func (db *DB) DeleteEntity(ctx context.Context, entity Entity) (sql.Result, error) {
	return db.Delete(ctx, entity.Table, ConditionsFromEntity(entity)...)
}

// Get finds one row by field equality.
func (db *DB) Get(ctx context.Context, table, field string, value any) (Entity, bool, error) {
	sqlText, args, err := NewBuilder(WithDialect(db.dialect), WithWrapper(db.wrapper)).Select("*").From(table).Where(Eq(field, value)).Page(NewPage(1, 1)).SQL()
	if err != nil {
		return Entity{}, false, err
	}
	return db.QueryOne(ctx, sqlText, args...)
}

// Find runs a Query.
func (db *DB) Find(ctx context.Context, q Query) ([]Entity, error) {
	b := NewBuilder(WithDialect(db.dialect), WithWrapper(db.wrapper)).Query(q)
	sqlText, args, err := b.SQL()
	if err != nil {
		return nil, err
	}
	return db.Query(ctx, sqlText, args...)
}

// FindAll returns all rows from table.
func (db *DB) FindAll(ctx context.Context, table string) ([]Entity, error) {
	return db.Find(ctx, NewQuery(table))
}

// FindBy returns rows where field equals value.
func (db *DB) FindBy(ctx context.Context, table, field string, value any) ([]Entity, error) {
	return db.Find(ctx, NewQuery(table).Where(Eq(field, value)))
}

// FindLike returns rows where field matches a LIKE pattern.
func (db *DB) FindLike(ctx context.Context, table, field string, value any, mode string) ([]Entity, error) {
	return db.Find(ctx, NewQuery(table).Where(Like(field, BuildLikeValue(value, mode))))
}

// Count counts rows in table matching conditions.
func (db *DB) Count(ctx context.Context, table string, conds ...Condition) (int64, error) {
	sqlText, args, err := buildCountSQL(db.dialect, db.wrapper, []string{table}, conds...)
	if err != nil {
		return 0, err
	}
	return scanInt64(db.sqlDB.QueryRowContext(ctx, sqlText, args...))
}

// Page runs a paged query and count query.
func (db *DB) Page(ctx context.Context, q Query, page Page) (PageResult[Entity], error) {
	q.Page = &page
	items, err := db.Find(ctx, q)
	if err != nil {
		return PageResult[Entity]{}, err
	}
	countSQL, countArgs, err := buildCountSQL(db.dialect, db.wrapper, q.Tables, q.Conditions...)
	if err != nil {
		return PageResult[Entity]{}, err
	}
	total, err := scanInt64(db.sqlDB.QueryRowContext(ctx, countSQL, countArgs...))
	if err != nil {
		return PageResult[Entity]{}, err
	}
	return NewPageResult(page, total, items), nil
}

// Tx runs fn inside a transaction.
func (db *DB) Tx(ctx context.Context, opts *sql.TxOptions, fn func(*Session) error) error {
	tx, err := db.sqlDB.BeginTx(ctx, opts)
	if err != nil {
		return wrapInternal("db: begin transaction", err)
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			// Preserve panic semantics after rollback; bin/check_arch.sh allowlists
			// this transaction-boundary rethrow explicitly.
			panic(p)
		}
	}()
	s := &Session{tx: tx, parent: db}
	if err := fn(s); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("%w; rollback: %v", err, rbErr)
		}
		return err
	}
	if err := tx.Commit(); err != nil {
		return wrapInternal("db: commit transaction", err)
	}
	return nil
}

func execBatch(ctx context.Context, exec sqlExecutor, query string, batches ...[]any) ([]sql.Result, error) {
	results := make([]sql.Result, 0, len(batches))
	for _, args := range batches {
		result, err := exec.ExecContext(ctx, query, args...)
		if err != nil {
			return results, wrapInternal("db: execute batch SQL", err)
		}
		results = append(results, result)
	}
	return results, nil
}

func queryEntities(ctx context.Context, exec sqlExecutor, query string, args ...any) ([]Entity, error) {
	rows, err := exec.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, wrapInternal("db: query entities", err)
	}
	return ScanRows(rows)
}

func queryOne(ctx context.Context, exec sqlExecutor, query string, args ...any) (Entity, bool, error) {
	rows, err := exec.QueryContext(ctx, query, args...)
	if err != nil {
		return Entity{}, false, wrapInternal("db: query one", err)
	}
	return ScanOne(rows)
}

func queryScalar(ctx context.Context, exec sqlExecutor, query string, args ...any) (any, bool, error) {
	rows, err := exec.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, false, wrapInternal("db: query scalar", err)
	}
	defer func() { _ = rows.Close() }()
	if !rows.Next() {
		return nil, false, wrapInternal("db: iterate scalar rows", rows.Err())
	}
	var value any
	if err := rows.Scan(&value); err != nil {
		return nil, false, wrapInternal("db: scan scalar", err)
	}
	return normalizeDBValue(value), true, wrapInternal("db: iterate scalar rows", rows.Err())
}

func insertEntity(ctx context.Context, exec sqlExecutor, dialect Dialect, wrapper Wrapper, entity Entity) (sql.Result, error) {
	sqlText, args, err := NewBuilder(WithDialect(dialect), WithWrapper(wrapper)).Insert(entity).SQL()
	if err != nil {
		return nil, err
	}
	result, err := exec.ExecContext(ctx, sqlText, args...)
	if err != nil {
		return nil, wrapInternal("db: insert entity", err)
	}
	return result, nil
}

func updateEntity(ctx context.Context, exec sqlExecutor, dialect Dialect, wrapper Wrapper, entity Entity, conds ...Condition) (sql.Result, error) {
	sqlText, args, err := NewBuilder(WithDialect(dialect), WithWrapper(wrapper)).Update(entity).Where(conds...).SQL()
	if err != nil {
		return nil, err
	}
	result, err := exec.ExecContext(ctx, sqlText, args...)
	if err != nil {
		return nil, wrapInternal("db: update entity", err)
	}
	return result, nil
}

func deleteRows(ctx context.Context, exec sqlExecutor, dialect Dialect, wrapper Wrapper, table string, conds ...Condition) (sql.Result, error) {
	sqlText, args, err := NewBuilder(WithDialect(dialect), WithWrapper(wrapper)).Delete(table).Where(conds...).SQL()
	if err != nil {
		return nil, err
	}
	result, err := exec.ExecContext(ctx, sqlText, args...)
	if err != nil {
		return nil, wrapInternal("db: delete rows", err)
	}
	return result, nil
}

func buildCountSQL(dialect Dialect, wrapper Wrapper, tables []string, conds ...Condition) (string, []any, error) {
	if len(tables) == 0 {
		return "", nil, invalidInputf("db: COUNT requires table")
	}
	if err := validateIdentifierList(tables, "COUNT tables", false); err != nil {
		return "", nil, err
	}
	parts := []string{"SELECT COUNT(*) FROM", wrapList(tables, wrapper)}
	if len(conds) == 0 {
		return strings.Join(parts, " "), nil, nil
	}
	where, args, _, err := buildConditions(conds, dialect, wrapper, 1)
	if err != nil {
		return "", nil, err
	}
	if where != "" {
		parts = append(parts, "WHERE", where)
	}
	return strings.Join(parts, " "), args, nil
}

func scanInt64(row *sql.Row) (int64, error) {
	var n int64
	if err := row.Scan(&n); err != nil {
		return 0, wrapInternal("db: scan int64", err)
	}
	return n, nil
}

func buildUpsertSQL(dialect Dialect, wrapper Wrapper, entity Entity, conflictFields []string, updateFields ...string) (string, []any, error) {
	insertSQL, args, err := NewBuilder(WithDialect(dialect), WithWrapper(wrapper)).Insert(entity).SQL()
	if err != nil {
		return "", nil, err
	}
	keys := entity.sortedKeys()
	if len(updateFields) == 0 {
		conflict := map[string]struct{}{}
		for _, field := range conflictFields {
			conflict[field] = struct{}{}
		}
		for _, key := range keys {
			if _, ok := conflict[key]; !ok {
				updateFields = append(updateFields, key)
			}
		}
	}
	if len(updateFields) == 0 {
		switch dialect {
		case DialectPostgres, DialectSQLite:
			if len(conflictFields) == 0 {
				return "", nil, invalidInputf("db: upsert conflict fields are required for %s", dialect)
			}
			return insertSQL + " ON CONFLICT (" + wrapList(conflictFields, wrapper) + ") DO NOTHING", args, nil
		default:
			return insertSQL, args, nil
		}
	}
	sets := make([]string, 0, len(updateFields))
	for _, field := range updateFields {
		if strings.TrimSpace(field) == "" {
			continue
		}
		if err := validateIdentifier(field, "upsert update field"); err != nil {
			return "", nil, err
		}
		switch dialect {
		case DialectMySQL:
			sets = append(sets, wrapper.Wrap(field)+" = VALUES("+wrapper.Wrap(field)+")")
		default:
			sets = append(sets, wrapper.Wrap(field)+" = excluded."+wrapper.Wrap(field))
		}
	}
	if len(sets) == 0 {
		return insertSQL, args, nil
	}
	switch dialect {
	case DialectMySQL:
		return insertSQL + " ON DUPLICATE KEY UPDATE " + strings.Join(sets, ", "), args, nil
	case DialectPostgres, DialectSQLite:
		if len(conflictFields) == 0 {
			return "", nil, invalidInputf("db: upsert conflict fields are required for %s", dialect)
		}
		if err := validateIdentifierList(conflictFields, "upsert conflict fields", false); err != nil {
			return "", nil, err
		}
		return insertSQL + " ON CONFLICT (" + wrapList(conflictFields, wrapper) + ") DO UPDATE SET " + strings.Join(sets, ", "), args, nil
	default:
		return "", nil, unsupportedf("db: upsert is not implemented for dialect %q", dialect)
	}
}
