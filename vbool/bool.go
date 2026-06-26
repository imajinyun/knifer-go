package vbool

import boolimpl "github.com/imajinyun/knifer-go/internal/boolean"

func Negate(b bool) bool  { return boolimpl.Negate(b) }
func ToInt(b bool) int    { return boolimpl.ToInt(b) }
func And(bs ...bool) bool { return boolimpl.And(bs...) }
func Or(bs ...bool) bool  { return boolimpl.Or(bs...) }
