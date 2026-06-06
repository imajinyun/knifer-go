package vcaptcha_test

import (
	"image/color"
	"os"
	"path/filepath"
	"testing"

	"github.com/imajinyun/go-knifer/vcaptcha"
)

type fixedGenerator struct{ code string }

func (g fixedGenerator) Generate() string { return g.code }

func (g fixedGenerator) Verify(code, userInput string) bool { return code == userInput }

func TestFacadeRandomGenerator(t *testing.T) {
	g := vcaptcha.NewRandomGenerator(4)
	code := g.Generate()
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
	g := vcaptcha.NewRandomGeneratorWithBase("abcd", 4)
	idx := 0
	code := vcaptcha.GenRandomGeneratorWithOptions(g, vcaptcha.WithGeneratorRandomInt(func(max int) int {
		v := idx
		idx++
		return v % max
	}))
	if code != "abcd" {
		t.Fatalf("GenRandomGeneratorWithOptions = %q, want abcd", code)
	}
}

func TestFacadeMathGenerator(t *testing.T) {
	g := vcaptcha.NewMathGenerator()
	code := g.Generate()
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
	g := vcaptcha.NewMathGeneratorWith(1, false)
	values := []int{1, 7, 3}
	idx := 0
	code := vcaptcha.GenMathGeneratorWithOptions(g, vcaptcha.WithGeneratorRandomInt(func(max int) int {
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
	if !vcaptcha.VerifyCaptchaIgnoreCase("ABC", "abc") {
		t.Fatal("expected case-insensitive verification to pass")
	}
	if vcaptcha.VerifyCaptchaIgnoreCase("ABC", "def") {
		t.Fatal("expected different code to fail verification")
	}
}

func TestFacadeLineCaptcha(t *testing.T) {
	c := vcaptcha.NewLineCaptcha(100, 40)
	if c == nil {
		t.Fatal("expected non-nil line captcha")
	}
	c.CreateCode()
	code := c.Code()
	if len(code) == 0 {
		t.Fatal("expected line captcha to have non-empty code")
	}
}

func TestFacadeCircleCaptcha(t *testing.T) {
	c := vcaptcha.NewCircleCaptcha(100, 40)
	if c == nil {
		t.Fatal("expected non-nil circle captcha")
	}
	c.CreateCode()
	code := c.Code()
	if len(code) == 0 {
		t.Fatal("expected circle captcha to have non-empty code")
	}
}

func TestFacadeCaptchaOptions(t *testing.T) {
	colorCalls := 0
	line := vcaptcha.NewLineCaptchaWithOptions(100, 40,
		vcaptcha.WithGenerator(fixedGenerator{code: "ABCD"}),
		vcaptcha.WithBackground(color.Black),
		vcaptcha.WithInterfereCount(0),
		vcaptcha.WithRandomInt(func(max int) int { return 0 }),
		vcaptcha.WithColorFunc(func() color.Color {
			colorCalls++
			return color.RGBA{R: 1, G: 2, B: 3, A: 255}
		}),
	)
	if got := line.Code(); got != "ABCD" {
		t.Fatalf("line captcha code = %q, want ABCD", got)
	}
	if !line.Verify("ABCD") {
		t.Fatal("line captcha should verify fixed code")
	}
	if colorCalls != len("ABCD") {
		t.Fatalf("custom color func calls=%d, want %d", colorCalls, len("ABCD"))
	}

	circle := vcaptcha.NewCircleCaptchaWithOptions(100, 40, vcaptcha.WithGenerator(fixedGenerator{code: "WXYZ"}))
	if got := circle.Code(); got != "WXYZ" {
		t.Fatalf("circle captcha code = %q, want WXYZ", got)
	}
	shear := vcaptcha.NewShearCaptchaWithOptions(100, 40, vcaptcha.WithGenerator(fixedGenerator{code: "EFGH"}))
	if got := shear.Code(); got != "EFGH" {
		t.Fatalf("shear captcha code = %q, want EFGH", got)
	}
	gif := vcaptcha.NewGifCaptchaWithOptions(100, 40, vcaptcha.WithGenerator(fixedGenerator{code: "IJKL"}), vcaptcha.WithGIFRepeat(1), vcaptcha.WithGIFDelay(5))
	if got := gif.Code(); got != "IJKL" {
		t.Fatalf("gif captcha code = %q, want IJKL", got)
	}
	if gif.Repeat != 1 || gif.Delay != 5 {
		t.Fatalf("gif options not applied: repeat=%d delay=%d", gif.Repeat, gif.Delay)
	}
}

func TestFacadeCaptchaWriteOptions(t *testing.T) {
	c := vcaptcha.NewLineCaptchaWithOptions(100, 40, vcaptcha.WithGenerator(fixedGenerator{code: "ABCD"}))
	c.CreateCode()
	path := filepath.Join(t.TempDir(), "nested", "captcha.png")
	if err := c.WriteToFileWithOptions(path, vcaptcha.WithFilePerm(0o600), vcaptcha.WithDirPerm(0o700)); err != nil {
		t.Fatalf("WriteToFileWithOptions: %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat captcha file: %v", err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("captcha file perm = %o, want 600", got)
	}
	if err := c.WriteToFileWithOptions(path, vcaptcha.WithOverwrite(false)); err == nil {
		t.Fatal("WriteToFileWithOptions should reject overwrite=false for existing file")
	}
}
