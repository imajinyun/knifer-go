package net

import (
	"encoding/binary"
	"fmt"
	"math"
	"math/big"
	stdnet "net"
	"regexp"
	"strconv"
	"strings"
)

type wildcardConfig struct {
	compile  func(string) (*regexp.Regexp, error)
	parseIP  func(string) stdnet.IP
	parseInt func(string) (int, error)
}

type ipConfig struct {
	parseIP   func(string) stdnet.IP
	parseCIDR func(string) (stdnet.IP, *stdnet.IPNet, error)
	parseInt  func(string) (int, error)
}

// WildcardOption customizes wildcard IP matching per call.
type WildcardOption func(*wildcardConfig)

// IPOption customizes IP parsing helpers per call.
type IPOption func(*ipConfig)

// WithWildcardCompileFunc sets the compiler used by MatchesWildcardWithOptions.
func WithWildcardCompileFunc(compile func(string) (*regexp.Regexp, error)) WildcardOption {
	return func(c *wildcardConfig) {
		if compile != nil {
			c.compile = compile
		}
	}
}

// WithWildcardIPParser sets the IP parser used by MatchesWildcardWithOptions.
func WithWildcardIPParser(parse func(string) stdnet.IP) WildcardOption {
	return func(c *wildcardConfig) {
		if parse != nil {
			c.parseIP = parse
		}
	}
}

// WithWildcardIntParser sets the integer parser used by MatchesWildcardWithOptions.
func WithWildcardIntParser(parse func(string) (int, error)) WildcardOption {
	return func(c *wildcardConfig) {
		if parse != nil {
			c.parseInt = parse
		}
	}
}

// WithIPParser sets the IP parser used by IP helpers.
func WithIPParser(parse func(string) stdnet.IP) IPOption {
	return func(c *ipConfig) {
		if parse != nil {
			c.parseIP = parse
		}
	}
}

// WithCIDRParser sets the CIDR parser used by IP range helpers.
func WithCIDRParser(parse func(string) (stdnet.IP, *stdnet.IPNet, error)) IPOption {
	return func(c *ipConfig) {
		if parse != nil {
			c.parseCIDR = parse
		}
	}
}

// WithIPIntParser sets the integer parser used by IP helpers.
func WithIPIntParser(parse func(string) (int, error)) IPOption {
	return func(c *ipConfig) {
		if parse != nil {
			c.parseInt = parse
		}
	}
}

func applyWildcardOptions(opts []WildcardOption) wildcardConfig {
	cfg := wildcardConfig{compile: regexp.Compile, parseIP: stdnet.ParseIP, parseInt: strconv.Atoi}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if cfg.compile == nil {
		cfg.compile = regexp.Compile
	}
	if cfg.parseIP == nil {
		cfg.parseIP = stdnet.ParseIP
	}
	if cfg.parseInt == nil {
		cfg.parseInt = strconv.Atoi
	}
	return cfg
}

func applyIPOptions(opts []IPOption) ipConfig {
	cfg := ipConfig{parseIP: stdnet.ParseIP, parseCIDR: stdnet.ParseCIDR, parseInt: strconv.Atoi}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if cfg.parseIP == nil {
		cfg.parseIP = stdnet.ParseIP
	}
	if cfg.parseCIDR == nil {
		cfg.parseCIDR = stdnet.ParseCIDR
	}
	if cfg.parseInt == nil {
		cfg.parseInt = strconv.Atoi
	}
	return cfg
}

const (
	// LocalIP is the IPv4 loopback address.
	LocalIP = "127.0.0.1"
	// IPSplitMark separates an IPv4 range.
	IPSplitMark = "-"
	// IPMaskSplitMark separates an IPv4 address and mask bit.
	IPMaskSplitMark = "/"
	// IPMaskMax is the maximum IPv4 mask bit.
	IPMaskMax = 32
)

// LongToIPv4 converts a uint32 IPv4 value to dotted string form.
func LongToIPv4(longIP uint32) string {
	var b [4]byte
	binary.BigEndian.PutUint32(b[:], longIP)
	return stdnet.IP(b[:]).String()
}

