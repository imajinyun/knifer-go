package vimg

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"strings"
	"testing"
)

func TestFacadeQRCode(t *testing.T) {
	const content = "facade qr payload"
	pngBytes, err := QRCodePNG(content,
		WithQRCodeSize(150),
		WithQRCodeMargin(2),
		WithQRCodeErrorCorrection(QRErrorCorrectionQuartile),
	)
	if err != nil {
		t.Fatalf("QRCodePNG: %v", err)
	}
	cfg, err := png.DecodeConfig(bytes.NewReader(pngBytes))
	if err != nil {
		t.Fatalf("png config: %v", err)
	}
	if cfg.Width != 150 || cfg.Height != 150 {
		t.Fatalf("QRCodePNG size = %dx%d, want 150x150", cfg.Width, cfg.Height)
	}
	result, err := DecodeQRCode(bytes.NewReader(pngBytes), WithDecodeTryHarder(true))
	if err != nil {
		t.Fatalf("DecodeQRCode: %v", err)
	}
	if result.Text != content || result.Format != BarcodeFormatQRCode {
		t.Fatalf("DecodeQRCode = (%q, %v), want (%q, %v)", result.Text, result.Format, content, BarcodeFormatQRCode)
	}
}

func TestFacadeBarcodeRenderers(t *testing.T) {
	pngBytes, err := BarcodePNG("123456789012", BarcodeFormatEAN13, WithBarcodeSize(220, 90))
	if err != nil {
		t.Fatalf("BarcodePNG: %v", err)
	}
	if _, err := png.DecodeConfig(bytes.NewReader(pngBytes)); err != nil {
		t.Fatalf("png config: %v", err)
	}

	svg, err := QRCodeSVG("facade svg", WithQRCodeForeground(color.Black))
	if err != nil {
		t.Fatalf("QRCodeSVG: %v", err)
	}
	if !strings.Contains(svg, "<svg") || !strings.Contains(svg, "<path") {
		t.Fatalf("QRCodeSVG missing expected tags: %q", svg[:min(len(svg), 80)])
	}

	ascii, err := QRCodeASCII("facade ascii")
	if err != nil {
		t.Fatalf("QRCodeASCII: %v", err)
	}
	if !strings.Contains(ascii, "██") {
		t.Fatalf("QRCodeASCII missing block chars: %q", ascii[:min(len(ascii), 80)])
	}

	data, err := BarcodeBase64Data("facade data", BarcodeFormatQRCode)
	if err != nil {
		t.Fatalf("BarcodeBase64Data: %v", err)
	}
	if !strings.HasPrefix(data, "data:image/png;base64,") {
		t.Fatalf("BarcodeBase64Data prefix = %q", data[:min(len(data), 30)])
	}

	customASCII, err := QRCodeASCIIWithChars("facade ascii chars", "##", "..")
	if err != nil {
		t.Fatalf("QRCodeASCIIWithChars: %v", err)
	}
	if !strings.Contains(customASCII, "##") || !strings.Contains(customASCII, "..") {
		t.Fatalf("QRCodeASCIIWithChars missing custom chars: %q", customASCII[:min(len(customASCII), 80)])
	}

	output, err := QRCodeBytes("facade output", BarcodeOutputFormatSVG, WithQRCodeTransparentBackground())
	if err != nil {
		t.Fatalf("QRCodeBytes: %v", err)
	}
	if !strings.Contains(string(output), `<svg`) || !strings.Contains(string(output), `rgba(0,0,0,0.000)`) {
		t.Fatalf("QRCodeBytes output missing expected SVG content: %q", string(output[:min(len(output), 120)]))
	}
}

