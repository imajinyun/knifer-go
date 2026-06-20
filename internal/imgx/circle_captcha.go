package imgx

import (
	"bytes"
	"image"
	"image/png"
)

// CircleCaptcha mirrors CircleCaptcha and uses interference circles.
type CircleCaptcha struct {
	AbstractCaptcha
}

// NewCircleCaptcha creates a captcha with 5 characters and 15 circles by default.
func NewCircleCaptcha(width, height int) *CircleCaptcha {
	return NewCircleCaptchaWithOptions(width, height)
}

// NewCircleCaptchaWithOptions creates a circle captcha customized by options.
func NewCircleCaptchaWithOptions(width, height int, opts ...CaptchaOption) *CircleCaptcha {
	c := &CircleCaptcha{}
	c.Width = width
	c.Height = height
	c.InterfereCount = 15
	c.SetGenerator(NewRandomGenerator(5))
	applyCaptchaOptions(&c.AbstractCaptcha, opts)
	return c
}

// NewCircleCaptchaWith creates a captcha with custom character and circle counts.
func NewCircleCaptchaWith(width, height, codeCount, circleCount int) *CircleCaptcha {
	return NewCircleCaptchaWithOptions(width, height, WithGenerator(NewRandomGenerator(codeCount)), WithInterfereCount(circleCount))
}

// CreateCode generates a new captcha text and image.
func (c *CircleCaptcha) CreateCode() {
	c.generateCode()
	img := image.NewRGBA(image.Rect(0, 0, c.Width, c.Height))
	fillBackground(img, c.bg())
	half := c.Height >> 1
	for i := 0; i < c.InterfereCount; i++ {
		cx := c.randInt(c.Width)
		cy := c.randInt(c.Height)
		rx := c.randInt(atLeastOne(half))
		ry := c.randInt(atLeastOne(half))
		drawOval(img, cx, cy, rx, ry, c.randColor())
	}
	drawString(img, c.code, c.Width, c.Height, computeScale(c.Height), c.randColor)

	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	c.setImageBytes(buf.Bytes())
}

// ImageBytes renders the image lazily.
func (c *CircleCaptcha) ImageBytes() []byte {
	if c.imgBytes == nil {
		c.CreateCode()
	}
	return c.imageBytesCopy()
}

// Code generates the captcha lazily.
func (c *CircleCaptcha) Code() string {
	if c.code == "" {
		c.CreateCode()
	}
	return c.code
}

var _ ICaptcha = (*CircleCaptcha)(nil)
