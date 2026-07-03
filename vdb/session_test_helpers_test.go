package vdb

import (
	"database/sql"
	"database/sql/driver"
	"io"
)

var scriptedRowsForTest driver.Rows

func init() {
	sql.Register("vdb_pool_test", poolTestDriver{})
	sql.Register("vdb_scripted_rows_test", scriptedRowsDriver{})
}

type poolTestDriver struct{}

func (poolTestDriver) Open(string) (driver.Conn, error) { return poolTestConn{}, nil }

type poolTestConn struct{}

func (poolTestConn) Prepare(string) (driver.Stmt, error) { return poolTestStmt{}, nil }
func (poolTestConn) Close() error                        { return nil }
func (poolTestConn) Begin() (driver.Tx, error)           { return poolTestTx{}, nil }

type poolTestStmt struct{}

func (poolTestStmt) Close() error                               { return nil }
func (poolTestStmt) NumInput() int                              { return -1 }
func (poolTestStmt) Exec([]driver.Value) (driver.Result, error) { return driver.RowsAffected(0), nil }
func (poolTestStmt) Query([]driver.Value) (driver.Rows, error)  { return poolTestRows{}, nil }

type poolTestRows struct{}

func (poolTestRows) Columns() []string         { return []string{"id"} }
func (poolTestRows) Close() error              { return nil }
func (poolTestRows) Next([]driver.Value) error { return io.EOF }

type poolTestTx struct{}

func (poolTestTx) Commit() error   { return nil }
func (poolTestTx) Rollback() error { return nil }

type scriptedRowsDriver struct{}

func (scriptedRowsDriver) Open(string) (driver.Conn, error) { return scriptedRowsConn{}, nil }

type scriptedRowsConn struct{}

func (scriptedRowsConn) Prepare(string) (driver.Stmt, error) { return scriptedRowsStmt{}, nil }
func (scriptedRowsConn) Close() error                        { return nil }
func (scriptedRowsConn) Begin() (driver.Tx, error)           { return poolTestTx{}, nil }

type scriptedRowsStmt struct{}

func (scriptedRowsStmt) Close() error  { return nil }
func (scriptedRowsStmt) NumInput() int { return -1 }
func (scriptedRowsStmt) Exec([]driver.Value) (driver.Result, error) {
	return driver.RowsAffected(0), nil
}

func (scriptedRowsStmt) Query([]driver.Value) (driver.Rows, error) {
	if scriptedRowsForTest != nil {
		return scriptedRowsForTest, nil
	}
	return poolTestRows{}, nil
}

type scriptedRows struct {
	cols    []string
	data    [][]driver.Value
	pos     int
	nextErr error
	closed  *bool
}

func (r *scriptedRows) Columns() []string { return r.cols }

func (r *scriptedRows) Close() error {
	if r.closed != nil {
		*r.closed = true
	}
	return nil
}

func (r *scriptedRows) Next(dest []driver.Value) error {
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
