package vimg_test

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"io/fs"
	"os"
	"strings"

	"github.com/imajinyun/knifer-go/vimg"
)

func ExampleBarcodeASCII() {
	ascii, err := vimg.BarcodeASCII("example", vimg.BarcodeFormatQRCode)
	fmt.Println(strings.Contains(ascii, "██"), err)
	// Output: true <nil>
}

func ExampleBarcodeASCIIWithChars() {
	ascii, err := vimg.BarcodeASCIIWithChars("example", vimg.BarcodeFormatQRCode, "##", "..")
	fmt.Println(strings.Contains(ascii, "##"), strings.Contains(ascii, ".."), err)
	// Output: true true <nil>
}

func ExampleBarcodeBase64Data() {
	data, err := vimg.BarcodeBase64Data("example", vimg.BarcodeFormatQRCode)
	fmt.Println(strings.HasPrefix(data, "data:image/png;base64,"), err)
	// Output: true <nil>
}

func ExampleBarcodeBytes() {
	b, err := vimg.BarcodeBytes("example", vimg.BarcodeFormatQRCode, vimg.BarcodeOutputFormatSVG)
	fmt.Println(strings.Contains(string(b), "<svg"), err)
	// Output: true <nil>
}

func ExampleBarcodeImage() {
	img, err := vimg.BarcodeImage(
		"123456789012",
		vimg.BarcodeFormatEAN13,
		vimg.WithBarcodeSize(220, 90),
	)
	fmt.Println(img.Bounds().Dx(), img.Bounds().Dy(), err)
	// Output: 220 90 <nil>
}

func ExampleBarcodePNG() {
	b, err := vimg.BarcodePNG("example", vimg.BarcodeFormatQRCode, vimg.WithBarcodeSize(64, 64))
	cfg, _ := png.DecodeConfig(bytes.NewReader(b))
	fmt.Println(cfg.Width, cfg.Height, err)
	// Output: 64 64 <nil>
}

func ExampleBarcodeSVG() {
	svg, err := vimg.BarcodeSVG("example", vimg.BarcodeFormatQRCode)
	fmt.Println(strings.Contains(svg, "<svg"), strings.Contains(svg, "<path"), err)
	// Output: true true <nil>
}

func ExampleCanDecodeBarcodeFormat() {
	fmt.Println(vimg.CanDecodeBarcodeFormat(vimg.BarcodeFormatQRCode))
	fmt.Println(vimg.CanDecodeBarcodeFormat(vimg.BarcodeFormatUnknown))
	// Output:
	// true
	// false
}

func ExampleCanEncodeBarcodeFormat() {
	fmt.Println(vimg.CanEncodeBarcodeFormat(vimg.BarcodeFormatQRCode))
	fmt.Println(vimg.CanEncodeBarcodeFormat(vimg.BarcodeFormatUnknown))
	// Output:
	// true
	// false
}

func ExampleConvertFormat() {
	var out bytes.Buffer
	err := vimg.ConvertFormat(&out, bytes.NewReader(examplePNG(4, 3)), "jpeg")
	fmt.Println(out.Len() > 0, err)
	// Output: true <nil>
}

func ExampleDecodeBarcode() {
	b, _ := vimg.BarcodePNG("123456789012", vimg.BarcodeFormatEAN13, vimg.WithBarcodeSize(220, 90))
	result, err := vimg.DecodeBarcode(bytes.NewReader(b), vimg.WithDecodeFormats(vimg.BarcodeFormatEAN13))
	fmt.Println(result.Format, result.Text, err)
	// Output: EAN_13 1234567890128 <nil>
}

func ExampleDecodeBarcodeImage() {
	img, _ := vimg.BarcodeImage("123456789012", vimg.BarcodeFormatEAN13, vimg.WithBarcodeSize(220, 90))
	result, err := vimg.DecodeBarcodeImage(img, vimg.WithDecodeFormats(vimg.BarcodeFormatEAN13))
	fmt.Println(result.Format, result.Text, err)
	// Output: EAN_13 1234567890128 <nil>
}