// IPv4ToLong converts a dotted IPv4 string to uint32.
func IPv4ToLong(strIP string) (uint32, error) {
	return IPv4ToLongWithOptions(strIP)
}

// IPv4ToLongWithOptions converts a dotted IPv4 string to uint32 using custom providers.
func IPv4ToLongWithOptions(strIP string, opts ...IPOption) (uint32, error) {
	cfg := applyIPOptions(opts)
	ip := cfg.parseIP(strings.TrimSpace(strIP)).To4()
	if ip == nil {
		return 0, fmt.Errorf("invalid IPv4 address: %s", strIP)
	}
	return binary.BigEndian.Uint32(ip), nil
}

// IPv4ToLongDefault converts a dotted IPv4 string to uint32, returning defaultValue when invalid.
func IPv4ToLongDefault(strIP string, defaultValue uint32) uint32 {
	return IPv4ToLongDefaultWithOptions(strIP, defaultValue)
}

// IPv4ToLongDefaultWithOptions converts a dotted IPv4 string to uint32 using custom providers, returning defaultValue when invalid.
func IPv4ToLongDefaultWithOptions(strIP string, defaultValue uint32, opts ...IPOption) uint32 {
	v, err := IPv4ToLongWithOptions(strIP, opts...)
	if err != nil {
		return defaultValue
	}
	return v
}

// IPv6ToBigInt converts an IPv6 string to a big integer.
func IPv6ToBigInt(ipv6Str string) (*big.Int, error) {
	return IPv6ToBigIntWithOptions(ipv6Str)
}

// IPv6ToBigIntWithOptions converts an IPv6 string to a big integer using custom providers.
func IPv6ToBigIntWithOptions(ipv6Str string, opts ...IPOption) (*big.Int, error) {
	cfg := applyIPOptions(opts)
	ip := cfg.parseIP(strings.TrimSpace(ipv6Str))
	if ip == nil || ip.To4() != nil {
		return nil, fmt.Errorf("invalid IPv6 address: %s", ipv6Str)
	}
	return new(big.Int).SetBytes(ip.To16()), nil
}

// BigIntToIPv6 converts a big integer in the IPv6 address range to a string.
func BigIntToIPv6(n *big.Int) (string, error) {
	if n == nil || n.Sign() < 0 || n.BitLen() > 128 {
		return "", fmt.Errorf("IPv6 integer out of range")
	}
	b := n.Bytes()
	buf := make([]byte, 16)
	copy(buf[16-len(b):], b)
	return stdnet.IP(buf).String(), nil
}

// IsIP reports whether s is an IPv4 or IPv6 address.
func IsIP(s string) bool { return IsIPWithOptions(s) }

// IsIPWithOptions reports whether s is an IPv4 or IPv6 address using custom providers.
func IsIPWithOptions(s string, opts ...IPOption) bool {
	cfg := applyIPOptions(opts)
	return cfg.parseIP(strings.TrimSpace(s)) != nil
}

// IsIPv4 reports whether s is an IPv4 address.
func IsIPv4(s string) bool {
	return IsIPv4WithOptions(s)
}

// IsIPv4WithOptions reports whether s is an IPv4 address using custom providers.
func IsIPv4WithOptions(s string, opts ...IPOption) bool {
	cfg := applyIPOptions(opts)
	ip := cfg.parseIP(strings.TrimSpace(s))
	return ip != nil && ip.To4() != nil
}

// IsIPv6 reports whether s is an IPv6 address.
func IsIPv6(s string) bool {
	return IsIPv6WithOptions(s)
}

// IsIPv6WithOptions reports whether s is an IPv6 address using custom providers.
func IsIPv6WithOptions(s string, opts ...IPOption) bool {
	cfg := applyIPOptions(opts)
	ip := cfg.parseIP(strings.TrimSpace(s))
	return ip != nil && ip.To4() == nil && ip.To16() != nil
}

// IsInnerIP reports whether ipAddress belongs to common private IPv4 ranges.
func IsInnerIP(ipAddress string) bool {
	return IsInnerIPWithOptions(ipAddress)
}

