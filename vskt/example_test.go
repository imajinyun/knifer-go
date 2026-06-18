package vskt_test

import (
	"fmt"
	"net"

	"github.com/imajinyun/go-knifer/vskt"
)

func ExampleGetRemoteAddress() {
	conn, _ := net.Dial("tcp", "example.com:80")
	addr := vskt.GetRemoteAddress(conn)
	fmt.Println(addr != nil)
	// Output: true
}
