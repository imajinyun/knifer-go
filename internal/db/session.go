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
	result, err := s.tx.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, wrapInternal("db: execute transaction SQL", err)
	}
	return result, nil
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
	result, err := s.tx.ExecContext(ctx, sqlText, args...)
	if err != nil {
		return nil, wrapInternal("db: execute transaction upsert", err)
	}
	return result, nil
}

// Update updates rows in the transaction.
func (s *Session) Update(ctx context.Context, entity Entity, conds ...Condition) (sql.Result, error) {
	if len(conds) == 0 {
		return nil, invalidInputf("db: UPDATE without conditions is unsafe; use UpdateAll to update every row explicitly")
	}
	return updateEntity(ctx, s.tx, s.parent.dialect, s.parent.wrapper, entity, conds...)
}

// UpdateAll updates all rows in entity.Table within the transaction.
func (s *Session) UpdateAll(ctx context.Context, entity Entity) (sql.Result, error) {
	return updateEntity(ctx, s.tx, s.parent.dialect, s.parent.wrapper, entity)
}

// Delete deletes rows in the transaction.
func (s *Session) Delete(ctx context.Context, table string, conds ...Condition) (sql.Result, error) {
	if len(conds) == 0 {
		return nil, invalidInputf("db: DELETE without conditions is unsafe; use DeleteAll to delete every row explicitly")
	}
	return deleteRows(ctx, s.tx, s.parent.dialect, s.parent.wrapper, table, conds...)
}

// DeleteAll deletes all rows from table within the transaction.
func (s *Session) DeleteAll(ctx context.Context, table string) (sql.Result, error) {
	return deleteRows(ctx, s.tx, s.parent.dialect, s.parent.wrapper, table)
}

// Savepoint creates a savepoint when the selected driver supports the SQL syntax.
func (s *Session) Savepoint(ctx context.Context, name string) error {
	if err := validateIdentifier(name, "savepoint name"); err != nil {
		return err
	}
	_, err := s.tx.ExecContext(ctx, "SAVEPOINT "+s.parent.wrapper.Wrap(name))
	return wrapInternal("db: create savepoint", err)
}

// RollbackTo rolls back to a savepoint when supported by the selected driver.
func (s *Session) RollbackTo(ctx context.Context, name string) error {
	if err := validateIdentifier(name, "savepoint name"); err != nil {
		return err
	}
	_, err := s.tx.ExecContext(ctx, "ROLLBACK TO SAVEPOINT "+s.parent.wrapper.Wrap(name))
	return wrapInternal("db: rollback to savepoint", err)
}
