package vimg_test

import (
	"bytes"
	"fmt"
	"image"
	"image/png"

	"github.com/imajinyun/go-knifer/vimg"
)

func ExampleVerifyIgnoreCase() {
	matched := vimg.VerifyIgnoreCase("AbC4", "abc4")
	fmt.Println(matched)
	// Output: true
}

func ExampleNewMathGeneratorWith() {
	gen := vimg.NewMathGeneratorWith(2, false)
	result := vimg.GenMathGeneratorWithOptions(gen)
	fmt.Println(len(result) > 0)
	// Output: true
}

func ExampleCanEncodeBarcodeFormat() {
	fmt.Println(vimg.CanEncodeBarcodeFormat(vimg.BarcodeFormatQRCode))
	fmt.Println(vimg.CanEncodeBarcodeFormat(vimg.BarcodeFormatUnknown))
	// Output:
	// true
	// false
}

func ExampleInfo() {
	var buf bytes.Buffer
	_ = png.Encode(&buf, image.NewRGBA(image.Rect(0, 0, 2, 3)))

	width, height, format, err := vimg.Info(&buf)
	fmt.Println(width, height, format, err)
	// Output: 2 3 png <nil>
}

func ExampleQRCodeImage() {
	img, err := vimg.QRCodeImage("hello", vimg.WithQRCodeSize(64))
	fmt.Println(img.Bounds().Dx(), img.Bounds().Dy(), err)
	// Output: 64 64 <nil>
}

func ExampleLineCaptcha_ImageBytes() {
	c := vimg.NewLineCaptcha(100, 40)
	first := c.ImageBytes()
	want := append([]byte(nil), first...)
	first[0] ^= 0xff

	fmt.Println(bytes.Equal(c.ImageBytes(), want))
	// Output: true
}
