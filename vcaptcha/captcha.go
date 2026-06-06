package vcaptcha

import (
	"image/color"
	"io/fs"

	"github.com/imajinyun/go-knifer/internal/captcha"
)

// Captcha is the common captcha interface.
type Captcha = captcha.ICaptcha

// ICaptcha is the common captcha interface.
type ICaptcha = captcha.ICaptcha

// CodeGenerator generates captcha verification code.
type CodeGenerator = captcha.CodeGenerator

// CaptchaOption customizes captcha construction.
type CaptchaOption = captcha.CaptchaOption

// WriteOption customizes captcha file output.
type WriteOption = captcha.WriteOption

// GeneratorOption customizes captcha code generation.
type GeneratorOption = captcha.GeneratorOption

// AbstractCaptcha contains common captcha fields and behavior.
type AbstractCaptcha = captcha.AbstractCaptcha

// LineCaptcha draws line-interference captcha images.
type LineCaptcha = captcha.LineCaptcha

// CircleCaptcha draws circle-interference captcha images.
type CircleCaptcha = captcha.CircleCaptcha

// ShearCaptcha draws sheared captcha images.
type ShearCaptcha = captcha.ShearCaptcha

// GifCaptcha draws animated GIF captcha images.
type GifCaptcha = captcha.GifCaptcha

// RandomGenerator generates random captcha strings.
type RandomGenerator = captcha.RandomGenerator

// MathGenerator generates math-expression captcha strings.
type MathGenerator = captcha.MathGenerator

// VerifyCaptchaIgnoreCase verifies code ignoring case.
func VerifyCaptchaIgnoreCase(code, input string) bool { return captcha.VerifyIgnoreCase(code, input) }

// WithGenerator sets the captcha code generator.
func WithGenerator(generator CodeGenerator) CaptchaOption { return captcha.WithGenerator(generator) }

// WithBackground sets the captcha background color.
func WithBackground(background color.Color) CaptchaOption { return captcha.WithBackground(background) }

// WithFontSize sets the font size ratio against captcha height.
func WithFontSize(fontSize float64) CaptchaOption { return captcha.WithFontSize(fontSize) }

// WithInterfereCount sets the number of interference elements.
func WithInterfereCount(count int) CaptchaOption { return captcha.WithInterfereCount(count) }

// WithGIFRepeat sets the animated GIF repeat count.
func WithGIFRepeat(repeat int) CaptchaOption { return captcha.WithGIFRepeat(repeat) }

// WithGIFDelay sets the animated GIF frame delay in 1/100 second units.
func WithGIFDelay(delay int) CaptchaOption { return captcha.WithGIFDelay(delay) }

// WithRandomInt sets the random integer function used while rendering captcha images.
func WithRandomInt(randomInt func(max int) int) CaptchaOption {
	return captcha.WithRandomInt(randomInt)
}

// WithGeneratorRandomInt sets the random integer function used while generating captcha codes.
func WithGeneratorRandomInt(randomInt func(max int) int) GeneratorOption {
	return captcha.WithGeneratorRandomInt(randomInt)
}

// WithColorFunc sets the color function used while rendering captcha images.
func WithColorFunc(colorFunc func() color.Color) CaptchaOption {
	return captcha.WithColorFunc(colorFunc)
}

// WithFilePerm sets the file permission used by WriteToFileWithOptions.
func WithFilePerm(perm fs.FileMode) WriteOption { return captcha.WithFilePerm(perm) }

// WithDirPerm sets the parent directory permission used by WriteToFileWithOptions.
func WithDirPerm(perm fs.FileMode) WriteOption { return captcha.WithDirPerm(perm) }

// WithOverwrite controls whether WriteToFileWithOptions may replace an existing file.
func WithOverwrite(overwrite bool) WriteOption { return captcha.WithOverwrite(overwrite) }

// WithCreateParents controls whether WriteToFileWithOptions creates parent directories.
func WithCreateParents(create bool) WriteOption { return captcha.WithCreateParents(create) }

// NewRandomGenerator creates a random captcha generator.
func NewRandomGenerator(length int) *RandomGenerator { return captcha.NewRandomGenerator(length) }

// NewRandomGeneratorWithBase creates a random captcha generator using base.
func NewRandomGeneratorWithBase(base string, length int) *RandomGenerator {
	return captcha.NewRandomGeneratorWithBase(base, length)
}

// NewMathGenerator creates a math captcha generator.
func NewMathGenerator() *MathGenerator { return captcha.NewMathGenerator() }

// GenRandomGeneratorWithOptions generates a random captcha string with per-call options.
func GenRandomGeneratorWithOptions(generator *RandomGenerator, opts ...GeneratorOption) string {
	return generator.GenWithOptions(opts...)
}

// GenMathGeneratorWithOptions generates a math captcha string with per-call options.
func GenMathGeneratorWithOptions(generator *MathGenerator, opts ...GeneratorOption) string {
	return generator.GenWithOptions(opts...)
}

// NewLineCaptcha creates a line-interference captcha.
func NewLineCaptcha(width, height int) *LineCaptcha { return captcha.NewLineCaptcha(width, height) }

// NewLineCaptchaWithOptions creates a line-interference captcha customized by options.
func NewLineCaptchaWithOptions(width, height int, opts ...CaptchaOption) *LineCaptcha {
	return captcha.NewLineCaptchaWithOptions(width, height, opts...)
}

// NewLineCaptchaWith creates a line-interference captcha with options.
func NewLineCaptchaWith(width, height, codeCount, lineCount int) *LineCaptcha {
	return captcha.NewLineCaptchaWith(width, height, codeCount, lineCount)
}

// NewCircleCaptcha creates a circle-interference captcha.
func NewCircleCaptcha(width, height int) *CircleCaptcha {
	return captcha.NewCircleCaptcha(width, height)
}

// NewCircleCaptchaWithOptions creates a circle-interference captcha customized by options.
func NewCircleCaptchaWithOptions(width, height int, opts ...CaptchaOption) *CircleCaptcha {
	return captcha.NewCircleCaptchaWithOptions(width, height, opts...)
}

// NewCircleCaptchaWith creates a circle-interference captcha with options.
func NewCircleCaptchaWith(width, height, codeCount, circleCount int) *CircleCaptcha {
	return captcha.NewCircleCaptchaWith(width, height, codeCount, circleCount)
}

// NewShearCaptcha creates a sheared captcha.
func NewShearCaptcha(width, height int) *ShearCaptcha { return captcha.NewShearCaptcha(width, height) }

// NewShearCaptchaWithOptions creates a sheared captcha customized by options.
func NewShearCaptchaWithOptions(width, height int, opts ...CaptchaOption) *ShearCaptcha {
	return captcha.NewShearCaptchaWithOptions(width, height, opts...)
}

// NewShearCaptchaWith creates a sheared captcha with options.
func NewShearCaptchaWith(width, height, codeCount, thickness int) *ShearCaptcha {
	return captcha.NewShearCaptchaWith(width, height, codeCount, thickness)
}

// NewGifCaptcha creates a GIF captcha.
func NewGifCaptcha(width, height int) *GifCaptcha { return captcha.NewGifCaptcha(width, height) }

// NewGifCaptchaWithOptions creates a GIF captcha customized by options.
func NewGifCaptchaWithOptions(width, height int, opts ...CaptchaOption) *GifCaptcha {
	return captcha.NewGifCaptchaWithOptions(width, height, opts...)
}

// NewGifCaptchaWith creates a GIF captcha with options.
func NewGifCaptchaWith(width, height, codeCount int) *GifCaptcha {
	return captcha.NewGifCaptchaWith(width, height, codeCount)
}
