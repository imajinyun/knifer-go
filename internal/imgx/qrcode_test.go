package imgx

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"strings"
	"testing"

	knifer "github.com/imajinyun/go-knifer"
)

func TestQRCodePNGAndDecode(t *testing.T) {
	const content = "https://github.com/imajinyun/go-knifer"
	pngBytes, err := QRCodePNG(content,
		WithQRCodeSize(180),
		WithQRCodeMargin(2),
		WithQRCodeErrorCorrection(QRErrorCorrectionMedium),
	)
	if err != nil {
		t.Fatalf("QRCodePNG: %v", err)
	}
	if len(pngBytes) == 0 {
		t.Fatal("QRCodePNG returned empty bytes")
	}
	cfg, err := png.DecodeConfig(bytes.NewReader(pngBytes))
	if err != nil {
		t.Fatalf("png config: %v", err)
	}
	if cfg.Width != 180 || cfg.Height != 180 {
		t.Fatalf("QRCodePNG size = %dx%d, want 180x180", cfg.Width, cfg.Height)
	}

	result, err := DecodeQRCode(bytes.NewReader(pngBytes), WithDecodeTryHarder(true))
	if err != nil {
		t.Fatalf("DecodeQRCode: %v", err)
	}
	if result.Text != content {
		t.Fatalf("decoded text = %q, want %q", result.Text, content)
	}
	if result.Format != BarcodeFormatQRCode {
		t.Fatalf("decoded format = %v, want %v", result.Format, BarcodeFormatQRCode)
	}
}

func TestBarcodeCode128PNGAndDecode(t *testing.T) {
	const content = "GO-KNIFER-128"
	pngBytes, err := BarcodePNG(content, BarcodeFormatCode128,
		WithBarcodeSize(260, 90),
		WithBarcodeMargin(4),
	)
	if err != nil {
		t.Fatalf("BarcodePNG: %v", err)
	}
	cfg, err := png.DecodeConfig(bytes.NewReader(pngBytes))
	if err != nil {
		t.Fatalf("png config: %v", err)
	}
	if cfg.Width != 260 || cfg.Height != 90 {
		t.Fatalf("BarcodePNG size = %dx%d, want 260x90", cfg.Width, cfg.Height)
	}

	result, err := DecodeBarcode(bytes.NewReader(pngBytes),
		WithDecodeFormats(BarcodeFormatCode128),
		WithDecodeTryHarder(true),
	)
	if err != nil {
		t.Fatalf("DecodeBarcode: %v", err)
	}
	if result.Text != content {
		t.Fatalf("decoded text = %q, want %q", result.Text, content)
	}
	if result.Format != BarcodeFormatCode128 {
		t.Fatalf("decoded format = %v, want %v", result.Format, BarcodeFormatCode128)
	}
}

func TestBarcodeFormatCapabilities(t *testing.T) {
	allFormats := []BarcodeFormat{
		BarcodeFormatUnknown,
		BarcodeFormatAztec,
		BarcodeFormatCodabar,
		BarcodeFormatCode39,
		BarcodeFormatCode93,
		BarcodeFormatCode128,
		BarcodeFormatDataMatrix,
		BarcodeFormatEAN8,
		BarcodeFormatEAN13,
		BarcodeFormatITF,
		BarcodeFormatMaxiCode,
		BarcodeFormatPDF417,
		BarcodeFormatQRCode,
		BarcodeFormatRSS14,
		BarcodeFormatRSSExpanded,
		BarcodeFormatUPCA,
		BarcodeFormatUPCE,
	}
	encodable := map[BarcodeFormat]bool{
		BarcodeFormatCodabar:    true,
		BarcodeFormatCode39:     true,
		BarcodeFormatCode93:     true,
		BarcodeFormatCode128:    true,
		BarcodeFormatDataMatrix: true,
		BarcodeFormatEAN8:       true,
		BarcodeFormatEAN13:      true,
		BarcodeFormatITF:        true,
		BarcodeFormatQRCode:     true,
		BarcodeFormatUPCA:       true,
		BarcodeFormatUPCE:       true,
	}
	decodable := map[BarcodeFormat]bool{
		BarcodeFormatAztec:      true,
		BarcodeFormatCodabar:    true,
		BarcodeFormatCode39:     true,
		BarcodeFormatCode93:     true,
		BarcodeFormatCode128:    true,
		BarcodeFormatDataMatrix: true,
		BarcodeFormatEAN8:       true,
		BarcodeFormatEAN13:      true,
		BarcodeFormatITF:        true,
		BarcodeFormatQRCode:     true,
		BarcodeFormatRSS14:      true,
		BarcodeFormatUPCA:       true,
		BarcodeFormatUPCE:       true,
	}

	for _, format := range allFormats {
		t.Run(format.String(), func(t *testing.T) {
			if got := CanEncodeBarcodeFormat(format); got != encodable[format] {
				t.Fatalf("CanEncodeBarcodeFormat(%v) = %v, want %v", format, got, encodable[format])
			}
			if got := CanDecodeBarcodeFormat(format); got != decodable[format] {
				t.Fatalf("CanDecodeBarcodeFormat(%v) = %v, want %v", format, got, decodable[format])
			}
		})
	}

	if got, want := formatSet(SupportedEncodeBarcodeFormats()), encodable; !sameFormatSet(got, want) {
		t.Fatalf("SupportedEncodeBarcodeFormats = %v, want %v", got, want)
	}
	if got, want := formatSet(SupportedDecodeBarcodeFormats()), decodable; !sameFormatSet(got, want) {
		t.Fatalf("SupportedDecodeBarcodeFormats = %v, want %v", got, want)
	}
}

