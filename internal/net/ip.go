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
	compile func(string) (*regexp.Regexp, error)
}

// WildcardOption customizes wildcard IP matching per call.
type WildcardOption func(*wildcardConfig)

// WithWildcardCompileFunc sets the compiler used by MatchesWildcardWithOptions.
func WithWildcardCompileFunc(compile func(string) (*regexp.Regexp, error)) WildcardOption {
	return func(c *wildcardConfig) { c.compile = compile }
}

func applyWildcardOptions(opts []WildcardOption) wildcardConfig {
	cfg := wildcardConfig{compile: regexp.Compile}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if cfg.compile == nil {
		cfg.compile = regexp.Compile
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
	ip := stdnet.ParseIP(strings.TrimSpace(strIP)).To4()
	if ip == nil {
		return 0, fmt.Errorf("invalid IPv4 address: %s", strIP)
	}
	return binary.BigEndian.Uint32(ip), nil
}

// IPv4ToLongDefault converts a dotted IPv4 string to uint32, returning defaultValue when invalid.
func IPv4ToLongDefault(strIP string, defaultValue uint32) uint32 {
	v, err := IPv4ToLong(strIP)
	if err != nil {
		return defaultValue
	}
	return v
}

// IPv6ToBigInt converts an IPv6 string to a big integer.
func IPv6ToBigInt(ipv6Str string) (*big.Int, error) {
	ip := stdnet.ParseIP(strings.TrimSpace(ipv6Str))
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
func IsIP(s string) bool { return stdnet.ParseIP(strings.TrimSpace(s)) != nil }

// IsIPv4 reports whether s is an IPv4 address.
func IsIPv4(s string) bool {
	ip := stdnet.ParseIP(strings.TrimSpace(s))
	return ip != nil && ip.To4() != nil
}

// IsIPv6 reports whether s is an IPv6 address.
func IsIPv6(s string) bool {
	ip := stdnet.ParseIP(strings.TrimSpace(s))
	return ip != nil && ip.To4() == nil && ip.To16() != nil
}

// IsInnerIP reports whether ipAddress belongs to common private IPv4 ranges.
func IsInnerIP(ipAddress string) bool {
	ip := stdnet.ParseIP(strings.TrimSpace(ipAddress))
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
	bit, err := MaskBitByMask(mask)
	if err != nil {
		return "", err
	}
	return ip + IPMaskSplitMark + strconv.Itoa(bit), nil
}

// BeginIP returns the first IPv4 address in an ip/maskBit block.
func BeginIP(ip string, maskBit int) (string, error) {
	v, err := BeginIPLong(ip, maskBit)
	if err != nil {
		return "", err
	}
	return LongToIPv4(v), nil
}

// BeginIPLong returns the first IPv4 value in an ip/maskBit block.
func BeginIPLong(ip string, maskBit int) (uint32, error) {
	base, err := IPv4ToLong(ip)
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
	v, err := EndIPLong(ip, maskBit)
	if err != nil {
		return "", err
	}
	return LongToIPv4(v), nil
}

// EndIPLong returns the last IPv4 value in an ip/maskBit block.
func EndIPLong(ip string, maskBit int) (uint32, error) {
	begin, err := BeginIPLong(ip, maskBit)
	if err != nil {
		return 0, err
	}
	mask, _ := maskLong(maskBit)
	return begin | ^mask, nil
}

// MaskBitByMask converts a dotted IPv4 mask to mask bits.
func MaskBitByMask(mask string) (int, error) {
	ip := stdnet.ParseIP(strings.TrimSpace(mask)).To4()
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
	from, err := IPv4ToLong(fromIP)
	if err != nil {
		return "", err
	}
	to, err := IPv4ToLong(toIP)
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
	from, err := IPv4ToLong(fromIP)
	if err != nil {
		return 0, err
	}
	to, err := IPv4ToLong(toIP)
	if err != nil {
		return 0, err
	}
	if from > to {
		from, to = to, from
	}
	return uint64(to-from) + 1, nil
}

// IsMaskValid reports whether mask is a contiguous IPv4 mask.
func IsMaskValid(mask string) bool { _, err := MaskBitByMask(mask); return err == nil }

// IsMaskBitValid reports whether maskBit is in [0, 32].
func IsMaskBitValid(maskBit int) bool { return maskBit >= 0 && maskBit <= IPMaskMax }

// ListIPs expands an IPv4 range expression: single IP, from-to, or ip/maskBit.
func ListIPs(ipRange string, isAll bool) ([]string, error) {
	if strings.Contains(ipRange, IPSplitMark) {
		parts := strings.SplitN(ipRange, IPSplitMark, 2)
		return ListIPRange(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
	}
	if strings.Contains(ipRange, IPMaskSplitMark) {
		parts := strings.SplitN(ipRange, IPMaskSplitMark, 2)
		bit, err := strconv.Atoi(strings.TrimSpace(parts[1]))
		if err != nil {
			return nil, err
		}
		return ListIPCIDR(strings.TrimSpace(parts[0]), bit, isAll)
	}
	if !IsIPv4(ipRange) {
		return nil, fmt.Errorf("invalid IPv4 address: %s", ipRange)
	}
	return []string{ipRange}, nil
}

// ListIPCIDR expands an ip/maskBit block into IPv4 strings.
func ListIPCIDR(ip string, maskBit int, isAll bool) ([]string, error) {
	start, err := BeginIPLong(ip, maskBit)
	if err != nil {
		return nil, err
	}
	end, err := EndIPLong(ip, maskBit)
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
	from, err := IPv4ToLong(fromIP)
	if err != nil {
		return nil, err
	}
	to, err := IPv4ToLong(toIP)
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
	if !IsIPv4(ipAddress) {
		return false
	}
	cfg := applyWildcardOptions(opts)
	parts := strings.Split(wildcard, ".")
	if len(parts) != 4 {
		return false
	}
	for i, p := range parts {
		if p == "*" {
			parts[i] = `\d{1,3}`
			continue
		}
		if n, err := strconv.Atoi(p); err != nil || n < 0 || n > 255 {
			return false
		}
		parts[i] = regexp.QuoteMeta(p)
	}
	re, err := cfg.compile(`^` + strings.Join(parts, `\.`) + `$`)
	if err != nil {
		return false
	}
	return re.MatchString(ipAddress)
}

// IsInRange reports whether ip belongs to cidr.
func IsInRange(ip, cidr string) bool {
	parsed := stdnet.ParseIP(strings.TrimSpace(ip))
	if parsed == nil {
		return false
	}
	_, network, err := stdnet.ParseCIDR(strings.TrimSpace(cidr))
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
