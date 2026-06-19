package vmask_test

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vmask"
)

func ExampleMobilePhone() {
	fmt.Println(vmask.MobilePhone("13812345678"))
	// Output: 138****5678
}

func ExampleEmail() {
	fmt.Println(vmask.Email("test@example.com"))
	// Output: t***@example.com
}

func ExampleIPv4() {
	fmt.Println(vmask.IPv4("192.0.2.15"))
	// Output: 192.*.*.*
}

func ExampleChineseName() {
	fmt.Println(vmask.ChineseName("张三丰"))
	// Output: 张**
}

func ExampleBankCard() {
	fmt.Println(vmask.BankCard("1234 5678 9012 3456"))
	// Output: 1234 **** **** 3456
}
