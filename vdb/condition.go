package vdb

import dbimpl "github.com/imajinyun/go-knifer/internal/db"

func NewQuery(tables ...string) Query { return dbimpl.NewQuery(tables...) }

func Eq(field string, value any) Condition { return dbimpl.Eq(field, value) }

func Ne(field string, value any) Condition { return dbimpl.Ne(field, value) }

func Gt(field string, value any) Condition { return dbimpl.Gt(field, value) }

func Gte(field string, value any) Condition { return dbimpl.Gte(field, value) }

func Lt(field string, value any) Condition { return dbimpl.Lt(field, value) }

func Lte(field string, value any) Condition { return dbimpl.Lte(field, value) }

func Like(field string, value any) Condition { return dbimpl.Like(field, value) }

func In(field string, values ...any) Condition { return dbimpl.In(field, values...) }

func Between(field string, begin, end any) Condition { return dbimpl.Between(field, begin, end) }

func IsNull(field string) Condition { return dbimpl.IsNull(field) }

func IsNotNull(field string) Condition { return dbimpl.IsNotNull(field) }

func OrWith(c Condition) Condition { return dbimpl.OrWith(c) }

func AndGroup(conds ...Condition) Condition { return dbimpl.AndGroup(conds...) }

func OrGroup(conds ...Condition) Condition { return dbimpl.OrGroup(conds...) }

func ConditionsFromEntity(e Entity) []Condition { return dbimpl.ConditionsFromEntity(e) }

func BuildConditions(conds ...Condition) (string, []any, error) {
	return dbimpl.BuildConditions(conds...)
}

func BuildLikeValue(value any, mode string) string { return dbimpl.BuildLikeValue(value, mode) }
