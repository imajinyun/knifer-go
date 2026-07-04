package net

import (
	"context"
	"fmt"
	stdnet "net"
	"net/http"
	"os"
	"slices"
	"strings"
	"time"

	"golang.org/x/net/idna"
)

const (
	// PortRangeMin is the default minimum dynamic helper port.
	PortRangeMin = 1024
	// PortRangeMax is the maximum TCP/UDP port.
	PortRangeMax = 0xFFFF
)

// Dialer is the minimal interface used by PingWithOptions to open network connections.
type Dialer interface {
	DialContext(ctx context.Context, network, address string) (stdnet.Conn, error)
}

type pingConfig struct {
	ctx     context.Context
	timeout time.Duration
	ports   []int
	network string
	dialer  Dialer
}

type connectConfig struct {
	ctx     context.Context
	timeout time.Duration
	network string
	dialer  Dialer
}

// PingOption customizes PingWithOptions.
type PingOption func(*pingConfig)

// ConnectOption customizes ConnectWithOptions, NetCatWithOptions, and IsOpenWithOptions.
type ConnectOption func(*connectConfig)

// WithPingContext sets the context used by PingWithOptions.
func WithPingContext(ctx context.Context) PingOption { return func(c *pingConfig) { c.ctx = ctx } }

// WithPingTimeout sets the timeout for each connection attempt made by PingWithOptions.
func WithPingTimeout(timeout time.Duration) PingOption {
	return func(c *pingConfig) { c.timeout = timeout }
}

// WithPingPorts sets the destination ports PingWithOptions probes.
func WithPingPorts(ports ...int) PingOption {
	return func(c *pingConfig) { c.ports = slices.Clone(ports) }
}

// WithPingNetwork sets the network used by PingWithOptions, such as tcp, tcp4, or tcp6.
func WithPingNetwork(network string) PingOption { return func(c *pingConfig) { c.network = network } }

// WithPingDialer sets the dialer used by PingWithOptions.
func WithPingDialer(d Dialer) PingOption {
	return func(c *pingConfig) {
		if d != nil {
			c.dialer = d
		}
	}
}

// WithConnectContext sets the context used by connection helpers.
func WithConnectContext(ctx context.Context) ConnectOption {
	return func(c *connectConfig) { c.ctx = ctx }
}

// WithConnectTimeout bounds connection attempts made by connection helpers.
func WithConnectTimeout(timeout time.Duration) ConnectOption {
	return func(c *connectConfig) { c.timeout = timeout }
}

// WithConnectNetwork sets the network used by connection helpers, such as tcp, tcp4, or tcp6.
func WithConnectNetwork(network string) ConnectOption {
	return func(c *connectConfig) { c.network = network }
}

// WithConnectDialer sets the dialer used by connection helpers.
func WithConnectDialer(d Dialer) ConnectOption {
	return func(c *connectConfig) {
		if d != nil {
			c.dialer = d
		}
	}
}

func applyPingOptions(opts []PingOption) pingConfig {
	cfg := pingConfig{
		ctx:     context.Background(),
		timeout: 3 * time.Second,
		ports:   []int{80, 443},
		network: "tcp",
	}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if cfg.ctx == nil {
		cfg.ctx = context.Background()
	}
	if cfg.timeout <= 0 {
		cfg.timeout = 3 * time.Second
	}
	if len(cfg.ports) == 0 {
		cfg.ports = []int{80, 443}
	}
	cfg.network = strings.TrimSpace(cfg.network)
	if cfg.network == "" {
		cfg.network = "tcp"
	}
	if cfg.dialer == nil {
		cfg.dialer = &stdnet.Dialer{Timeout: cfg.timeout}
	}
	return cfg
}

func applyConnectOptions(opts []ConnectOption) (connectConfig, context.CancelFunc) {
	cfg := connectConfig{ctx: context.Background(), network: "tcp"}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if cfg.ctx == nil {
		cfg.ctx = context.Background()
	}
	cancel := func() {}
	if cfg.timeout > 0 {
		cfg.ctx, cancel = context.WithTimeout(cfg.ctx, cfg.timeout)
	}
	cfg.network = strings.TrimSpace(cfg.network)
	if cfg.network == "" {
		cfg.network = "tcp"
	}
	if cfg.dialer == nil {
		cfg.dialer = &stdnet.Dialer{}
	}
	return cfg, cancel
}

