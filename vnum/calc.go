package vnum

import numimpl "github.com/imajinyun/knifer-go/internal/num"

func Calculate(expression string) (float64, error) { return numimpl.Calculate(expression) }

func CalculateWithOptions(expression string, opts ...ParseOption) (float64, error) {
	return numimpl.CalculateWithOptions(expression, opts...)
}
