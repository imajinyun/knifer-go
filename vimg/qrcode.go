package vimg

import (
	"image"
	"image/color"
	"io"

	"github.com/imajinyun/knifer-go/internal/imgx"
)

// BarcodeFormat identifies a QR code or barcode format.
type BarcodeFormat = imgx.BarcodeFormat

const (
	BarcodeFormatUnknown     = imgx.BarcodeFormatUnknown
	BarcodeFormatAztec       = imgx.BarcodeFormatAztec
	BarcodeFormatCodabar     = imgx.BarcodeFormatCodabar
	BarcodeFormatCode39      = imgx.BarcodeFormatCode39
	BarcodeFormatCode93      = imgx.BarcodeFormatCode93
	BarcodeFormatCode128     = imgx.BarcodeFormatCode128
	BarcodeFormatDataMatrix  = imgx.BarcodeFormatDataMatrix
	BarcodeFormatEAN8        = imgx.BarcodeFormatEAN8
	BarcodeFormatEAN13       = imgx.BarcodeFormatEAN13
	BarcodeFormatITF         = imgx.BarcodeFormatITF
	BarcodeFormatMaxiCode    = imgx.BarcodeFormatMaxiCode
	BarcodeFormatPDF417      = imgx.BarcodeFormatPDF417
	BarcodeFormatQRCode      = imgx.BarcodeFormatQRCode
	BarcodeFormatRSS14       = imgx.BarcodeFormatRSS14
	BarcodeFormatRSSExpanded = imgx.BarcodeFormatRSSExpanded
	BarcodeFormatUPCA        = imgx.BarcodeFormatUPCA
	BarcodeFormatUPCE        = imgx.BarcodeFormatUPCE
)

// BarcodeOutputFormat identifies a barcode rendering output format.
type BarcodeOutputFormat = imgx.BarcodeOutputFormat

const (
	BarcodeOutputFormatUnknown    = imgx.BarcodeOutputFormatUnknown
	BarcodeOutputFormatPNG        = imgx.BarcodeOutputFormatPNG
	BarcodeOutputFormatSVG        = imgx.BarcodeOutputFormatSVG
	BarcodeOutputFormatASCII      = imgx.BarcodeOutputFormatASCII
	BarcodeOutputFormatBase64Data = imgx.BarcodeOutputFormatBase64Data
)

// QRErrorCorrectionLevel identifies the QR code error correction level.
type QRErrorCorrectionLevel = imgx.QRErrorCorrectionLevel

const (
	QRErrorCorrectionUnknown  = imgx.QRErrorCorrectionUnknown
	QRErrorCorrectionLow      = imgx.QRErrorCorrectionLow
	QRErrorCorrectionMedium   = imgx.QRErrorCorrectionMedium
	QRErrorCorrectionQuartile = imgx.QRErrorCorrectionQuartile
	QRErrorCorrectionHigh     = imgx.QRErrorCorrectionHigh
)

// BarcodeOption customizes barcode generation.
type BarcodeOption = imgx.BarcodeOption

// QRCodeOption customizes QR code generation.
type QRCodeOption = imgx.QRCodeOption

// DecodeOption customizes barcode decoding.
type DecodeOption = imgx.DecodeOption

// DecodeResult contains decoded barcode payload and metadata.
type DecodeResult = imgx.DecodeResult

// WithBarcodeSize sets the generated barcode image size in pixels.
func WithBarcodeSize(width, height int) BarcodeOption { return imgx.WithBarcodeSize(width, height) }

// WithQRCodeSize sets a square QR code image size in pixels.
func WithQRCodeSize(size int) QRCodeOption { return imgx.WithQRCodeSize(size) }

// WithBarcodeMargin sets the barcode quiet-zone margin.
func WithBarcodeMargin(margin int) BarcodeOption { return imgx.WithBarcodeMargin(margin) }

// WithQRCodeMargin sets the QR code quiet-zone margin.
func WithQRCodeMargin(margin int) QRCodeOption { return imgx.WithQRCodeMargin(margin) }

// WithBarcodeForeground sets the color used for dark modules or bars.
func WithBarcodeForeground(foreground color.Color) BarcodeOption {
	return imgx.WithBarcodeForeground(foreground)
}

// WithQRCodeForeground sets the color used for dark QR modules.
func WithQRCodeForeground(foreground color.Color) QRCodeOption {
	return imgx.WithQRCodeForeground(foreground)
}

// WithBarcodeBackground sets the color used for light modules or spaces.
func WithBarcodeBackground(background color.Color) BarcodeOption {
	return imgx.WithBarcodeBackground(background)
}

// WithQRCodeBackground sets the color used for light QR modules.
func WithQRCodeBackground(background color.Color) QRCodeOption {
	return imgx.WithQRCodeBackground(background)
}

