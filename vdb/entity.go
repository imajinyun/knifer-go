package vdb

import (
	"database/sql"

	dbimpl "github.com/imajinyun/knifer-go/internal/db"
)

func NewEntity(table string) Entity { return dbimpl.NewEntity(table) }

func EntityFromMap(table string, values map[string]any) Entity {
	return dbimpl.EntityFromMap(table, values)
}

func ScanRows(rows *sql.Rows) ([]Entity, error) { return dbimpl.ScanRows(rows) }

func ScanOne(rows *sql.Rows) (Entity, bool, error) { return dbimpl.ScanOne(rows) }

func AssignEntity(entity Entity, dst any) error { return dbimpl.AssignEntity(entity, dst) }
