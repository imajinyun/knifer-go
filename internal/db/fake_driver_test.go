package db

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
)

// fakeBehavior is a programmable script for the in-memory test driver. It lets
// each test inject deterministic results and errors for Exec/Query/transaction
// boundaries without depending on a real database driver.
type fakeBehavior struct {
	execResult  driver.Result
	execErr     error
	execErrFunc func() error
	queryErr    error
	beginErr    error
	commitErr   error
	rollbackErr error
	queryFunc   func(query string) (driver.Rows, error)
}

func (b *fakeBehavior) exec(string, []driver.NamedValue) (driver.Result, error) {
	if b.execErrFunc != nil {
		if err := b.execErrFunc(); err != nil {
			return nil, err
		}
	}
	if b.execErr != nil {
		return nil, b.execErr
	}
	if b.execResult != nil {
		return b.execResult, nil
	}
	return fakeResult{}, nil
}

func (b *fakeBehavior) query(query string, _ []driver.NamedValue) (driver.Rows, error) {
	if b.queryErr != nil {
		return nil, b.queryErr
	}
	if b.queryFunc != nil {
		return b.queryFunc(query)
	}
	return &fakeRows{}, nil
}

type fakeConnector struct{ b *fakeBehavior }

func (c *fakeConnector) Connect(context.Context) (driver.Conn, error) {
	return &fakeConn{b: c.b}, nil
}

func (c *fakeConnector) Driver() driver.Driver { return fakeDriver{} }

type fakeDriver struct{}

func (fakeDriver) Open(string) (driver.Conn, error) { return &fakeConn{b: &fakeBehavior{}}, nil }

type fakeConn struct{ b *fakeBehavior }

func (c *fakeConn) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("prepare not supported by fake driver")
}

func (c *fakeConn) Close() error { return nil }

func (c *fakeConn) Begin() (driver.Tx, error) {
	if c.b.beginErr != nil {
		return nil, c.b.beginErr
	}
	return &fakeTx{b: c.b}, nil
}

func (c *fakeConn) ExecContext(_ context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	return c.b.exec(query, args)
}

func (c *fakeConn) QueryContext(_ context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	return c.b.query(query, args)
}

type fakeTx struct{ b *fakeBehavior }

func (t *fakeTx) Commit() error   { return t.b.commitErr }
func (t *fakeTx) Rollback() error { return t.b.rollbackErr }

type fakeResult struct {
	lastID    int64
	affected  int64
	lastIDErr error
	affectErr error
}

func (r fakeResult) LastInsertId() (int64, error) { return r.lastID, r.lastIDErr }
func (r fakeResult) RowsAffected() (int64, error) { return r.affected, r.affectErr }

type fakeRows struct {
	cols    []string
	data    [][]driver.Value
	pos     int
	nextErr error
	closeFn func()
}

func (r *fakeRows) Columns() []string { return r.cols }
func (r *fakeRows) Close() error {
	if r.closeFn != nil {
		r.closeFn()
	}
	return nil
}

func (r *fakeRows) Next(dest []driver.Value) error {
	if r.pos >= len(r.data) {
		if r.nextErr != nil {
			return r.nextErr
		}
		return io.EOF
	}
	copy(dest, r.data[r.pos])
	r.pos++
	return nil
}

// mkRows builds a fakeRows with the given columns and row values.
func mkRows(cols []string, data ...[]driver.Value) *fakeRows {
	return &fakeRows{cols: cols, data: data}
}

// newFakeDB wires a *DB backed by the programmable fake driver.
func newFakeDB(b *fakeBehavior, opts ...Option) *DB {
	sqlDB := sql.OpenDB(&fakeConnector{b: b})
	return Use(sqlDB, opts...)
}
