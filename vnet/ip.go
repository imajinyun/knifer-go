package vnet

import (
	"math/big"

	netimpl "github.com/imajinyun/go-knifer/internal/net"
)

func LongToIPv4(longIP uint32) string { return netimpl.LongToIPv4(longIP) }

func IPv4ToLong(strIP string) (uint32, error) { return netimpl.IPv4ToLong(strIP) }

func IPv4ToLongDefault(strIP string, defaultValue uint32) uint32 {
	return netimpl.IPv4ToLongDefault(strIP, defaultValue)
}

func IPv6ToBigInt(ipv6Str string) (*big.Int, error) { return netimpl.IPv6ToBigInt(ipv6Str) }

func BigIntToIPv6(n *big.Int) (string, error) { return netimpl.BigIntToIPv6(n) }

func IsIP(s string) bool { return netimpl.IsIP(s) }

func IsIPv4(s string) bool { return netimpl.IsIPv4(s) }

func IsIPv6(s string) bool { return netimpl.IsIPv6(s) }

func IsInnerIP(ipAddress string) bool { return netimpl.IsInnerIP(ipAddress) }

func FormatIPBlock(ip, mask string) (string, error) { return netimpl.FormatIPBlock(ip, mask) }

func BeginIP(ip string, maskBit int) (string, error) { return netimpl.BeginIP(ip, maskBit) }

func BeginIPLong(ip string, maskBit int) (uint32, error) { return netimpl.BeginIPLong(ip, maskBit) }

func EndIP(ip string, maskBit int) (string, error) { return netimpl.EndIP(ip, maskBit) }

func EndIPLong(ip string, maskBit int) (uint32, error) { return netimpl.EndIPLong(ip, maskBit) }

func MaskBitByMask(mask string) (int, error) { return netimpl.MaskBitByMask(mask) }

func CountByMaskBit(maskBit int, isAll bool) (uint64, error) {
	return netimpl.CountByMaskBit(maskBit, isAll)
}

func MaskByMaskBit(maskBit int) (string, error) { return netimpl.MaskByMaskBit(maskBit) }

func MaskByIPRange(fromIP, toIP string) (string, error) { return netimpl.MaskByIPRange(fromIP, toIP) }

func CountByIPRange(fromIP, toIP string) (uint64, error) {
	return netimpl.CountByIPRange(fromIP, toIP)
}

func IsMaskValid(mask string) bool { return netimpl.IsMaskValid(mask) }

func IsMaskBitValid(maskBit int) bool { return netimpl.IsMaskBitValid(maskBit) }

func ListIPs(ipRange string, isAll bool) ([]string, error) { return netimpl.ListIPs(ipRange, isAll) }

func ListIPCIDR(ip string, maskBit int, isAll bool) ([]string, error) {
	return netimpl.ListIPCIDR(ip, maskBit, isAll)
}

func ListIPRange(fromIP, toIP string) ([]string, error) { return netimpl.ListIPRange(fromIP, toIP) }

func MatchesWildcard(wildcard, ipAddress string) bool {
	return netimpl.MatchesWildcard(wildcard, ipAddress)
}

func IsInRange(ip, cidr string) bool { return netimpl.IsInRange(ip, cidr) }

func HideIPPart(ip string) string { return netimpl.HideIPPart(ip) }

func HideIPPartLong(ip uint32) string { return netimpl.HideIPPartLong(ip) }
