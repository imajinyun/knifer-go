package net

import (
	"math/big"
	"testing"
)

func TestIPv4Conversion(t *testing.T) {
	v, err := IPv4ToLong("127.0.0.1")
	if err != nil || v != 2130706433 {
		t.Fatalf("IPv4ToLong = %d %v", v, err)
	}
	if got := LongToIPv4(v); got != "127.0.0.1" {
		t.Fatalf("LongToIPv4 = %q", got)
	}
}

func TestIPv6BigInt(t *testing.T) {
	v, err := IPv6ToBigInt("::1")
	if err != nil || v.Cmp(big.NewInt(1)) != 0 {
		t.Fatalf("IPv6ToBigInt = %v %v", v, err)
	}
	if got, err := BigIntToIPv6(big.NewInt(1)); err != nil || got != "::1" {
		t.Fatalf("BigIntToIPv6 = %q %v", got, err)
	}
}
