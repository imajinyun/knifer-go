package db

import (
	"context"
	"database/sql"
)

// Session wraps a database transaction.
type Session struct {
	tx     *sql.Tx
	parent *DB
}

// Exec executes SQL in the transaction.
func (s *Session) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return s.tx.ExecContext(ctx, query, args...)
}

// ExecNamed executes named-parameter SQL in the transaction.
func (s *Session) ExecNamed(ctx context.Context, query string, args map[string]any) (sql.Result, error) {
	named, err := ParseNamed(query, args, s.parent.dialect)
	if err != nil {
		return nil, err
	}
	return s.Exec(ctx, named.SQL, named.Params...)
}

// ExecBatch executes the same SQL with multiple argument sets.
func (s *Session) ExecBatch(ctx context.Context, query string, batches ...[]any) ([]sql.Result, error) {
	return execBatch(ctx, s.tx, query, batches...)
}

// Query executes SQL in the transaction.
func (s *Session) Query(ctx context.Context, query string, args ...any) ([]Entity, error) {
	return queryEntities(ctx, s.tx, query, args...)
}

// QueryOne returns the first row in the transaction.
func (s *Session) QueryOne(ctx context.Context, query string, args ...any) (Entity, bool, error) {
	return queryOne(ctx, s.tx, query, args...)
}

// QueryScalar returns the first column from the first row in the transaction.
func (s *Session) QueryScalar(ctx context.Context, query string, args ...any) (any, bool, error) {
	return queryScalar(ctx, s.tx, query, args...)
}

// Insert inserts entity values in the transaction.
func (s *Session) Insert(ctx context.Context, entity Entity) (sql.Result, error) {
	return insertEntity(ctx, s.tx, s.parent.dialect, s.parent.wrapper, entity)
}

// Upsert inserts entity values or updates existing rows in the transaction.
func (s *Session) Upsert(ctx context.Context, entity Entity, conflictFields []string, updateFields ...string) (sql.Result, error) {
	sqlText, args, err := buildUpsertSQL(s.parent.dialect, s.parent.wrapper, entity, conflictFields, updateFields...)
	if err != nil {
		return nil, err
	}
	return s.tx.ExecContext(ctx, sqlText, args...)
}

// Update updates rows in the transaction.
func (s *Session) Update(ctx context.Context, entity Entity, conds ...Condition) (sql.Result, error) {
	return updateEntity(ctx, s.tx, s.parent.dialect, s.parent.wrapper, entity, conds...)
}

// Delete deletes rows in the transaction.
func (s *Session) Delete(ctx context.Context, table string, conds ...Condition) (sql.Result, error) {
	return deleteRows(ctx, s.tx, s.parent.dialect, s.parent.wrapper, table, conds...)
}

// Savepoint creates a savepoint when the selected driver supports the SQL syntax.
func (s *Session) Savepoint(ctx context.Context, name string) error {
	_, err := s.tx.ExecContext(ctx, "SAVEPOINT "+s.parent.wrapper.Wrap(name))
	return err
}

// RollbackTo rolls back to a savepoint when supported by the selected driver.
func (s *Session) RollbackTo(ctx context.Context, name string) error {
	_, err := s.tx.ExecContext(ctx, "ROLLBACK TO SAVEPOINT "+s.parent.wrapper.Wrap(name))
	return err
}
