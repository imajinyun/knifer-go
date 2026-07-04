package vskt_test

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/imajinyun/knifer-go/vskt"
)

func TestFacadeSocketFactories(t *testing.T) {
	addr := &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 9999}
	listener := &facadeFakeListener{addr: facadeFakeAddr("facade-aio")}
	aio, err := vskt.NewAioServerAddrWithOptions(addr, nil, vskt.WithListenerFactory(func(got *net.TCPAddr) (net.Listener, error) {
		if got != addr {
			return nil, errors.New("unexpected aio addr")
		}
		return listener, nil
	}))
	if err != nil {
		t.Fatalf("NewAioServerAddrWithOptions: %v", err)
	}
	if aio.Listener() != listener || aio.LocalAddr().String() != "facade-aio" {
		t.Fatalf("aio listener = %#v addr=%v", aio.Listener(), aio.LocalAddr())
	}
	_ = aio.Close()

	listener = &facadeFakeListener{addr: facadeFakeAddr("facade-nio")}
	nio, err := vskt.NewNioServerAddrWithOptions(addr, nil, vskt.WithListenerFactory(func(got *net.TCPAddr) (net.Listener, error) {
		if got != addr {
			return nil, errors.New("unexpected nio addr")
		}
		return listener, nil
	}))
	if err != nil {
		t.Fatalf("NewNioServerAddrWithOptions: %v", err)
	}
	if nio.Listener() != listener || nio.LocalAddr().String() != "facade-nio" {
		t.Fatalf("nio listener = %#v addr=%v", nio.Listener(), nio.LocalAddr())
	}
	_ = nio.Close()

	client, server := net.Pipe()
	defer func() { _ = server.Close() }()
	nioClient, err := vskt.NewNioClientAddrWithOptions(addr, vskt.WithConnFactory(func(got *net.TCPAddr) (net.Conn, error) {
		if got != addr {
			return nil, errors.New("unexpected client addr")
		}
		return client, nil
	}))
	if err != nil {
		t.Fatalf("NewNioClientAddrWithOptions: %v", err)
	}
	if nioClient.Channel() != client {
		t.Fatalf("nio client channel = %#v", nioClient.Channel())
	}
	_ = nioClient.Close()
}

func TestFacadeGeneratedSocketHelpers(t *testing.T) {
	client, server := net.Pipe()
	accepted := false
	action := &vskt.SimpleIoAction{OnAccept: func(session *vskt.AioSession) {
		accepted = session != nil
	}}
	aioClient := vskt.NewAioClientWithConn(client, action, vskt.NewSocketConfig())
	defer func() { _ = aioClient.Close() }()
	defer func() { _ = server.Close() }()
	if aioClient.Session() == nil || !accepted {
		t.Fatalf("NewAioClientWithConn session=%#v accepted=%v", aioClient.Session(), accepted)
	}

	addr := &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 7777}
	dialer := &facadeFakeDialer{}
	conn, err := vskt.ChannelUtilDial(nil, 0)
	if err == nil {
		_ = conn.Close()
		t.Fatal("ChannelUtilDial with nil address should fail")
	}
	conn, err = vskt.ChannelDialWithOptions(addr, vskt.WithConnectDialer(dialer))
	if err != nil {
		t.Fatalf("ChannelDialWithOptions with fake dialer failed: %v", err)
	}
	_ = conn.Close()
	_ = dialer.server.Close()

	remoteClient, remoteServer := net.Pipe()
	defer func() { _ = remoteClient.Close() }()
	defer func() { _ = remoteServer.Close() }()
	if got := vskt.GetRemoteAddress(remoteClient); got == nil {
		t.Fatal("GetRemoteAddress should return the remote endpoint for net.Pipe")
	}
	if got := vskt.SocketRemoteAddress(remoteClient); got == nil {
		t.Fatal("SocketRemoteAddress should return the remote endpoint for net.Pipe")
	}
}

