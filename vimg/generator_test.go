package vimg_test

import (
	"testing"

	"github.com/imajinyun/knifer-go/vimg"
)

func TestFacadeRandomGenerator(t *testing.T) {
	g := vimg.NewRandomGenerator(4)
	code := g.Gen()
	if len(code) != 4 {
		t.Fatalf("expected code length 4, got %d", len(code))
	}
	if !g.Verify(code, code) {
		t.Fatal("expected generated code to verify")
	}
	if g.Verify(code, "wrong") {
		t.Fatal("expected wrong code to fail verification")
	}
}

func TestFacadeRandomGeneratorOptions(t *testing.T) {
	g := vimg.NewRandomGeneratorWithBase("abcd", 4)
	idx := 0
	code := vimg.GenRandomGeneratorWithOptions(g, vimg.WithGeneratorRandomInt(func(max int) int {
		v := idx
		idx++
		return v % max
	}))
	if code != "abcd" {
		t.Fatalf("GenRandomGeneratorWithOptions = %q, want abcd", code)
	}
}

func TestFacadeMathGenerator(t *testing.T) {
	g := vimg.NewMathGenerator()
	code := g.Gen()
	if len(code) == 0 {
		t.Fatal("expected non-empty math code")
	}
	// MathGenerator produces expressions like "1+2="; Verify needs the computed answer.
	// We just smoke-test that generation and verification accept a correct answer.
	if !g.Verify("1+1=", "2") {
		t.Fatal("expected 1+1= to verify with answer 2")
	}
}

func TestFacadeMathGeneratorOptions(t *testing.T) {
	g := vimg.NewMathGeneratorWith(1, false)
	values := []int{1, 7, 3}
	idx := 0
	code := vimg.GenMathGeneratorWithOptions(g, vimg.WithGeneratorRandomInt(func(max int) int {
		v := values[idx]
		idx++
		return v % max
	}))
	if code != "7-3=" {
		t.Fatalf("GenMathGeneratorWithOptions = %q, want 7-3=", code)
	}
	if !g.Verify(code, "4") {
		t.Fatalf("generated math code should verify: %q", code)
	}
}

func TestFacadeVerifyIgnoreCase(t *testing.T) {
	if !vimg.VerifyCaptchaIgnoreCase("ABC", "abc") {
		t.Fatal("expected case-insensitive verification to pass")
	}
	if vimg.VerifyCaptchaIgnoreCase("ABC", "def") {
		t.Fatal("expected different code to fail verification")
	}
}
