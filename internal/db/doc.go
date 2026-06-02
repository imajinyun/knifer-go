// Package db provides database helpers built on top of database/sql.
//
// It owns SQL-specific concepts such as entities, conditions, query builders,
// named parameters, pagination, transactions, and lightweight metadata lookup.
// Callers pass an existing *sql.DB so connection pooling remains controlled by
// the standard library and the selected driver.
package db
