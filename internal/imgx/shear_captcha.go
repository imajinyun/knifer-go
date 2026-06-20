package imgx

import (
	"bytes"
	"image"
	"image/png"
)

// ShearCaptcha mirrors the utility toolkit ShearCaptcha and applies distortion.
type ShearCaptcha struct {
	AbstractCaptcha
}

// NewShearCaptcha creates a captcha with 5 characters and line width 4 by default.
func NewShearCaptcha(width, height int) *ShearCaptcha {
	return NewShearCaptchaWithOptions(width, height)
}

// NewShearCaptchaWithOptions creates a shear captcha customized by options.
func NewShearCaptchaWithOptions(width, height int, opts ...CaptchaOption) *ShearCaptcha {
	c := &ShearCaptcha{}
	c.Width = width
	c.Height = height
	c.InterfereCount = 4
	c.SetGenerator(NewRandomGenerator(5))
	applyCaptchaOptions(&c.AbstractCaptcha, opts)
	return c
}

// NewShearCaptchaWith creates a captcha with custom character count and line width.
func NewShearCaptchaWith(width, height, codeCount, thickness int) *ShearCaptcha {
	return NewShearCaptchaWithOptions(width, height, WithGenerator(NewRandomGenerator(codeCount)), WithInterfereCount(thickness))
}

// CreateCode generates a captcha text and image.
func (c *ShearCaptcha) CreateCode() {
	c.generateCode()
	img := image.NewRGBA(image.Rect(0, 0, c.Width, c.Height))
	fillBackground(img, c.bg())

	// 1) Characters.
	drawString(img, c.code, c.Width, c.Height, computeScale(c.Height), c.randColor)

	// 2) Distortion.
	shearX(img, c.bg(), c.randIntRange, c.randInt)
	shearY(img, c.bg(), c.randIntRange)

	// 3) Thick interference line.
	thickness := c.InterfereCount
	if thickness <= 0 {
		thickness = 4
	}
	x1 := 0
	y1 := c.randInt(c.Height) + 1
	x2 := c.Width
	y2 := c.randInt(c.Height) + 1
	drawThickLine(img, x1, y1, x2, y2, thickness, c.randColor())

	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	c.setImageBytes(buf.Bytes())
}

// ImageBytes renders the image lazily.
func (c *ShearCaptcha) ImageBytes() []byte {
	if c.imgBytes == nil {
		c.CreateCode()
	}
	return c.imageBytesCopy()
}

// Code generates the captcha lazily.
func (c *ShearCaptcha) Code() string {
	if c.code == "" {
		c.CreateCode()
	}
	return c.code
}

var _ ICaptcha = (*ShearCaptcha)(nil)
