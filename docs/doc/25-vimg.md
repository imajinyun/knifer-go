# vimg Quickstart

`vimg` provides image processing, QR/barcode, and captcha helpers. It covers image metadata reads, format conversion, thumbnail generation, ZXing-backed QR/barcode generation and decoding, PNG/SVG/ASCII/Base64 data URI output, QR logo embedding, transparent backgrounds, captcha generation/verification, and file writes.

Captcha image bytes are returned as defensive copies, so callers can inspect or transform the returned slice without mutating the captcha's cached image. File overwrite failures from captcha writers preserve `fs.ErrExist` and carry the go-knifer invalid-input error code for consistent error inspection.

## Which helper should I use?

| Goal | Start with | Notes |
| --- | --- | --- |
| Read dimensions and format | `Info` | Decodes image configuration without transforming pixels. |
| Convert image format | `ConvertFormat` | Use when the original dimensions should be preserved and only encoding changes. |
| Generate thumbnails | `Thumbnail` | Resizes to a maximum edge and writes in the requested format. |
| Generate QR codes | `QRCodePNG`, `QRCodeSVG`, `QRCodeBytes`, `QRCodeImage` | QR helpers wrap barcode helpers with `BarcodeFormatQRCode`. |
| Generate non-QR barcodes | `BarcodePNG`, `BarcodeSVG`, `BarcodeASCII`, `BarcodeBytes` | Check `CanEncodeBarcodeFormat` before exposing user-selected formats. |
| Add a QR logo | `WithQRCodeLogo`, `WithQRCodeLogoSize`, `WithQRCodeLogoRatio` | Use high error correction and test scanability after adding logos. |
| Decode QR only | `DecodeQRCode`, `DecodeQRCodeImage` | Restricts decode results to QR content. |
| Decode selected formats | `DecodeBarcode`, `DecodeBarcodeImage`, `WithDecodeFormats` | Restrict accepted formats when image input is untrusted or ambiguous. |
| Generate captchas | `NewLineCaptchaWithOptions`, `NewCircleCaptchaWithOptions`, `NewShearCaptchaWithOptions`, `NewGifCaptchaWithOptions` | Inject generators and random functions for deterministic tests. |
| Write captcha files safely | `WriteToFileWithOptions` with `WithCreateParents`, `WithOverwrite`, `WithOpenFile` | Use file and directory permission options for local storage contracts. |

## Image safety checklist

- Bound uploaded image size before passing readers to decoders; image decoding can allocate based on dimensions and format.
- Validate barcode formats with `CanEncodeBarcodeFormat` and `CanDecodeBarcodeFormat` before accepting user configuration.
- Keep QR logo size conservative and use `QRErrorCorrectionHigh` when embedding logos, then verify with real scanners.
- Use `WithDecodeFormats` to limit barcode decoding to expected symbologies instead of accepting every supported reader.
- Inject `WithRandomInt`, `WithGeneratorRandomInt`, and `WithOpenFile` in tests to avoid nondeterministic captchas or filesystem writes.
- Review `WithOverwrite`, `WithCreateParents`, `WithFilePerm`, and `WithDirPerm` before writing files from user-controlled paths.
- Treat captcha verification as a UX friction layer, not as a replacement for rate limiting or abuse detection.

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

## When not to use vimg

- Use a dedicated image pipeline when you need streaming transforms, EXIF preservation, color management, or advanced resampling controls.
- Use external services or libraries when barcode requirements include formats or GS1 variants not supported by the facade.
- Use server-side session/rate-limit controls for abuse prevention; captchas alone are not an authorization or anti-fraud system.
- Avoid QR logo embedding when the code must be read under poor lighting, damaged print, or very small display sizes.

## Related packages

- Use `vfile` when image paths, temporary files, or directory traversal need filesystem policy checks.
- Use `vhttp` or `vresty` when images are fetched from remote URLs before processing.
- Use `vrand` when captcha or placeholder generation requires random bytes or deterministic test sources.

## Benchmarks and trade-offs

- `Info` is cheaper than full decode/encode paths because it only reads metadata. Use it for validation before expensive work.
- `Thumbnail` and `ConvertFormat` decode and re-encode images, so CPU and memory scale with pixel count and output format.
- SVG and ASCII barcode outputs are useful for text or vector contexts; PNG and image outputs are better for raster workflows.
- `DecodeBarcode` tries multiple binarizers for robustness, which costs more than reading a clean, known-format QR image.
- Captcha image byte access returns defensive copies, trading allocation for protection against callers mutating cached captcha output.

## FAQ

### Why can QR logos make codes unreadable?

The logo covers modules that scanners need. Use high error correction, keep the logo small with `WithQRCodeLogoRatio` or `WithQRCodeLogoSize`, and test the generated code on target devices.

### Why are encode and decode format lists different?

The underlying ZXing-backed readers and writers do not support identical format sets. Use `SupportedEncodeBarcodeFormats` and `SupportedDecodeBarcodeFormats` separately.

### How do I make captcha tests deterministic?

Pass a fixed generator with `WithGenerator`, or inject randomness through `WithRandomInt` and `WithGeneratorRandomInt`. For file writes, inject `WithOpenFile` or `WithMkdirAll`.

### Should I use `BarcodeBytes` or format-specific helpers?

Use format-specific helpers when the output type is fixed at compile time. Use `BarcodeBytes` or `QRCodeBytes` when output format comes from configuration.