type resolveConfig struct {
	ctx       context.Context
	timeout   time.Duration
	network   string
	resolver  *stdnet.Resolver
	attrNames []string
}

type addressConfig struct {
	network  string
	resolver func(network, address string) (*stdnet.TCPAddr, error)
}

// ResolveOption customizes DNS and host resolution helpers.
type ResolveOption func(*resolveConfig)

// AddressOption customizes TCP address construction helpers.
type AddressOption func(*addressConfig)

// WithResolveContext sets the context used by DNS lookups.
func WithResolveContext(ctx context.Context) ResolveOption {
	return func(c *resolveConfig) { c.ctx = ctx }
}

// WithResolveTimeout bounds DNS lookups with a timeout.
func WithResolveTimeout(timeout time.Duration) ResolveOption {
	return func(c *resolveConfig) { c.timeout = timeout }
}

// WithResolveNetwork sets the IP lookup network, such as ip, ip4, or ip6.
func WithResolveNetwork(network string) ResolveOption {
	return func(c *resolveConfig) { c.network = network }
}

// WithResolver sets the resolver used by DNS lookups.
func WithResolver(resolver *stdnet.Resolver) ResolveOption {
	return func(c *resolveConfig) { c.resolver = resolver }
}

// WithDNSTypes sets the DNS record types looked up by GetDNSInfoWithOptions.
func WithDNSTypes(attrNames ...string) ResolveOption {
	return func(c *resolveConfig) { c.attrNames = slices.Clone(attrNames) }
}

// WithAddressNetwork sets the network used by TCP address construction helpers.
func WithAddressNetwork(network string) AddressOption {
	return func(c *addressConfig) { c.network = network }
}

// WithTCPAddrResolver sets the resolver used by TCP address construction helpers.
func WithTCPAddrResolver(resolver func(network, address string) (*stdnet.TCPAddr, error)) AddressOption {
	return func(c *addressConfig) {
		if resolver != nil {
			c.resolver = resolver
		}
	}
}

func applyResolveOptions(opts []ResolveOption) (resolveConfig, context.CancelFunc) {
	cfg := resolveConfig{ctx: context.Background(), network: "ip", resolver: stdnet.DefaultResolver}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if cfg.ctx == nil {
		cfg.ctx = context.Background()
	}
	cancel := func() {}
	if cfg.timeout > 0 {
		cfg.ctx, cancel = context.WithTimeout(cfg.ctx, cfg.timeout)
	}
	cfg.network = strings.TrimSpace(cfg.network)
	if cfg.network == "" {
		cfg.network = "ip"
	}
	if cfg.resolver == nil {
		cfg.resolver = stdnet.DefaultResolver
	}
	return cfg, cancel
}

func applyAddressOptions(opts []AddressOption) addressConfig {
	cfg := addressConfig{network: "tcp", resolver: stdnet.ResolveTCPAddr}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	cfg.network = strings.TrimSpace(cfg.network)
	if cfg.network == "" {
		cfg.network = "tcp"
	}
	if cfg.resolver == nil {
		cfg.resolver = stdnet.ResolveTCPAddr
	}
	return cfg
}

type portConfig struct {
	network         string
	host            string
	listenerFactory func(network, address string) (stdnet.Listener, error)
}

// PortOption customizes local port probing helpers.
type PortOption func(*portConfig)

// WithPortNetwork sets the network used by local port probes, such as tcp, tcp4, or tcp6.
func WithPortNetwork(network string) PortOption { return func(c *portConfig) { c.network = network } }

// WithPortHost sets the local host/address used by local port probes.
func WithPortHost(host string) PortOption { return func(c *portConfig) { c.host = host } }

// WithPortListenerFactory sets the listener factory used to probe local ports.
func WithPortListenerFactory(factory func(network, address string) (stdnet.Listener, error)) PortOption {
	return func(c *portConfig) {
		if factory != nil {
			c.listenerFactory = factory
		}
	}
}

func applyPortOptions(opts []PortOption) portConfig {
	cfg := portConfig{network: "tcp", host: "127.0.0.1", listenerFactory: stdnet.Listen}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	cfg.network = strings.TrimSpace(cfg.network)
	if cfg.network == "" {
		cfg.network = "tcp"
	}
	cfg.host = strings.TrimSpace(cfg.host)
	if cfg.host == "" {
		cfg.host = "127.0.0.1"
	}
	if cfg.listenerFactory == nil {
		cfg.listenerFactory = stdnet.Listen
	}
	return cfg
}

