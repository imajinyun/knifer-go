package vnet_test

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vnet"
)

func ExampleCreateAddress() {
	addr := vnet.CreateAddress("127.0.0.1", 8080)
	fmt.Println(addr.String())
	// Output: 127.0.0.1:8080
}