func TestQRCodeSVGAndASCII(t *testing.T) {
	svg, err := QRCodeSVG("svg payload", WithQRCodeSize(64), WithQRCodeForeground(color.RGBA{R: 1, G: 2, B: 3, A: 255}))
	if err != nil {
		t.Fatalf("QRCodeSVG: %v", err)
	}
	for _, want := range []string{`<svg xmlns="http://www.w3.org/2000/svg"`, `fill="#010203"`, `<path`} {
		want = strings.ReplaceAll(want, `\"`, `"`)
		if !strings.Contains(svg, want) {
			t.Fatalf("QRCodeSVG missing %q in %q", want, svg[:min(len(svg), 120)])
		}
	}

	ascii, err := QRCodeASCII("ascii payload", WithQRCodeSize(33), WithQRCodeMargin(1))
	if err != nil {
		t.Fatalf("QRCodeASCII: %v", err)
	}
	if !strings.Contains(ascii, "██") || !strings.Contains(ascii, "\n") {
		t.Fatalf("QRCodeASCII did not contain expected blocks/newlines: %q", ascii[:min(len(ascii), 80)])
	}

	customASCII, err := QRCodeASCIIWithChars("ascii payload", "##", "..", WithQRCodeSize(33), WithQRCodeMargin(1))
	if err != nil {
		t.Fatalf("QRCodeASCIIWithChars: %v", err)
	}
	if !strings.Contains(customASCII, "##") || !strings.Contains(customASCII, "..") {
		t.Fatalf("QRCodeASCIIWithChars missing custom chars: %q", customASCII[:min(len(customASCII), 80)])
	}
}

func TestQRCodeOutputs(t *testing.T) {
	cases := []struct {
		name   string
		format BarcodeOutputFormat
		want   string
	}{
		{"png", BarcodeOutputFormatPNG, "\x89PNG"},
		{"svg", BarcodeOutputFormatSVG, "<svg"},
		{"ascii", BarcodeOutputFormatASCII, "██"},
		{"base64 data", BarcodeOutputFormatBase64Data, "data:image/png;base64,"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := QRCodeBytes("output payload", tc.format, WithQRCodeSize(64))
			if err != nil {
				t.Fatalf("QRCodeBytes: %v", err)
			}
			if !strings.Contains(string(data), tc.want) {
				t.Fatalf("QRCodeBytes missing %q in %q", tc.want, string(data[:min(len(data), 80)]))
			}
		})
	}
}

func TestQRCodeTransparentBackground(t *testing.T) {
	img, err := QRCodeImage("transparent payload", WithQRCodeSize(80), WithQRCodeTransparentBackground())
	if err != nil {
		t.Fatalf("QRCodeImage: %v", err)
	}
	_, _, _, a := img.At(0, 0).RGBA()
	if a != 0 {
		t.Fatalf("background alpha = %d, want transparent", a)
	}

	svg, err := QRCodeSVG("transparent payload", WithQRCodeSize(80), WithQRCodeTransparentBackground())
	if err != nil {
		t.Fatalf("QRCodeSVG: %v", err)
	}
	if !strings.Contains(svg, `fill="rgba(0,0,0,0.000)"`) {
		t.Fatalf("QRCodeSVG missing transparent fill: %q", svg[:min(len(svg), 160)])
	}
}

