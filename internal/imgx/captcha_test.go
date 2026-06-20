package imgx

import (
	"bytes"
	"errors"
	"image/color"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	knifer "github.com/imajinyun/go-knifer"
)

type fixedGenerator struct{ code string }

func (g fixedGenerator) Gen() string { return g.code }

func (g fixedGenerator) Verify(code, userInput string) bool { return code == userInput }

func TestCaptchaOptionsAndWriteOptions(t *testing.T) {
	colorCalls := 0
	randomCalls := 0
	c := NewLineCaptchaWithOptions(100, 40,
		WithGenerator(fixedGenerator{code: "ABCD"}),
		WithBackground(color.Black),
		WithInterfereCount(0),
		WithRandomInt(func(max int) int {
			randomCalls++
			return 0
		}),
		WithColorFunc(func() color.Color {
			colorCalls++
			return color.RGBA{R: 1, G: 2, B: 3, A: 255}
		}),
	)
	c.CreateCode()
	if c.Code() != "ABCD" || !c.Verify("ABCD") || c.Verify("abcd") {
		t.Fatalf("custom generator not applied: code=%q", c.Code())
	}
	if colorCalls != len("ABCD") {
		t.Fatalf("custom color func calls=%d, want %d", colorCalls, len("ABCD"))
	}
	if randomCalls != 0 {
		t.Fatalf("custom random func should not be called when interference is disabled and color func is set, got %d", randomCalls)
	}
	path := filepath.Join(t.TempDir(), "nested", "captcha.png")
	if err := c.WriteToFileWithOptions(path, WithFilePerm(0o600), WithDirPerm(0o700)); err != nil {
		t.Fatalf("WriteToFileWithOptions: %v", err)
	}
	if err := c.WriteToFileWithOptions(path, WithOverwrite(false)); err == nil {
		t.Fatal("WriteToFileWithOptions should reject overwrite when disabled")
	} else {
		if !errors.Is(err, fs.ErrExist) {
			t.Fatalf("WriteToFileWithOptions overwrite error = %v, want fs.ErrExist", err)
		}
		if !errors.Is(err, knifer.ErrCodeInvalidInput) {
			t.Fatalf("WriteToFileWithOptions overwrite error = %v, want ErrCodeInvalidInput", err)
		}
	}

	g := NewGifCaptchaWithOptions(100, 40, WithGenerator(fixedGenerator{code: "XYZ"}), WithGIFRepeat(1), WithGIFDelay(5))
	if g.Repeat != 1 || g.Delay != 5 || g.Code() != "XYZ" {
		t.Fatalf("gif options not applied: repeat=%d delay=%d code=%q", g.Repeat, g.Delay, g.Code())
	}
}

func TestCaptchaWriteProviderOptions(t *testing.T) {
	c := NewLineCaptchaWithOptions(100, 40, WithGenerator(fixedGenerator{code: "ABCD"}), WithInterfereCount(0))
	c.CreateCode()

	var mkdirPath string
	var mkdirPerm fs.FileMode
	var openPath string
	var openFlag int
	var openPerm fs.FileMode
	var written bytes.Buffer
	err := c.WriteToFileWithOptions("/virtual/captcha.png",
		WithMkdirAll(func(path string, perm fs.FileMode) error {
			mkdirPath, mkdirPerm = path, perm
			return nil
		}),
		WithOpenFile(func(path string, flag int, perm fs.FileMode) (io.WriteCloser, error) {
			openPath, openFlag, openPerm = path, flag, perm
			return nopWriteCloser{Writer: &written}, nil
		}),
		WithDirPerm(0o700), WithFilePerm(0o600),
	)
	if err != nil {
		t.Fatalf("WriteToFileWithOptions provider: %v", err)
	}
	if mkdirPath != "/virtual" || mkdirPerm != 0o700 || openPath != "/virtual/captcha.png" || openPerm != 0o600 || openFlag&os.O_CREATE == 0 || written.Len() == 0 {
		t.Fatalf("providers mkdir=%q/%v open=%q flag=%#x perm=%v bytes=%d", mkdirPath, mkdirPerm, openPath, openFlag, openPerm, written.Len())
	}
}

type nopWriteCloser struct{ io.Writer }

func (w nopWriteCloser) Close() error { return nil }

func TestICaptchaInterface(t *testing.T) {
	var _ ICaptcha = NewLineCaptcha(100, 40)
	var _ ICaptcha = NewCircleCaptcha(100, 40)
	var _ ICaptcha = NewShearCaptcha(100, 40)
	var _ ICaptcha = NewGifCaptcha(100, 40)
}

func TestImageBase64Data(t *testing.T) {
	c := NewLineCaptcha(100, 40)
	s := c.ImageBase64Data()
	if !strings.HasPrefix(s, "data:image/png;base64,") {
		t.Fatalf("unexpected data uri prefix: %q", s[:30])
	}
}

func TestCaptchaImageBytesDefensiveCopy(t *testing.T) {
	captchas := []struct {
		name string
		new  func() ICaptcha
	}{
		{name: "line", new: func() ICaptcha { return NewLineCaptcha(100, 40) }},
		{name: "circle", new: func() ICaptcha { return NewCircleCaptcha(100, 40) }},
		{name: "shear", new: func() ICaptcha { return NewShearCaptcha(100, 40) }},
		{name: "gif", new: func() ICaptcha { return NewGifCaptcha(100, 40) }},
	}
	for _, tc := range captchas {
		t.Run(tc.name, func(t *testing.T) {
			c := tc.new()
			original := c.ImageBytes()
			if len(original) == 0 {
				t.Fatal("ImageBytes returned empty data")
			}
			wantFirst := original[0]
			original[0] ^= 0xff
			got := c.ImageBytes()
			if got[0] != wantFirst {
				t.Fatalf("ImageBytes exposed internal backing array: got first byte %d, want %d", got[0], wantFirst)
			}
		})
	}
}

func TestCaptchaWriteErrorContract(t *testing.T) {
	c := NewLineCaptchaWithOptions(100, 40, WithGenerator(fixedGenerator{code: "ABCD"}))
	if err := c.Write(nil); !errors.Is(err, knifer.ErrCodeInvalidInput) {
		t.Fatalf("Write(nil) error = %v, want ErrCodeInvalidInput", err)
	}
	if err := c.Write(io.Discard); !errors.Is(err, knifer.ErrCodeInvalidInput) {
		t.Fatalf("Write empty image error = %v, want ErrCodeInvalidInput", err)
	}
}
