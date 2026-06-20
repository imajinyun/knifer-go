# vimg Quickstart

`vimg` provides image processing, QR/barcode, and captcha helpers. It covers image metadata reads, format conversion, thumbnail generation, ZXing-backed QR/barcode generation and decoding, PNG/SVG/ASCII/Base64 data URI output, QR logo embedding, transparent backgrounds, captcha generation/verification, and file writes.

Captcha image bytes are returned as defensive copies, so callers can inspect or transform the returned slice without mutating the captcha's cached image. File overwrite failures from captcha writers preserve `fs.ErrExist` and carry the go-knifer invalid-input error code for consistent error inspection.

## Read image information

```go
package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"

	"github.com/imajinyun/go-knifer/vimg"
)

func main() {
	img := image.NewRGBA(image.Rect(0, 0, 80, 40))
	img.Set(0, 0, color.RGBA{R: 255, A: 255})

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		panic(err)
	}

	width, height, format, err := vimg.Info(bytes.NewReader(buf.Bytes()))
	if err != nil {
		panic(err)
	}
	fmt.Println(width, height, format)
}
```

## Convert formats and generate thumbnails

```go
package main

import (
	"bytes"
	"fmt"
	"image"
	"image/png"

	"github.com/imajinyun/go-knifer/vimg"
)

func main() {
	img := image.NewRGBA(image.Rect(0, 0, 160, 80))
	var src bytes.Buffer
	if err := png.Encode(&src, img); err != nil {
		panic(err)
	}

	var jpegOut bytes.Buffer
	if err := vimg.ConvertFormat(&jpegOut, bytes.NewReader(src.Bytes()), "jpeg"); err != nil {
		panic(err)
	}

	var thumb bytes.Buffer
	if err := vimg.Thumbnail(&thumb, bytes.NewReader(src.Bytes()), 64, "png"); err != nil {
		panic(err)
	}
fmt.Println(jpegOut.Len() > 0, thumb.Len() > 0)
}
```

## Generate QR codes and barcodes

```go
package main

import (
	"fmt"
	"image/color"
	"strings"

	"github.com/imajinyun/go-knifer/vimg"
)

func main() {
	pngBytes, err := vimg.QRCodePNG("https://github.com/imajinyun/go-knifer",
		vimg.WithQRCodeSize(180),
		vimg.WithQRCodeMargin(2),
		vimg.WithQRCodeErrorCorrection(vimg.QRErrorCorrectionMedium),
	)
	if err != nil {
		panic(err)
	}

	svg, err := vimg.QRCodeSVG("svg payload",
		vimg.WithQRCodeSize(120),
		vimg.WithQRCodeForeground(color.Black),
		vimg.WithQRCodeTransparentBackground(),
	)
	if err != nil {
		panic(err)
	}

	barcode, err := vimg.BarcodePNG("123456789012", vimg.BarcodeFormatEAN13,
		vimg.WithBarcodeSize(220, 90),
	)
	if err != nil {
		panic(err)
	}

	fmt.Println(len(pngBytes) > 0, strings.Contains(svg, "<svg"), len(barcode) > 0)
}
```

QR-specific helpers are shortcuts around the generic barcode APIs:

- `QRCodePNG`, `QRCodeSVG`, `QRCodeASCII`, `QRCodeBase64Data`, `QRCodeBytes`, and `QRCodeImage` always use `BarcodeFormatQRCode`.
- `BarcodePNG`, `BarcodeSVG`, `BarcodeASCII`, `BarcodeBase64Data`, `BarcodeBytes`, and `BarcodeImage` accept an explicit `BarcodeFormat`.
- Use `WithBarcodeSize` / `WithQRCodeSize`, `WithBarcodeMargin` / `WithQRCodeMargin`, and color options to control output size, quiet-zone margin, and foreground/background colors.
- Use `WithBarcodeTransparentBackground` or `WithQRCodeTransparentBackground` when the light modules should be transparent in raster and SVG output.

## Embed QR logos

```go
package main

import (
	"fmt"
	"image"
	"image/color"
	"strings"

	"github.com/imajinyun/go-knifer/vimg"
)

func main() {
	logo := image.NewRGBA(image.Rect(0, 0, 16, 16))
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			logo.Set(x, y, color.RGBA{R: 255, A: 255})
		}
	}

	svg, err := vimg.QRCodeSVG("logo payload",
		vimg.WithQRCodeSize(160),
		vimg.WithQRCodeErrorCorrection(vimg.QRErrorCorrectionHigh),
		vimg.WithQRCodeLogo(logo),
		vimg.WithQRCodeLogoRatio(6),
	)
	if err != nil {
		panic(err)
	}
	fmt.Println(strings.Contains(svg, `<image `))
}
```

