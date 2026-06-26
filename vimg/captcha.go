package vimg

import (
	"image/color"
	"io"
	"io/fs"

	"github.com/imajinyun/knifer-go/internal/imgx"
)

// Captcha is the common captcha interface.
type Captcha = imgx.ICaptcha

// ICaptcha is the common captcha interface.
type ICaptcha = imgx.ICaptcha

// CodeGenerator generates captcha verification code.
type CodeGenerator = imgx.CodeGenerator

// CaptchaOption customizes captcha construction.
type CaptchaOption = imgx.CaptchaOption

// WriteOption customizes captcha file output.
type WriteOption = imgx.WriteOption

// GeneratorOption customizes captcha code generation.
type GeneratorOption = imgx.GeneratorOption

// AbstractCaptcha contains common captcha fields and behavior.
type AbstractCaptcha = imgx.AbstractCaptcha

// LineCaptcha draws line-interference captcha images.
type LineCaptcha = imgx.LineCaptcha

// CircleCaptcha draws circle-interference captcha images.
type CircleCaptcha = imgx.CircleCaptcha

// ShearCaptcha draws sheared captcha images.
type ShearCaptcha = imgx.ShearCaptcha

// GifCaptcha draws animated GIF captcha images.
type GifCaptcha = imgx.GifCaptcha

// RandomGenerator generates random captcha strings.
type RandomGenerator = imgx.RandomGenerator

// MathGenerator generates math-expression captcha strings.
type MathGenerator = imgx.MathGenerator

// VerifyCaptchaIgnoreCase verifies code ignoring case.
func VerifyCaptchaIgnoreCase(code, input string) bool { return imgx.VerifyIgnoreCase(code, input) }

// WithGenerator sets the captcha code generator.
func WithGenerator(generator CodeGenerator) CaptchaOption { return imgx.WithGenerator(generator) }

// WithBackground sets the captcha background color.
func WithBackground(background color.Color) CaptchaOption { return imgx.WithBackground(background) }

// WithFontSize sets the font size ratio against captcha height.
func WithFontSize(fontSize float64) CaptchaOption { return imgx.WithFontSize(fontSize) }

// WithInterfereCount sets the number of interference elements.
func WithInterfereCount(count int) CaptchaOption { return imgx.WithInterfereCount(count) }

// WithGIFRepeat sets the animated GIF repeat count.
func WithGIFRepeat(repeat int) CaptchaOption { return imgx.WithGIFRepeat(repeat) }

// WithGIFDelay sets the animated GIF frame delay in 1/100 second units.
func WithGIFDelay(delay int) CaptchaOption { return imgx.WithGIFDelay(delay) }

// WithRandomInt sets the random integer function used while rendering captcha images.
func WithRandomInt(randomInt func(max int) int) CaptchaOption {
	return imgx.WithRandomInt(randomInt)
}

// WithGeneratorRandomInt sets the random integer function used while generating captcha codes.
func WithGeneratorRandomInt(randomInt func(max int) int) GeneratorOption {
	return imgx.WithGeneratorRandomInt(randomInt)
}

// WithGeneratorIntParser sets the integer parser used by math captcha verification.
func WithGeneratorIntParser(parser func(string) (int, error)) GeneratorOption {
	return imgx.WithGeneratorIntParser(parser)
}

// WithColorFunc sets the color function used while rendering captcha images.
func WithColorFunc(colorFunc func() color.Color) CaptchaOption {
	return imgx.WithColorFunc(colorFunc)
}

// WithFilePerm sets the file permission used by WriteToFileWithOptions.
func WithFilePerm(perm fs.FileMode) WriteOption { return imgx.WithFilePerm(perm) }

// WithDirPerm sets the parent directory permission used by WriteToFileWithOptions.
func WithDirPerm(perm fs.FileMode) WriteOption { return imgx.WithDirPerm(perm) }

// WithOverwrite controls whether WriteToFileWithOptions may replace an existing file.
func WithOverwrite(overwrite bool) WriteOption { return imgx.WithOverwrite(overwrite) }

// WithCreateParents controls whether WriteToFileWithOptions creates parent directories.
func WithCreateParents(create bool) WriteOption { return imgx.WithCreateParents(create) }

