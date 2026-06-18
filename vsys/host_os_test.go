package vsys_test

import (
	"net"
	"testing"

	"github.com/imajinyun/go-knifer/vsys"
)

func TestFacadeHostInfo(t *testing.T) {
	info := vsys.SystemHostInfo()
	if info == nil {
		t.Fatal("expected non-nil host info")
	}

	_, ipNet, err := net.ParseCIDR("10.1.2.3/24")
	if err != nil {
		t.Fatal(err)
	}
	ipNet.IP = net.ParseIP("10.1.2.3")
	info = vsys.SysHostInfoWithOptions(
		vsys.WithHostNameFunc(func() (string, error) { return "facade-host", nil }),
		vsys.WithHostInterfaceAddrsFunc(func() ([]net.Addr, error) { return []net.Addr{ipNet}, nil }),
	)
	if info.GetName() != "facade-host" || info.GetAddress() != "10.1.2.3" {
		t.Fatalf("SysHostInfoWithOptions = %#v", info)
	}

	info = vsys.NewHostInfoWithOptions(vsys.WithHostAddressFunc(func() string { return "198.51.100.2" }))
	if info.GetAddress() != "198.51.100.2" {
		t.Fatalf("NewHostInfoWithOptions address = %q", info.GetAddress())
	}
}

func TestFacadeNewHostInfo(t *testing.T) {
	info := vsys.NewHostInfo()
	if info == nil {
		t.Fatal("expected non-nil host info from NewHostInfo")
	}
	if info.GetName() == "" {
		t.Log("host name is empty (expected in some environments)")
	}
}

func TestFacadeGetHostInfo(t *testing.T) {
	info := vsys.GetHostInfo()
	if info == nil {
		t.Fatal("expected non-nil host info from GetHostInfo")
	}
	infoWithOpts := vsys.GetHostInfoWithOptions(vsys.WithHostNameFunc(func() (string, error) { return "get-host", nil }))
	if infoWithOpts.GetName() != "get-host" {
		t.Fatalf("GetHostInfoWithOptions name = %q", infoWithOpts.GetName())
	}
}

func TestFacadeNewOsInfo(t *testing.T) {
	info := vsys.NewOsInfo()
	if info == nil {
		t.Fatal("expected non-nil os info from NewOsInfo")
	}
}

func TestFacadeGetOsInfo(t *testing.T) {
	info := vsys.GetOsInfo()
	if info == nil {
		t.Fatal("expected non-nil os info from GetOsInfo")
	}
	infoWithOpts := vsys.GetOsInfoWithOptions(vsys.WithOSNameFunc(func() string { return "get-os" }))
	if infoWithOpts.GetName() != "get-os" {
		t.Fatalf("GetOsInfoWithOptions name = %q", infoWithOpts.GetName())
	}
}

func TestFacadeOsInfo(t *testing.T) {
	info := vsys.SystemOsInfo()
	if info == nil {
		t.Fatal("expected non-nil os info")
	}
	info = vsys.SysOsInfoWithOptions(vsys.WithOSNameFunc(func() string { return "linux" }))
	if info.GetName() != "linux" {
		t.Fatalf("SysOsInfoWithOptions name = %q", info.GetName())
	}
	info = vsys.SystemOsInfoWithOptions(vsys.WithOSNameFunc(func() string { return "windows" }))
	if info.GetName() != "windows" {
		t.Fatalf("SystemOsInfoWithOptions name = %q", info.GetName())
	}
}

func TestFacadeOsInfoOptions(t *testing.T) {
	info := vsys.NewOsInfoWithOptions(
		vsys.WithOSNameFunc(func() string { return "windows" }),
		vsys.WithOSArchFunc(func() string { return "amd64" }),
		vsys.WithOSVersionFunc(func() string { return "11" }),
		vsys.WithOSFileSeparatorFunc(func() string { return "\\" }),
		vsys.WithOSLineSeparatorFunc(func() string { return "\r\n" }),
		vsys.WithOSPathSeparatorFunc(func() string { return ";" }),
	)
	if info.GetName() != "windows" || info.GetArch() != "amd64" || info.GetVersion() != "11" || info.GetFileSeparator() != "\\" || info.GetLineSeparator() != "\r\n" || info.GetPathSeparator() != ";" {
		t.Fatalf("NewOsInfoWithOptions = %#v", info)
	}
	if !info.IsWindows() || info.IsLinux() {
		t.Fatalf("NewOsInfoWithOptions OS helpers = %#v", info)
	}
}
