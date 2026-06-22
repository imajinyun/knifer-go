package vskt_test

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/imajinyun/go-knifer/vskt"
)

func ExampleGetRemoteAddress() {
	fmt.Println(vskt.GetRemoteAddress(nil) == nil)
	// Output: true
}

func ExampleFuncEncoder() {
	encoder := vskt.FuncEncoder[string](func(_ *vskt.AioSession, b *bytes.Buffer, data string) {
		b.WriteString("encoded:" + data)
	})
	var out bytes.Buffer

	encoder.Encode(nil, &out, "payload")
	fmt.Println(out.String())
	// Output: encoded:payload
}

func ExampleNewSocketErrorf() {
	err := vskt.NewSocketErrorf("socket %s", "closed")

	fmt.Println(err.Error())
	// Output: socket closed
}

func ExampleNewSocketErrorMsg() {
	err := vskt.NewSocketErrorMsg("connection reset")

	fmt.Println(err.Error())
	// Output: connection reset
}

func ExampleWrapSocketError() {
	cause := errors.New("dial failed")
	err := vskt.WrapSocketError(cause, "connect")

	fmt.Println(err.Error())
	fmt.Println(errors.Is(err, cause))
	// Output:
	// connect: dial failed
	// true
}

func ExampleFuncDecoder() {
	decoder := vskt.FuncDecoder[string](func(_ *vskt.AioSession, b *bytes.Buffer) (string, bool) {
		return b.String(), true
	})

	value, ok := decoder.Decode(nil, bytes.NewBufferString("payload"))

	fmt.Println(value)
	fmt.Println(ok)
	// Output:
	// payload
	// true
}

func ExampleNewSocketConfig() {
	cfg := vskt.NewSocketConfig()

	fmt.Println(cfg != nil)
	// Output: true
}

func ExampleNewSocketError() {
	err := vskt.NewSocketError(fmt.Errorf("closed"))

	fmt.Println(err.Error())
	// Output: closed: closed
}