// IsInnerIPWithOptions reports whether ipAddress belongs to common private IPv4 ranges using custom providers.
func IsInnerIPWithOptions(ipAddress string, opts ...IPOption) bool {
	cfg := applyIPOptions(opts)
	ip := cfg.parseIP(strings.TrimSpace(ipAddress))
	if ip == nil {
		return false
	}
	if ip.IsLoopback() || ip.IsPrivate() {
		return true
	}
	v4 := ip.To4()
	return v4 != nil && v4[0] == 169 && v4[1] == 254
}

// FormatIPBlock formats ip and mask as ip/maskBit.
func FormatIPBlock(ip, mask string) (string, error) {
	return FormatIPBlockWithOptions(ip, mask)
}

// FormatIPBlockWithOptions formats ip and mask as ip/maskBit using custom providers.
func FormatIPBlockWithOptions(ip, mask string, opts ...IPOption) (string, error) {
	bit, err := MaskBitByMaskWithOptions(mask, opts...)
	if err != nil {
		return "", err
	}
	return ip + IPMaskSplitMark + strconv.Itoa(bit), nil
}

// BeginIP returns the first IPv4 address in an ip/maskBit block.
func BeginIP(ip string, maskBit int) (string, error) {
	return BeginIPWithOptions(ip, maskBit)
}

// BeginIPWithOptions returns the first IPv4 address in an ip/maskBit block using custom providers.
func BeginIPWithOptions(ip string, maskBit int, opts ...IPOption) (string, error) {
	v, err := BeginIPLongWithOptions(ip, maskBit, opts...)
	if err != nil {
		return "", err
	}
	return LongToIPv4(v), nil
}

// BeginIPLong returns the first IPv4 value in an ip/maskBit block.
func BeginIPLong(ip string, maskBit int) (uint32, error) {
	return BeginIPLongWithOptions(ip, maskBit)
}

// BeginIPLongWithOptions returns the first IPv4 value in an ip/maskBit block using custom providers.
func BeginIPLongWithOptions(ip string, maskBit int, opts ...IPOption) (uint32, error) {
	base, err := IPv4ToLongWithOptions(ip, opts...)
	if err != nil {
		return 0, err
	}
	mask, err := maskLong(maskBit)
	if err != nil {
		return 0, err
	}
	return base & mask, nil
}

// EndIP returns the last IPv4 address in an ip/maskBit block.
func EndIP(ip string, maskBit int) (string, error) {
	return EndIPWithOptions(ip, maskBit)
}

// EndIPWithOptions returns the last IPv4 address in an ip/maskBit block using custom providers.
func EndIPWithOptions(ip string, maskBit int, opts ...IPOption) (string, error) {
	v, err := EndIPLongWithOptions(ip, maskBit, opts...)
	if err != nil {
		return "", err
	}
	return LongToIPv4(v), nil
}

// EndIPLong returns the last IPv4 value in an ip/maskBit block.
func EndIPLong(ip string, maskBit int) (uint32, error) {
	return EndIPLongWithOptions(ip, maskBit)
}

// EndIPLongWithOptions returns the last IPv4 value in an ip/maskBit block using custom providers.
func EndIPLongWithOptions(ip string, maskBit int, opts ...IPOption) (uint32, error) {
	begin, err := BeginIPLongWithOptions(ip, maskBit, opts...)
	if err != nil {
		return 0, err
	}
	mask, _ := maskLong(maskBit)
	return begin | ^mask, nil
}

// MaskBitByMask converts a dotted IPv4 mask to mask bits.
func MaskBitByMask(mask string) (int, error) {
	return MaskBitByMaskWithOptions(mask)
}

// MaskBitByMaskWithOptions converts a dotted IPv4 mask to mask bits using custom providers.
func MaskBitByMaskWithOptions(mask string, opts ...IPOption) (int, error) {
	cfg := applyIPOptions(opts)
	ip := cfg.parseIP(strings.TrimSpace(mask)).To4()
	if ip == nil {
		return 0, fmt.Errorf("invalid IPv4 mask: %s", mask)
	}
	ones, bits := stdnet.IPMask(ip).Size()
	if bits != 32 || ones < 0 {
		return 0, fmt.Errorf("non-contiguous IPv4 mask: %s", mask)
	}
	return ones, nil
}

