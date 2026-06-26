package vnet_test

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"mime/multipart"
	"net"
	"net/http"

	"github.com/imajinyun/knifer-go/vnet"
)

func ExampleCreateAddress() {
	addr := vnet.CreateAddress("127.0.0.1", 8080)
	fmt.Println(addr.String())
	// Output: 127.0.0.1:8080
}

func ExampleIPv4ToLong() {
	n, err := vnet.IPv4ToLong("192.0.2.1")

	fmt.Println(n)
	fmt.Println(vnet.LongToIPv4(n))
	fmt.Println(err)
	// Output:
	// 3221225985
	// 192.0.2.1
	// <nil>
}

func ExampleLongToIPv4() {
	fmt.Println(vnet.LongToIPv4(3221225985))
	// Output: 192.0.2.1
}

func ExampleIPv4ToLongDefault() {
	fmt.Println(vnet.IPv4ToLongDefault("not-an-ip", 42))
	// Output: 42
}

func ExampleIPv6ToBigInt() {
	n, err := vnet.IPv6ToBigInt("2001:db8::1")
	fmt.Println(n.String())
	fmt.Println(err)
	// Output:
	// 42540766411282592856903984951653826561
	// <nil>
}

func ExampleBigIntToIPv6() {
	n, _ := vnet.IPv6ToBigInt("2001:db8::1")
	ip, err := vnet.BigIntToIPv6(n)
	fmt.Println(ip)
	fmt.Println(err)
	// Output:
	// 2001:db8::1
	// <nil>
}

func ExampleIsIP() {
	fmt.Println(vnet.IsIP("192.0.2.1"))
	fmt.Println(vnet.IsIP("example"))
	// Output:
	// true
	// false
}

func ExampleIsIPv4() {
	fmt.Println(vnet.IsIPv4("192.0.2.1"))
	fmt.Println(vnet.IsIPv4("2001:db8::1"))
	// Output:
	// true
	// false
}

func ExampleIsIPv6() {
	fmt.Println(vnet.IsIPv6("2001:db8::1"))
	fmt.Println(vnet.IsIPv6("192.0.2.1"))
	// Output:
	// true
	// false
}

func ExampleIsInnerIP() {
	fmt.Println(vnet.IsInnerIP("10.0.0.1"))
	fmt.Println(vnet.IsInnerIP("203.0.113.1"))
	// Output:
	// true
	// false
}

func ExampleFormatIPBlock() {
	block, err := vnet.FormatIPBlock("192.0.2.9", "255.255.255.0")
	fmt.Println(block)
	fmt.Println(err)
	// Output:
	// 192.0.2.9/24
	// <nil>
}

func ExampleBeginIP() {
	begin, beginErr := vnet.BeginIP("192.0.2.9", 24)
	end, endErr := vnet.EndIP("192.0.2.9", 24)
	count, countErr := vnet.CountByMaskBit(24, true)

	fmt.Println(begin, beginErr)
	fmt.Println(end, endErr)
	fmt.Println(count, countErr)
	// Output:
	// 192.0.2.0 <nil>
	// 192.0.2.255 <nil>
	// 256 <nil>
}

func ExampleBeginIPLong() {
	n, err := vnet.BeginIPLong("192.0.2.9", 24)
	fmt.Println(n)
	fmt.Println(vnet.LongToIPv4(n))
	fmt.Println(err)
	// Output:
	// 3221225984
	// 192.0.2.0
	// <nil>
}

func ExampleEndIP() {
	ip, err := vnet.EndIP("192.0.2.9", 24)
	fmt.Println(ip)
	fmt.Println(err)
	// Output:
	// 192.0.2.255
	// <nil>
}

func ExampleEndIPLong() {
	n, err := vnet.EndIPLong("192.0.2.9", 24)
	fmt.Println(n)
	fmt.Println(vnet.LongToIPv4(n))
	fmt.Println(err)
	// Output:
	// 3221226239
	// 192.0.2.255
	// <nil>
}

func ExampleMaskBitByMask() {
	maskBit, err := vnet.MaskBitByMask("255.255.255.0")
	fmt.Println(maskBit, err)
	// Output: 24 <nil>
}

func ExampleCountByMaskBit() {
	all, _ := vnet.CountByMaskBit(30, true)
	usable, _ := vnet.CountByMaskBit(30, false)
	fmt.Println(all)
	fmt.Println(usable)
	// Output:
	// 4
	// 2
}

func ExampleMaskByMaskBit() {
	mask, err := vnet.MaskByMaskBit(24)
	fmt.Println(mask)
	fmt.Println(err)
	// Output:
	// 255.255.255.0
	// <nil>
}

