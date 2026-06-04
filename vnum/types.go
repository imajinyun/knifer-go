package vnum

import numimpl "github.com/imajinyun/go-knifer/internal/num"

type (
	Number       = numimpl.Number
	Ordered      = numimpl.Ordered
	RoundingMode = numimpl.RoundingMode
)

const (
	RoundHalfUp   = numimpl.RoundHalfUp
	RoundHalfEven = numimpl.RoundHalfEven
	RoundDown     = numimpl.RoundDown
)