func TestFacadeBarcodeCapabilities(t *testing.T) {
	if !CanEncodeBarcodeFormat(BarcodeFormatQRCode) || !CanDecodeBarcodeFormat(BarcodeFormatQRCode) {
		t.Fatal("qr code should be supported for encode and decode")
	}
	if CanEncodeBarcodeFormat(BarcodeFormatAztec) || !CanDecodeBarcodeFormat(BarcodeFormatAztec) {
		t.Fatal("aztec should be decode-only")
	}
	if len(SupportedEncodeBarcodeFormats()) == 0 || len(SupportedDecodeBarcodeFormats()) == 0 {
		t.Fatal("supported format lists must not be empty")
	}
}

func TestFacadeBarcodeOptionSetters(t *testing.T) {
	// BarcodeOption setters
	if opt := WithBarcodeMargin(4); opt == nil {
		t.Fatal("WithBarcodeMargin nil")
	}
	if opt := WithBarcodeForeground(color.Black); opt == nil {
		t.Fatal("WithBarcodeForeground nil")
	}
	if opt := WithBarcodeBackground(color.White); opt == nil {
		t.Fatal("WithBarcodeBackground nil")
	}
	if opt := WithBarcodeTransparentBackground(); opt == nil {
		t.Fatal("WithBarcodeTransparentBackground nil")
	}
	if opt := WithBarcodeColors(color.Black, color.White); opt == nil {
		t.Fatal("WithBarcodeColors nil")
	}
	if opt := WithBarcodeCharacterSet("ISO-8859-1"); opt == nil {
		t.Fatal("WithBarcodeCharacterSet nil")
	}
	if opt := WithBarcodeForceCodeSet("B"); opt == nil {
		t.Fatal("WithBarcodeForceCodeSet nil")
	}
	if opt := WithBarcodeGS1Format(true); opt == nil {
		t.Fatal("WithBarcodeGS1Format nil")
	}
	logo := image.NewRGBA(image.Rect(0, 0, 10, 10))
	if opt := WithBarcodeLogo(logo); opt == nil {
		t.Fatal("WithBarcodeLogo nil")
	}
	if opt := WithBarcodeLogoSize(20, 20); opt == nil {
		t.Fatal("WithBarcodeLogoSize nil")
	}
	// QRCodeOption setters
	if opt := WithQRCodeBackground(color.White); opt == nil {
		t.Fatal("WithQRCodeBackground nil")
	}
	if opt := WithQRCodeColors(color.Black, color.White); opt == nil {
		t.Fatal("WithQRCodeColors nil")
	}
	if opt := WithQRCodeVersion(5); opt == nil {
		t.Fatal("WithQRCodeVersion nil")
	}
	if opt := WithQRCodeMaskPattern(2); opt == nil {
		t.Fatal("WithQRCodeMaskPattern nil")
	}
	if opt := WithQRCodeLogo(logo); opt == nil {
		t.Fatal("WithQRCodeLogo nil")
	}
	if opt := WithQRCodeLogoSize(15, 15); opt == nil {
		t.Fatal("WithQRCodeLogoSize nil")
	}
	if opt := WithQRCodeLogoRatio(4); opt == nil {
		t.Fatal("WithQRCodeLogoRatio nil")
	}
	// DecodeOption setters
	if opt := WithDecodeFormats(BarcodeFormatQRCode); opt == nil {
		t.Fatal("WithDecodeFormats nil")
	}
	if opt := WithDecodePureBarcode(true); opt == nil {
		t.Fatal("WithDecodePureBarcode nil")
	}
	if opt := WithDecodeAlsoInverted(true); opt == nil {
		t.Fatal("WithDecodeAlsoInverted nil")
	}
	if opt := WithDecodeCharacterSet("UTF-8"); opt == nil {
		t.Fatal("WithDecodeCharacterSet nil")
	}
}

