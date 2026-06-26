package vnet

import (
	"math/big"
	stdnet "net"
	"regexp"

	netimpl "github.com/imajinyun/knifer-go/internal/net"
)

type WildcardOption = netimpl.WildcardOption

type IPOption = netimpl.IPOption

func WithWildcardCompileFunc(compile func(string) (*regexp.Regexp, error)) WildcardOption {
	return netimpl.WithWildcardCompileFunc(compile)
}

func WithWildcardIPParser(parse func(string) stdnet.IP) WildcardOption {
	return netimpl.WithWildcardIPParser(parse)
}

func WithWildcardIntParser(parse func(string) (int, error)) WildcardOption {
	return netimpl.WithWildcardIntParser(parse)
}

func WithIPParser(parse func(string) stdnet.IP) IPOption { return netimpl.WithIPParser(parse) }

func WithCIDRParser(parse func(string) (stdnet.IP, *stdnet.IPNet, error)) IPOption {
	return netimpl.WithCIDRParser(parse)
}

func WithIPIntParser(parse func(string) (int, error)) IPOption { return netimpl.WithIPIntParser(parse) }

func LongToIPv4(longIP uint32) string { return netimpl.LongToIPv4(longIP) }

func IPv4ToLong(strIP string) (uint32, error) { return netimpl.IPv4ToLong(strIP) }

func IPv4ToLongWithOptions(strIP string, opts ...IPOption) (uint32, error) {
	return netimpl.IPv4ToLongWithOptions(strIP, opts...)
}

func IPv4ToLongDefault(strIP string, defaultValue uint32) uint32 {
	return netimpl.IPv4ToLongDefault(strIP, defaultValue)
}

func IPv4ToLongDefaultWithOptions(strIP string, defaultValue uint32, opts ...IPOption) uint32 {
	return netimpl.IPv4ToLongDefaultWithOptions(strIP, defaultValue, opts...)
}

func IPv6ToBigInt(ipv6Str string) (*big.Int, error) { return netimpl.IPv6ToBigInt(ipv6Str) }

func IPv6ToBigIntWithOptions(ipv6Str string, opts ...IPOption) (*big.Int, error) {
	return netimpl.IPv6ToBigIntWithOptions(ipv6Str, opts...)
}

func BigIntToIPv6(n *big.Int) (string, error) { return netimpl.BigIntToIPv6(n) }

func IsIP(s string) bool { return netimpl.IsIP(s) }

func IsIPWithOptions(s string, opts ...IPOption) bool { return netimpl.IsIPWithOptions(s, opts...) }

func IsIPv4(s string) bool { return netimpl.IsIPv4(s) }

func IsIPv4WithOptions(s string, opts ...IPOption) bool { return netimpl.IsIPv4WithOptions(s, opts...) }

func IsIPv6(s string) bool { return netimpl.IsIPv6(s) }

func IsIPv6WithOptions(s string, opts ...IPOption) bool { return netimpl.IsIPv6WithOptions(s, opts...) }

func IsInnerIP(ipAddress string) bool { return netimpl.IsInnerIP(ipAddress) }

func IsInnerIPWithOptions(ipAddress string, opts ...IPOption) bool {
	return netimpl.IsInnerIPWithOptions(ipAddress, opts...)
}

func FormatIPBlock(ip, mask string) (string, error) { return netimpl.FormatIPBlock(ip, mask) }

func FormatIPBlockWithOptions(ip, mask string, opts ...IPOption) (string, error) {
	return netimpl.FormatIPBlockWithOptions(ip, mask, opts...)
}

func BeginIP(ip string, maskBit int) (string, error) { return netimpl.BeginIP(ip, maskBit) }

func BeginIPWithOptions(ip string, maskBit int, opts ...IPOption) (string, error) {
	return netimpl.BeginIPWithOptions(ip, maskBit, opts...)
}

func BeginIPLong(ip string, maskBit int) (uint32, error) { return netimpl.BeginIPLong(ip, maskBit) }

