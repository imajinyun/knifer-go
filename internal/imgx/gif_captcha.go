package imgx

import (
	"bytes"
	"image"
	"image/color"
	"image/color/palette"
	"image/gif"
)

// GifCaptcha mirrors the utility toolkit GifCaptcha and renders animated GIF captchas.
//
// Each frame highlights one character and draws the others in pale colors to
// create a blinking effect.
type GifCaptcha struct {
	AbstractCaptcha

	// Repeat is the frame loop count; 0 means infinite looping.
	Repeat int
	// Delay is the frame delay in 1/100 second units; default is 10.
	Delay int
}

// NewGifCaptcha creates a captcha with 5 characters and 10 interference elements by default.
func NewGifCaptcha(width, height int) *GifCaptcha {
	return NewGifCaptchaWithOptions(width, height)
}

// NewGifCaptchaWithOptions creates a GIF captcha customized by options.
func NewGifCaptchaWithOptions(width, height int, opts ...CaptchaOption) *GifCaptcha {
	c := &GifCaptcha{}
	c.Width = width
	c.Height = height
	c.InterfereCount = 10
	c.Repeat = 0
	c.Delay = 10
	c.SetGenerator(NewRandomGenerator(5))
	cfg := applyCaptchaOptions(&c.AbstractCaptcha, opts)
	if cfg.setGIFRepeat {
		c.Repeat = cfg.gifRepeat
	}
	if cfg.setGIFDelay {
		c.Delay = cfg.gifDelay
	}
	return c
}

// NewGifCaptchaWith creates a captcha with a custom character count.
func NewGifCaptchaWith(width, height, codeCount int) *GifCaptcha {
	return NewGifCaptchaWithOptions(width, height, WithGenerator(NewRandomGenerator(codeCount)))
}

// CreateCode generates a new captcha text and GIF image.
func (c *GifCaptcha) CreateCode() {
	c.generateCode()

	frames := make([]*image.Paletted, 0, len(c.code))
	delays := make([]int, 0, len(c.code))
	disposals := make([]byte, 0, len(c.code))

	pal := append(color.Palette{}, palette.Plan9...)
	for hi := 0; hi < len(c.code); hi++ {
		rgba := image.NewRGBA(image.Rect(0, 0, c.Width, c.Height))
		fillBackground(rgba, c.bg())
		// Interference circles.
		half := c.Height >> 1
		for i := 0; i < c.InterfereCount; i++ {
			cx := c.randInt(c.Width)
			cy := c.randInt(c.Height)
			rx := c.randInt(atLeastOne(half))
			ry := c.randInt(atLeastOne(half))
			drawOval(rgba, cx, cy, rx, ry, c.randColor())
		}
		// Characters: highlight the current hi position.
		drawCodeFrame(rgba, c.code, hi, c.Width, c.Height, c.randInt, c.randColor)

		// Convert to a paletted frame.
		p := image.NewPaletted(rgba.Bounds(), pal)
		for y := 0; y < c.Height; y++ {
			for x := 0; x < c.Width; x++ {
				p.Set(x, y, rgba.RGBAAt(x, y))
			}
		}
		frames = append(frames, p)
		delays = append(delays, c.Delay)
		disposals = append(disposals, gif.DisposalBackground)
	}

	var buf bytes.Buffer
	_ = gif.EncodeAll(&buf, &gif.GIF{
		Image:     frames,
		Delay:     delays,
		LoopCount: c.Repeat,
		Disposal:  disposals,
	})
	c.setImageBytes(buf.Bytes())
}

// ImageBytes renders the image lazily.
func (c *GifCaptcha) ImageBytes() []byte {
	if c.imgBytes == nil {
		c.CreateCode()
	}
	return c.imgBytes
}

// Code generates the captcha lazily.
func (c *GifCaptcha) Code() string {
	if c.code == "" {
		c.CreateCode()
	}
	return c.code
}

// ImageBase64Data overrides the data URI MIME type for GIF.
func (c *GifCaptcha) ImageBase64Data() string {
	if c.imgBytes == nil {
		c.CreateCode()
	}
	return "data:image/gif;base64," + c.ImageBase64()
}

var _ ICaptcha = (*GifCaptcha)(nil)

// drawCodeFrame draws one GIF frame; the highlighted index uses a vivid color,
// while other positions use pale colors.
func drawCodeFrame(img *image.RGBA, code string, highlight, w, h int, randomInt func(max int) int, colorFunc func() color.Color) {
	scale := computeScale(h)
	charW := fontWidth*scale + scale
	totalW := charW * len(code)
	startX := (w - totalW) / 2
	charH := fontHeight * scale
	startY := (h - charH) / 2
	for i := 0; i < len(code); i++ {
		var c color.Color
		if i == highlight {
			c = colorFunc()
		} else {
			r := paleColorComponent(randomInt)
			g := paleColorComponent(randomInt)
			b := paleColorComponent(randomInt)
			c = color.RGBA{R: r, G: g, B: b, A: 255}
		}
		drawChar(img, code[i], startX+i*charW, startY, scale, c)
	}
}

func paleColorComponent(randomInt func(max int) int) uint8 {
	n := 160 + randomInt(80)
	if n < 0 {
		return 0
	}
	if n > 255 {
		return 255
	}
	return uint8(n)
}
