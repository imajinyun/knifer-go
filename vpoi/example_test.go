package vpoi_test

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vpoi"
)

func ExampleSheetNames() {
	names, err := vpoi.SheetNames("nonexistent.xlsx")
	fmt.Println(names)
	fmt.Println(err != nil)
	// Output:
	// []
	// true
}
