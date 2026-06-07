package socket

import (
	"net"
	"sync"
	"sync/atomic"
)

// NioClient is an event-driven TCP client aligned with the utility NIO NioClient.
type NioClient struct {
	conn    net.Conn
	handler ChannelHandler
	config  *SocketConfig

	closed atomic.Bool
	mu     sync.Mutex
	wg     sync.WaitGroup
}

// NewNioClient creates a client and connects to the specified host and port.
func NewNioClient(host string, port int) (*NioClient, error) {
	return NewNioClientWithOptions(host, port)
}

// NewNioClientWithOptions creates a client and connects to the specified host and port with custom config options.
func NewNioClientWithOptions(host string, port int, opts ...ConfigOption) (*NioClient, error) {
	return NewNioClientAddrWithOptions(&net.TCPAddr{IP: net.ParseIP(host), Port: port}, opts...)
}

// NewNioClientAddr creates a client and connects to the specified address.
func NewNioClientAddr(addr *net.TCPAddr) (*NioClient, error) {
	return NewNioClientAddrWithOptions(addr)
}

// NewNioClientAddrWithOptions creates a client and connects to the specified address with custom config options.
func NewNioClientAddrWithOptions(addr *net.TCPAddr, opts ...ConfigOption) (*NioClient, error) {
	if addr == nil {
		return nil, NewSocketErrorMsg("address must not be nil")
	}
	c := &NioClient{config: NewSocketConfigWithOptions(opts...)}
	if err := c.init(addr); err != nil {
		return nil, err
	}
	return c, nil
}

// init initializes the connection.
func (c *NioClient) init(addr *net.TCPAddr) error {
	connFactory := func(addr *net.TCPAddr) (net.Conn, error) { return net.DialTCP("tcp", nil, addr) }
	if c.config != nil && c.config.ConnFactory != nil {
		connFactory = c.config.ConnFactory
	}
	conn, err := connFactory(addr)
	if err != nil {
		return NewSocketError(err)
	}
	c.conn = conn
	return nil
}

// SetChannelHandler sets the data handler.
func (c *NioClient) SetChannelHandler(handler ChannelHandler) *NioClient {
	c.handler = handler
	return c
}

// Channel returns the underlying net.Conn.
func (c *NioClient) Channel() net.Conn {
	return c.conn
}

// Listen asynchronously listens for data pushed by the server.
func (c *NioClient) Listen() {
	c.wg.Add(1)
	runWithConfig(c.config, func() {
		defer c.wg.Done()
		for {
			if c.closed.Load() {
				return
			}
			if c.handler == nil {
				return
			}
			if err := c.handler.Handle(c.conn); err != nil {
				// Stop listening on errors such as closed connections or read failures.
				return
			}
		}
	})
}

// Write writes multiple data fragments.
func (c *NioClient) Write(datas ...[]byte) (*NioClient, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, d := range datas {
		if len(d) == 0 {
			continue
		}
		if _, err := c.conn.Write(d); err != nil {
			return c, NewSocketError(err)
		}
	}
	return c, nil
}

// Close closes the client.
func (c *NioClient) Close() error {
	if c.closed.Swap(true) {
		return nil
	}
	if c.conn != nil {
		_ = c.conn.Close()
	}
	c.wg.Wait()
	return nil
}