type interfaceConfig struct {
	interfaceByName func(string) (*stdnet.Interface, error)
	interfaces      func() ([]stdnet.Interface, error)
	interfaceAddrs  func(stdnet.Interface) ([]stdnet.Addr, error)
	lookupAddr      func(string) ([]string, error)
	hostname        func() (string, error)
}

// InterfaceOption customizes local network interface helpers.
type InterfaceOption func(*interfaceConfig)

// WithInterfaceByNameFunc sets the function used to find a network interface by name.
func WithInterfaceByNameFunc(fn func(string) (*stdnet.Interface, error)) InterfaceOption {
	return func(c *interfaceConfig) {
		if fn != nil {
			c.interfaceByName = fn
		}
	}
}

// WithInterfacesFunc sets the function used to list local network interfaces.
func WithInterfacesFunc(fn func() ([]stdnet.Interface, error)) InterfaceOption {
	return func(c *interfaceConfig) {
		if fn != nil {
			c.interfaces = fn
		}
	}
}

// WithInterfaceAddrsFunc sets the function used to collect addresses for an interface.
func WithInterfaceAddrsFunc(fn func(stdnet.Interface) ([]stdnet.Addr, error)) InterfaceOption {
	return func(c *interfaceConfig) {
		if fn != nil {
			c.interfaceAddrs = fn
		}
	}
}

// WithReverseLookupFunc sets the function used for reverse host lookup.
func WithReverseLookupFunc(fn func(string) ([]string, error)) InterfaceOption {
	return func(c *interfaceConfig) {
		if fn != nil {
			c.lookupAddr = fn
		}
	}
}

// WithNetHostnameFunc sets the function used to collect the local hostname fallback.
func WithNetHostnameFunc(fn func() (string, error)) InterfaceOption {
	return func(c *interfaceConfig) {
		if fn != nil {
			c.hostname = fn
		}
	}
}