func TestFacadeBarcodeImageAndWrite(t *testing.T) {
	img, err := BarcodeImage("123456789012", BarcodeFormatEAN13, WithBarcodeSize(220, 90))
	if err != nil {
		t.Fatalf("BarcodeImage: %v", err)
	}
	if img.Bounds().Dx() != 220 || img.Bounds().Dy() != 90 {
		t.Fatalf("BarcodeImage size = %dx%d", img.Bounds().Dx(), img.Bounds().Dy())
	}
	qImg, err := QRCodeImage("facade image")
	if err != nil {
		t.Fatalf("QRCodeImage: %v", err)
	}
	if qImg.Bounds().Dx() <= 0 || qImg.Bounds().Dy() <= 0 {
		t.Fatal("QRCodeImage has invalid bounds")
	}
	var buf bytes.Buffer
	if err := WriteBarcode(&buf, "987654321098", BarcodeFormatEAN13); err != nil {
		t.Fatalf("WriteBarcode: %v", err)
	}
	if buf.Len() == 0 {
		t.Fatal("WriteBarcode empty")
	}
	buf.Reset()
	if err := WriteQRCode(&buf, "write qr"); err != nil {
		t.Fatalf("WriteQRCode: %v", err)
	}
	if buf.Len() == 0 {
		t.Fatal("WriteQRCode empty")
	}

	// BarcodeBytes and QRCodeBase64Data
	bmpBytes, err := BarcodeBytes("test", BarcodeFormatQRCode, BarcodeOutputFormatPNG)
	if err != nil {
		t.Fatalf("BarcodeBytes: %v", err)
	}
	if len(bmpBytes) == 0 {
		t.Fatal("BarcodeBytes empty")
	}
	qrData, err := QRCodeBase64Data("base64 data")
	if err != nil {
		t.Fatalf("QRCodeBase64Data: %v", err)
	}
	if !strings.HasPrefix(qrData, "data:image/png;base64,") {
		t.Fatalf("QRCodeBase64Data prefix = %q", qrData[:30])
	}

	// BarcodeSVG
	svg, err := BarcodeSVG("svg content", BarcodeFormatQRCode)
	if err != nil {
		t.Fatalf("BarcodeSVG: %v", err)
	}
	if !strings.Contains(svg, "<svg") {
		t.Fatal("BarcodeSVG missing <svg>")
	}

	// BarcodeASCII and BarcodeASCIIWithChars
	ascii, err := BarcodeASCII("ascii content", BarcodeFormatQRCode)
	if err != nil {
		t.Fatalf("BarcodeASCII: %v", err)
	}
	if len(ascii) == 0 {
		t.Fatal("BarcodeASCII empty")
	}
	customASCII, err := BarcodeASCIIWithChars("custom", BarcodeFormatQRCode, "##", "..")
	if err != nil {
		t.Fatalf("BarcodeASCIIWithChars: %v", err)
	}
	if !strings.Contains(customASCII, "##") || !strings.Contains(customASCII, "..") {
		t.Fatalf("BarcodeASCIIWithChars missing custom chars")
	}

	// DecodeBarcodeImage
	decoded, err := DecodeBarcodeImage(img, WithDecodeFormats(BarcodeFormatEAN13))
	if err != nil {
		t.Fatalf("DecodeBarcodeImage: %v", err)
	}
	if decoded.Format != BarcodeFormatEAN13 {
		t.Fatalf("DecodeBarcodeImage format = %v, want EAN_13", decoded.Format)
	}
	if len(decoded.Text) == 0 {
		t.Fatal("DecodeBarcodeImage text is empty")
	}
	qDecoded, err := DecodeQRCodeImage(qImg)
	if err != nil {
		t.Fatalf("DecodeQRCodeImage: %v", err)
	}
	if qDecoded.Format != BarcodeFormatQRCode {
		t.Fatalf("DecodeQRCodeImage format = %v", qDecoded.Format)
	}

	// DecodeBarcode
	decodedFromReader, err := DecodeBarcode(bytes.NewReader(imgToPNG(img)), WithDecodeFormats(BarcodeFormatEAN13))
	if err != nil {
		t.Fatalf("DecodeBarcode: %v", err)
	}
	if decodedFromReader.Format != BarcodeFormatEAN13 {
		t.Fatalf("DecodeBarcode format = %v", decodedFromReader.Format)
	}
}

func imgToPNG(img image.Image) []byte {
	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	return buf.Bytes()
}