func BeginIPLongWithOptions(ip string, maskBit int, opts ...IPOption) (uint32, error) {
	return netimpl.BeginIPLongWithOptions(ip, maskBit, opts...)
}

func EndIP(ip string, maskBit int) (string, error) { return netimpl.EndIP(ip, maskBit) }

func EndIPWithOptions(ip string, maskBit int, opts ...IPOption) (string, error) {
	return netimpl.EndIPWithOptions(ip, maskBit, opts...)
}

func EndIPLong(ip string, maskBit int) (uint32, error) { return netimpl.EndIPLong(ip, maskBit) }

func EndIPLongWithOptions(ip string, maskBit int, opts ...IPOption) (uint32, error) {
	return netimpl.EndIPLongWithOptions(ip, maskBit, opts...)
}

func MaskBitByMask(mask string) (int, error) { return netimpl.MaskBitByMask(mask) }

func MaskBitByMaskWithOptions(mask string, opts ...IPOption) (int, error) {
	return netimpl.MaskBitByMaskWithOptions(mask, opts...)
}

func CountByMaskBit(maskBit int, isAll bool) (uint64, error) {
	return netimpl.CountByMaskBit(maskBit, isAll)
}

func MaskByMaskBit(maskBit int) (string, error) { return netimpl.MaskByMaskBit(maskBit) }

func MaskByIPRange(fromIP, toIP string) (string, error) { return netimpl.MaskByIPRange(fromIP, toIP) }

func MaskByIPRangeWithOptions(fromIP, toIP string, opts ...IPOption) (string, error) {
	return netimpl.MaskByIPRangeWithOptions(fromIP, toIP, opts...)
}

func CountByIPRange(fromIP, toIP string) (uint64, error) {
	return netimpl.CountByIPRange(fromIP, toIP)
}

func CountByIPRangeWithOptions(fromIP, toIP string, opts ...IPOption) (uint64, error) {
	return netimpl.CountByIPRangeWithOptions(fromIP, toIP, opts...)
}

func IsMaskValid(mask string) bool { return netimpl.IsMaskValid(mask) }

func IsMaskValidWithOptions(mask string, opts ...IPOption) bool {
	return netimpl.IsMaskValidWithOptions(mask, opts...)
}

func IsMaskBitValid(maskBit int) bool { return netimpl.IsMaskBitValid(maskBit) }

func ListIPs(ipRange string, isAll bool) ([]string, error) { return netimpl.ListIPs(ipRange, isAll) }

func ListIPsWithOptions(ipRange string, isAll bool, opts ...IPOption) ([]string, error) {
	return netimpl.ListIPsWithOptions(ipRange, isAll, opts...)
}

func ListIPCIDR(ip string, maskBit int, isAll bool) ([]string, error) {
	return netimpl.ListIPCIDR(ip, maskBit, isAll)
}

func ListIPCIDRWithOptions(ip string, maskBit int, isAll bool, opts ...IPOption) ([]string, error) {
	return netimpl.ListIPCIDRWithOptions(ip, maskBit, isAll, opts...)
}

func ListIPRange(fromIP, toIP string) ([]string, error) { return netimpl.ListIPRange(fromIP, toIP) }

func ListIPRangeWithOptions(fromIP, toIP string, opts ...IPOption) ([]string, error) {
	return netimpl.ListIPRangeWithOptions(fromIP, toIP, opts...)
}

func MatchesWildcard(wildcard, ipAddress string) bool {
	return netimpl.MatchesWildcard(wildcard, ipAddress)
}

func MatchesWildcardWithOptions(wildcard, ipAddress string, opts ...WildcardOption) bool {
	return netimpl.MatchesWildcardWithOptions(wildcard, ipAddress, opts...)
}

func IsInRange(ip, cidr string) bool { return netimpl.IsInRange(ip, cidr) }

func IsInRangeWithOptions(ip, cidr string, opts ...IPOption) bool {
	return netimpl.IsInRangeWithOptions(ip, cidr, opts...)
}

func HideIPPart(ip string) string { return netimpl.HideIPPart(ip) }

func HideIPPartLong(ip uint32) string { return netimpl.HideIPPartLong(ip) }
