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

	knifer "github.com/imajinyun/knifer-go"
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

func TestNilCaptchaWriteProviderOptionsDoNotOverwriteConfiguredProviders(t *testing.T) {
	mkdirAll := func(string, fs.FileMode) error { return nil }
	openFile := func(string, int, fs.FileMode) (io.WriteCloser, error) {
		return nopWriteCloser{Writer: io.Discard}, nil
	}
	cfg := applyWriteOptions([]WriteOption{
		WithMkdirAll(mkdirAll),
		WithMkdirAll(nil),
		WithOpenFile(openFile),
		WithOpenFile(nil),
	})
	if cfg.mkdirAll == nil || cfg.openFile == nil {
		t.Fatalf("nil write provider option overwrote configured provider: %#v", cfg)
	}
}

type nopWriteCloser struct{ io.Writer }

func (w nopWriteCloser) Close() error { return nil }

type closeErrorWriteCloser struct {
	io.Writer
	err error
}

func (w closeErrorWriteCloser) Close() error { return w.err }

func TestCaptchaWriteToFileReturnsCloseError(t *testing.T) {
	closeErr := errors.New("close failed")
	c := NewLineCaptchaWithOptions(100, 40, WithGenerator(fixedGenerator{code: "ABCD"}), WithInterfereCount(0))
	c.CreateCode()

	err := c.WriteToFileWithOptions("/virtual/captcha.png",
		WithMkdirAll(func(string, fs.FileMode) error { return nil }),
		WithOpenFile(func(string, int, fs.FileMode) (io.WriteCloser, error) {
			return closeErrorWriteCloser{Writer: io.Discard, err: closeErr}, nil
		}),
	)
	if !errors.Is(err, closeErr) {
		t.Fatalf("WriteToFileWithOptions close error = %v, want close cause", err)
	}
}

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

func TestCaptchaAdditionalOptionsAndHelpers(t *testing.T) {
	c := NewLineCaptchaWithOptions(100, 40,
		nil,
		WithFontSize(0.6),
		WithGenerator(fixedGenerator{code: "AbCd"}),
		WithInterfereCount(0),
		WithRandomInt(func(max int) int { return max + 1 }),
	)
	if c.FontSize != 0.6 {
		t.Fatalf("FontSize = %v, want 0.6", c.FontSize)
	}
	if c.Generator() == nil {
		t.Fatal("Generator should return configured generator")
	}
	if got := c.randInt(5); got != 1 {
		t.Fatalf("randInt wraps provider output = %d, want 1", got)
	}
	if got := (&AbstractCaptcha{}).randInt(0); got != 0 {
		t.Fatalf("randInt(0) = %d, want 0", got)
	}
	c.randomInt = func(max int) int { return -3 }
	if got := c.randInt(5); got != 0 {
		t.Fatalf("randInt clamps negative provider output = %d, want 0", got)
	}
	if got := c.randIntRange(7, 7); got != 7 {
		t.Fatalf("randIntRange equal bounds = %d, want 7", got)
	}
	if !VerifyIgnoreCase(" AbCd ", "abcd") || VerifyIgnoreCase("AbCd", "abce") {
		t.Fatal("VerifyIgnoreCase should trim and compare case-insensitively")
	}
}

func TestCaptchaWriteToFileWithOptionsSkipsParentCreation(t *testing.T) {
	c := NewLineCaptchaWithOptions(100, 40, WithGenerator(fixedGenerator{code: "ABCD"}), WithInterfereCount(0))
	c.CreateCode()

	mkdirCalled := false
	var openFlag int
	var written bytes.Buffer
	err := c.WriteToFileWithOptions("/virtual/captcha.png",
		WithCreateParents(false),
		WithMkdirAll(func(path string, perm fs.FileMode) error {
			mkdirCalled = true
			return nil
		}),
		WithOpenFile(func(path string, flag int, perm fs.FileMode) (io.WriteCloser, error) {
			openFlag = flag
			return nopWriteCloser{Writer: &written}, nil
		}),
		nil,
	)
	if err != nil {
		t.Fatalf("WriteToFileWithOptions: %v", err)
	}
	if mkdirCalled {
		t.Fatal("mkdirAll should not be called when parent creation is disabled")
	}
	if openFlag&os.O_EXCL != 0 || written.Len() == 0 {
		t.Fatalf("open flag = %#x written=%d, want no O_EXCL and non-empty data", openFlag, written.Len())
	}
}

func TestAbstractCaptchaImageBytesDefensiveCopy(t *testing.T) {
	a := &AbstractCaptcha{}
	if got := a.ImageBytes(); got != nil {
		t.Fatalf("empty ImageBytes = %v, want nil", got)
	}
	a.setImageBytes([]byte{1, 2, 3})
	got := a.ImageBytes()
	got[0] = 9
	if a.ImageBytes()[0] != 1 {
		t.Fatal("AbstractCaptcha.ImageBytes exposed internal backing array")
	}

	path := filepath.Join(t.TempDir(), "captcha.png")
	if err := a.WriteToFile(path); err != nil {
		t.Fatalf("WriteToFile: %v", err)
	}
}
