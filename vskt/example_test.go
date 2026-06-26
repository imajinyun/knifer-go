package vskt_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/imajinyun/knifer-go/vskt"
)

type exampleDialer struct {
	err error
}

func (d exampleDialer) DialContext(context.Context, string, string) (net.Conn, error) {
	return nil, d.err
}

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

func ExampleNewSocketConfigWithOptions() {
	cfg := vskt.NewSocketConfigWithOptions(
		vskt.WithThreadPoolSize(2),
		vskt.WithReadTimeout(100),
		vskt.WithWriteBufferSize(4096),
	)

	fmt.Println(cfg.ThreadPoolSize)
	fmt.Println(cfg.ReadTimeout)
	fmt.Println(cfg.WriteBufferSize)
	// Output:
	// 2
	// 100
	// 4096
}

func ExampleSocketConfig_SetReadBufferSize() {
	cfg := vskt.NewSocketConfig().SetReadBufferSize(2048).SetWriteBufferSize(4096)

	fmt.Println(cfg.ReadBufferSize)
	fmt.Println(cfg.WriteBufferSize)
	// Output:
	// 2048
	// 4096
}

func ExampleSocketConnectWithOptions() {
	cause := errors.New("offline")
	conn, err := vskt.SocketConnectWithOptions(
		"example.invalid",
		80,
		vskt.WithConnectNetwork("tcp4"),
		vskt.WithConnectDialer(exampleDialer{err: cause}),
	)

	fmt.Println(conn == nil)
	fmt.Println(errors.Is(err, cause))
	// Output:
	// true
	// true
}

func ExampleSocketConnectAddrWithOptions() {
	cause := errors.New("offline")
	addr := &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 80}
	conn, err := vskt.SocketConnectAddrWithOptions(addr, vskt.WithConnectDialer(exampleDialer{err: cause}))

	fmt.Println(conn == nil)
	fmt.Println(errors.Is(err, cause))
	// Output:
	// true
	// true
}

func ExampleSocketRemoteAddress() {
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()

	fmt.Println(vskt.SocketRemoteAddress(client) != nil)
	fmt.Println(vskt.SocketIsConnected(client))
	// Output:
	// true
	// true
}

func ExampleSimpleIoAction_DoAction() {
	action := &vskt.SimpleIoAction{
		OnDoAction: func(_ *vskt.AioSession, data *bytes.Buffer) {
			fmt.Println(data.String())
		},
	}

	action.DoAction(nil, bytes.NewBufferString("payload"))
	// Output: payload
}

func ExampleNewAioSession() {
	cfg := vskt.NewSocketConfigWithOptions(
		vskt.WithReadBufferSize(128),
		vskt.WithWriteBufferSize(256),
		vskt.WithClock(func() time.Time { return time.Unix(0, 0) }),
	)
	session := vskt.NewAioSession(nil, &vskt.SimpleIoAction{}, cfg)

	fmt.Println(session.ReadBuffer().Cap())
	fmt.Println(session.WriteBuffer().Cap())
	fmt.Println(session.IsOpen())
	// Output:
	// 128
	// 256
	// false
}
