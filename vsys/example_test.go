package vsys_test

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vsys"
)

func ExampleGetCurrentPID() {
	pid := vsys.GetCurrentPID()
	fmt.Println(pid > 0)
	// Output: true
}
