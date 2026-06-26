package vimg_test

import (
	"testing"

	"github.com/imajinyun/knifer-go/vimg"
)

func TestFacadeLineCaptcha(t *testing.T) {
	c := vimg.NewLineCaptcha(100, 40)
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
	c := vimg.NewCircleCaptcha(100, 40)
	if c == nil {
		t.Fatal("expected non-nil circle captcha")
	}
	c.CreateCode()
	code := c.Code()
	if len(code) == 0 {
		t.Fatal("expected circle captcha to have non-empty code")
	}
}

func TestFacadeLineCaptchaWith(t *testing.T) {
	c := vimg.NewLineCaptchaWith(100, 40, 4, 3)
	if c == nil {
		t.Fatal("expected non-nil line captcha with params")
	}
	c.CreateCode()
	if len(c.Code()) == 0 {
		t.Fatal("expected line captcha with to have non-empty code")
	}
}

func TestFacadeCircleCaptchaWith(t *testing.T) {
	c := vimg.NewCircleCaptchaWith(100, 40, 4, 3)
	if c == nil {
		t.Fatal("expected non-nil circle captcha with params")
	}
	c.CreateCode()
	if len(c.Code()) == 0 {
		t.Fatal("expected circle captcha with to have non-empty code")
	}
}

func TestFacadeShearCaptcha(t *testing.T) {
	c := vimg.NewShearCaptcha(100, 40)
	if c == nil {
		t.Fatal("expected non-nil shear captcha")
	}
	c.CreateCode()
	if len(c.Code()) == 0 {
		t.Fatal("expected shear captcha to have non-empty code")
	}
}

func TestFacadeShearCaptchaWith(t *testing.T) {
	c := vimg.NewShearCaptchaWith(100, 40, 4, 2)
	if c == nil {
		t.Fatal("expected non-nil shear captcha with params")
	}
	c.CreateCode()
	if len(c.Code()) == 0 {
		t.Fatal("expected shear captcha with to have non-empty code")
	}
}

func TestFacadeGifCaptcha(t *testing.T) {
	c := vimg.NewGifCaptcha(100, 40)
	if c == nil {
		t.Fatal("expected non-nil gif captcha")
	}
	c.CreateCode()
	if len(c.Code()) == 0 {
		t.Fatal("expected gif captcha to have non-empty code")
	}
}

func TestFacadeGifCaptchaWith(t *testing.T) {
	c := vimg.NewGifCaptchaWith(100, 40, 4)
	if c == nil {
		t.Fatal("expected non-nil gif captcha with params")
	}
	c.CreateCode()
	if len(c.Code()) == 0 {
		t.Fatal("expected gif captcha with to have non-empty code")
	}
}
