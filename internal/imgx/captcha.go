package imgx

import (
	"encoding/base64"
	"image/color"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	knifer "github.com/imajinyun/go-knifer"
)

// ICaptcha mirrors the ICaptcha interface.
type ICaptcha interface {
	// CreateCode generates the captcha text and renders the image.
	CreateCode()
	// Code returns the captcha text.
	Code() string
	// Verify reports whether the user input is valid, usually case-insensitively.
	Verify(userInputCode string) bool
	// ImageBytes returns the encoded image bytes.
	ImageBytes() []byte
	// ImageBase64 returns the Base64-encoded image.
	ImageBase64() string
	// ImageBase64Data returns the Base64 image with a data URI prefix.
	ImageBase64Data() string
	// Write writes the image to an io.Writer.
	Write(w io.Writer) error
	// WriteToFile writes the image to a file path.
	WriteToFile(path string) error
}

// AbstractCaptcha mirrors the utility captcha AbstractCaptcha and holds shared captcha state.
type AbstractCaptcha struct {
	Width          int         // Image width.
	Height         int         // Image height.
	InterfereCount int         // Number of interference elements.
	FontSize       float64     // Font size ratio against Height; default is 0.75.
	Background     color.Color // Background color; nil means white.

	generator CodeGenerator
	randomInt func(max int) int
	colorFunc func() color.Color
	code      string
	imgBytes  []byte
}

type captchaConfig struct {
	generator      CodeGenerator
	background     color.Color
	fontSize       float64
	interfereCount int
	setInterfere   bool
	gifRepeat      int
	setGIFRepeat   bool
	gifDelay       int
	setGIFDelay    bool
	randomInt      func(max int) int
	colorFunc      func() color.Color
}

// CaptchaOption customizes captcha construction.
type CaptchaOption func(*captchaConfig)

// WithGenerator sets the captcha code generator.
func WithGenerator(generator CodeGenerator) CaptchaOption {
	return func(c *captchaConfig) { c.generator = generator }
}

// WithBackground sets the captcha background color.
func WithBackground(background color.Color) CaptchaOption {
	return func(c *captchaConfig) { c.background = background }
}

// WithFontSize sets the font size ratio against captcha height.
func WithFontSize(fontSize float64) CaptchaOption {
	return func(c *captchaConfig) { c.fontSize = fontSize }
}

// WithInterfereCount sets the number of interference elements.
func WithInterfereCount(count int) CaptchaOption {
	return func(c *captchaConfig) {
		c.interfereCount = count
		c.setInterfere = true
	}
}

// WithGIFRepeat sets the animated GIF repeat count.
func WithGIFRepeat(repeat int) CaptchaOption {
	return func(c *captchaConfig) {
		c.gifRepeat = repeat
		c.setGIFRepeat = true
	}
}

// WithGIFDelay sets the animated GIF frame delay in 1/100 second units.
func WithGIFDelay(delay int) CaptchaOption {
	return func(c *captchaConfig) {
		c.gifDelay = delay
		c.setGIFDelay = true
	}
}

// WithRandomInt sets the random integer function used while rendering captcha images.
func WithRandomInt(randomInt func(max int) int) CaptchaOption {
	return func(c *captchaConfig) { c.randomInt = randomInt }
}

// WithColorFunc sets the color function used while rendering captcha images.
func WithColorFunc(colorFunc func() color.Color) CaptchaOption {
	return func(c *captchaConfig) { c.colorFunc = colorFunc }
}

func applyCaptchaOptions(c *AbstractCaptcha, opts []CaptchaOption) captchaConfig {
	cfg := captchaConfig{}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if cfg.generator != nil {
		c.SetGenerator(cfg.generator)
	}
	if cfg.background != nil {
		c.SetBackground(cfg.background)
	}
	if cfg.fontSize > 0 {
		c.FontSize = cfg.fontSize
	}
	if cfg.setInterfere {
		c.InterfereCount = cfg.interfereCount
	}
	if cfg.randomInt != nil {
		c.randomInt = cfg.randomInt
	}
	if cfg.colorFunc != nil {
		c.colorFunc = cfg.colorFunc
	}
	return cfg
}

type writeConfig struct {
	filePerm      fs.FileMode
	dirPerm       fs.FileMode
	overwrite     bool
	createParents bool
	mkdirAll      func(string, fs.FileMode) error
	openFile      func(string, int, fs.FileMode) (io.WriteCloser, error)
}

// WriteOption customizes captcha file output.
type WriteOption func(*writeConfig)

// WithFilePerm sets the file permission used by WriteToFileWithOptions.
func WithFilePerm(perm fs.FileMode) WriteOption { return func(c *writeConfig) { c.filePerm = perm } }

// WithDirPerm sets the parent directory permission used by WriteToFileWithOptions.
func WithDirPerm(perm fs.FileMode) WriteOption { return func(c *writeConfig) { c.dirPerm = perm } }

// WithOverwrite controls whether WriteToFileWithOptions may replace an existing file.
func WithOverwrite(overwrite bool) WriteOption {
	return func(c *writeConfig) { c.overwrite = overwrite }
}

// WithCreateParents controls whether WriteToFileWithOptions creates parent directories.
func WithCreateParents(create bool) WriteOption {
	return func(c *writeConfig) { c.createParents = create }
}

// WithMkdirAll sets the directory creator used by WriteToFileWithOptions.
func WithMkdirAll(mkdirAll func(string, fs.FileMode) error) WriteOption {
	return func(c *writeConfig) { c.mkdirAll = mkdirAll }
}

// WithOpenFile sets the file opener used by WriteToFileWithOptions.
func WithOpenFile(openFile func(string, int, fs.FileMode) (io.WriteCloser, error)) WriteOption {
	return func(c *writeConfig) { c.openFile = openFile }
}