func ExampleDecodeQRCode() {
	b, _ := vimg.QRCodePNG("payload", vimg.WithQRCodeSize(96))
	result, err := vimg.DecodeQRCode(bytes.NewReader(b))
	fmt.Println(result.Format, result.Text, err)
	// Output: QR_CODE payload <nil>
}

func ExampleDecodeQRCodeImage() {
	img, _ := vimg.QRCodeImage("payload", vimg.WithQRCodeSize(96))
	result, err := vimg.DecodeQRCodeImage(img)
	fmt.Println(result.Format, result.Text, err)
	// Output: QR_CODE payload <nil>
}

func ExampleGenMathGeneratorWithOptions() {
	gen := vimg.NewMathGeneratorWith(1, false)
	values := []int{1, 7, 3}
	idx := 0
	code := vimg.GenMathGeneratorWithOptions(gen, vimg.WithGeneratorRandomInt(func(max int) int {
		v := values[idx]
		idx++
		return v % max
	}))
	fmt.Println(code, gen.Verify(code, "4"))
	// Output: 7-3= true
}

func ExampleGenRandomGeneratorWithOptions() {
	gen := vimg.NewRandomGeneratorWithBase("abcd", 4)
	idx := 0
	code := vimg.GenRandomGeneratorWithOptions(gen, vimg.WithGeneratorRandomInt(func(max int) int {
		v := idx
		idx++
		return v % max
	}))
	fmt.Println(code)
	// Output: abcd
}

func ExampleInfo() {
	width, height, format, err := vimg.Info(bytes.NewReader(examplePNG(2, 3)))
	fmt.Println(width, height, format, err)
	// Output: 2 3 png <nil>
}

func ExampleNewCircleCaptcha() {
	c := vimg.NewCircleCaptcha(100, 40)
	fmt.Println(c.Width, c.Height, len(c.Code()))
	// Output: 100 40 5
}

func ExampleNewCircleCaptchaWith() {
	c := vimg.NewCircleCaptchaWith(100, 40, 4, 2)
	fmt.Println(c.Width, c.Height, len(c.Code()), c.InterfereCount)
	// Output: 100 40 4 2
}

func ExampleNewCircleCaptchaWithOptions() {
	c := vimg.NewCircleCaptchaWithOptions(100, 40, vimg.WithGenerator(fixedGenerator{code: "WXYZ"}))
	fmt.Println(c.Code(), c.Verify("WXYZ"))
	// Output: WXYZ true
}

func ExampleNewGifCaptcha() {
	c := vimg.NewGifCaptcha(100, 40)
	fmt.Println(c.Width, c.Height, len(c.Code()))
	// Output: 100 40 5
}

func ExampleNewGifCaptchaWith() {
	c := vimg.NewGifCaptchaWith(100, 40, 4)
	fmt.Println(c.Width, c.Height, len(c.Code()))
	// Output: 100 40 4
}

func ExampleNewGifCaptchaWithOptions() {
	c := vimg.NewGifCaptchaWithOptions(
		100,
		40,
		vimg.WithGenerator(fixedGenerator{code: "IJKL"}),
		vimg.WithGIFRepeat(1),
		vimg.WithGIFDelay(5),
	)
	fmt.Println(c.Code(), c.Repeat, c.Delay)
	// Output: IJKL 1 5
}

func ExampleNewLineCaptcha() {
	c := vimg.NewLineCaptcha(100, 40)
	fmt.Println(c.Width, c.Height, len(c.Code()))
	// Output: 100 40 5
}

func ExampleNewLineCaptchaWith() {
	c := vimg.NewLineCaptchaWith(100, 40, 4, 3)
	fmt.Println(c.Width, c.Height, len(c.Code()), c.InterfereCount)
	// Output: 100 40 4 3
}

func ExampleNewLineCaptchaWithOptions() {
	c := vimg.NewLineCaptchaWithOptions(100, 40, vimg.WithGenerator(fixedGenerator{code: "ABCD"}))
	fmt.Println(c.Code(), c.Verify("ABCD"))
	// Output: ABCD true
}

func ExampleNewMathGenerator() {
	gen := vimg.NewMathGenerator()
	fmt.Println(gen.Verify("1+1=", "2"))
	// Output: true
}

