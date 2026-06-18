package vlog_test

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vlog"
)

func ExampleGetDefault() {
	log := vlog.GetDefault()
	fmt.Println(log != nil)
	// Output: true
}
