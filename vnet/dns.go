package vnet

import (
	"net/http"

	netimpl "github.com/imajinyun/go-knifer/internal/net"
)

func IDNToASCII(unicode string) (string, error) { return netimpl.IDNToASCII(unicode) }

func GetMultistageReverseProxyIP(ip string) string { return netimpl.GetMultistageReverseProxyIP(ip) }

func IsUnknown(checkString string) bool { return netimpl.IsUnknown(checkString) }

func ParseCookies(cookieStr string) []*http.Cookie { return netimpl.ParseCookies(cookieStr) }

func GetDNSInfo(hostName string, attrNames ...string) ([]string, error) {
	return netimpl.GetDNSInfo(hostName, attrNames...)
}