func TestQRCodeSVGWithLogo(t *testing.T) {
	logo := solidTestLogo(color.RGBA{G: 255, A: 255})
	svg, err := QRCodeSVG("svg logo payload",
		WithQRCodeSize(120),
		WithQRCodeErrorCorrection(QRErrorCorrectionHigh),
		WithQRCodeLogo(logo),
		WithQRCodeLogoRatio(6),
	)
	if err != nil {
		t.Fatalf("QRCodeSVG: %v", err)
	}
	for _, want := range []string{`<image `, `href="data:image/png;base64,`, `<rect`} {
		want = strings.ReplaceAll(want, `\"`, `"`)
		if !strings.Contains(svg, want) {
			t.Fatalf("QRCodeSVG missing %q in %q", want, svg[:min(len(svg), 160)])
		}
	}
}

func TestQRCodeLogoRendering(t *testing.T) {
	logo := solidTestLogo(color.RGBA{R: 255, A: 255})
	img, err := QRCodeImage("logo payload",
		WithQRCodeSize(120),
		WithQRCodeErrorCorrection(QRErrorCorrectionHigh),
		WithQRCodeLogo(logo),
		WithQRCodeLogoSize(18, 18),
	)
	if err != nil {
		t.Fatalf("QRCodeImage: %v", err)
	}
	r, g, b, a := img.At(60, 60).RGBA()
	if r>>8 != 255 || g>>8 != 0 || b>>8 != 0 || a>>8 != 255 {
		t.Fatalf("center pixel = rgba(%d,%d,%d,%d), want red", r>>8, g>>8, b>>8, a>>8)
	}
}

func TestBarcodeBase64Data(t *testing.T) {
	data, err := QRCodeBase64Data("data uri payload", WithQRCodeSize(80))
	if err != nil {
		t.Fatalf("QRCodeBase64Data: %v", err)
	}
	if !strings.HasPrefix(data, "data:image/png;base64,") {
		t.Fatalf("QRCodeBase64Data prefix = %q", data[:min(len(data), 30)])
	}
}

func TestQRCodeInvalidArgsClassified(t *testing.T) {
	cases := []struct {
		name string
		err  error
		code knifer.ErrCode
	}{
		{"empty content", func() error { _, err := QRCodePNG(""); return err }(), knifer.ErrCodeInvalidInput},
		{"bad size", func() error { _, err := QRCodePNG("x", WithQRCodeSize(0)); return err }(), knifer.ErrCodeInvalidInput},
		{"bad margin", func() error { _, err := QRCodePNG("x", WithQRCodeMargin(-1)); return err }(), knifer.ErrCodeInvalidInput},
		{"unsupported writer", func() error { _, err := BarcodePNG("x", BarcodeFormatAztec); return err }(), knifer.ErrCodeUnsupported},
		{"non qr logo unsupported", func() error {
			_, err := BarcodePNG("123456789012", BarcodeFormatEAN13, WithBarcodeLogo(image.NewRGBA(image.Rect(0, 0, 1, 1))))
			return err
		}(), knifer.ErrCodeUnsupported},
		{"bad logo ratio", func() error { _, err := QRCodePNG("x", WithQRCodeLogoRatio(0)); return err }(), knifer.ErrCodeInvalidInput},
		{"bad ascii chars", func() error { _, err := QRCodeASCIIWithChars("x", "", " "); return err }(), knifer.ErrCodeInvalidInput},
		{"bad output format", func() error { _, err := QRCodeBytes("x", BarcodeOutputFormatUnknown); return err }(), knifer.ErrCodeInvalidInput},
		{"unsupported decode format", func() error {
			_, err := DecodeBarcode(strings.NewReader("not an image"), WithDecodeFormats(BarcodeFormatPDF417))
			return err
		}(), knifer.ErrCodeUnsupported},
		{"decode invalid", func() error { _, err := DecodeQRCode(strings.NewReader("not an image")); return err }(), knifer.ErrCodeInvalidInput},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.err == nil {
				t.Fatal("expected error, got nil")
			}
			code, ok := knifer.CodeOf(tc.err)
			if !ok {
				t.Fatalf("expected classified error, got %v", tc.err)
			}
			if code != tc.code {
				t.Fatalf("error code = %v, want %v", code, tc.code)
			}
		})
	}
}

func solidTestLogo(c color.Color) image.Image {
	logo := image.NewRGBA(image.Rect(0, 0, 12, 12))
	for y := 0; y < 12; y++ {
		for x := 0; x < 12; x++ {
			logo.Set(x, y, c)
		}
	}
	return logo
}

func formatSet(formats []BarcodeFormat) map[BarcodeFormat]bool {
	out := make(map[BarcodeFormat]bool, len(formats))
	for _, format := range formats {
		out[format] = true
	}
	return out
}

func sameFormatSet(got, want map[BarcodeFormat]bool) bool {
	if len(got) != len(want) {
		return false
	}
	for format, supported := range want {
		if got[format] != supported {
			return false
		}
	}
	return true
}