func ExampleNewMathGeneratorWith() {
	gen := vimg.NewMathGeneratorWith(2, false)
	result := vimg.GenMathGeneratorWithOptions(gen)
	fmt.Println(len(result) > 0)
	// Output: true
}

func ExampleNewRandomGenerator() {
	gen := vimg.NewRandomGenerator(4)
	code := gen.Gen()
	fmt.Println(len(code), gen.Verify(code, code))
	// Output: 4 true
}

func ExampleNewRandomGeneratorWithBase() {
	gen := vimg.NewRandomGeneratorWithBase("ab", 4)
	code := vimg.GenRandomGeneratorWithOptions(gen, vimg.WithGeneratorRandomInt(func(int) int { return 1 }))
	fmt.Println(code)
	// Output: bbbb
}

func ExampleNewShearCaptcha() {
	c := vimg.NewShearCaptcha(100, 40)
	fmt.Println(c.Width, c.Height, len(c.Code()))
	// Output: 100 40 5
}

func ExampleNewShearCaptchaWith() {
	c := vimg.NewShearCaptchaWith(100, 40, 4, 2)
	fmt.Println(c.Width, c.Height, len(c.Code()), c.InterfereCount)
	// Output: 100 40 4 2
}

func ExampleNewShearCaptchaWithOptions() {
	c := vimg.NewShearCaptchaWithOptions(100, 40, vimg.WithGenerator(fixedGenerator{code: "EFGH"}))
	fmt.Println(c.Code(), c.Verify("EFGH"))
	// Output: EFGH true
}

func ExampleQRCodeASCII() {
	ascii, err := vimg.QRCodeASCII("example")
	fmt.Println(strings.Contains(ascii, "██"), err)
	// Output: true <nil>
}

func ExampleQRCodeASCIIWithChars() {
	ascii, err := vimg.QRCodeASCIIWithChars("example", "##", "..")
	fmt.Println(strings.Contains(ascii, "##"), strings.Contains(ascii, ".."), err)
	// Output: true true <nil>
}

func ExampleQRCodeBase64Data() {
	data, err := vimg.QRCodeBase64Data("example")
	fmt.Println(strings.HasPrefix(data, "data:image/png;base64,"), err)
	// Output: true <nil>
}

func ExampleQRCodeBytes() {
	b, err := vimg.QRCodeBytes("example", vimg.BarcodeOutputFormatSVG)
	fmt.Println(strings.Contains(string(b), "<svg"), err)
	// Output: true <nil>
}

func ExampleQRCodeImage() {
	img, err := vimg.QRCodeImage("hello", vimg.WithQRCodeSize(64))
	fmt.Println(img.Bounds().Dx(), img.Bounds().Dy(), err)
	// Output: 64 64 <nil>
}

func ExampleQRCodePNG() {
	b, err := vimg.QRCodePNG("example", vimg.WithQRCodeSize(64))
	cfg, _ := png.DecodeConfig(bytes.NewReader(b))
	fmt.Println(cfg.Width, cfg.Height, err)
	// Output: 64 64 <nil>
}

func ExampleQRCodeSVG() {
	svg, err := vimg.QRCodeSVG("example")
	fmt.Println(strings.Contains(svg, "<svg"), strings.Contains(svg, "<path"), err)
	// Output: true true <nil>
}

func ExampleSupportedDecodeBarcodeFormats() {
	formats := vimg.SupportedDecodeBarcodeFormats()
	fmt.Println(len(formats) > 0, formats[0])
	// Output: true QR_CODE
}

func ExampleSupportedEncodeBarcodeFormats() {
	formats := vimg.SupportedEncodeBarcodeFormats()
	fmt.Println(len(formats) > 0, formats[0])
	// Output: true CODABAR
}

func ExampleThumbnail() {
	var out bytes.Buffer
	err := vimg.Thumbnail(&out, bytes.NewReader(examplePNG(20, 10)), 5, "png")
	cfg, _ := png.DecodeConfig(bytes.NewReader(out.Bytes()))
	fmt.Println(cfg.Width, cfg.Height, err)
	// Output: 5 2 <nil>
}

func ExampleVerifyCaptchaIgnoreCase() {
	fmt.Println(vimg.VerifyCaptchaIgnoreCase("Captcha", "captcha"))
	// Output: true
}