// WithBarcodeTransparentBackground sets a transparent background for raster and SVG output.
func WithBarcodeTransparentBackground() BarcodeOption { return imgx.WithBarcodeTransparentBackground() }

// WithQRCodeTransparentBackground sets a transparent background for QR raster and SVG output.
func WithQRCodeTransparentBackground() QRCodeOption { return imgx.WithQRCodeTransparentBackground() }

// WithBarcodeColors sets both foreground and background colors.
func WithBarcodeColors(foreground, background color.Color) BarcodeOption {
	return imgx.WithBarcodeColors(foreground, background)
}

// WithQRCodeColors sets both QR foreground and background colors.
func WithQRCodeColors(foreground, background color.Color) QRCodeOption {
	return imgx.WithQRCodeColors(foreground, background)
}

// WithBarcodeCharacterSet sets the encoder character set hint.
func WithBarcodeCharacterSet(characterSet string) BarcodeOption {
	return imgx.WithBarcodeCharacterSet(characterSet)
}

// WithQRCodeErrorCorrection sets the QR error correction level.
func WithQRCodeErrorCorrection(level QRErrorCorrectionLevel) QRCodeOption {
	return imgx.WithQRCodeErrorCorrection(level)
}

// WithQRCodeVersion pins the QR code version from 1 to 40.
func WithQRCodeVersion(version int) QRCodeOption { return imgx.WithQRCodeVersion(version) }

// WithQRCodeMaskPattern pins the QR mask pattern from 0 to 7.
func WithQRCodeMaskPattern(mask int) QRCodeOption { return imgx.WithQRCodeMaskPattern(mask) }

// WithBarcodeForceCodeSet forces Code 128 code set A, B or C.
func WithBarcodeForceCodeSet(codeSet string) BarcodeOption {
	return imgx.WithBarcodeForceCodeSet(codeSet)
}

// WithBarcodeGS1Format enables the GS1 format hint.
func WithBarcodeGS1Format(enabled bool) BarcodeOption { return imgx.WithBarcodeGS1Format(enabled) }

// WithBarcodeLogo embeds logo at the center of generated PNG/raster output.
func WithBarcodeLogo(logo image.Image) BarcodeOption { return imgx.WithBarcodeLogo(logo) }

// WithQRCodeLogo embeds logo at the center of generated QR PNG/raster output.
func WithQRCodeLogo(logo image.Image) QRCodeOption { return imgx.WithQRCodeLogo(logo) }

// WithBarcodeLogoSize sets the embedded logo size in pixels.
func WithBarcodeLogoSize(width, height int) BarcodeOption {
	return imgx.WithBarcodeLogoSize(width, height)
}

// WithQRCodeLogoSize sets the embedded QR logo size in pixels.
func WithQRCodeLogoSize(width, height int) QRCodeOption {
	return imgx.WithQRCodeLogoSize(width, height)
}

// WithQRCodeLogoRatio sets the default QR logo long-edge ratio when explicit logo size is not set.
func WithQRCodeLogoRatio(ratio int) QRCodeOption { return imgx.WithQRCodeLogoRatio(ratio) }

// WithDecodeFormats restricts decoding to the provided formats.
func WithDecodeFormats(formats ...BarcodeFormat) DecodeOption {
	return imgx.WithDecodeFormats(formats...)
}

// WithDecodeTryHarder spends more time looking for a barcode.
func WithDecodeTryHarder(enabled bool) DecodeOption { return imgx.WithDecodeTryHarder(enabled) }

// WithDecodePureBarcode hints that the image is a pure monochrome barcode.
func WithDecodePureBarcode(enabled bool) DecodeOption { return imgx.WithDecodePureBarcode(enabled) }

// WithDecodeAlsoInverted tries decoding an inverted image as a fallback.
func WithDecodeAlsoInverted(enabled bool) DecodeOption { return imgx.WithDecodeAlsoInverted(enabled) }

// WithDecodeCharacterSet sets the decoder character set hint.
func WithDecodeCharacterSet(characterSet string) DecodeOption {
	return imgx.WithDecodeCharacterSet(characterSet)
}

// BarcodeImage returns a raster image for content encoded with format.
func BarcodeImage(content string, format BarcodeFormat, opts ...BarcodeOption) (image.Image, error) {
	return imgx.BarcodeImage(content, format, opts...)
}

// CanEncodeBarcodeFormat reports whether format is supported for generation.
func CanEncodeBarcodeFormat(format BarcodeFormat) bool { return imgx.CanEncodeBarcodeFormat(format) }

// CanDecodeBarcodeFormat reports whether format is supported for decoding.
func CanDecodeBarcodeFormat(format BarcodeFormat) bool { return imgx.CanDecodeBarcodeFormat(format) }

// SupportedEncodeBarcodeFormats returns the barcode formats supported for generation.
func SupportedEncodeBarcodeFormats() []BarcodeFormat { return imgx.SupportedEncodeBarcodeFormats() }

