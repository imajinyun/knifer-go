package vnum

import numimpl "github.com/imajinyun/knifer-go/internal/num"

type (
	Number             = numimpl.Number
	Ordered            = numimpl.Ordered
	RoundingMode       = numimpl.RoundingMode
	RandomNumberOption = numimpl.RandomNumberOption
	ParseOption        = numimpl.ParseOption
	FormatOption       = numimpl.FormatOption
	DoubleOption       = numimpl.DoubleOption
)

const (
	RoundHalfUp   = numimpl.RoundHalfUp
	RoundHalfEven = numimpl.RoundHalfEven
	RoundDown     = numimpl.RoundDown
)