func ExampleVerifyIgnoreCase() {
	matched := vimg.VerifyIgnoreCase("AbC4", "abc4")
	fmt.Println(matched)
	// Output: true
}

func ExampleWithBackground() {
	c := vimg.NewLineCaptchaWithOptions(
		100,
		40,
		vimg.WithGenerator(fixedGenerator{code: "ABCD"}),
		vimg.WithBackground(color.Black),
	)
	fmt.Println(c.Code(), c.Background != nil)
	// Output: ABCD true
}

func ExampleWithBarcodeBackground() {
	svg, err := vimg.BarcodeSVG(
		"example",
		vimg.BarcodeFormatQRCode,
		vimg.WithBarcodeBackground(color.White),
	)
	fmt.Println(strings.Contains(svg, "<svg"), err)
	// Output: true <nil>
}

func ExampleWithBarcodeCharacterSet() {
	b, err := vimg.BarcodePNG(
		"example",
		vimg.BarcodeFormatQRCode,
		vimg.WithBarcodeCharacterSet("UTF-8"),
	)
	fmt.Println(len(b) > 0, err)
	// Output: true <nil>
}

func ExampleWithBarcodeColors() {
	svg, err := vimg.BarcodeSVG(
		"example",
		vimg.BarcodeFormatQRCode,
		vimg.WithBarcodeColors(color.Black, color.White),
	)
	fmt.Println(strings.Contains(svg, "<svg"), err)
	// Output: true <nil>
}

func ExampleWithBarcodeForceCodeSet() {
	b, err := vimg.BarcodePNG(
		"ABC123",
		vimg.BarcodeFormatCode128,
		vimg.WithBarcodeForceCodeSet("B"),
		vimg.WithBarcodeSize(220, 90),
	)
	fmt.Println(len(b) > 0, err)
	// Output: true <nil>
}

func ExampleWithBarcodeForeground() {
	svg, err := vimg.BarcodeSVG(
		"example",
		vimg.BarcodeFormatQRCode,
		vimg.WithBarcodeForeground(color.Black),
	)
	fmt.Println(strings.Contains(svg, "<svg"), err)
	// Output: true <nil>
}

func ExampleWithBarcodeGS1Format() {
	b, err := vimg.BarcodePNG(
		"0101234567890128",
		vimg.BarcodeFormatCode128,
		vimg.WithBarcodeGS1Format(true),
	)
	fmt.Println(len(b) > 0, err)
	// Output: true <nil>
}

func ExampleWithBarcodeLogo() {
	b, err := vimg.BarcodePNG(
		"example",
		vimg.BarcodeFormatQRCode,
		vimg.WithBarcodeLogo(exampleLogo()),
	)
	fmt.Println(len(b) > 0, err)
	// Output: true <nil>
}

func ExampleWithBarcodeLogoSize() {
	b, err := vimg.BarcodePNG(
		"example",
		vimg.BarcodeFormatQRCode,
		vimg.WithBarcodeLogo(exampleLogo()),
		vimg.WithBarcodeLogoSize(8, 8),
	)
	fmt.Println(len(b) > 0, err)
	// Output: true <nil>
}

func ExampleWithBarcodeMargin() {
	b, err := vimg.BarcodePNG(
		"example",
		vimg.BarcodeFormatQRCode,
		vimg.WithBarcodeMargin(1),
	)
	fmt.Println(len(b) > 0, err)
	// Output: true <nil>
}

func ExampleWithBarcodeSize() {
	img, err := vimg.BarcodeImage(
		"123456789012",
		vimg.BarcodeFormatEAN13,
		vimg.WithBarcodeSize(220, 90),
	)
	fmt.Println(img.Bounds().Dx(), img.Bounds().Dy(), err)
	// Output: 220 90 <nil>
}

func ExampleWithBarcodeTransparentBackground() {
	svg, err := vimg.BarcodeSVG(
		"example",
		vimg.BarcodeFormatQRCode,
		vimg.WithBarcodeTransparentBackground(),
	)
	fmt.Println(strings.Contains(svg, "rgba(0,0,0,0.000)"), err)
	// Output: true <nil>
}

