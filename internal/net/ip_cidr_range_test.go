package net

import (
	"reflect"
	"testing"
)

func TestIPCIDRHelpers(t *testing.T) {
	if got, _ := BeginIP("192.168.1.9", 24); got != "192.168.1.0" {
		t.Fatalf("BeginIP = %q", got)
	}
	if got, _ := EndIP("192.168.1.9", 24); got != "192.168.1.255" {
		t.Fatalf("EndIP = %q", got)
	}
	if bit, _ := MaskBitByMask("255.255.255.0"); bit != 24 {
		t.Fatalf("MaskBitByMask = %d", bit)
	}
	if mask, _ := MaskByMaskBit(24); mask != "255.255.255.0" {
		t.Fatalf("MaskByMaskBit = %q", mask)
	}
	if count, _ := CountByMaskBit(30, false); count != 2 {
		t.Fatalf("CountByMaskBit = %d", count)
	}
	if ips, _ := ListIPCIDR("192.168.1.0", 30, false); !reflect.DeepEqual(ips, []string{"192.168.1.1", "192.168.1.2"}) {
		t.Fatalf("ListIPCIDR = %#v", ips)
	}
}

func TestIPRangeMatching(t *testing.T) {
	if !IsInRange("192.168.1.2", "192.168.1.0/24") {
		t.Fatal("IsInRange failed")
	}
}

func TestIPRangeMaskAndListHelpers(t *testing.T) {
	if got := IPv4ToLongDefault("bad", 99); got != 99 {
		t.Fatalf("IPv4ToLongDefault invalid = %d", got)
	}
	if got := IPv4ToLongDefault("127.0.0.1", 99); got == 99 {
		t.Fatalf("IPv4ToLongDefault valid returned default")
	}
	if got, err := FormatIPBlock("192.168.1.10", "255.255.255.0"); err != nil || got != "192.168.1.10/24" {
		t.Fatalf("FormatIPBlock = %q, %v", got, err)
	}
	if got, err := BeginIPLong("192.168.1.10", 24); err != nil || LongToIPv4(got) != "192.168.1.0" {
		t.Fatalf("BeginIPLong = %d/%s, %v", got, LongToIPv4(got), err)
	}
	if got, err := EndIPLong("192.168.1.10", 24); err != nil || LongToIPv4(got) != "192.168.1.255" {
		t.Fatalf("EndIPLong = %d/%s, %v", got, LongToIPv4(got), err)
	}
	if got, err := MaskByIPRange("192.168.1.255", "192.168.1.0"); err != nil || got != "255.255.255.0" {
		t.Fatalf("MaskByIPRange = %q, %v", got, err)
	}
	if got, err := CountByIPRange("192.168.1.3", "192.168.1.1"); err != nil || got != 3 {
		t.Fatalf("CountByIPRange = %d, %v", got, err)
	}
	if !IsMaskValid("255.255.255.0") || IsMaskValid("255.0.255.0") {
		t.Fatalf("IsMaskValid mismatch")
	}
	if got, err := ListIPs("192.168.1.1", false); err != nil || !reflect.DeepEqual(got, []string{"192.168.1.1"}) {
		t.Fatalf("ListIPs(single) = %#v, %v", got, err)
	}
	if got, err := ListIPs("192.168.1.1-192.168.1.3", false); err != nil || !reflect.DeepEqual(got, []string{"192.168.1.1", "192.168.1.2", "192.168.1.3"}) {
		t.Fatalf("ListIPs(range) = %#v, %v", got, err)
	}
	if got, err := ListIPs("192.168.1.0/30", false); err != nil || !reflect.DeepEqual(got, []string{"192.168.1.1", "192.168.1.2"}) {
		t.Fatalf("ListIPs(cidr) = %#v, %v", got, err)
	}
	if got, err := ListIPRange("192.168.1.3", "192.168.1.1"); err != nil || !reflect.DeepEqual(got, []string{"192.168.1.1", "192.168.1.2", "192.168.1.3"}) {
		t.Fatalf("ListIPRange = %#v, %v", got, err)
	}
}

func TestIPRangeErrorBoundaries(t *testing.T) {
	if _, err := FormatIPBlock("192.168.1.1", "bad"); err == nil {
		t.Fatalf("FormatIPBlock should reject invalid mask")
	}
	if _, err := BeginIP("bad", 24); err == nil {
		t.Fatalf("BeginIP should reject invalid IP")
	}
	if _, err := EndIP("192.168.1.1", 33); err == nil {
		t.Fatalf("EndIP should reject invalid mask bit")
	}
	if _, err := MaskByIPRange("bad", "192.168.1.1"); err == nil {
		t.Fatalf("MaskByIPRange should reject invalid from IP")
	}
	if _, err := CountByIPRange("192.168.1.1", "bad"); err == nil {
		t.Fatalf("CountByIPRange should reject invalid to IP")
	}
	if _, err := ListIPs("bad", false); err == nil {
		t.Fatalf("ListIPs should reject invalid single IP")
	}
	if _, err := ListIPs("192.168.1.0/bad", false); err == nil {
		t.Fatalf("ListIPs should reject invalid cidr bit")
	}
	if got, err := ListIPCIDR("192.168.1.0", 31, false); err != nil || !reflect.DeepEqual(got, []string{"192.168.1.0", "192.168.1.1"}) {
		t.Fatalf("ListIPCIDR /31 = %#v, %v", got, err)
	}
	if _, err := ListIPRange("192.0.0.0", "192.16.0.0"); err == nil {
		t.Fatalf("ListIPRange should reject oversized ranges")
	}
}