func applyInterfaceOptions(opts []InterfaceOption) interfaceConfig {
	cfg := interfaceConfig{
		interfaceByName: stdnet.InterfaceByName,
		interfaces:      stdnet.Interfaces,
		interfaceAddrs:  func(iface stdnet.Interface) ([]stdnet.Addr, error) { return iface.Addrs() },
		lookupAddr:      stdnet.LookupAddr,
		hostname:        osHostname,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if cfg.interfaceByName == nil {
		cfg.interfaceByName = stdnet.InterfaceByName
	}
	if cfg.interfaces == nil {
		cfg.interfaces = stdnet.Interfaces
	}
	if cfg.interfaceAddrs == nil {
		cfg.interfaceAddrs = func(iface stdnet.Interface) ([]stdnet.Addr, error) { return iface.Addrs() }
	}
	if cfg.lookupAddr == nil {
		cfg.lookupAddr = stdnet.LookupAddr
	}
	if cfg.hostname == nil {
		cfg.hostname = osHostname
	}
	return cfg
}

// IsValidPort reports whether port is a valid TCP/UDP port number.
func IsValidPort(port int) bool { return port >= 0 && port <= PortRangeMax }

// IsUsableLocalPort reports whether port can be bound locally on TCP.
func IsUsableLocalPort(port int) bool {
	return IsUsableLocalPortWithOptions(port)
}

// IsUsableLocalPortWithOptions reports whether port can be bound locally with custom probe options.
func IsUsableLocalPortWithOptions(port int, opts ...PortOption) bool {
	if !IsValidPort(port) || port == 0 {
		return false
	}
	cfg := applyPortOptions(opts)
	ln, err := cfg.listenerFactory(cfg.network, stdnet.JoinHostPort(cfg.host, strconvPort(port)))
	if err != nil {
		return false
	}
	_ = ln.Close()
	return true
}

// GetUsableLocalPort returns an available local TCP port in the default range.
func GetUsableLocalPort() (int, error) { return GetUsableLocalPortInRange(PortRangeMin, PortRangeMax) }

// GetUsableLocalPortWithOptions returns an available local port in the default range with custom probe options.
func GetUsableLocalPortWithOptions(opts ...PortOption) (int, error) {
	return GetUsableLocalPortInRangeWithOptions(PortRangeMin, PortRangeMax, opts...)
}

// GetUsableLocalPortFrom returns an available local TCP port from minPort to max.
func GetUsableLocalPortFrom(minPort int) (int, error) {
	return GetUsableLocalPortInRange(minPort, PortRangeMax)
}

// GetUsableLocalPortFromWithOptions returns an available local port from minPort to max with custom probe options.
func GetUsableLocalPortFromWithOptions(minPort int, opts ...PortOption) (int, error) {
	return GetUsableLocalPortInRangeWithOptions(minPort, PortRangeMax, opts...)
}

// GetUsableLocalPortInRange returns an available local TCP port in [minPort, maxPort].
func GetUsableLocalPortInRange(minPort, maxPort int) (int, error) {
	return GetUsableLocalPortInRangeWithOptions(minPort, maxPort)
}

// GetUsableLocalPortInRangeWithOptions returns an available local port in [minPort, maxPort] with custom probe options.
func GetUsableLocalPortInRangeWithOptions(minPort, maxPort int, opts ...PortOption) (int, error) {
	if minPort < 0 || maxPort > PortRangeMax || minPort > maxPort {
		return 0, fmt.Errorf("invalid port range: %d-%d", minPort, maxPort)
	}
	for port := minPort; port <= maxPort; port++ {
		if IsUsableLocalPortWithOptions(port, opts...) {
			return port, nil
		}
	}
	return 0, fmt.Errorf("no usable local port in range %d-%d", minPort, maxPort)
}

// GetUsableLocalPorts returns up to numRequested available ports in [minPort, maxPort].
func GetUsableLocalPorts(numRequested, minPort, maxPort int) ([]int, error) {
	return GetUsableLocalPortsWithOptions(numRequested, minPort, maxPort)
}

// GetUsableLocalPortsWithOptions returns up to numRequested available ports in [minPort, maxPort] with custom probe options.
func GetUsableLocalPortsWithOptions(numRequested, minPort, maxPort int, opts ...PortOption) ([]int, error) {
	if numRequested <= 0 {
		return nil, nil
	}
	ports := make([]int, 0, numRequested)
	for port := minPort; port <= maxPort && len(ports) < numRequested; port++ {
		if IsUsableLocalPortWithOptions(port, opts...) {
			ports = append(ports, port)
		}
	}
	if len(ports) < numRequested {
		return ports, fmt.Errorf("only found %d usable local ports", len(ports))
	}
	return ports, nil
}

// LocalPortGenerator generates available local ports from a moving cursor.
type LocalPortGenerator struct {
	next int
	opts []PortOption
}

// NewLocalPortGenerator creates a local port generator.
func NewLocalPortGenerator(beginPort int) *LocalPortGenerator {
	return NewLocalPortGeneratorWithOptions(beginPort)
}

// NewLocalPortGeneratorWithOptions creates a local port generator with custom probe options.
func NewLocalPortGeneratorWithOptions(beginPort int, opts ...PortOption) *LocalPortGenerator {
	return &LocalPortGenerator{next: beginPort, opts: slices.Clone(opts)}
}

// Gen returns the next available local port.
func (g *LocalPortGenerator) Gen() (int, error) {
	return g.GenWithOptions()
}

// GenWithOptions returns the next available local port with per-call probe options.
func (g *LocalPortGenerator) GenWithOptions(opts ...PortOption) (int, error) {
	if g == nil {
		return 0, fmt.Errorf("nil local port generator")
	}
	allOpts := append(slices.Clone(g.opts), opts...)
	port, err := GetUsableLocalPortInRangeWithOptions(g.next, PortRangeMax, allOpts...)
	if err != nil {
		return 0, err
	}
	g.next = port + 1
	return port, nil
}

// HideIPPart hides the last IPv4 segment.
func HideIPPart(ip string) string {
	idx := strings.LastIndex(ip, ".")
	if idx < 0 {
		return ip
	}
	return ip[:idx+1] + "*"
}

// HideIPPartLong hides the last segment of an IPv4 integer.
func HideIPPartLong(ip uint32) string { return HideIPPart(LongToIPv4(ip)) }

// BuildInetSocketAddress builds a TCP address with a default port when host contains none.
func BuildInetSocketAddress(host string, defaultPort int) (*stdnet.TCPAddr, error) {
	return BuildInetSocketAddressWithOptions(host, defaultPort)
}

// BuildInetSocketAddressWithOptions builds a TCP address with custom address resolution options.
func BuildInetSocketAddressWithOptions(host string, defaultPort int, opts ...AddressOption) (*stdnet.TCPAddr, error) {
	cfg := applyAddressOptions(opts)
	if _, _, err := stdnet.SplitHostPort(host); err == nil {
		return cfg.resolver(cfg.network, host)
	}
	return cfg.resolver(cfg.network, stdnet.JoinHostPort(host, strconvPort(defaultPort)))
}

// CreateAddress builds a TCP address from host and port.
func CreateAddress(host string, port int) *stdnet.TCPAddr {
	addr, _ := CreateAddressWithOptions(host, port)
	return addr
}

// CreateAddressWithOptions builds a TCP address from host and port with custom address resolution options.
func CreateAddressWithOptions(host string, port int, opts ...AddressOption) (*stdnet.TCPAddr, error) {
	cfg := applyAddressOptions(opts)
	return cfg.resolver(cfg.network, stdnet.JoinHostPort(host, strconvPort(port)))
}

// GetIPByHost resolves hostName to the first IP string.
func GetIPByHost(hostName string) string {
	ips, err := GetIPByHostWithOptions(hostName)
	if err != nil || len(ips) == 0 {
		return hostName
	}
	return ips[0]
}

// GetIPByHostWithOptions resolves hostName to IP strings with custom resolver options.
func GetIPByHostWithOptions(hostName string, opts ...ResolveOption) ([]string, error) {
	cfg, cancel := applyResolveOptions(opts)
	defer cancel()
	ips, err := cfg.resolver.LookupIP(cfg.ctx, cfg.network, hostName)
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(ips))
	for _, ip := range ips {
		out = append(out, ip.String())
	}
	return out, nil
}