func ExampleWithColorFunc() {
	colorCalls := 0
	c := vimg.NewLineCaptchaWithOptions(
		100,
		40,
		vimg.WithGenerator(fixedGenerator{code: "ABCD"}),
		vimg.WithInterfereCount(0),
		vimg.WithColorFunc(func() color.Color {
			colorCalls++
			return color.Black
		}),
	)
	_ = c.ImageBytes()
	fmt.Println(c.Code(), colorCalls)
	// Output: ABCD 4
}

func ExampleWithCreateParents() {
	c := exampleCaptcha()
	var mkdirCalled bool
	var written bytes.Buffer
	err := c.WriteToFileWithOptions(
		"captcha.png",
		vimg.WithCreateParents(false),
		vimg.WithMkdirAll(func(string, fs.FileMode) error {
			mkdirCalled = true
			return nil
		}),
		vimg.WithOpenFile(func(string, int, fs.FileMode) (io.WriteCloser, error) {
			return nopWriteCloser{Writer: &written}, nil
		}),
	)
	fmt.Println(mkdirCalled, written.Len() > 0, err)
	// Output: false true <nil>
}

func ExampleWithDecodeAlsoInverted() {
	fmt.Println(vimg.WithDecodeAlsoInverted(true) != nil)
	// Output: true
}

func ExampleWithDecodeCharacterSet() {
	fmt.Println(vimg.WithDecodeCharacterSet("UTF-8") != nil)
	// Output: true
}

func ExampleWithDecodeFormats() {
	b, _ := vimg.BarcodePNG("123456789012", vimg.BarcodeFormatEAN13, vimg.WithBarcodeSize(220, 90))
	result, err := vimg.DecodeBarcode(bytes.NewReader(b), vimg.WithDecodeFormats(vimg.BarcodeFormatEAN13))
	fmt.Println(result.Format, err)
	// Output: EAN_13 <nil>
}

func ExampleWithDecodePureBarcode() {
	fmt.Println(vimg.WithDecodePureBarcode(true) != nil)
	// Output: true
}

func ExampleWithDecodeTryHarder() {
	b, _ := vimg.QRCodePNG("payload")
	result, err := vimg.DecodeQRCode(bytes.NewReader(b), vimg.WithDecodeTryHarder(true))
	fmt.Println(result.Text, err)
	// Output: payload <nil>
}

func ExampleWithDirPerm() {
	c := exampleCaptcha()
	var mkdirPerm fs.FileMode
	err := c.WriteToFileWithOptions(
		"/virtual/captcha.png",
		vimg.WithDirPerm(0o700),
		vimg.WithMkdirAll(func(_ string, perm fs.FileMode) error {
			mkdirPerm = perm
			return nil
		}),
		vimg.WithOpenFile(func(string, int, fs.FileMode) (io.WriteCloser, error) {
			return nopWriteCloser{Writer: io.Discard}, nil
		}),
	)
	fmt.Printf("%o %v\n", mkdirPerm, err)
	// Output: 700 <nil>
}

func ExampleWithFilePerm() {
	c := exampleCaptcha()
	var filePerm fs.FileMode
	err := c.WriteToFileWithOptions(
		"/virtual/captcha.png",
		vimg.WithFilePerm(0o600),
		vimg.WithMkdirAll(func(string, fs.FileMode) error { return nil }),
		vimg.WithOpenFile(func(_ string, _ int, perm fs.FileMode) (io.WriteCloser, error) {
			filePerm = perm
			return nopWriteCloser{Writer: io.Discard}, nil
		}),
	)
	fmt.Printf("%o %v\n", filePerm, err)
	// Output: 600 <nil>
}

func ExampleWithFontSize() {
	c := vimg.NewLineCaptchaWithOptions(100, 40, vimg.WithFontSize(0.8))
	fmt.Println(c.FontSize)
	// Output: 0.8
}

func ExampleWithGIFDelay() {
	c := vimg.NewGifCaptchaWithOptions(100, 40, vimg.WithGIFDelay(5))
	fmt.Println(c.Delay)
	// Output: 5
}

