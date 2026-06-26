package imgx

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"strings"
	"testing"

	knifer "github.com/imajinyun/knifer-go"
	"github.com/makiuchi-d/gozxing"
)

func TestQRCodePNGAndDecode(t *testing.T) {
	const content = "https://github.com/imajinyun/knifer-go"
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
		{"nil foreground", func() error { _, err := QRCodePNG("x", WithQRCodeForeground(nil)); return err }(), knifer.ErrCodeInvalidInput},
		{"nil background", func() error { _, err := QRCodePNG("x", WithQRCodeBackground(nil)); return err }(), knifer.ErrCodeInvalidInput},
		{"nil colors foreground", func() error { _, err := QRCodePNG("x", WithQRCodeColors(nil, color.White)); return err }(), knifer.ErrCodeInvalidInput},
		{"nil colors background", func() error { _, err := QRCodePNG("x", WithQRCodeColors(color.Black, nil)); return err }(), knifer.ErrCodeInvalidInput},
		{"empty character set", func() error { _, err := QRCodePNG("x", WithBarcodeCharacterSet(" ")); return err }(), knifer.ErrCodeInvalidInput},
		{"bad qr version", func() error { _, err := QRCodePNG("x", WithQRCodeVersion(41)); return err }(), knifer.ErrCodeInvalidInput},
		{"bad qr mask", func() error { _, err := QRCodePNG("x", WithQRCodeMaskPattern(8)); return err }(), knifer.ErrCodeInvalidInput},
		{"bad force code set", func() error {
			_, err := BarcodePNG("x", BarcodeFormatCode128, WithBarcodeForceCodeSet("D"))
			return err
		}(), knifer.ErrCodeInvalidInput},
		{"nil logo", func() error { _, err := QRCodePNG("x", WithQRCodeLogo(nil)); return err }(), knifer.ErrCodeInvalidInput},
		{"bad logo size", func() error { _, err := QRCodePNG("x", WithQRCodeLogoSize(0, 1)); return err }(), knifer.ErrCodeInvalidInput},
		{"bad error correction", func() error {
			_, err := QRCodePNG("x", WithQRCodeErrorCorrection(QRErrorCorrectionUnknown))
			return err
		}(), knifer.ErrCodeInvalidInput},
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
		{"empty decode formats", func() error {
			_, err := DecodeBarcode(strings.NewReader("not an image"), WithDecodeFormats())
			return err
		}(), knifer.ErrCodeInvalidInput},
		{"empty decode character set", func() error {
			_, err := DecodeBarcode(strings.NewReader("not an image"), WithDecodeCharacterSet(" "))
			return err
		}(), knifer.ErrCodeInvalidInput},
		{"nil writer", func() error { return WriteQRCode(nil, "x") }(), knifer.ErrCodeInvalidInput},
		{"nil image decode", func() error { _, err := DecodeQRCodeImage(nil); return err }(), knifer.ErrCodeInvalidInput},
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

func TestQRCodeAdditionalOptionsAndWrappers(t *testing.T) {
	logo := solidTestLogo(color.RGBA{B: 255, A: 255})
	data, err := QRCodePNG("configured payload",
		WithQRCodeSize(96),
		WithQRCodeColors(color.RGBA{R: 10, A: 255}, color.RGBA{G: 20, A: 255}),
		WithBarcodeCharacterSet("UTF-8"),
		WithQRCodeVersion(4),
		WithQRCodeMaskPattern(2),
		WithQRCodeLogo(logo),
		WithQRCodeLogoSize(12, 12),
		WithBarcodeEncodeHint(gozxing.EncodeHintType_MARGIN, 1),
		nil,
	)
	if err != nil {
		t.Fatalf("QRCodePNG with options: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("QRCodePNG returned empty data")
	}

	var buf bytes.Buffer
	if err := WriteQRCode(&buf, "write qr payload", WithQRCodeSize(64)); err != nil {
		t.Fatalf("WriteQRCode: %v", err)
	}
	if !bytes.HasPrefix(buf.Bytes(), []byte("\x89PNG")) {
		t.Fatalf("WriteQRCode prefix = %q, want PNG", buf.Bytes()[:min(buf.Len(), 4)])
	}

	img, err := QRCodeImage("decode image payload", WithQRCodeSize(160), WithQRCodeMargin(2))
	if err != nil {
		t.Fatalf("QRCodeImage: %v", err)
	}
	result, err := DecodeQRCodeImage(img, WithDecodeTryHarder(true), WithDecodePureBarcode(true), WithDecodeAlsoInverted(true), WithDecodeCharacterSet("UTF-8"), WithDecodeHint(gozxing.DecodeHintType_OTHER, "ignored"))
	if err != nil {
		t.Fatalf("DecodeQRCodeImage: %v", err)
	}
	if result.Text != "decode image payload" || result.Format != BarcodeFormatQRCode {
		t.Fatalf("DecodeQRCodeImage result = %+v", result)
	}
}

func TestBarcodeOneDimensionalOptions(t *testing.T) {
	pngBytes, err := BarcodePNG("123456789012", BarcodeFormatCode128,
		WithBarcodeSize(240, 80),
		WithBarcodeColors(color.Black, color.White),
		WithBarcodeCharacterSet("UTF-8"),
		WithBarcodeForceCodeSet(" b "),
		WithBarcodeGS1Format(false),
	)
	if err != nil {
		t.Fatalf("BarcodePNG: %v", err)
	}
	if !bytes.HasPrefix(pngBytes, []byte("\x89PNG")) {
		t.Fatalf("BarcodePNG prefix = %q, want PNG", pngBytes[:min(len(pngBytes), 4)])
	}
}

func TestBarcodeInternalFormatMappings(t *testing.T) {
	formats := []BarcodeFormat{
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
	for _, format := range formats {
		t.Run(format.String(), func(t *testing.T) {
			goFormat, err := toGozxingBarcodeFormat(format)
			if err != nil {
				t.Fatalf("toGozxingBarcodeFormat: %v", err)
			}
			if got := fromGozxingBarcodeFormat(goFormat); got != format {
				t.Fatalf("fromGozxingBarcodeFormat = %v, want %v", got, format)
			}
		})
	}
	if _, err := toGozxingBarcodeFormat(BarcodeFormatUnknown); !errorsIsCode(err, knifer.ErrCodeUnsupported) {
		t.Fatalf("toGozxingBarcodeFormat unknown error = %v, want unsupported", err)
	}
	if got := fromGozxingBarcodeFormat(gozxing.BarcodeFormat(-1)); got != BarcodeFormatUnknown {
		t.Fatalf("fromGozxingBarcodeFormat unknown = %v, want unknown", got)
	}

	levels := []QRErrorCorrectionLevel{
		QRErrorCorrectionLow,
		QRErrorCorrectionMedium,
		QRErrorCorrectionQuartile,
		QRErrorCorrectionHigh,
	}
	for _, level := range levels {
		if _, err := toGozxingQRECLevel(level); err != nil {
			t.Fatalf("toGozxingQRECLevel(%v): %v", level, err)
		}
	}
}

func TestBarcodeWriterAndReaderDispatch(t *testing.T) {
	writerFormats := []BarcodeFormat{
		BarcodeFormatCodabar,
		BarcodeFormatCode39,
		BarcodeFormatCode93,
		BarcodeFormatCode128,
		BarcodeFormatDataMatrix,
		BarcodeFormatEAN8,
		BarcodeFormatEAN13,
		BarcodeFormatITF,
		BarcodeFormatQRCode,
		BarcodeFormatUPCA,
		BarcodeFormatUPCE,
	}
	for _, format := range writerFormats {
		t.Run("writer_"+format.String(), func(t *testing.T) {
			if _, err := barcodeWriter(format); err != nil {
				t.Fatalf("barcodeWriter(%v): %v", format, err)
			}
		})
	}
	if _, err := barcodeWriter(BarcodeFormatPDF417); !errorsIsCode(err, knifer.ErrCodeUnsupported) {
		t.Fatalf("barcodeWriter unsupported = %v", err)
	}

	readers := barcodeReaders([]BarcodeFormat{BarcodeFormatQRCode, BarcodeFormatDataMatrix, BarcodeFormatCode128, BarcodeFormatCode39, BarcodeFormatCode93, BarcodeFormatEAN13, BarcodeFormatEAN8, BarcodeFormatUPCA, BarcodeFormatUPCE, BarcodeFormatITF, BarcodeFormatCodabar, BarcodeFormatRSS14, BarcodeFormatAztec, BarcodeFormatPDF417})
	if got, want := len(readers), 13; got != want {
		t.Fatalf("barcodeReaders count = %d, want %d", got, want)
	}
	if got := len(barcodeReaders(nil)); got != len(SupportedDecodeBarcodeFormats()) {
		t.Fatalf("barcodeReaders(nil) = %d, want supported format count", got)
	}
}

func errorsIsCode(err error, code knifer.ErrCode) bool {
	got, ok := knifer.CodeOf(err)
	return ok && got == code
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