// GetNetworkInterface returns a network interface by name.
func GetNetworkInterface(name string) (*stdnet.Interface, error) {
	return GetNetworkInterfaceWithOptions(name)
}

// GetNetworkInterfaceWithOptions returns a network interface by name using custom providers.
func GetNetworkInterfaceWithOptions(name string, opts ...InterfaceOption) (*stdnet.Interface, error) {
	return applyInterfaceOptions(opts).interfaceByName(name)
}

// GetNetworkInterfaces returns all network interfaces.
func GetNetworkInterfaces() ([]stdnet.Interface, error) { return GetNetworkInterfacesWithOptions() }

// GetNetworkInterfacesWithOptions returns all network interfaces using custom providers.
func GetNetworkInterfacesWithOptions(opts ...InterfaceOption) ([]stdnet.Interface, error) {
	return applyInterfaceOptions(opts).interfaces()
}

// LocalIPv4s returns local IPv4 addresses.
func LocalIPv4s() []string { return LocalIPv4sWithOptions() }

// LocalIPv4sWithOptions returns local IPv4 addresses using custom providers.
func LocalIPv4sWithOptions(opts ...InterfaceOption) []string {
	return ToIPList(LocalAddressListWithOptions(nil, opts...))
}

// LocalIPv6s returns local IPv6 addresses.
func LocalIPv6s() []string {
	return LocalIPv6sWithOptions()
}

// LocalIPv6sWithOptions returns local IPv6 addresses using custom providers.
func LocalIPv6sWithOptions(opts ...InterfaceOption) []string {
	return ToIPList(LocalAddressListWithOptions(func(ip stdnet.IP) bool { return ip.To4() == nil && ip.To16() != nil }, opts...))
}

// LocalIPs returns all local IP addresses.
func LocalIPs() []string {
	return LocalIPsWithOptions()
}

// LocalIPsWithOptions returns all local IP addresses using custom providers.
func LocalIPsWithOptions(opts ...InterfaceOption) []string {
	return ToIPList(LocalAddressListWithOptions(func(ip stdnet.IP) bool { return ip != nil }, opts...))
}

// ToIPList converts IP addresses to strings.
func ToIPList(addressList []stdnet.IP) []string {
	out := make([]string, 0, len(addressList))
	seen := map[string]struct{}{}
	for _, ip := range addressList {
		if ip == nil {
			continue
		}
		s := ip.String()
		if _, ok := seen[s]; !ok {
			seen[s] = struct{}{}
			out = append(out, s)
		}
	}
	return out
}