func ExampleWithGIFRepeat() {
	c := vimg.NewGifCaptchaWithOptions(100, 40, vimg.WithGIFRepeat(1))
	fmt.Println(c.Repeat)
	// Output: 1
}

func ExampleWithGenerator() {
	c := vimg.NewLineCaptchaWithOptions(100, 40, vimg.WithGenerator(fixedGenerator{code: "ABCD"}))
	fmt.Println(c.Code())
	// Output: ABCD
}

func ExampleWithGeneratorIntParser() {
	gen := vimg.NewMathGenerator()
	code := "one+two="
	parser := func(s string) (int, error) {
		switch s {
		case "one":
			return 1, nil
		case "two":
			return 2, nil
		case "3":
			return 3, nil
		default:
			return 0, nil
		}
	}
	fmt.Println(gen.VerifyWithOptions(code, "3", vimg.WithGeneratorIntParser(parser)))
	// Output: true
}

func ExampleWithGeneratorRandomInt() {
	gen := vimg.NewRandomGeneratorWithBase("abcd", 4)
	code := vimg.GenRandomGeneratorWithOptions(gen, vimg.WithGeneratorRandomInt(func(int) int { return 2 }))
	fmt.Println(code)
	// Output: cccc
}

func ExampleWithInterfereCount() {
	c := vimg.NewLineCaptchaWithOptions(100, 40, vimg.WithInterfereCount(2))
	fmt.Println(c.InterfereCount)
	// Output: 2
}

func ExampleWithMkdirAll() {
	c := exampleCaptcha()
	var mkdirPath string
	err := c.WriteToFileWithOptions(
		"/virtual/captcha.png",
		vimg.WithMkdirAll(func(path string, perm fs.FileMode) error {
			mkdirPath = path
			return nil
		}),
		vimg.WithOpenFile(func(string, int, fs.FileMode) (io.WriteCloser, error) {
			return nopWriteCloser{Writer: io.Discard}, nil
		}),
	)
	fmt.Println(mkdirPath, err)
	// Output: /virtual <nil>
}

func ExampleWithOpenFile() {
	c := exampleCaptcha()
	var openedPath string
	err := c.WriteToFileWithOptions(
		"/virtual/captcha.png",
		vimg.WithMkdirAll(func(string, fs.FileMode) error { return nil }),
		vimg.WithOpenFile(func(path string, flag int, perm fs.FileMode) (io.WriteCloser, error) {
			openedPath = path
			return nopWriteCloser{Writer: io.Discard}, nil
		}),
	)
	fmt.Println(openedPath, err)
	// Output: /virtual/captcha.png <nil>
}

func ExampleWithOverwrite() {
	c := exampleCaptcha()
	var flag int
	err := c.WriteToFileWithOptions(
		"/virtual/captcha.png",
		vimg.WithOverwrite(false),
		vimg.WithMkdirAll(func(string, fs.FileMode) error { return nil }),
		vimg.WithOpenFile(func(_ string, f int, _ fs.FileMode) (io.WriteCloser, error) {
			flag = f
			return nopWriteCloser{Writer: io.Discard}, nil
		}),
	)
	fmt.Println(flag&os.O_EXCL != 0, err)
	// Output: true <nil>
}

func ExampleWithQRCodeBackground() {
	svg, err := vimg.QRCodeSVG("example", vimg.WithQRCodeBackground(color.White))
	fmt.Println(strings.Contains(svg, "<svg"), err)
	// Output: true <nil>
}

func ExampleWithQRCodeColors() {
	svg, err := vimg.QRCodeSVG("example", vimg.WithQRCodeColors(color.Black, color.White))
	fmt.Println(strings.Contains(svg, "<svg"), err)
	// Output: true <nil>
}

func ExampleWithQRCodeErrorCorrection() {
	b, err := vimg.QRCodePNG("example", vimg.WithQRCodeErrorCorrection(vimg.QRErrorCorrectionHigh))
	fmt.Println(len(b) > 0, err)
	// Output: true <nil>
}

func ExampleWithQRCodeForeground() {
	svg, err := vimg.QRCodeSVG("example", vimg.WithQRCodeForeground(color.Black))
	fmt.Println(strings.Contains(svg, "<svg"), err)
	// Output: true <nil>
}

