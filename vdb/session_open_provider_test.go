package vdb

import (
	"database/sql"
	"errors"
	"testing"

	knifer "github.com/imajinyun/knifer-go"
)

func TestFacadeOpenProvider(t *testing.T) {
	opened := false
	customDB, err := Open("ignored", "ignored", WithSQLOpenFunc(func(driverName, dsn string) (*sql.DB, error) {
		opened = true
		return sql.Open("vdb_pool_test", "")
	}))
	if err != nil {
		t.Fatalf("Open with custom SQLOpen: %v", err)
	}
	defer func() { _ = customDB.Close() }()
	if !opened {
		t.Fatal("WithSQLOpenFunc provider was not called")
	}
}

func TestFacadeOpenProviderErrorContract(t *testing.T) {
	cause := errors.New("driver unavailable")
	_, err := Open("ignored", "ignored", WithSQLOpenFunc(func(string, string) (*sql.DB, error) {
		return nil, cause
	}))
	if !errors.Is(err, knifer.ErrCodeInternal) {
		t.Fatalf("Open provider error = %v, want ErrCodeInternal", err)
	}
	if !errors.Is(err, cause) {
		t.Fatalf("Open provider error = %v, want provider cause", err)
	}
	var dbErr *DBError
	if !errors.As(err, &dbErr) {
		t.Fatalf("errors.As(err, *DBError) = false: %v", err)
	}
}