// LocalAddressList returns local IP addresses matching addressFilter. nil means non-loopback IPv4.
func LocalAddressList(addressFilter func(stdnet.IP) bool) []stdnet.IP {
	return LocalAddressListWithOptions(addressFilter)
}

// LocalAddressListWithOptions returns local IP addresses matching addressFilter using custom providers.
func LocalAddressListWithOptions(addressFilter func(stdnet.IP) bool, opts ...InterfaceOption) []stdnet.IP {
	return LocalAddressListByInterfaceWithOptions(nil, addressFilter, opts...)
}

// LocalAddressListByInterface returns local IP addresses matching interface and address filters.
func LocalAddressListByInterface(interfaceFilter func(stdnet.Interface) bool, addressFilter func(stdnet.IP) bool) []stdnet.IP {
	return LocalAddressListByInterfaceWithOptions(interfaceFilter, addressFilter)
}

// LocalAddressListByInterfaceWithOptions returns local IP addresses matching interface and address filters using custom providers.
func LocalAddressListByInterfaceWithOptions(interfaceFilter func(stdnet.Interface) bool, addressFilter func(stdnet.IP) bool, opts ...InterfaceOption) []stdnet.IP {
	cfg := applyInterfaceOptions(opts)
	interfaces, err := cfg.interfaces()
	if err != nil {
		return nil
	}
	out := make([]stdnet.IP, 0)
	for _, iface := range interfaces {
		if interfaceFilter != nil && !interfaceFilter(iface) {
			continue
		}
		addrs, err := cfg.interfaceAddrs(iface)
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			ip := addrIP(addr)
			if ip == nil {
				continue
			}
			if addressFilter == nil {
				if ip.IsLoopback() || ip.To4() == nil {
					continue
				}
			} else if !addressFilter(ip) {
				continue
			}
			out = append(out, ip)
		}
	}
	return out
}

// GetLocalhostStr returns a preferred local host IP string.
func GetLocalhostStr() string {
	return GetLocalhostStrWithOptions()
}

// GetLocalhostStrWithOptions returns a preferred local host IP string using custom providers.
func GetLocalhostStrWithOptions(opts ...InterfaceOption) string {
	ips := LocalAddressListWithOptions(nil, opts...)
	if len(ips) > 0 {
		return ips[0].String()
	}
	return LocalIP
}

// GetLocalhost returns a preferred local host IP.
func GetLocalhost() stdnet.IP { return GetLocalhostWithOptions() }

// GetLocalhostWithOptions returns a preferred local host IP using custom providers.
func GetLocalhostWithOptions(opts ...InterfaceOption) stdnet.IP {
	return stdnet.ParseIP(GetLocalhostStrWithOptions(opts...))
}

// GetLocalHostName returns the OS host name.
func GetLocalHostName() string {
	return GetLocalHostNameWithOptions()
}

// GetLocalHostNameWithOptions returns the OS host name using custom providers.
func GetLocalHostNameWithOptions(opts ...InterfaceOption) string {
	cfg := applyInterfaceOptions(opts)
	host, err := cfg.lookupAddr(GetLocalhostStrWithOptions(opts...))
	if err == nil && len(host) > 0 {
		return strings.TrimSuffix(host[0], ".")
	}
	name, _ := cfg.hostname()
	return name
}

// GetLocalMACAddress returns the first non-empty local hardware address.
func GetLocalMACAddress(separator ...string) string {
	return GetLocalMACAddressWithOptions(nil, separator...)
}

// GetLocalMACAddressWithOptions returns the first non-empty local hardware address using custom providers.
func GetLocalMACAddressWithOptions(opts []InterfaceOption, separator ...string) string {
	hw := GetLocalHardwareAddressWithOptions(opts...)
	if hw == nil {
		return ""
	}
	sep := ":"
	if len(separator) > 0 {
		sep = separator[0]
	}
	return formatHardwareAddress(hw, sep)
}

// GetMACAddress returns the hardware address of the interface owning inetAddress.
func GetMACAddress(inetAddress stdnet.IP, separator ...string) string {
	return GetMACAddressWithOptions(inetAddress, nil, separator...)
}

