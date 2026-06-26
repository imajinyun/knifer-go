package vnum

import (
	"math/big"

	numimpl "github.com/imajinyun/knifer-go/internal/num"
)

func Factorial(n uint64) (uint64, error) { return numimpl.Factorial(n) }

func FactorialRange(start, end uint64) (uint64, error) { return numimpl.FactorialRange(start, end) }

func FactorialBig(n *big.Int) *big.Int { return numimpl.FactorialBig(n) }

func FactorialBigRange(start, end *big.Int) *big.Int { return numimpl.FactorialBigRange(start, end) }

func Sqrt(x uint64) uint64 { return numimpl.Sqrt(x) }

func ProcessMultiple(selectNum, minNum int) int { return numimpl.ProcessMultiple(selectNum, minNum) }

func Divisor(m, n int) int { return numimpl.Divisor(m, n) }

func Multiple(m, n int) int { return numimpl.Multiple(m, n) }