// WithMkdirAll sets the directory creator used by WriteToFileWithOptions.
func WithMkdirAll(mkdirAll func(string, fs.FileMode) error) WriteOption {
	return imgx.WithMkdirAll(mkdirAll)
}

// WithOpenFile sets the file opener used by WriteToFileWithOptions.
func WithOpenFile(openFile func(string, int, fs.FileMode) (io.WriteCloser, error)) WriteOption {
	return imgx.WithOpenFile(openFile)
}

// NewRandomGenerator creates a random captcha generator.
func NewRandomGenerator(length int) *RandomGenerator { return imgx.NewRandomGenerator(length) }

// NewRandomGeneratorWithBase creates a random captcha generator using base.
func NewRandomGeneratorWithBase(base string, length int) *RandomGenerator {
	return imgx.NewRandomGeneratorWithBase(base, length)
}

// NewMathGenerator creates a math captcha generator.
func NewMathGenerator() *MathGenerator { return imgx.NewMathGenerator() }

// GenRandomGeneratorWithOptions generates a random captcha string with per-call options.
func GenRandomGeneratorWithOptions(generator *RandomGenerator, opts ...GeneratorOption) string {
	return generator.GenWithOptions(opts...)
}

// GenMathGeneratorWithOptions generates a math captcha string with per-call options.
func GenMathGeneratorWithOptions(generator *MathGenerator, opts ...GeneratorOption) string {
	return generator.GenWithOptions(opts...)
}

// NewLineCaptcha creates a line-interference captcha.
func NewLineCaptcha(width, height int) *LineCaptcha { return imgx.NewLineCaptcha(width, height) }

// NewLineCaptchaWithOptions creates a line-interference captcha customized by options.
func NewLineCaptchaWithOptions(width, height int, opts ...CaptchaOption) *LineCaptcha {
	return imgx.NewLineCaptchaWithOptions(width, height, opts...)
}

// NewLineCaptchaWith creates a line-interference captcha with options.
func NewLineCaptchaWith(width, height, codeCount, lineCount int) *LineCaptcha {
	return imgx.NewLineCaptchaWith(width, height, codeCount, lineCount)
}

// NewCircleCaptcha creates a circle-interference captcha.
func NewCircleCaptcha(width, height int) *CircleCaptcha {
	return imgx.NewCircleCaptcha(width, height)
}

// NewCircleCaptchaWithOptions creates a circle-interference captcha customized by options.
func NewCircleCaptchaWithOptions(width, height int, opts ...CaptchaOption) *CircleCaptcha {
	return imgx.NewCircleCaptchaWithOptions(width, height, opts...)
}

// NewCircleCaptchaWith creates a circle-interference captcha with options.
func NewCircleCaptchaWith(width, height, codeCount, circleCount int) *CircleCaptcha {
	return imgx.NewCircleCaptchaWith(width, height, codeCount, circleCount)
}

// NewShearCaptcha creates a sheared captcha.
func NewShearCaptcha(width, height int) *ShearCaptcha { return imgx.NewShearCaptcha(width, height) }

// NewShearCaptchaWithOptions creates a sheared captcha customized by options.
func NewShearCaptchaWithOptions(width, height int, opts ...CaptchaOption) *ShearCaptcha {
	return imgx.NewShearCaptchaWithOptions(width, height, opts...)
}

// NewShearCaptchaWith creates a sheared captcha with options.
func NewShearCaptchaWith(width, height, codeCount, thickness int) *ShearCaptcha {
	return imgx.NewShearCaptchaWith(width, height, codeCount, thickness)
}

// NewGifCaptcha creates a GIF captcha.
func NewGifCaptcha(width, height int) *GifCaptcha { return imgx.NewGifCaptcha(width, height) }

// NewGifCaptchaWithOptions creates a GIF captcha customized by options.
func NewGifCaptchaWithOptions(width, height int, opts ...CaptchaOption) *GifCaptcha {
	return imgx.NewGifCaptchaWithOptions(width, height, opts...)
}

// NewGifCaptchaWith creates a GIF captcha with options.
func NewGifCaptchaWith(width, height, codeCount int) *GifCaptcha {
	return imgx.NewGifCaptchaWith(width, height, codeCount)
}