func ExampleWithQRCodeLogo() {
	b, err := vimg.QRCodePNG("example", vimg.WithQRCodeLogo(exampleLogo()))
	fmt.Println(len(b) > 0, err)
	// Output: true <nil>
}

func ExampleWithQRCodeLogoRatio() {
	b, err := vimg.QRCodePNG(
		"example",
		vimg.WithQRCodeLogo(exampleLogo()),
		vimg.WithQRCodeLogoRatio(4),
	)
	fmt.Println(len(b) > 0, err)
	// Output: true <nil>
}

func ExampleWithQRCodeLogoSize() {
	b, err := vimg.QRCodePNG(
		"example",
		vimg.WithQRCodeLogo(exampleLogo()),
		vimg.WithQRCodeLogoSize(8, 8),
	)
	fmt.Println(len(b) > 0, err)
	// Output: true <nil>
}

func ExampleWithQRCodeMargin() {
	b, err := vimg.QRCodePNG("example", vimg.WithQRCodeMargin(1))
	fmt.Println(len(b) > 0, err)
	// Output: true <nil>
}

func ExampleWithQRCodeMaskPattern() {
	b, err := vimg.QRCodePNG("A", vimg.WithQRCodeVersion(1), vimg.WithQRCodeMaskPattern(2))
	fmt.Println(len(b) > 0, err)
	// Output: true <nil>
}

func ExampleWithQRCodeSize() {
	img, err := vimg.QRCodeImage("example", vimg.WithQRCodeSize(64))
	fmt.Println(img.Bounds().Dx(), img.Bounds().Dy(), err)
	// Output: 64 64 <nil>
}

func ExampleWithQRCodeTransparentBackground() {
	svg, err := vimg.QRCodeSVG("example", vimg.WithQRCodeTransparentBackground())
	fmt.Println(strings.Contains(svg, "rgba(0,0,0,0.000)"), err)
	// Output: true <nil>
}

func ExampleWithQRCodeVersion() {
	b, err := vimg.QRCodePNG("A", vimg.WithQRCodeVersion(1))
	fmt.Println(len(b) > 0, err)
	// Output: true <nil>
}

func ExampleWithRandomInt() {
	c := vimg.NewLineCaptchaWithOptions(
		100,
		40,
		vimg.WithGenerator(fixedGenerator{code: "ABCD"}),
		vimg.WithInterfereCount(0),
		vimg.WithRandomInt(func(int) int { return 0 }),
	)
	fmt.Println(c.Code(), len(c.ImageBytes()) > 0)
	// Output: ABCD true
}

func ExampleWriteBarcode() {
	var buf bytes.Buffer
	err := vimg.WriteBarcode(&buf, "123456789012", vimg.BarcodeFormatEAN13, vimg.WithBarcodeSize(220, 90))
	fmt.Println(buf.Len() > 0, err)
	// Output: true <nil>
}

func ExampleWriteQRCode() {
	var buf bytes.Buffer
	err := vimg.WriteQRCode(&buf, "payload")
	fmt.Println(buf.Len() > 0, err)
	// Output: true <nil>
}

func ExampleLineCaptcha_ImageBytes() {
	c := vimg.NewLineCaptcha(100, 40)
	first := c.ImageBytes()
	want := append([]byte(nil), first...)
	first[0] ^= 0xff

	fmt.Println(bytes.Equal(c.ImageBytes(), want))
	// Output: true
}

func examplePNG(width, height int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.SetRGBA(x, y, color.RGBA{R: uint8(x), G: uint8(y), B: 128, A: 255})
		}
	}

	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	return buf.Bytes()
}

func exampleLogo() image.Image {
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			img.Set(x, y, color.Black)
		}
	}
	return img
}

func exampleCaptcha() *vimg.LineCaptcha {
	c := vimg.NewLineCaptchaWithOptions(
		100,
		40,
		vimg.WithGenerator(fixedGenerator{code: "ABCD"}),
		vimg.WithInterfereCount(0),
		vimg.WithRandomInt(func(int) int { return 0 }),
		vimg.WithColorFunc(func() color.Color { return color.Black }),
	)
	_ = c.ImageBytes()
	return c
}