Logo options are intentionally QR-only. Passing `WithBarcodeLogo` to non-QR formats returns `ErrCodeUnsupported`, because 1D barcodes and other 2D formats have different readability constraints. If no explicit `WithQRCodeLogoSize` is provided, `WithQRCodeLogoRatio` controls the logo long edge relative to the QR size; the default ratio is `6`.

## Decode QR codes and barcodes

```go
package main

import (
	"bytes"
	"fmt"

	"github.com/imajinyun/go-knifer/vimg"
)

func main() {
	pngBytes, err := vimg.QRCodePNG("decode payload")
	if err != nil {
		panic(err)
	}

	result, err := vimg.DecodeQRCode(bytes.NewReader(pngBytes), vimg.WithDecodeTryHarder(true))
	if err != nil {
		panic(err)
	}
	fmt.Println(result.Text, result.Format == vimg.BarcodeFormatQRCode)
}
```

Use `DecodeQRCode` / `DecodeQRCodeImage` for QR-only reads. Use `DecodeBarcode` / `DecodeBarcodeImage` with `WithDecodeFormats` when you need to restrict accepted formats. Decoding tries both ZXing hybrid and global histogram binarizers, which helps with different contrast and thresholding conditions.

## Inspect supported barcode formats

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vimg"
)

func main() {
	fmt.Println(vimg.CanEncodeBarcodeFormat(vimg.BarcodeFormatQRCode))
	fmt.Println(vimg.CanDecodeBarcodeFormat(vimg.BarcodeFormatAztec))
	fmt.Println(len(vimg.SupportedEncodeBarcodeFormats()) > 0)
	fmt.Println(len(vimg.SupportedDecodeBarcodeFormats()) > 0)
}
```

Encode and decode support are not identical. For example, Aztec currently has a reader but no writer in this facade. Check `SupportedEncodeBarcodeFormats` and `SupportedDecodeBarcodeFormats` before exposing a user-selectable format list.

## Render custom ASCII or dispatch output by type

```go
package main

import (
	"fmt"
	"strings"

	"github.com/imajinyun/go-knifer/vimg"
)

func main() {
	ascii, err := vimg.QRCodeASCIIWithChars("ascii payload", "##", "..")
	if err != nil {
		panic(err)
	}

	svgBytes, err := vimg.QRCodeBytes("output payload", vimg.BarcodeOutputFormatSVG)
	if err != nil {
		panic(err)
	}

	fmt.Println(strings.Contains(ascii, "##"), strings.Contains(string(svgBytes), "<svg"))
}
```

`BarcodeBytes` and `QRCodeBytes` are useful when a caller stores the output format in configuration. Supported output formats are `BarcodeOutputFormatPNG`, `BarcodeOutputFormatSVG`, `BarcodeOutputFormatASCII`, and `BarcodeOutputFormatBase64Data`.

## Generate and verify captchas

```go
package main

import (
	"fmt"
	"strings"

	"github.com/imajinyun/go-knifer/vimg"
)

func main() {
	captcha := vimg.NewLineCaptchaWithOptions(120, 48,
		vimg.WithGenerator(vimg.NewRandomGeneratorWithBase("ABC123", 4)),
		vimg.WithInterfereCount(8),
	)
	captcha.CreateCode()

	code := captcha.Code()
	fmt.Println(captcha.Verify(strings.ToLower(code)))
	fmt.Println(strings.HasPrefix(captcha.ImageBase64Data(), "data:image/png;base64,"))
}
```

`ImageBytes` returns a copy of the encoded image bytes each time. Mutating that returned slice will not corrupt later `ImageBytes`, `ImageBase64`, `Write`, or file-write calls.

## Use math captchas and write options

```go
package main

import (
	"fmt"
	"path/filepath"

	"github.com/imajinyun/go-knifer/vimg"
)

func main() {
	generator := vimg.NewMathGeneratorWith(1, false)
	fmt.Println(generator.Verify("1+1=", "2"))

	captcha := vimg.NewGifCaptchaWithOptions(120, 48, vimg.WithGenerator(generator))
	captcha.CreateCode()
	path := filepath.Join("tmp", "captcha.gif")
	_ = captcha.WriteToFileWithOptions(path,
		vimg.WithCreateParents(true),
		vimg.WithFilePerm(0o600),
		vimg.WithOverwrite(true),
	)
}
```
