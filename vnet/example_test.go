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

func ExampleIPv4ToLong() {
	n, err := vnet.IPv4ToLong("192.0.2.1")

	fmt.Println(n)
	fmt.Println(vnet.LongToIPv4(n))
	fmt.Println(err)
	// Output:
	// 3221225985
	// 192.0.2.1
	// <nil>
}

func ExampleBeginIP() {
	begin, beginErr := vnet.BeginIP("192.0.2.9", 24)
	end, endErr := vnet.EndIP("192.0.2.9", 24)
	count, countErr := vnet.CountByMaskBit(24, true)

	fmt.Println(begin, beginErr)
	fmt.Println(end, endErr)
	fmt.Println(count, countErr)
	// Output:
	// 192.0.2.0 <nil>
	// 192.0.2.255 <nil>
	// 256 <nil>
}

func ExampleMaskBitByMask() {
	maskBit, err := vnet.MaskBitByMask("255.255.255.0")
	fmt.Println(maskBit, err)
	// Output: 24 <nil>
}

func ExampleListIPRange() {
	ips, err := vnet.ListIPRange("192.0.2.1", "192.0.2.2")
	fmt.Println(ips)
	fmt.Println(err)
	// Output:
	// [192.0.2.1 192.0.2.2]
	// <nil>
}
