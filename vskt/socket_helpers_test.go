package vskt_test

import (
	"context"
	"net"
	"sync/atomic"
)

type facadeFakeDialer struct {
	calls   atomic.Int32
	network string
	address string
	server  net.Conn
}

type facadeFakeAddr string

func (a facadeFakeAddr) Network() string { return "tcp" }
func (a facadeFakeAddr) String() string  { return string(a) }

type facadeFakeListener struct {
	addr net.Addr
}

func (l *facadeFakeListener) Accept() (net.Conn, error) { return nil, net.ErrClosed }
func (l *facadeFakeListener) Close() error              { return nil }
func (l *facadeFakeListener) Addr() net.Addr            { return l.addr }

func (d *facadeFakeDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	d.calls.Add(1)
	d.network = network
	d.address = address
	client, server := net.Pipe()
	d.server = server
	return client, nil
}
