package captcha

import (
	"bytes"
	"image"
	"image/png"
)

// LineCaptcha mirrors the utility toolkit LineCaptcha and uses interference lines.
type LineCaptcha struct {
	AbstractCaptcha
}

// NewLineCaptcha creates a captcha with 5 characters and 150 lines by default.
func NewLineCaptcha(width, height int) *LineCaptcha {
	return NewLineCaptchaWith(width, height, 5, 150)
}

// NewLineCaptchaWithOptions creates a line captcha customized by options.
func NewLineCaptchaWithOptions(width, height int, opts ...CaptchaOption) *LineCaptcha {
	c := NewLineCaptcha(width, height)
	applyCaptchaOptions(&c.AbstractCaptcha, opts)
	return c
}

// NewLineCaptchaWith creates a captcha with custom character and line counts.
func NewLineCaptchaWith(width, height, codeCount, lineCount int) *LineCaptcha {
	c := &LineCaptcha{}
	c.Width = width
	c.Height = height
	c.InterfereCount = lineCount
	c.SetGenerator(NewRandomGenerator(codeCount))
	return c
}

// CreateCode generates a new captcha text and image.
func (c *LineCaptcha) CreateCode() {
	c.generateCode()
	c.setImageBytes(c.renderPNG(c.code))
}

// renderPNG renders PNG bytes from code.
func (c *LineCaptcha) renderPNG(code string) []byte {
	img := image.NewRGBA(image.Rect(0, 0, c.Width, c.Height))
	fillBackground(img, c.bg())
	// Interference lines.
	for i := 0; i < c.InterfereCount; i++ {
		xs := c.randInt(c.Width)
		ys := c.randInt(c.Height)
		xe := xs + c.randInt(atLeastOne(c.Width/8))
		ye := ys + c.randInt(atLeastOne(c.Height/8))
		drawLine(img, xs, ys, xe, ye, c.randColor())
	}
	// Characters.
	scale := computeScale(c.Height)
	drawString(img, code, c.Width, c.Height, scale, c.randColor)

	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	return buf.Bytes()
}

// ImageBytes overrides the embedded method to render lazily.
func (c *LineCaptcha) ImageBytes() []byte {
	if c.imgBytes == nil {
		c.CreateCode()
	}
	return c.imgBytes
}

// Code overrides the embedded method to generate lazily.
func (c *LineCaptcha) Code() string {
	if c.code == "" {
		c.CreateCode()
	}
	return c.code
}

// Embedding provides Verify, Write, and WriteToFile for ICaptcha.
var _ ICaptcha = (*LineCaptcha)(nil)

func atLeastOne(v int) int {
	if v > 1 {
		return v
	}
	return 1
}

// computeScale calculates the 5x7 bitmap font scale from image height.
func computeScale(height int) int {
	scale := height / (fontHeight + 4)
	if scale < 1 {
		scale = 1
	}
	return scale
}
