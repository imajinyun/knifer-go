package vnum

import numimpl "github.com/imajinyun/knifer-go/internal/num"

func Range(start, end, step int) []int { return numimpl.Range(start, end, step) }

func RangeClosed(start, stop, step int) []int { return numimpl.RangeClosed(start, stop, step) }

func AppendRange(start, stop, step int, values []int) []int {
	return numimpl.AppendRange(start, stop, step, values)
}
