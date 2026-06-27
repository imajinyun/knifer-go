package vstr

import "testing"

func TestBOMAndCharsetFacade(t *testing.T) {
	if HasBOM([]byte{0xFF, 0xFE, 0x00}) != BOMUTF16LE {
		t.Fatal("HasBOM failed")
	}
	gbk, err := FromUTF8([]byte("中文"), "gbk")
	if err != nil {
		t.Fatalf("FromUTF8 error = %v", err)
	}
	utf8, err := ToUTF8(gbk, "gbk")
	if err != nil || string(utf8) != "中文" {
		t.Fatalf("ToUTF8 = %q, %v", utf8, err)
	}
}