// SupportedDecodeBarcodeFormats returns the barcode formats supported for decoding.
func SupportedDecodeBarcodeFormats() []BarcodeFormat { return imgx.SupportedDecodeBarcodeFormats() }

// QRCodeImage returns a raster QR image.
func QRCodeImage(content string, opts ...QRCodeOption) (image.Image, error) {
	return imgx.QRCodeImage(content, opts...)
}

// WriteBarcode writes a PNG-encoded barcode to w.
func WriteBarcode(w io.Writer, content string, format BarcodeFormat, opts ...BarcodeOption) error {
	return imgx.WriteBarcode(w, content, format, opts...)
}

// WriteQRCode writes a PNG-encoded QR code to w.
func WriteQRCode(w io.Writer, content string, opts ...QRCodeOption) error {
	return imgx.WriteQRCode(w, content, opts...)
}

// BarcodePNG returns PNG bytes for content encoded with format.
func BarcodePNG(content string, format BarcodeFormat, opts ...BarcodeOption) ([]byte, error) {
	return imgx.BarcodePNG(content, format, opts...)
}

// QRCodePNG returns PNG bytes for a QR code.
func QRCodePNG(content string, opts ...QRCodeOption) ([]byte, error) {
	return imgx.QRCodePNG(content, opts...)
}

// BarcodeBase64Data returns a PNG data URI for content encoded with format.
func BarcodeBase64Data(content string, format BarcodeFormat, opts ...BarcodeOption) (string, error) {
	return imgx.BarcodeBase64Data(content, format, opts...)
}

// QRCodeBase64Data returns a PNG data URI for a QR code.
func QRCodeBase64Data(content string, opts ...QRCodeOption) (string, error) {
	return imgx.QRCodeBase64Data(content, opts...)
}

// BarcodeBytes renders content encoded with format to the requested output bytes.
func BarcodeBytes(content string, format BarcodeFormat, output BarcodeOutputFormat, opts ...BarcodeOption) ([]byte, error) {
	return imgx.BarcodeBytes(content, format, output, opts...)
}

// QRCodeBytes renders QR content to the requested output bytes.
func QRCodeBytes(content string, output BarcodeOutputFormat, opts ...QRCodeOption) ([]byte, error) {
	return imgx.QRCodeBytes(content, output, opts...)
}

// BarcodeSVG returns an SVG rendering for content encoded with format.
func BarcodeSVG(content string, format BarcodeFormat, opts ...BarcodeOption) (string, error) {
	return imgx.BarcodeSVG(content, format, opts...)
}

// QRCodeSVG returns an SVG rendering for a QR code.
func QRCodeSVG(content string, opts ...QRCodeOption) (string, error) {
	return imgx.QRCodeSVG(content, opts...)
}

// BarcodeASCII returns an ASCII rendering for content encoded with format.
func BarcodeASCII(content string, format BarcodeFormat, opts ...BarcodeOption) (string, error) {
	return imgx.BarcodeASCII(content, format, opts...)
}

// BarcodeASCIIWithChars returns a text rendering using custom set and unset strings.
func BarcodeASCIIWithChars(
	content string,
	format BarcodeFormat,
	setString string,
	unsetString string,
	opts ...BarcodeOption,
) (string, error) {
	return imgx.BarcodeASCIIWithChars(content, format, setString, unsetString, opts...)
}

// QRCodeASCII returns an ASCII rendering for a QR code.
func QRCodeASCII(content string, opts ...QRCodeOption) (string, error) {
	return imgx.QRCodeASCII(content, opts...)
}

// QRCodeASCIIWithChars returns a QR text rendering using custom set and unset strings.
func QRCodeASCIIWithChars(content string, setString string, unsetString string, opts ...QRCodeOption) (string, error) {
	return imgx.QRCodeASCIIWithChars(content, setString, unsetString, opts...)
}

// DecodeBarcode decodes one barcode from a raster image stream.
func DecodeBarcode(r io.Reader, opts ...DecodeOption) (*DecodeResult, error) {
	return imgx.DecodeBarcode(r, opts...)
}

// DecodeQRCode decodes one QR code from a raster image stream.
func DecodeQRCode(r io.Reader, opts ...DecodeOption) (*DecodeResult, error) {
	return imgx.DecodeQRCode(r, opts...)
}

// DecodeBarcodeImage decodes one barcode from img.
func DecodeBarcodeImage(img image.Image, opts ...DecodeOption) (*DecodeResult, error) {
	return imgx.DecodeBarcodeImage(img, opts...)
}

// DecodeQRCodeImage decodes one QR code from img.
func DecodeQRCodeImage(img image.Image, opts ...DecodeOption) (*DecodeResult, error) {
	return imgx.DecodeQRCodeImage(img, opts...)
}