// GetMACAddressWithOptions returns the hardware address of the interface owning inetAddress using custom providers.
func GetMACAddressWithOptions(inetAddress stdnet.IP, opts []InterfaceOption, separator ...string) string {
	hw := GetHardwareAddressWithOptions(inetAddress, opts...)
	if hw == nil {
		return ""
	}
	sep := ":"
	if len(separator) > 0 {
		sep = separator[0]
	}
	return formatHardwareAddress(hw, sep)
}

// GetHardwareAddress returns the hardware address of the interface owning inetAddress.
func GetHardwareAddress(inetAddress stdnet.IP) stdnet.HardwareAddr {
	return GetHardwareAddressWithOptions(inetAddress)
}

// GetHardwareAddressWithOptions returns the hardware address of the interface owning inetAddress using custom providers.
func GetHardwareAddressWithOptions(inetAddress stdnet.IP, opts ...InterfaceOption) stdnet.HardwareAddr {
	cfg := applyInterfaceOptions(opts)
	interfaces, err := cfg.interfaces()
	if err != nil {
		return nil
	}
	for _, iface := range interfaces {
		addrs, err := cfg.interfaceAddrs(iface)
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			if ip := addrIP(addr); ip != nil && ip.Equal(inetAddress) {
				return iface.HardwareAddr
			}
		}
	}
	return nil
}

// GetLocalHardwareAddress returns the first non-empty local hardware address.
func GetLocalHardwareAddress() stdnet.HardwareAddr {
	return GetLocalHardwareAddressWithOptions()
}

// GetLocalHardwareAddressWithOptions returns the first non-empty local hardware address using custom providers.
func GetLocalHardwareAddressWithOptions(opts ...InterfaceOption) stdnet.HardwareAddr {
	interfaces, err := applyInterfaceOptions(opts).interfaces()
	if err != nil {
		return nil
	}
	for _, iface := range interfaces {
		if len(iface.HardwareAddr) > 0 && iface.Flags&stdnet.FlagLoopback == 0 {
			return iface.HardwareAddr
		}
	}
	return nil
}

// NetCat sends data to host:port over TCP.
func NetCat(host string, port int, data []byte, timeout time.Duration) error {
	return NetCatWithOptions(host, port, data, WithConnectTimeout(timeout))
}

// NetCatWithOptions sends data to host:port using custom connection options.
func NetCatWithOptions(host string, port int, data []byte, opts ...ConnectOption) error {
	conn, err := ConnectWithOptions(host, port, opts...)
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()
	_, err = conn.Write(data)
	return err
}

// Ping checks whether an IP or host is reachable by opening a TCP connection to common ports.
func Ping(ip string, timeout time.Duration) bool {
	return PingWithOptions(ip, WithPingTimeout(timeout))
}

// PingWithOptions checks whether an IP or host is reachable with custom probe options.
func PingWithOptions(ip string, opts ...PingOption) bool {
	cfg := applyPingOptions(opts)
	for _, port := range cfg.ports {
		if !IsValidPort(port) {
			continue
		}
		ctx := cfg.ctx
		cancel := func() {}
		if cfg.timeout > 0 {
			ctx, cancel = context.WithTimeout(cfg.ctx, cfg.timeout)
		}
		conn, err := cfg.dialer.DialContext(ctx, cfg.network, stdnet.JoinHostPort(ip, strconvPort(port)))
		cancel()
		if err == nil {
			_ = conn.Close()
			return true
		}
	}
	return false
}

// IsOpen reports whether address can be opened within timeout.
func IsOpen(address *stdnet.TCPAddr, timeout time.Duration) bool {
	return IsOpenWithOptions(address, WithConnectTimeout(timeout))
}

