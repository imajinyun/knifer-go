package vcsv_test

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vcsv"
)

func ExampleReadString() {
	records, _ := vcsv.ReadString("a,b,c\n1,2,3\n")
	fmt.Println(records)
	// Output: [[a b c] [1 2 3]]
}