func TestFacadePortServerConstructorsUseListenerFactory(t *testing.T) {
	plainNio, err := vskt.NewNioServer(0)
	if err != nil {
		t.Fatalf("NewNioServer: %v", err)
	}
	_ = plainNio.Close()

	nioListener := &facadeFakeListener{addr: facadeFakeAddr("nio-port")}
	nio, err := vskt.NewNioServerWithOptions(0, vskt.WithListenerFactory(func(addr *net.TCPAddr) (net.Listener, error) {
		if addr == nil || addr.Port != 0 {
			return nil, errors.New("unexpected nio port address")
		}
		return nioListener, nil
	}))
	if err != nil {
		t.Fatalf("NewNioServerWithOptions: %v", err)
	}
	if nio.Listener() != nioListener || nio.LocalAddr().String() != "nio-port" {
		t.Fatalf("nio listener = %#v addr=%v", nio.Listener(), nio.LocalAddr())
	}
	_ = nio.Close()

	nioConfigListener := &facadeFakeListener{addr: facadeFakeAddr("nio-config")}
	nioWithConfig, err := vskt.NewNioServerWithConfig(0, vskt.NewSocketConfigWithOptions(vskt.WithListenerFactory(func(addr *net.TCPAddr) (net.Listener, error) {
		if addr == nil || addr.Port != 0 {
			return nil, errors.New("unexpected nio config address")
		}
		return nioConfigListener, nil
	})))
	if err != nil {
		t.Fatalf("NewNioServerWithConfig: %v", err)
	}
	if nioWithConfig.Listener() != nioConfigListener {
		t.Fatalf("nio config listener = %#v", nioWithConfig.Listener())
	}
	_ = nioWithConfig.Close()

	nioAddr, err := vskt.NewNioServerAddr(&net.TCPAddr{Port: 0})
	if err != nil {
		t.Fatalf("NewNioServerAddr: %v", err)
	}
	_ = nioAddr.Close()

	nioAddrConfigListener := &facadeFakeListener{addr: facadeFakeAddr("nio-addr-config")}
	nioAddrConfig, err := vskt.NewNioServerAddrWithConfig(&net.TCPAddr{Port: 0}, vskt.NewSocketConfigWithOptions(vskt.WithListenerFactory(func(addr *net.TCPAddr) (net.Listener, error) {
		return nioAddrConfigListener, nil
	})))
	if err != nil {
		t.Fatalf("NewNioServerAddrWithConfig: %v", err)
	}
	if nioAddrConfig.Listener() != nioAddrConfigListener {
		t.Fatalf("nio addr config listener = %#v", nioAddrConfig.Listener())
	}
	_ = nioAddrConfig.Close()

	plainAio, err := vskt.NewAioServer(0)
	if err != nil {
		t.Fatalf("NewAioServer: %v", err)
	}
	_ = plainAio.Close()

	aioListener := &facadeFakeListener{addr: facadeFakeAddr("aio-port")}
	aio, err := vskt.NewAioServerWithOptions(0, vskt.WithListenerFactory(func(addr *net.TCPAddr) (net.Listener, error) {
		if addr == nil || addr.Port != 0 {
			return nil, errors.New("unexpected aio port address")
		}
		return aioListener, nil
	}))
	if err != nil {
		t.Fatalf("NewAioServerWithOptions: %v", err)
	}
	if aio.Listener() != aioListener || aio.LocalAddr().String() != "aio-port" {
		t.Fatalf("aio listener = %#v addr=%v", aio.Listener(), aio.LocalAddr())
	}
	_ = aio.Close()

	aioAddrListener := &facadeFakeListener{addr: facadeFakeAddr("aio-addr")}
	aioAddr, err := vskt.NewAioServerAddr(&net.TCPAddr{Port: 0}, vskt.NewSocketConfigWithOptions(vskt.WithListenerFactory(func(addr *net.TCPAddr) (net.Listener, error) {
		return aioAddrListener, nil
	})))
	if err != nil {
		t.Fatalf("NewAioServerAddr: %v", err)
	}
	if aioAddr.Listener() != aioAddrListener {
		t.Fatalf("aio addr listener = %#v", aioAddr.Listener())
	}
	_ = aioAddr.Close()
}

func TestFacadeServerContextStartMethodsReturnOnCancel(t *testing.T) {
	nio, err := vskt.NewNioServerWithOptions(0, vskt.WithListenerFactory(func(*net.TCPAddr) (net.Listener, error) {
		return &facadeFakeListener{addr: facadeFakeAddr("nio-context")}, nil
	}))
	if err != nil {
		t.Fatalf("NewNioServerWithOptions: %v", err)
	}
	nioCtx, cancelNio := context.WithCancel(context.Background())
	nioDone := make(chan struct{})
	go func() {
		defer close(nioDone)
		nio.ListenContext(nioCtx)
	}()
	cancelNio()
	select {
	case <-nioDone:
	case <-time.After(time.Second):
		t.Fatal("NioServer.ListenContext did not return after cancel")
	}

	aio, err := vskt.NewAioServerWithOptions(0, vskt.WithListenerFactory(func(*net.TCPAddr) (net.Listener, error) {
		return &facadeFakeListener{addr: facadeFakeAddr("aio-context")}, nil
	}))
	if err != nil {
		t.Fatalf("NewAioServerWithOptions: %v", err)
	}
	aioCtx, cancelAio := context.WithCancel(context.Background())
	aioDone := make(chan struct{})
	go func() {
		defer close(aioDone)
		aio.StartContext(aioCtx, true)
	}()
	cancelAio()
	select {
	case <-aioDone:
	case <-time.After(time.Second):
		t.Fatal("AioServer.StartContext did not return after cancel")
	}
}

func TestFacadeClientConstructorsWithProviders(t *testing.T) {
	dialed := false
	client, server := net.Pipe()
	defer func() { _ = server.Close() }()
	nioClient, err := vskt.NewNioClientWithOptions("service.local", 8080,
		vskt.WithSocketIPParser(func(host string) net.IP {
			if host != "service.local" {
				t.Fatalf("host = %q, want service.local", host)
			}
			return net.IPv4(127, 0, 0, 1)
		}),
		vskt.WithConnFactory(func(addr *net.TCPAddr) (net.Conn, error) {
			dialed = true
			if addr.Port != 8080 {
				return nil, errors.New("unexpected nio client port")
			}
			return client, nil
		}),
	)
	if err != nil {
		t.Fatalf("NewNioClientWithOptions: %v", err)
	}
	if !dialed || nioClient.Channel() != client {
		t.Fatalf("NewNioClientWithOptions dialed=%v channel=%#v", dialed, nioClient.Channel())
	}
	_ = nioClient.Close()

	if _, err := vskt.NewAioClient(nil, &vskt.SimpleIoAction{}); err == nil {
		t.Fatal("NewAioClient with nil address should fail")
	}
	if _, err := vskt.NewAioClientWithConfig(nil, &vskt.SimpleIoAction{}, vskt.NewSocketConfig()); err == nil {
		t.Fatal("NewAioClientWithConfig with nil address should fail")
	}

	sessionClient, sessionServer := net.Pipe()
	defer func() { _ = sessionServer.Close() }()
	session := vskt.NewAioSession(sessionClient, &vskt.SimpleIoAction{}, vskt.NewSocketConfig())
	if session == nil {
		t.Fatal("NewAioSession returned nil")
	}
	_ = session.Close()
}