// CountByMaskBit returns the number of addresses represented by maskBit.
func CountByMaskBit(maskBit int, isAll bool) (uint64, error) {
	if !IsMaskBitValid(maskBit) {
		return 0, fmt.Errorf("invalid mask bit: %d", maskBit)
	}
	count := uint64(1) << uint(IPMaskMax-maskBit)
	if !isAll && maskBit < 31 && count >= 2 {
		count -= 2
	}
	return count, nil
}

// MaskByMaskBit converts mask bits to a dotted IPv4 mask.
func MaskByMaskBit(maskBit int) (string, error) {
	mask, err := maskLong(maskBit)
	if err != nil {
		return "", err
	}
	return LongToIPv4(mask), nil
}

// MaskByIPRange returns the common IPv4 mask for an inclusive range.
func MaskByIPRange(fromIP, toIP string) (string, error) {
	return MaskByIPRangeWithOptions(fromIP, toIP)
}

// MaskByIPRangeWithOptions returns the common IPv4 mask for an inclusive range using custom providers.
func MaskByIPRangeWithOptions(fromIP, toIP string, opts ...IPOption) (string, error) {
	from, err := IPv4ToLongWithOptions(fromIP, opts...)
	if err != nil {
		return "", err
	}
	to, err := IPv4ToLongWithOptions(toIP, opts...)
	if err != nil {
		return "", err
	}
	if from > to {
		from, to = to, from
	}
	diff := from ^ to
	bits := IPMaskMax
	for diff > 0 {
		bits--
		diff >>= 1
	}
	return MaskByMaskBit(bits)
}

// CountByIPRange returns the inclusive number of IPv4 addresses in a range.
func CountByIPRange(fromIP, toIP string) (uint64, error) {
	return CountByIPRangeWithOptions(fromIP, toIP)
}

// CountByIPRangeWithOptions returns the inclusive number of IPv4 addresses in a range using custom providers.
func CountByIPRangeWithOptions(fromIP, toIP string, opts ...IPOption) (uint64, error) {
	from, err := IPv4ToLongWithOptions(fromIP, opts...)
	if err != nil {
		return 0, err
	}
	to, err := IPv4ToLongWithOptions(toIP, opts...)
	if err != nil {
		return 0, err
	}
	if from > to {
		from, to = to, from
	}
	return uint64(to-from) + 1, nil
}

// IsMaskValid reports whether mask is a contiguous IPv4 mask.
func IsMaskValid(mask string) bool { return IsMaskValidWithOptions(mask) }

// IsMaskValidWithOptions reports whether mask is a contiguous IPv4 mask using custom providers.
func IsMaskValidWithOptions(mask string, opts ...IPOption) bool {
	_, err := MaskBitByMaskWithOptions(mask, opts...)
	return err == nil
}

// IsMaskBitValid reports whether maskBit is in [0, 32].
func IsMaskBitValid(maskBit int) bool { return maskBit >= 0 && maskBit <= IPMaskMax }

// ListIPs expands an IPv4 range expression: single IP, from-to, or ip/maskBit.
func ListIPs(ipRange string, isAll bool) ([]string, error) {
	return ListIPsWithOptions(ipRange, isAll)
}

// ListIPsWithOptions expands an IPv4 range expression using custom providers: single IP, from-to, or ip/maskBit.
func ListIPsWithOptions(ipRange string, isAll bool, opts ...IPOption) ([]string, error) {
	cfg := applyIPOptions(opts)
	if strings.Contains(ipRange, IPSplitMark) {
		parts := strings.SplitN(ipRange, IPSplitMark, 2)
		return ListIPRangeWithOptions(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]), opts...)
	}
	if strings.Contains(ipRange, IPMaskSplitMark) {
		parts := strings.SplitN(ipRange, IPMaskSplitMark, 2)
		bit, err := cfg.parseInt(strings.TrimSpace(parts[1]))
		if err != nil {
			return nil, err
		}
		return ListIPCIDRWithOptions(strings.TrimSpace(parts[0]), bit, isAll, opts...)
	}
	if !IsIPv4WithOptions(ipRange, opts...) {
		return nil, fmt.Errorf("invalid IPv4 address: %s", ipRange)
	}
	return []string{ipRange}, nil
}

