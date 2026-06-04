package vnum

import numimpl "github.com/imajinyun/go-knifer/internal/num"

func GenerateRandomNumber(begin, end, size int) []int {
	return numimpl.GenerateRandomNumber(begin, end, size)
}

func GenerateRandomNumberWithSeed(begin, end, size int, seed []int) []int {
	return numimpl.GenerateRandomNumberWithSeed(begin, end, size, seed)
}

func GenerateBySet(begin, end, size int) []int { return numimpl.GenerateBySet(begin, end, size) }