func applyWriteOptions(opts []WriteOption) writeConfig {
	cfg := writeConfig{filePerm: 0o644, dirPerm: 0o750, overwrite: true, createParents: true, mkdirAll: os.MkdirAll, openFile: defaultOpenWriteFile}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if cfg.mkdirAll == nil {
		cfg.mkdirAll = os.MkdirAll
	}
	if cfg.openFile == nil {
		cfg.openFile = defaultOpenWriteFile
	}
	return cfg
}

func defaultOpenWriteFile(path string, flag int, perm fs.FileMode) (io.WriteCloser, error) {
	// #nosec G304 -- captcha file output intentionally writes to the caller-provided destination path.
	return os.OpenFile(path, flag, perm)
}

// Code returns the current captcha text.
func (a *AbstractCaptcha) Code() string {
	if a.code == "" {
		a.code = a.ensureGenerator().Gen()
	}
	return a.code
}

// Verify uses the generator to validate user input.
func (a *AbstractCaptcha) Verify(userInputCode string) bool {
	return a.ensureGenerator().Verify(a.Code(), userInputCode)
}

// ImageBytes returns image bytes, or nil if not generated yet.
func (a *AbstractCaptcha) ImageBytes() []byte { return slices.Clone(a.imgBytes) }

func (a *AbstractCaptcha) imageBytesCopy() []byte { return slices.Clone(a.imgBytes) }

// ImageBase64 returns the Base64-encoded image.
func (a *AbstractCaptcha) ImageBase64() string {
	return base64.StdEncoding.EncodeToString(a.getImageBytes())
}

// ImageBase64Data returns a PNG data URI containing the Base64 image.
func (a *AbstractCaptcha) ImageBase64Data() string {
	return "data:image/png;base64," + a.ImageBase64()
}

// Write writes the image to an io.Writer.
func (a *AbstractCaptcha) Write(w io.Writer) error {
	if w == nil {
		return &knifer.Error{Code: knifer.ErrCodeInvalidInput, Message: "gkcaptcha: nil writer"}
	}
	b := a.getImageBytes()
	if len(b) == 0 {
		return &knifer.Error{Code: knifer.ErrCodeInvalidInput, Message: "gkcaptcha: empty image, call CreateCode first"}
	}
	_, err := w.Write(b)
	return err
}

// WriteToFile writes the image to a file.
func (a *AbstractCaptcha) WriteToFile(path string) error {
	return a.WriteToFileWithOptions(path)
}

// WriteToFileWithOptions writes the image to a file with custom filesystem options.
func (a *AbstractCaptcha) WriteToFileWithOptions(path string, opts ...WriteOption) error {
	b := a.getImageBytes()
	if len(b) == 0 {
		return &knifer.Error{Code: knifer.ErrCodeInvalidInput, Message: "gkcaptcha: empty image, call CreateCode first"}
	}
	cfg := applyWriteOptions(opts)
	if cfg.createParents {
		if err := cfg.mkdirAll(filepath.Dir(path), cfg.dirPerm); err != nil {
			return err
		}
	}
	flag := os.O_CREATE | os.O_WRONLY | os.O_TRUNC
	if !cfg.overwrite {
		flag |= os.O_EXCL
	}
	f, err := cfg.openFile(path, flag, cfg.filePerm) // #nosec G304 -- caller controls destination path.
	if err != nil {
		if os.IsExist(err) {
			return knifer.WrapError(knifer.ErrCodeInvalidInput, "gkcaptcha: file already exists", fs.ErrExist)
		}
		return err
	}
	defer func() { _ = f.Close() }()
	_, err = f.Write(b)
	return err
}

// Generator returns the underlying CodeGenerator.
func (a *AbstractCaptcha) Generator() CodeGenerator { return a.generator }

// SetGenerator replaces the CodeGenerator and resets generated state.
func (a *AbstractCaptcha) SetGenerator(g CodeGenerator) {
	a.generator = g
	a.code = ""
	a.imgBytes = nil
}

// SetBackground sets the background color.
func (a *AbstractCaptcha) SetBackground(bg color.Color) { a.Background = bg }

// getImageBytes returns lazily generated image bytes.
func (a *AbstractCaptcha) getImageBytes() []byte {
	return a.imgBytes
}

func (a *AbstractCaptcha) ensureGenerator() CodeGenerator {
	if a.generator == nil {
		a.generator = NewRandomGenerator(5)
	}
	return a.generator
}

func (a *AbstractCaptcha) generateCode() {
	a.code = a.ensureGenerator().Gen()
}

func (a *AbstractCaptcha) setImageBytes(b []byte) {
	a.imgBytes = slices.Clone(b)
}

func (a *AbstractCaptcha) bg() color.Color {
	if a.Background == nil {
		return color.White
	}
	return a.Background
}

func (a *AbstractCaptcha) randInt(max int) int {
	if max <= 0 {
		return 0
	}
	if a.randomInt != nil {
		v := a.randomInt(max)
		if v < 0 {
			return 0
		}
		if v >= max {
			return v % max
		}
		return v
	}
	return defaultRandomInt(max)
}

func (a *AbstractCaptcha) randIntRange(min, max int) int {
	if max <= min {
		return min
	}
	return min + a.randInt(max-min)
}

func (a *AbstractCaptcha) randColor() color.Color {
	if a.colorFunc != nil {
		return a.colorFunc()
	}
	return randomColor(a.randInt)
}

// VerifyIgnoreCase compares code and input case-insensitively for compatibility helpers.
func VerifyIgnoreCase(code, input string) bool {
	return strings.EqualFold(strings.TrimSpace(code), strings.TrimSpace(input))
}