// ListIPCIDR expands an ip/maskBit block into IPv4 strings.
func ListIPCIDR(ip string, maskBit int, isAll bool) ([]string, error) {
	return ListIPCIDRWithOptions(ip, maskBit, isAll)
}

// ListIPCIDRWithOptions expands an ip/maskBit block into IPv4 strings using custom providers.
func ListIPCIDRWithOptions(ip string, maskBit int, isAll bool, opts ...IPOption) ([]string, error) {
	start, err := BeginIPLongWithOptions(ip, maskBit, opts...)
	if err != nil {
		return nil, err
	}
	end, err := EndIPLongWithOptions(ip, maskBit, opts...)
	if err != nil {
		return nil, err
	}
	if !isAll && maskBit < 31 && end > start {
		start++
		end--
	}
	return listIPLongRange(start, end)
}

// ListIPRange expands an inclusive IPv4 range into strings.
func ListIPRange(fromIP, toIP string) ([]string, error) {
	return ListIPRangeWithOptions(fromIP, toIP)
}

// ListIPRangeWithOptions expands an inclusive IPv4 range into strings using custom providers.
func ListIPRangeWithOptions(fromIP, toIP string, opts ...IPOption) ([]string, error) {
	from, err := IPv4ToLongWithOptions(fromIP, opts...)
	if err != nil {
		return nil, err
	}
	to, err := IPv4ToLongWithOptions(toIP, opts...)
	if err != nil {
		return nil, err
	}
	if from > to {
		from, to = to, from
	}
	return listIPLongRange(from, to)
}

// MatchesWildcard reports whether ipAddress matches a wildcard such as 192.168.*.*.
func MatchesWildcard(wildcard, ipAddress string) bool {
	return MatchesWildcardWithOptions(wildcard, ipAddress)
}

// MatchesWildcardWithOptions reports whether ipAddress matches a wildcard with options.
func MatchesWildcardWithOptions(wildcard, ipAddress string, opts ...WildcardOption) bool {
	cfg := applyWildcardOptions(opts)
	ip := cfg.parseIP(strings.TrimSpace(ipAddress))
	v4 := ip.To4()
	if v4 == nil {
		return false
	}
	parsedIP := v4.String()
	parts := strings.Split(wildcard, ".")
	if len(parts) != 4 {
		return false
	}
	for i, p := range parts {
		if p == "*" {
			parts[i] = `\d{1,3}`
			continue
		}
		n, err := cfg.parseInt(p)
		if err != nil || n < 0 || n > 255 {
			return false
		}
		parts[i] = strconv.Itoa(n)
	}
	re, err := cfg.compile(`^` + strings.Join(parts, `\.`) + `$`)
	if err != nil {
		return false
	}
	return re.MatchString(parsedIP)
}

// IsInRange reports whether ip belongs to cidr.
func IsInRange(ip, cidr string) bool {
	return IsInRangeWithOptions(ip, cidr)
}

// IsInRangeWithOptions reports whether ip belongs to cidr using custom providers.
func IsInRangeWithOptions(ip, cidr string, opts ...IPOption) bool {
	cfg := applyIPOptions(opts)
	parsed := cfg.parseIP(strings.TrimSpace(ip))
	if parsed == nil {
		return false
	}
	_, network, err := cfg.parseCIDR(strings.TrimSpace(cidr))
	return err == nil && network.Contains(parsed)
}

func maskLong(maskBit int) (uint32, error) {
	if !IsMaskBitValid(maskBit) {
		return 0, fmt.Errorf("invalid mask bit: %d", maskBit)
	}
	if maskBit == 0 {
		return 0, nil
	}
	return math.MaxUint32 << uint(IPMaskMax-maskBit), nil
}

func listIPLongRange(from, to uint32) ([]string, error) {
	if from > to {
		return nil, nil
	}
	count := uint64(to-from) + 1
	if count > 1_000_000 {
		return nil, fmt.Errorf("IP range too large: %d", count)
	}
	out := make([]string, 0, count)
	for i := from; i <= to; i++ {
		out = append(out, LongToIPv4(i))
		if i == math.MaxUint32 {
			break
		}
	}
	return out, nil
}
