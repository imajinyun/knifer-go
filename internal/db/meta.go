package db

import (
	"context"
	"database/sql"
	"fmt"
)

// Column describes a table column.
type Column struct {
	TableName     string
	Name          string
	TypeName      string
	Nullable      bool
	DefaultValue  any
	PrimaryKey    bool
	AutoIncrement bool
}

// Table describes table metadata.
type Table struct {
	Catalog     string
	Schema      string
	Name        string
	Comment     string
	Columns     []Column
	PrimaryKeys []string
}

// ListTables lists user tables for supported dialects.
func (db *DB) ListTables(ctx context.Context) ([]string, error) {
	query, err := listTablesSQL(db.dialect)
	if err != nil {
		return nil, err
	}
	rows, err := db.sqlDB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	out := []string{}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		out = append(out, name)
	}
	return out, rows.Err()
}

// ListColumns lists columns for supported dialects.
func (db *DB) ListColumns(ctx context.Context, table string) ([]Column, error) {
	query, args, scanner, err := listColumnsSQL(db.dialect, table)
	if err != nil {
		return nil, err
	}
	rows, err := db.sqlDB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	out := []Column{}
	for rows.Next() {
		col, err := scanner(rows, table)
		if err != nil {
			return nil, err
		}
		out = append(out, col)
	}
	return out, rows.Err()
}

// TableMeta returns table metadata with columns and primary keys when available.
func (db *DB) TableMeta(ctx context.Context, table string) (Table, error) {
	cols, err := db.ListColumns(ctx, table)
	if err != nil {
		return Table{}, err
	}
	meta := Table{Name: table, Columns: cols}
	for _, col := range cols {
		if col.PrimaryKey {
			meta.PrimaryKeys = append(meta.PrimaryKeys, col.Name)
		}
	}
	return meta, nil
}

// ColumnNames returns column names for table.
func (db *DB) ColumnNames(ctx context.Context, table string) ([]string, error) {
	cols, err := db.ListColumns(ctx, table)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(cols))
	for _, col := range cols {
		names = append(names, col.Name)
	}
	return names, nil
}

// PrimaryKeys returns primary key column names for table.
func (db *DB) PrimaryKeys(ctx context.Context, table string) ([]string, error) {
	meta, err := db.TableMeta(ctx, table)
	if err != nil {
		return nil, err
	}
	return meta.PrimaryKeys, nil
}

type columnScanner func(*sql.Rows, string) (Column, error)

func listTablesSQL(d Dialect) (string, error) {
	switch d {
	case DialectSQLite:
		return "SELECT name FROM sqlite_master WHERE type = 'table' AND name NOT LIKE 'sqlite_%' ORDER BY name", nil
	case DialectMySQL:
		return "SELECT table_name FROM information_schema.tables WHERE table_schema = DATABASE() AND table_type = 'BASE TABLE' ORDER BY table_name", nil
	case DialectPostgres:
		return "SELECT table_name FROM information_schema.tables WHERE table_schema = current_schema() AND table_type = 'BASE TABLE' ORDER BY table_name", nil
	default:
		return "", fmt.Errorf("db: ListTables is not implemented for dialect %q", d)
	}
}

func listColumnsSQL(d Dialect, table string) (string, []any, columnScanner, error) {
	switch d {
	case DialectSQLite:
		return "PRAGMA table_info(" + WrapperForDialect(DialectSQLite).Wrap(table) + ")", nil, scanSQLiteColumn, nil
	case DialectMySQL:
		return "SELECT column_name, data_type, is_nullable, column_default, column_key, extra FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = ? ORDER BY ordinal_position", []any{table}, scanInformationSchemaColumn, nil
	case DialectPostgres:
		return "SELECT column_name, data_type, is_nullable, column_default, '' AS column_key, '' AS extra FROM information_schema.columns WHERE table_schema = current_schema() AND table_name = $1 ORDER BY ordinal_position", []any{table}, scanInformationSchemaColumn, nil
	default:
		return "", nil, nil, fmt.Errorf("db: ListColumns is not implemented for dialect %q", d)
	}
}

func scanSQLiteColumn(rows *sql.Rows, table string) (Column, error) {
	var cid int
	var name, typeName string
	var notNull int
	var defaultValue any
	var pk int
	if err := rows.Scan(&cid, &name, &typeName, &notNull, &defaultValue, &pk); err != nil {
		return Column{}, err
	}
	return Column{TableName: table, Name: name, TypeName: typeName, Nullable: notNull == 0, DefaultValue: defaultValue, PrimaryKey: pk > 0}, nil
}

func scanInformationSchemaColumn(rows *sql.Rows, table string) (Column, error) {
	var name, typeName, nullable string
	var defaultValue any
	var key, extra string
	if err := rows.Scan(&name, &typeName, &nullable, &defaultValue, &key, &extra); err != nil {
		return Column{}, err
	}
	return Column{TableName: table, Name: name, TypeName: typeName, Nullable: nullable == "YES", DefaultValue: defaultValue, PrimaryKey: key == "PRI", AutoIncrement: extra == "auto_increment"}, nil
}