func ExampleMaskByIPRange() {
	mask, err := vnet.MaskByIPRange("192.0.2.0", "192.0.2.255")
	fmt.Println(mask)
	fmt.Println(err)
	// Output:
	// 255.255.255.0
	// <nil>
}

func ExampleCountByIPRange() {
	count, err := vnet.CountByIPRange("192.0.2.1", "192.0.2.4")
	fmt.Println(count)
	fmt.Println(err)
	// Output:
	// 4
	// <nil>
}

func ExampleIsMaskValid() {
	fmt.Println(vnet.IsMaskValid("255.255.255.0"))
	fmt.Println(vnet.IsMaskValid("255.0.255.0"))
	// Output:
	// true
	// false
}

func ExampleIsMaskBitValid() {
	fmt.Println(vnet.IsMaskBitValid(24))
	fmt.Println(vnet.IsMaskBitValid(40))
	// Output:
	// true
	// false
}

func ExampleListIPs() {
	ips, err := vnet.ListIPs("192.0.2.1-192.0.2.3", true)
	fmt.Println(ips)
	fmt.Println(err)
	// Output:
	// [192.0.2.1 192.0.2.2 192.0.2.3]
	// <nil>
}

func ExampleListIPCIDR() {
	ips, err := vnet.ListIPCIDR("192.0.2.0", 30, false)
	fmt.Println(ips)
	fmt.Println(err)
	// Output:
	// [192.0.2.1 192.0.2.2]
	// <nil>
}

func ExampleListIPRange() {
	ips, err := vnet.ListIPRange("192.0.2.1", "192.0.2.2")
	fmt.Println(ips)
	fmt.Println(err)
	// Output:
	// [192.0.2.1 192.0.2.2]
	// <nil>
}

func ExampleParseCookies() {
	cookies := vnet.ParseCookies("sid=abc; theme=dark")
	for _, cookie := range cookies {
		fmt.Println(cookie.Name, cookie.Value)
	}
	// Output:
	// sid abc
	// theme dark
}

func ExampleMatchesWildcard() {
	fmt.Println(vnet.MatchesWildcard("192.168.*.*", "192.168.1.2"))
	fmt.Println(vnet.MatchesWildcard("10.0.*.*", "192.168.1.2"))
	// Output:
	// true
	// false
}

func ExampleMatchesWildcardWithOptions() {
	parseIP := vnet.WithWildcardIPParser(func(string) net.IP {
		return net.ParseIP("203.0.113.9")
	})

	fmt.Println(vnet.MatchesWildcardWithOptions("203.0.113.*", "ignored", parseIP))
	// Output: true
}

func ExampleIsInRange() {
	fmt.Println(vnet.IsInRange("192.0.2.10", "192.0.2.0/24"))
	fmt.Println(vnet.IsInRange("198.51.100.10", "192.0.2.0/24"))
	// Output:
	// true
	// false
}

func ExampleIsInRangeWithOptions() {
	parseIP := vnet.WithIPParser(func(string) net.IP {
		return net.ParseIP("192.0.2.10")
	})

	fmt.Println(vnet.IsInRangeWithOptions("ignored", "192.0.2.0/24", parseIP))
	// Output: true
}

func ExampleHideIPPart() {
	fmt.Println(vnet.HideIPPart("192.0.2.99"))
	// Output: 192.0.2.*
}

func ExampleHideIPPartLong() {
	n, _ := vnet.IPv4ToLong("192.0.2.99")
	fmt.Println(vnet.HideIPPartLong(n))
	// Output: 192.0.2.*
}

func ExampleIDNToASCII() {
	domain, err := vnet.IDNToASCII("例子.测试")
	fmt.Println(domain)
	fmt.Println(err)
	// Output:
	// xn--fsqu00a.xn--0zwm56d
	// <nil>
}

func ExampleGetMultistageReverseProxyIP() {
	fmt.Println(vnet.GetMultistageReverseProxyIP("unknown, 198.51.100.7, 203.0.113.9"))
	// Output: 198.51.100.7
}

func ExampleIsUnknown() {
	fmt.Println(vnet.IsUnknown("unknown"))
	fmt.Println(vnet.IsUnknown("198.51.100.7"))
	// Output:
	// true
	// false
}

func ExampleIsValidPort() {
	fmt.Println(vnet.IsValidPort(443))
	fmt.Println(vnet.IsValidPort(70000))
	// Output:
	// true
	// false
}

