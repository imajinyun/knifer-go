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
	sqlDB, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return nil, err
	}
	return Use(sqlDB, append([]Option{WithDialect(NormalizeDialect(driverName))}, opts...)...), nil
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
	return db.sqlDB.ExecContext(ctx, query, args...)
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
	return result.LastInsertId()
}

// Upsert inserts entity values or updates existing rows when supported by the dialect.
func (db *DB) Upsert(ctx context.Context, entity Entity, conflictFields []string, updateFields ...string) (sql.Result, error) {
	sqlText, args, err := buildUpsertSQL(db.dialect, db.wrapper, entity, conflictFields, updateFields...)
	if err != nil {
		return nil, err
	}
	return db.sqlDB.ExecContext(ctx, sqlText, args...)
}

// Update updates rows matching conditions.
func (db *DB) Update(ctx context.Context, entity Entity, conds ...Condition) (sql.Result, error) {
	return updateEntity(ctx, db.sqlDB, db.dialect, db.wrapper, entity, conds...)
}

// Delete deletes rows matching conditions.
func (db *DB) Delete(ctx context.Context, table string, conds ...Condition) (sql.Result, error) {
	return deleteRows(ctx, db.sqlDB, db.dialect, db.wrapper, table, conds...)
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
	b := NewBuilder(WithDialect(db.dialect), WithWrapper(db.wrapper)).Select("COUNT(*)").From(table).Where(conds...)
	sqlText, args, err := b.SQL()
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
	countSQL, countArgs, err := NewBuilder(WithDialect(db.dialect), WithWrapper(db.wrapper)).Select("COUNT(*)").From(q.Tables...).Where(q.Conditions...).SQL()
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
		return err
	}
	s := &Session{tx: tx, parent: db}
	if err := fn(s); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("%w; rollback: %v", err, rbErr)
		}
		return err
	}
	return tx.Commit()
}

func execBatch(ctx context.Context, exec sqlExecutor, query string, batches ...[]any) ([]sql.Result, error) {
	results := make([]sql.Result, 0, len(batches))
	for _, args := range batches {
		result, err := exec.ExecContext(ctx, query, args...)
		if err != nil {
			return results, err
		}
		results = append(results, result)
	}
	return results, nil
}

func queryEntities(ctx context.Context, exec sqlExecutor, query string, args ...any) ([]Entity, error) {
	rows, err := exec.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return ScanRows(rows)
}

func queryOne(ctx context.Context, exec sqlExecutor, query string, args ...any) (Entity, bool, error) {
	rows, err := exec.QueryContext(ctx, query, args...)
	if err != nil {
		return Entity{}, false, err
	}
	return ScanOne(rows)
}

func queryScalar(ctx context.Context, exec sqlExecutor, query string, args ...any) (any, bool, error) {
	rows, err := exec.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, false, err
	}
	defer func() { _ = rows.Close() }()
	if !rows.Next() {
		return nil, false, rows.Err()
	}
	var value any
	if err := rows.Scan(&value); err != nil {
		return nil, false, err
	}
	return normalizeDBValue(value), true, rows.Err()
}

func insertEntity(ctx context.Context, exec sqlExecutor, dialect Dialect, wrapper Wrapper, entity Entity) (sql.Result, error) {
	sqlText, args, err := NewBuilder(WithDialect(dialect), WithWrapper(wrapper)).Insert(entity).SQL()
	if err != nil {
		return nil, err
	}
	return exec.ExecContext(ctx, sqlText, args...)
}

func updateEntity(ctx context.Context, exec sqlExecutor, dialect Dialect, wrapper Wrapper, entity Entity, conds ...Condition) (sql.Result, error) {
	sqlText, args, err := NewBuilder(WithDialect(dialect), WithWrapper(wrapper)).Update(entity).Where(conds...).SQL()
	if err != nil {
		return nil, err
	}
	return exec.ExecContext(ctx, sqlText, args...)
}

func deleteRows(ctx context.Context, exec sqlExecutor, dialect Dialect, wrapper Wrapper, table string, conds ...Condition) (sql.Result, error) {
	sqlText, args, err := NewBuilder(WithDialect(dialect), WithWrapper(wrapper)).Delete(table).Where(conds...).SQL()
	if err != nil {
		return nil, err
	}
	return exec.ExecContext(ctx, sqlText, args...)
}

func scanInt64(row *sql.Row) (int64, error) {
	var n int64
	if err := row.Scan(&n); err != nil {
		return 0, err
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
				return "", nil, fmt.Errorf("db: upsert conflict fields are required for %s", dialect)
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
			return "", nil, fmt.Errorf("db: upsert conflict fields are required for %s", dialect)
		}
		return insertSQL + " ON CONFLICT (" + wrapList(conflictFields, wrapper) + ") DO UPDATE SET " + strings.Join(sets, ", "), args, nil
	default:
		return "", nil, fmt.Errorf("db: upsert is not implemented for dialect %q", dialect)
	}
}