// IsOpenWithOptions reports whether address can be opened with custom connection options.
func IsOpenWithOptions(address *stdnet.TCPAddr, opts ...ConnectOption) bool {
	if address == nil {
		return false
	}
	cfg, cancel := applyConnectOptions(opts)
	defer cancel()
	conn, err := cfg.dialer.DialContext(cfg.ctx, cfg.network, address.String())
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

// IDNToASCII converts a Unicode domain name to ASCII.
func IDNToASCII(unicode string) (string, error) { return idna.ToASCII(unicode) }

// GetMultistageReverseProxyIP returns the first non-unknown IP in a comma-separated proxy header.
func GetMultistageReverseProxyIP(ip string) string {
	for _, part := range strings.Split(ip, ",") {
		part = strings.TrimSpace(part)
		if !IsUnknown(part) {
			return part
		}
	}
	return ""
}

// IsUnknown reports whether checkString is empty or equals unknown case-insensitively.
func IsUnknown(checkString string) bool {
	return strings.TrimSpace(checkString) == "" || strings.EqualFold(strings.TrimSpace(checkString), "unknown")
}

// ParseCookies parses a Cookie header value.
func ParseCookies(cookieStr string) []*http.Cookie {
	req := &http.Request{Header: http.Header{"Cookie": []string{cookieStr}}}
	return req.Cookies()
}

// GetDNSInfo looks up DNS records by attribute names such as A, CNAME, MX, NS, or TXT.
func GetDNSInfo(hostName string, attrNames ...string) ([]string, error) {
	return GetDNSInfoWithOptions(hostName, WithDNSTypes(attrNames...))
}

// GetDNSInfoWithOptions looks up DNS records with custom resolver options.
func GetDNSInfoWithOptions(hostName string, opts ...ResolveOption) ([]string, error) {
	cfg, cancel := applyResolveOptions(opts)
	defer cancel()
	attrNames := cfg.attrNames
	if len(attrNames) == 0 {
		attrNames = []string{"A"}
	}
	out := make([]string, 0)
	for _, attr := range attrNames {
		switch strings.ToUpper(attr) {
		case "A", "AAAA":
			network := "ip4"
			if strings.ToUpper(attr) == "AAAA" {
				network = "ip6"
			}
			ips, err := cfg.resolver.LookupIP(cfg.ctx, network, hostName)
			if err != nil {
				return out, err
			}
			for _, ip := range ips {
				out = append(out, ip.String())
			}
		case "CNAME":
			v, err := cfg.resolver.LookupCNAME(cfg.ctx, hostName)
			if err != nil {
				return out, err
			}
			out = append(out, v)
		case "MX":
			mxs, err := cfg.resolver.LookupMX(cfg.ctx, hostName)
			if err != nil {
				return out, err
			}
			for _, mx := range mxs {
				out = append(out, mx.Host)
			}
		case "NS":
			nss, err := cfg.resolver.LookupNS(cfg.ctx, hostName)
			if err != nil {
				return out, err
			}
			for _, ns := range nss {
				out = append(out, ns.Host)
			}
		case "TXT":
			txts, err := cfg.resolver.LookupTXT(cfg.ctx, hostName)
			if err != nil {
				return out, err
			}
			out = append(out, txts...)
		}
	}
	return out, nil
}

// Connect opens a TCP connection to host:port.
func Connect(hostname string, port int, timeout time.Duration) (stdnet.Conn, error) {
	return ConnectWithOptions(hostname, port, WithConnectTimeout(timeout))
}

// ConnectWithOptions opens a connection to host:port using custom connection options.
func ConnectWithOptions(hostname string, port int, opts ...ConnectOption) (stdnet.Conn, error) {
	cfg, cancel := applyConnectOptions(opts)
	defer cancel()
	addr := stdnet.JoinHostPort(hostname, strconvPort(port))
	return cfg.dialer.DialContext(cfg.ctx, cfg.network, addr)
}

// GetRemoteAddress returns conn's remote address string.
func GetRemoteAddress(conn stdnet.Conn) string {
	if conn == nil || conn.RemoteAddr() == nil {
		return ""
	}
	return conn.RemoteAddr().String()
}

// IsConnected reports whether conn appears open.
func IsConnected(conn stdnet.Conn) bool { return conn != nil && conn.RemoteAddr() != nil }

func addrIP(addr stdnet.Addr) stdnet.IP {
	switch v := addr.(type) {
	case *stdnet.IPNet:
		return v.IP
	case *stdnet.IPAddr:
		return v.IP
	default:
		return nil
	}
}

func formatHardwareAddress(hw stdnet.HardwareAddr, sep string) string {
	parts := make([]string, len(hw))
	for i, b := range hw {
		parts[i] = fmt.Sprintf("%02x", b)
	}
	return strings.Join(parts, sep)
}

func strconvPort(port int) string { return fmt.Sprintf("%d", port) }

func osHostname() (string, error) { return os.Hostname() }