func ExampleToIPList() {
	ips := vnet.ToIPList([]net.IP{
		net.ParseIP("192.0.2.1"),
		net.ParseIP("192.0.2.1"),
		net.ParseIP("2001:db8::1"),
	})

	fmt.Println(ips)
	// Output: [192.0.2.1 2001:db8::1]
}

func ExampleBuildInetSocketAddressWithOptions() {
	resolver := vnet.WithTCPAddrResolver(func(network, address string) (*net.TCPAddr, error) {
		fmt.Println(network, address)
		return &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080}, nil
	})

	addr, err := vnet.BuildInetSocketAddressWithOptions("example.test", 8080, resolver)
	fmt.Println(addr.String())
	fmt.Println(err)
	// Output:
	// tcp example.test:8080
	// 127.0.0.1:8080
	// <nil>
}

func ExampleCreateAddressWithOptions() {
	resolver := vnet.WithTCPAddrResolver(func(network, address string) (*net.TCPAddr, error) {
		fmt.Println(network, address)
		return &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 9000}, nil
	})

	addr, err := vnet.CreateAddressWithOptions("localhost", 9000, resolver)
	fmt.Println(addr.String())
	fmt.Println(err)
	// Output:
	// tcp localhost:9000
	// 127.0.0.1:9000
	// <nil>
}

func ExampleGetNetworkInterfacesWithOptions() {
	interfaces, err := vnet.GetNetworkInterfacesWithOptions(vnet.WithInterfacesFunc(func() ([]net.Interface, error) {
		return []net.Interface{{Name: "eth0", Index: 1}}, nil
	}))

	fmt.Println(interfaces[0].Name)
	fmt.Println(err)
	// Output:
	// eth0
	// <nil>
}

func ExampleLocalIPsWithOptions() {
	interfaces := vnet.WithInterfacesFunc(func() ([]net.Interface, error) {
		return []net.Interface{{Name: "eth0", Index: 1}}, nil
	})
	addrs := vnet.WithInterfaceAddrsFunc(func(net.Interface) ([]net.Addr, error) {
		_, ipNet, _ := net.ParseCIDR("192.0.2.10/24")
		ipNet.IP = net.ParseIP("192.0.2.10")
		return []net.Addr{ipNet}, nil
	})

	fmt.Println(vnet.LocalIPsWithOptions(interfaces, addrs))
	// Output: [192.0.2.10]
}

func ExampleGetLocalHostNameWithOptions() {
	reverse := vnet.WithReverseLookupFunc(func(string) ([]string, error) {
		return []string{"example.local."}, nil
	})
	interfaces := vnet.WithInterfacesFunc(func() ([]net.Interface, error) {
		return []net.Interface{{Name: "eth0", Index: 1}}, nil
	})
	addrs := vnet.WithInterfaceAddrsFunc(func(net.Interface) ([]net.Addr, error) {
		_, ipNet, _ := net.ParseCIDR("192.0.2.10/24")
		ipNet.IP = net.ParseIP("192.0.2.10")
		return []net.Addr{ipNet}, nil
	})

	fmt.Println(vnet.GetLocalHostNameWithOptions(reverse, interfaces, addrs))
	// Output: example.local
}

func ExampleTLSVersion() {
	fmt.Println(vnet.TLSVersion("TLSv1.2") == tls.VersionTLS12)
	fmt.Println(vnet.TLSVersion("TLSv1.3") == tls.VersionTLS13)
	// Output:
	// true
	// true
}

func ExampleCreateTLSConfig() {
	config := vnet.CreateTLSConfig()
	fmt.Println(config.MinVersion == tls.VersionTLS12)
	// Output: true
}

func ExampleNewCertPool() {
	pool := vnet.NewCertPool()
	fmt.Println(pool != nil)
	// Output: true
}

func ExampleNewUploadSetting() {
	setting := vnet.NewUploadSetting()
	fmt.Println(setting.MaxFileSize > 0)
	fmt.Println(setting.AllowFileExts)
	// Output:
	// true
	// true
}

func ExampleParseMultipartForm() {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	_ = writer.WriteField("name", "gopher")
	part, _ := writer.CreateFormFile("avatar", "avatar.txt")
	_, _ = part.Write([]byte("hello"))
	_ = writer.Close()

	request, _ := http.NewRequest(http.MethodPost, "/upload", &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	form, err := vnet.ParseMultipartForm(request, vnet.NewUploadSetting())

	fmt.Println(form.GetParam("name"))
	fmt.Println(vnet.UploadFileName(form.Form.File["avatar"][0]))
	fmt.Println(vnet.UploadFileSize(form.Form.File["avatar"][0]))
	fmt.Println(err)
	// Output:
	// gopher
	// avatar.txt
	// 5
	// <nil>
}
