package socket

import (
	"context"
	"net"
	"strconv"
	"time"

	netimpl "github.com/imajinyun/knifer-go/internal/net"
)

type connectConfig struct {
	ctx     context.Context
	timeout time.Duration
	network string
	dialer  netimpl.Dialer
}

// ConnectOption customizes socket connection helpers.
type ConnectOption func(*connectConfig)

// WithConnectContext sets the context used while dialing.
func WithConnectContext(ctx context.Context) ConnectOption {
	return func(c *connectConfig) { c.ctx = ctx }
}

// WithConnectTimeout sets the dial timeout.
func WithConnectTimeout(timeout time.Duration) ConnectOption {
	return func(c *connectConfig) { c.timeout = timeout }
}

// WithConnectNetwork sets the network used for dialing, such as tcp, tcp4, or tcp6.
func WithConnectNetwork(network string) ConnectOption {
	return func(c *connectConfig) { c.network = network }
}

// WithConnectDialer sets the dialer used by connection helpers.
func WithConnectDialer(dialer netimpl.Dialer) ConnectOption {
	return func(c *connectConfig) {
		if dialer != nil {
			c.dialer = dialer
		}
	}
}

func applyConnectOptions(timeout time.Duration, opts []ConnectOption) connectConfig {
	cfg := connectConfig{ctx: context.Background(), timeout: timeout, network: "tcp"}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if cfg.ctx == nil {
		cfg.ctx = context.Background()
	}
	if cfg.network == "" {
		cfg.network = "tcp"
	}
	if cfg.dialer == nil {
		cfg.dialer = &net.Dialer{Timeout: cfg.timeout}
	}
	return cfg
}

// Connect creates a socket and connects to the specified address.
// When timeout <= 0, the default connection behavior without timeout is used.
func Connect(hostname string, port int, timeout time.Duration) (net.Conn, error) {
	return ConnectWithOptions(hostname, port, WithConnectTimeout(timeout))
}

// ConnectWithOptions creates a socket connection with custom dial options.
func ConnectWithOptions(hostname string, port int, opts ...ConnectOption) (net.Conn, error) {
	cfg := applyConnectOptions(0, opts)
	ctx := cfg.ctx
	cancel := func() {}
	if cfg.timeout > 0 {
		ctx, cancel = context.WithTimeout(cfg.ctx, cfg.timeout)
	}
	defer cancel()
	conn, err := cfg.dialer.DialContext(ctx, cfg.network, net.JoinHostPort(hostname, strconvPort(port)))
	if err != nil {
		return nil, NewSocketError(err)
	}
	return conn, nil
}

// ConnectAddr creates a connection from net.TCPAddr.
func ConnectAddr(addr *net.TCPAddr, timeout time.Duration) (net.Conn, error) {
	return ConnectAddrWithOptions(addr, WithConnectTimeout(timeout))
}

// ConnectAddrWithOptions creates a connection from net.TCPAddr with custom dial options.
func ConnectAddrWithOptions(addr *net.TCPAddr, opts ...ConnectOption) (net.Conn, error) {
	if addr == nil {
		return nil, NewSocketErrorMsg("address must not be nil")
	}
	return ConnectWithOptions(addr.IP.String(), addr.Port, opts...)
}

// GetRemoteAddress returns the remote address, or nil when conn is nil or disconnected.
func GetRemoteAddress(conn net.Conn) net.Addr {
	if !netimpl.IsConnected(conn) {
		return nil
	}
	return conn.RemoteAddr()
}

// IsConnected reports whether the connection is established and has a remote address.
func IsConnected(conn net.Conn) bool {
	return netimpl.IsConnected(conn)
}

func strconvPort(port int) string { return strconv.Itoa(port) }
