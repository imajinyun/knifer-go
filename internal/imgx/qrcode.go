package imgx

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"html"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"slices"
	"strings"

	knifer "github.com/imajinyun/knifer-go"
	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/aztec"
	"github.com/makiuchi-d/gozxing/datamatrix"
	"github.com/makiuchi-d/gozxing/oned"
	"github.com/makiuchi-d/gozxing/oned/rss"
	"github.com/makiuchi-d/gozxing/qrcode"
	"github.com/makiuchi-d/gozxing/qrcode/decoder"
)

const (
	defaultBarcodeSize     = 256
	defaultBarcodeHeight   = 80
	defaultBarcodeMargin   = 4
	defaultQRCodeLogoRatio = 6
)

// BarcodeFormat identifies a QR code or barcode format.
type BarcodeFormat int

const (
	// BarcodeFormatUnknown is an invalid barcode format.
	BarcodeFormatUnknown BarcodeFormat = iota
	// BarcodeFormatAztec is the Aztec 2D barcode format.
	BarcodeFormatAztec
	// BarcodeFormatCodabar is the CODABAR 1D barcode format.
	BarcodeFormatCodabar
	// BarcodeFormatCode39 is the Code 39 1D barcode format.
	BarcodeFormatCode39
	// BarcodeFormatCode93 is the Code 93 1D barcode format.
	BarcodeFormatCode93
	// BarcodeFormatCode128 is the Code 128 1D barcode format.
	BarcodeFormatCode128
	// BarcodeFormatDataMatrix is the Data Matrix 2D barcode format.
	BarcodeFormatDataMatrix
	// BarcodeFormatEAN8 is the EAN-8 1D barcode format.
	BarcodeFormatEAN8
	// BarcodeFormatEAN13 is the EAN-13 1D barcode format.
	BarcodeFormatEAN13
	// BarcodeFormatITF is the Interleaved Two of Five 1D barcode format.
	BarcodeFormatITF
	// BarcodeFormatMaxiCode is the MaxiCode 2D barcode format.
	BarcodeFormatMaxiCode
	// BarcodeFormatPDF417 is the PDF417 2D barcode format.
	BarcodeFormatPDF417
	// BarcodeFormatQRCode is the QR Code 2D barcode format.
	BarcodeFormatQRCode
	// BarcodeFormatRSS14 is the RSS 14 1D barcode format.
	BarcodeFormatRSS14
	// BarcodeFormatRSSExpanded is the RSS Expanded 1D barcode format.
	BarcodeFormatRSSExpanded
	// BarcodeFormatUPCA is the UPC-A 1D barcode format.
	BarcodeFormatUPCA
	// BarcodeFormatUPCE is the UPC-E 1D barcode format.
	BarcodeFormatUPCE
)

// String returns the ZXing-compatible barcode format name.
func (f BarcodeFormat) String() string {
	switch f {
	case BarcodeFormatAztec:
		return "AZTEC"
	case BarcodeFormatCodabar:
		return "CODABAR"
	case BarcodeFormatCode39:
		return "CODE_39"
	case BarcodeFormatCode93:
		return "CODE_93"
	case BarcodeFormatCode128:
		return "CODE_128"
	case BarcodeFormatDataMatrix:
		return "DATA_MATRIX"
	case BarcodeFormatEAN8:
		return "EAN_8"
	case BarcodeFormatEAN13:
		return "EAN_13"
	case BarcodeFormatITF:
		return "ITF"
	case BarcodeFormatMaxiCode:
		return "MAXICODE"
	case BarcodeFormatPDF417:
		return "PDF_417"
	case BarcodeFormatQRCode:
		return "QR_CODE"
	case BarcodeFormatRSS14:
		return "RSS_14"
	case BarcodeFormatRSSExpanded:
		return "RSS_EXPANDED"
	case BarcodeFormatUPCA:
		return "UPC_A"
	case BarcodeFormatUPCE:
		return "UPC_E"
	default:
		return "UNKNOWN"
	}
}

// BarcodeOutputFormat identifies a barcode rendering output format.
type BarcodeOutputFormat int

const (
	// BarcodeOutputFormatUnknown is an invalid output format.
	BarcodeOutputFormatUnknown BarcodeOutputFormat = iota
	// BarcodeOutputFormatPNG renders PNG bytes.
	BarcodeOutputFormatPNG
	// BarcodeOutputFormatSVG renders SVG bytes.
	BarcodeOutputFormatSVG
	// BarcodeOutputFormatASCII renders text bytes.
	BarcodeOutputFormatASCII
	// BarcodeOutputFormatBase64Data renders a PNG data URI as bytes.
	BarcodeOutputFormatBase64Data
)

// QRErrorCorrectionLevel identifies the QR code error correction level.
type QRErrorCorrectionLevel int

const (
	// QRErrorCorrectionUnknown is an invalid error correction level.
	QRErrorCorrectionUnknown QRErrorCorrectionLevel = iota
	// QRErrorCorrectionLow recovers about 7% of codewords.
	QRErrorCorrectionLow
	// QRErrorCorrectionMedium recovers about 15% of codewords.
	QRErrorCorrectionMedium
	// QRErrorCorrectionQuartile recovers about 25% of codewords.
	QRErrorCorrectionQuartile
	// QRErrorCorrectionHigh recovers about 30% of codewords.
	QRErrorCorrectionHigh
)

// BarcodeOption customizes barcode generation.
type BarcodeOption func(*barcodeConfig) error

// QRCodeOption customizes QR code generation.
type QRCodeOption = BarcodeOption

type barcodeConfig struct {
	width        int
	height       int
	margin       int
	foreground   color.Color
	background   color.Color
	characterSet string
	qrECLevel    QRErrorCorrectionLevel
	qrVersion    int
	qrMask       int
	forceCodeSet string
	gs1Format    bool
	hints        map[gozxing.EncodeHintType]any
	logo         image.Image
	logoWidth    int
	logoHeight   int
	logoRatio    int
}

// DecodeOption customizes barcode decoding.
type DecodeOption func(*decodeConfig) error

type decodeConfig struct {
	formats []BarcodeFormat
	hints   map[gozxing.DecodeHintType]any
}

// DecodeResult contains decoded barcode payload and metadata.
type DecodeResult struct {
	Text     string
	Format   BarcodeFormat
	RawBytes []byte
	Metadata map[string]any
}

func defaultBarcodeConfig(format BarcodeFormat) barcodeConfig {
	width := defaultBarcodeSize
	height := defaultBarcodeSize
	if isOneDimensionalFormat(format) {
		width = defaultBarcodeSize
		height = defaultBarcodeHeight
	}
	return barcodeConfig{
		width:      width,
		height:     height,
		margin:     defaultBarcodeMargin,
		foreground: color.Black,
		background: color.White,
		qrECLevel:  QRErrorCorrectionLow,
		qrVersion:  -1,
		qrMask:     -1,
		logoWidth:  -1,
		logoHeight: -1,
		logoRatio:  defaultQRCodeLogoRatio,
	}
}

// WithBarcodeSize sets the generated barcode image size in pixels.
func WithBarcodeSize(width, height int) BarcodeOption {
	return func(c *barcodeConfig) error {
		if width <= 0 || height <= 0 {
			return &knifer.Error{Code: knifer.ErrCodeInvalidInput, Message: "image: barcode size must be positive"}
		}
		c.width = width
		c.height = height
		return nil
	}
}

// WithQRCodeSize sets a square QR code image size in pixels.
func WithQRCodeSize(size int) QRCodeOption { return WithBarcodeSize(size, size) }

// WithBarcodeMargin sets the barcode quiet-zone margin.
func WithBarcodeMargin(margin int) BarcodeOption {
	return func(c *barcodeConfig) error {
		if margin < 0 {
			return &knifer.Error{Code: knifer.ErrCodeInvalidInput, Message: "image: barcode margin must be non-negative"}
		}
		c.margin = margin
		return nil
	}
}

// WithQRCodeMargin sets the QR code quiet-zone margin.
func WithQRCodeMargin(margin int) QRCodeOption { return WithBarcodeMargin(margin) }

// WithBarcodeForeground sets the color used for dark modules or bars.
func WithBarcodeForeground(foreground color.Color) BarcodeOption {
	return func(c *barcodeConfig) error {
		if foreground == nil {
			return &knifer.Error{Code: knifer.ErrCodeInvalidInput, Message: "image: nil barcode foreground"}
		}
		c.foreground = foreground
		return nil
	}
}

// WithQRCodeForeground sets the color used for dark QR modules.
func WithQRCodeForeground(foreground color.Color) QRCodeOption {
	return WithBarcodeForeground(foreground)
}

// WithBarcodeBackground sets the color used for light modules or spaces.
func WithBarcodeBackground(background color.Color) BarcodeOption {
	return func(c *barcodeConfig) error {
		if background == nil {
			return &knifer.Error{Code: knifer.ErrCodeInvalidInput, Message: "image: nil barcode background"}
		}
		c.background = background
		return nil
	}
}

// WithQRCodeBackground sets the color used for light QR modules.
func WithQRCodeBackground(background color.Color) QRCodeOption {
	return WithBarcodeBackground(background)
}

// WithBarcodeTransparentBackground sets a transparent background for raster and SVG output.
func WithBarcodeTransparentBackground() BarcodeOption {
	return WithBarcodeBackground(color.Transparent)
}

// WithQRCodeTransparentBackground sets a transparent background for QR raster and SVG output.
func WithQRCodeTransparentBackground() QRCodeOption {
	return WithBarcodeTransparentBackground()
}

// WithBarcodeColors sets both foreground and background colors.
func WithBarcodeColors(foreground, background color.Color) BarcodeOption {
	return func(c *barcodeConfig) error {
		if err := WithBarcodeForeground(foreground)(c); err != nil {
			return err
		}
		return WithBarcodeBackground(background)(c)
	}
}

// WithQRCodeColors sets both QR foreground and background colors.
func WithQRCodeColors(foreground, background color.Color) QRCodeOption {
	return WithBarcodeColors(foreground, background)
}

// WithBarcodeCharacterSet sets the encoder character set hint.
func WithBarcodeCharacterSet(characterSet string) BarcodeOption {
	return func(c *barcodeConfig) error {
		characterSet = strings.TrimSpace(characterSet)
		if characterSet == "" {
			return &knifer.Error{Code: knifer.ErrCodeInvalidInput, Message: "image: barcode character set must not be empty"}
		}
		c.characterSet = characterSet
		return nil
	}
}

// WithQRCodeErrorCorrection sets the QR error correction level.
func WithQRCodeErrorCorrection(level QRErrorCorrectionLevel) QRCodeOption {
	return func(c *barcodeConfig) error {
		if _, err := toGozxingQRECLevel(level); err != nil {
			return err
		}
		c.qrECLevel = level
		return nil
	}
}

// WithQRCodeVersion pins the QR code version from 1 to 40.
func WithQRCodeVersion(version int) QRCodeOption {
	return func(c *barcodeConfig) error {
		if version < 1 || version > 40 {
			return &knifer.Error{Code: knifer.ErrCodeInvalidInput, Message: "image: qr version must be between 1 and 40"}
		}
		c.qrVersion = version
		return nil
	}
}

// WithQRCodeMaskPattern pins the QR mask pattern from 0 to 7.
func WithQRCodeMaskPattern(mask int) QRCodeOption {
	return func(c *barcodeConfig) error {
		if mask < 0 || mask > 7 {
			return &knifer.Error{Code: knifer.ErrCodeInvalidInput, Message: "image: qr mask pattern must be between 0 and 7"}
		}
		c.qrMask = mask
		return nil
	}
}

// WithBarcodeForceCodeSet forces Code 128 code set A, B or C.
func WithBarcodeForceCodeSet(codeSet string) BarcodeOption {
	return func(c *barcodeConfig) error {
		codeSet = strings.ToUpper(strings.TrimSpace(codeSet))
		if codeSet != "A" && codeSet != "B" && codeSet != "C" {
			return &knifer.Error{Code: knifer.ErrCodeInvalidInput, Message: "image: barcode code set must be A, B or C"}
		}
		c.forceCodeSet = codeSet
		return nil
	}
}

// WithBarcodeGS1Format enables the GS1 format hint.
func WithBarcodeGS1Format(enabled bool) BarcodeOption {
	return func(c *barcodeConfig) error {
		c.gs1Format = enabled
		return nil
	}
}

// WithBarcodeLogo embeds logo at the center of generated PNG/raster output.
func WithBarcodeLogo(logo image.Image) BarcodeOption {
	return func(c *barcodeConfig) error {
		if logo == nil {
			return &knifer.Error{Code: knifer.ErrCodeInvalidInput, Message: "image: nil barcode logo"}
		}
		if logo.Bounds().Dx() <= 0 || logo.Bounds().Dy() <= 0 {
			return &knifer.Error{Code: knifer.ErrCodeInvalidInput, Message: "image: empty barcode logo"}
		}
		c.logo = logo
		return nil
	}
}

// WithQRCodeLogo embeds logo at the center of generated QR PNG/raster output.
func WithQRCodeLogo(logo image.Image) QRCodeOption { return WithBarcodeLogo(logo) }

// WithBarcodeLogoSize sets the embedded logo size in pixels.
func WithBarcodeLogoSize(width, height int) BarcodeOption {
	return func(c *barcodeConfig) error {
		if width <= 0 || height <= 0 {
			return &knifer.Error{Code: knifer.ErrCodeInvalidInput, Message: "image: barcode logo size must be positive"}
		}
		c.logoWidth = width
		c.logoHeight = height
		return nil
	}
}

// WithQRCodeLogoSize sets the embedded QR logo size in pixels.
func WithQRCodeLogoSize(width, height int) QRCodeOption { return WithBarcodeLogoSize(width, height) }

// WithQRCodeLogoRatio sets the default QR logo long-edge ratio when explicit logo size is not set.
func WithQRCodeLogoRatio(ratio int) QRCodeOption {
	return func(c *barcodeConfig) error {
		if ratio <= 0 {
			return &knifer.Error{Code: knifer.ErrCodeInvalidInput, Message: "image: qr logo ratio must be positive"}
		}
		c.logoRatio = ratio
		return nil
	}
}

// WithBarcodeEncodeHint sets a raw gozxing encode hint.
func WithBarcodeEncodeHint(hint gozxing.EncodeHintType, value any) BarcodeOption {
	return func(c *barcodeConfig) error {
		if c.hints == nil {
			c.hints = make(map[gozxing.EncodeHintType]any)
		}
		c.hints[hint] = value
		return nil
	}
}

// WithDecodeFormats restricts decoding to the provided formats.
func WithDecodeFormats(formats ...BarcodeFormat) DecodeOption {
	return func(c *decodeConfig) error {
		if len(formats) == 0 {
			return &knifer.Error{Code: knifer.ErrCodeInvalidInput, Message: "image: decode formats must not be empty"}
		}
		for _, format := range formats {
			if !CanDecodeBarcodeFormat(format) {
				return &knifer.Error{Code: knifer.ErrCodeUnsupported, Message: fmt.Sprintf("image: barcode reader for %s is unsupported", format)}
			}
		}
		c.formats = slices.Clone(formats)
		return nil
	}
}

// WithDecodeTryHarder spends more time looking for a barcode.
func WithDecodeTryHarder(enabled bool) DecodeOption {
	return withDecodeBoolHint(gozxing.DecodeHintType_TRY_HARDER, enabled)
}

// WithDecodePureBarcode hints that the image is a pure monochrome barcode.
func WithDecodePureBarcode(enabled bool) DecodeOption {
	return withDecodeBoolHint(gozxing.DecodeHintType_PURE_BARCODE, enabled)
}

// WithDecodeAlsoInverted tries decoding an inverted image as a fallback.
func WithDecodeAlsoInverted(enabled bool) DecodeOption {
	return withDecodeBoolHint(gozxing.DecodeHintType_ALSO_INVERTED, enabled)
}

// WithDecodeCharacterSet sets the decoder character set hint.
func WithDecodeCharacterSet(characterSet string) DecodeOption {
	return func(c *decodeConfig) error {
		characterSet = strings.TrimSpace(characterSet)
		if characterSet == "" {
			return &knifer.Error{Code: knifer.ErrCodeInvalidInput, Message: "image: decode character set must not be empty"}
		}
		setDecodeHint(c, gozxing.DecodeHintType_CHARACTER_SET, characterSet)
		return nil
	}
}

// WithDecodeHint sets a raw gozxing decode hint.
func WithDecodeHint(hint gozxing.DecodeHintType, value any) DecodeOption {
	return func(c *decodeConfig) error {
		setDecodeHint(c, hint, value)
		return nil
	}
}

// BarcodeImage returns a raster image for content encoded with format.
func BarcodeImage(content string, format BarcodeFormat, opts ...BarcodeOption) (image.Image, error) {
	matrix, cfg, err := encodeBarcodeMatrix(content, format, opts...)
	if err != nil {
		return nil, err
	}
	return renderBarcodeImage(matrix, cfg), nil
}

// QRCodeImage returns a raster QR image.
func QRCodeImage(content string, opts ...QRCodeOption) (image.Image, error) {
	return BarcodeImage(content, BarcodeFormatQRCode, opts...)
}

// WriteBarcode writes a PNG-encoded barcode to w.
func WriteBarcode(w io.Writer, content string, format BarcodeFormat, opts ...BarcodeOption) error {
	if w == nil {
		return &knifer.Error{Code: knifer.ErrCodeInvalidInput, Message: "image: nil writer"}
	}
	img, err := BarcodeImage(content, format, opts...)
	if err != nil {
		return err
	}
	if err := png.Encode(w, img); err != nil {
		return &knifer.Error{Code: knifer.ErrCodeInternal, Message: "image: barcode png encode failed", Cause: err}
	}
	return nil
}

// WriteQRCode writes a PNG-encoded QR code to w.
func WriteQRCode(w io.Writer, content string, opts ...QRCodeOption) error {
	return WriteBarcode(w, content, BarcodeFormatQRCode, opts...)
}

// BarcodePNG returns PNG bytes for content encoded with format.
func BarcodePNG(content string, format BarcodeFormat, opts ...BarcodeOption) ([]byte, error) {
	buf := &bytes.Buffer{}
	if err := WriteBarcode(buf, content, format, opts...); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// QRCodePNG returns PNG bytes for a QR code.
func QRCodePNG(content string, opts ...QRCodeOption) ([]byte, error) {
	return BarcodePNG(content, BarcodeFormatQRCode, opts...)
}

// BarcodeBase64Data returns a PNG data URI for content encoded with format.
func BarcodeBase64Data(content string, format BarcodeFormat, opts ...BarcodeOption) (string, error) {
	pngBytes, err := BarcodePNG(content, format, opts...)
	if err != nil {
		return "", err
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(pngBytes), nil
}

// BarcodeBytes renders content encoded with format to the requested output bytes.
func BarcodeBytes(content string, format BarcodeFormat, output BarcodeOutputFormat, opts ...BarcodeOption) ([]byte, error) {
	switch output {
	case BarcodeOutputFormatPNG:
		return BarcodePNG(content, format, opts...)
	case BarcodeOutputFormatSVG:
		svg, err := BarcodeSVG(content, format, opts...)
		if err != nil {
			return nil, err
		}
		return []byte(svg), nil
	case BarcodeOutputFormatASCII:
		ascii, err := BarcodeASCII(content, format, opts...)
		if err != nil {
			return nil, err
		}
		return []byte(ascii), nil
	case BarcodeOutputFormatBase64Data:
		data, err := BarcodeBase64Data(content, format, opts...)
		if err != nil {
			return nil, err
		}
		return []byte(data), nil
	default:
		return nil, &knifer.Error{Code: knifer.ErrCodeInvalidInput, Message: "image: invalid barcode output format"}
	}
}

// QRCodeBytes renders QR content to the requested output bytes.
func QRCodeBytes(content string, output BarcodeOutputFormat, opts ...QRCodeOption) ([]byte, error) {
	return BarcodeBytes(content, BarcodeFormatQRCode, output, opts...)
}

// QRCodeBase64Data returns a PNG data URI for a QR code.
func QRCodeBase64Data(content string, opts ...QRCodeOption) (string, error) {
	return BarcodeBase64Data(content, BarcodeFormatQRCode, opts...)
}

// BarcodeSVG returns an SVG rendering for content encoded with format.
func BarcodeSVG(content string, format BarcodeFormat, opts ...BarcodeOption) (string, error) {
	matrix, cfg, err := encodeBarcodeMatrix(content, format, opts...)
	if err != nil {
		return "", err
	}
	svg, err := renderBarcodeSVG(matrix, cfg)
	if err != nil {
		return "", err
	}
	return svg, nil
}

// QRCodeSVG returns an SVG rendering for a QR code.
func QRCodeSVG(content string, opts ...QRCodeOption) (string, error) {
	return BarcodeSVG(content, BarcodeFormatQRCode, opts...)
}

// BarcodeASCII returns an ASCII rendering for content encoded with format.
func BarcodeASCII(content string, format BarcodeFormat, opts ...BarcodeOption) (string, error) {
	return BarcodeASCIIWithChars(content, format, "██", "  ", opts...)
}

// BarcodeASCIIWithChars returns a text rendering using custom set and unset strings.
func BarcodeASCIIWithChars(content string, format BarcodeFormat, setString, unsetString string, opts ...BarcodeOption) (string, error) {
	if setString == "" || unsetString == "" {
		return "", &knifer.Error{Code: knifer.ErrCodeInvalidInput, Message: "image: barcode ascii characters must not be empty"}
	}
	matrix, _, err := encodeBarcodeMatrix(content, format, opts...)
	if err != nil {
		return "", err
	}
	return renderBarcodeASCII(matrix, setString, unsetString), nil
}

// QRCodeASCII returns an ASCII rendering for a QR code.
func QRCodeASCII(content string, opts ...QRCodeOption) (string, error) {
	return BarcodeASCII(content, BarcodeFormatQRCode, opts...)
}

// QRCodeASCIIWithChars returns a QR text rendering using custom set and unset strings.
func QRCodeASCIIWithChars(content string, setString, unsetString string, opts ...QRCodeOption) (string, error) {
	return BarcodeASCIIWithChars(content, BarcodeFormatQRCode, setString, unsetString, opts...)
}

// DecodeBarcode decodes one barcode from a raster image stream.
func DecodeBarcode(r io.Reader, opts ...DecodeOption) (*DecodeResult, error) {
	if r == nil {
		return nil, &knifer.Error{Code: knifer.ErrCodeInvalidInput, Message: "image: nil reader"}
	}
	cfg := decodeConfig{}
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		if err := opt(&cfg); err != nil {
			return nil, err
		}
	}
	img, err := decodeAny(r)
	if err != nil {
		return nil, err
	}
	return DecodeBarcodeImage(img, opts...)
}

// DecodeQRCode decodes one QR code from a raster image stream.
func DecodeQRCode(r io.Reader, opts ...DecodeOption) (*DecodeResult, error) {
	allOpts := append([]DecodeOption{WithDecodeFormats(BarcodeFormatQRCode)}, opts...)
	return DecodeBarcode(r, allOpts...)
}

// DecodeBarcodeImage decodes one barcode from img.
func DecodeBarcodeImage(img image.Image, opts ...DecodeOption) (*DecodeResult, error) {
	if img == nil {
		return nil, &knifer.Error{Code: knifer.ErrCodeInvalidInput, Message: "image: nil image"}
	}
	cfg := decodeConfig{}
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		if err := opt(&cfg); err != nil {
			return nil, err
		}
	}
	hints := cfg.hints
	if len(cfg.formats) > 0 {
		formats := make(gozxing.BarcodeFormats, 0, len(cfg.formats))
		for _, format := range cfg.formats {
			goFormat, err := toGozxingBarcodeFormat(format)
			if err != nil {
				return nil, err
			}
			formats = append(formats, goFormat)
		}
		if hints == nil {
			hints = make(map[gozxing.DecodeHintType]any)
		}
		hints[gozxing.DecodeHintType_POSSIBLE_FORMATS] = formats
	}
	bitmaps, err := barcodeBitmaps(img)
	if err != nil {
		return nil, &knifer.Error{Code: knifer.ErrCodeInvalidInput, Message: "image: barcode bitmap creation failed", Cause: err}
	}
	readers := barcodeReaders(cfg.formats)
	var lastErr error
	for _, bmp := range bitmaps {
		for _, reader := range readers {
			result, err := reader.Decode(bmp, hints)
			if err == nil {
				return newDecodeResult(result), nil
			}
			lastErr = err
		}
	}
	return nil, &knifer.Error{Code: knifer.ErrCodeInvalidInput, Message: "image: barcode decode failed", Cause: lastErr}
}

// DecodeQRCodeImage decodes one QR code from img.
func DecodeQRCodeImage(img image.Image, opts ...DecodeOption) (*DecodeResult, error) {
	allOpts := append([]DecodeOption{WithDecodeFormats(BarcodeFormatQRCode)}, opts...)
	return DecodeBarcodeImage(img, allOpts...)
}

func encodeBarcodeMatrix(content string, format BarcodeFormat, opts ...BarcodeOption) (*gozxing.BitMatrix, barcodeConfig, error) {
	if strings.TrimSpace(content) == "" {
		return nil, barcodeConfig{}, &knifer.Error{Code: knifer.ErrCodeInvalidInput, Message: "image: barcode content must not be empty"}
	}
	cfg := defaultBarcodeConfig(format)
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		if err := opt(&cfg); err != nil {
			return nil, barcodeConfig{}, err
		}
	}
	goFormat, err := toGozxingBarcodeFormat(format)
	if err != nil {
		return nil, barcodeConfig{}, err
	}
	if cfg.logo != nil && format != BarcodeFormatQRCode {
		return nil, barcodeConfig{}, &knifer.Error{Code: knifer.ErrCodeUnsupported, Message: "image: barcode logo is only supported for qr code"}
	}
	writer, err := barcodeWriter(format)
	if err != nil {
		return nil, barcodeConfig{}, err
	}
	hints, err := encodeHints(cfg, format)
	if err != nil {
		return nil, barcodeConfig{}, err
	}
	matrix, err := writer.Encode(content, goFormat, cfg.width, cfg.height, hints)
	if err != nil {
		return nil, barcodeConfig{}, &knifer.Error{Code: knifer.ErrCodeInvalidInput, Message: "image: barcode encode failed", Cause: err}
	}
	return matrix, cfg, nil
}

func encodeHints(cfg barcodeConfig, format BarcodeFormat) (map[gozxing.EncodeHintType]any, error) {
	hints := make(map[gozxing.EncodeHintType]any, len(cfg.hints)+6)
	for k, v := range cfg.hints {
		hints[k] = v
	}
	hints[gozxing.EncodeHintType_MARGIN] = cfg.margin
	if cfg.characterSet != "" {
		hints[gozxing.EncodeHintType_CHARACTER_SET] = cfg.characterSet
	}
	if cfg.gs1Format {
		hints[gozxing.EncodeHintType_GS1_FORMAT] = true
	}
	if cfg.forceCodeSet != "" {
		hints[gozxing.EncodeHintType_FORCE_CODE_SET] = cfg.forceCodeSet
	}
	if format == BarcodeFormatQRCode {
		level, err := toGozxingQRECLevel(cfg.qrECLevel)
		if err != nil {
			return nil, err
		}
		hints[gozxing.EncodeHintType_ERROR_CORRECTION] = level
		if cfg.qrVersion > 0 {
			hints[gozxing.EncodeHintType_QR_VERSION] = cfg.qrVersion
		}
		if cfg.qrMask >= 0 {
			hints[gozxing.EncodeHintType_QR_MASK_PATTERN] = cfg.qrMask
		}
	}
	return hints, nil
}

// CanEncodeBarcodeFormat reports whether format has a barcode writer.
func CanEncodeBarcodeFormat(format BarcodeFormat) bool {
	switch format {
	case BarcodeFormatCodabar, BarcodeFormatCode39, BarcodeFormatCode93, BarcodeFormatCode128,
		BarcodeFormatDataMatrix, BarcodeFormatEAN8, BarcodeFormatEAN13, BarcodeFormatITF,
		BarcodeFormatQRCode, BarcodeFormatUPCA, BarcodeFormatUPCE:
		return true
	default:
		return false
	}
}

// SupportedEncodeBarcodeFormats returns the barcode formats supported for generation.
func SupportedEncodeBarcodeFormats() []BarcodeFormat {
	return []BarcodeFormat{
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
}

func barcodeWriter(format BarcodeFormat) (gozxing.Writer, error) {
	switch format {
	case BarcodeFormatCodabar:
		return oned.NewCodaBarWriter(), nil
	case BarcodeFormatCode39:
		return oned.NewCode39Writer(), nil
	case BarcodeFormatCode93:
		return oned.NewCode93Writer(), nil
	case BarcodeFormatCode128:
		return oned.NewCode128Writer(), nil
	case BarcodeFormatDataMatrix:
		return datamatrix.NewDataMatrixWriter(), nil
	case BarcodeFormatEAN8:
		return oned.NewEAN8Writer(), nil
	case BarcodeFormatEAN13:
		return oned.NewEAN13Writer(), nil
	case BarcodeFormatITF:
		return oned.NewITFWriter(), nil
	case BarcodeFormatQRCode:
		return qrcode.NewQRCodeWriter(), nil
	case BarcodeFormatUPCA:
		return oned.NewUPCAWriter(), nil
	case BarcodeFormatUPCE:
		return oned.NewUPCEWriter(), nil
	case BarcodeFormatAztec, BarcodeFormatMaxiCode, BarcodeFormatPDF417, BarcodeFormatRSS14, BarcodeFormatRSSExpanded:
		return nil, &knifer.Error{Code: knifer.ErrCodeUnsupported, Message: fmt.Sprintf("image: barcode writer for %s is unsupported", format)}
	default:
		return nil, &knifer.Error{Code: knifer.ErrCodeUnsupported, Message: fmt.Sprintf("image: unsupported barcode format %s", format)}
	}
}

// CanDecodeBarcodeFormat reports whether format has a barcode reader.
func CanDecodeBarcodeFormat(format BarcodeFormat) bool {
	switch format {
	case BarcodeFormatQRCode, BarcodeFormatDataMatrix, BarcodeFormatCode128, BarcodeFormatCode39,
		BarcodeFormatCode93, BarcodeFormatEAN13, BarcodeFormatEAN8, BarcodeFormatUPCA,
		BarcodeFormatUPCE, BarcodeFormatITF, BarcodeFormatCodabar, BarcodeFormatRSS14,
		BarcodeFormatAztec:
		return true
	default:
		return false
	}
}

// SupportedDecodeBarcodeFormats returns the barcode formats supported for decoding.
func SupportedDecodeBarcodeFormats() []BarcodeFormat {
	return []BarcodeFormat{
		BarcodeFormatQRCode,
		BarcodeFormatDataMatrix,
		BarcodeFormatCode128,
		BarcodeFormatCode39,
		BarcodeFormatCode93,
		BarcodeFormatEAN13,
		BarcodeFormatEAN8,
		BarcodeFormatUPCA,
		BarcodeFormatUPCE,
		BarcodeFormatITF,
		BarcodeFormatCodabar,
		BarcodeFormatRSS14,
		BarcodeFormatAztec,
	}
}

func barcodeReaders(formats []BarcodeFormat) []gozxing.Reader {
	if len(formats) == 0 {
		formats = SupportedDecodeBarcodeFormats()
	}
	readers := make([]gozxing.Reader, 0, len(formats))
	for _, format := range formats {
		switch format {
		case BarcodeFormatQRCode:
			readers = append(readers, qrcode.NewQRCodeReader())
		case BarcodeFormatDataMatrix:
			readers = append(readers, datamatrix.NewDataMatrixReader())
		case BarcodeFormatCode128:
			readers = append(readers, oned.NewCode128Reader())
		case BarcodeFormatCode39:
			readers = append(readers, oned.NewCode39Reader())
		case BarcodeFormatCode93:
			readers = append(readers, oned.NewCode93Reader())
		case BarcodeFormatEAN13:
			readers = append(readers, oned.NewEAN13Reader())
		case BarcodeFormatEAN8:
			readers = append(readers, oned.NewEAN8Reader())
		case BarcodeFormatUPCA:
			readers = append(readers, oned.NewUPCAReader())
		case BarcodeFormatUPCE:
			readers = append(readers, oned.NewUPCEReader())
		case BarcodeFormatITF:
			readers = append(readers, oned.NewITFReader())
		case BarcodeFormatCodabar:
			readers = append(readers, oned.NewCodaBarReader())
		case BarcodeFormatRSS14:
			readers = append(readers, rss.NewRSS14Reader())
		case BarcodeFormatAztec:
			readers = append(readers, aztec.NewAztecReader())
		}
	}
	return readers
}

func renderBarcodeImage(matrix *gozxing.BitMatrix, cfg barcodeConfig) image.Image {
	bounds := matrix.Bounds()
	dst := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if matrix.Get(x, y) {
				dst.Set(x, y, cfg.foreground)
			} else {
				dst.Set(x, y, cfg.background)
			}
		}
	}
	if cfg.logo != nil {
		drawBarcodeLogo(dst, cfg)
	}
	return dst
}

func drawBarcodeLogo(dst *image.RGBA, cfg barcodeConfig) {
	logo, left, top, logoWidth, logoHeight := prepareBarcodeLogo(cfg, dst.Bounds())
	if logoWidth <= 0 || logoHeight <= 0 {
		return
	}
	padding := max(2, min(logoWidth, logoHeight)/12)
	bg := image.Rect(left-padding, top-padding, left+logoWidth+padding, top+logoHeight+padding).Intersect(dst.Bounds())
	draw.Draw(dst, bg, &image.Uniform{cfg.background}, image.Point{}, draw.Src)
	draw.Draw(dst, image.Rect(left, top, left+logoWidth, top+logoHeight), logo, image.Point{}, draw.Over)
}

func prepareBarcodeLogo(cfg barcodeConfig, bounds image.Rectangle) (image.Image, int, int, int, int) {
	logoBounds := cfg.logo.Bounds()
	logoWidth := cfg.logoWidth
	logoHeight := cfg.logoHeight
	if logoWidth <= 0 || logoHeight <= 0 {
		maxWidth := bounds.Dx() / cfg.logoRatio
		maxHeight := bounds.Dy() / cfg.logoRatio
		logoWidth, logoHeight = fitLongEdge(logoBounds.Dx(), logoBounds.Dy(), min(maxWidth, maxHeight))
	}
	if logoWidth <= 0 || logoHeight <= 0 {
		return nil, 0, 0, 0, 0
	}
	logo := resizeNearest(cfg.logo, logoWidth, logoHeight)
	left := bounds.Min.X + (bounds.Dx()-logoWidth)/2
	top := bounds.Min.Y + (bounds.Dy()-logoHeight)/2
	return logo, left, top, logoWidth, logoHeight
}

func pngDataURI(img image.Image) (string, error) {
	buf := &bytes.Buffer{}
	if err := png.Encode(buf, img); err != nil {
		return "", &knifer.Error{Code: knifer.ErrCodeInternal, Message: "image: barcode logo png encode failed", Cause: err}
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

func resizeNearest(src image.Image, width, height int) image.Image {
	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	bounds := src.Bounds()
	srcWidth := bounds.Dx()
	srcHeight := bounds.Dy()
	for y := 0; y < height; y++ {
		sy := bounds.Min.Y + y*srcHeight/height
		for x := 0; x < width; x++ {
			sx := bounds.Min.X + x*srcWidth/width
			dst.Set(x, y, src.At(sx, sy))
		}
	}
	return dst
}

func renderBarcodeSVG(matrix *gozxing.BitMatrix, cfg barcodeConfig) (string, error) {
	width := matrix.GetWidth()
	height := matrix.GetHeight()
	fg := svgColor(cfg.foreground)
	bg := svgColor(cfg.background)
	var b strings.Builder
	b.Grow(width*height/2 + 256)
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %d %d" width="%d" height="%d" shape-rendering="crispEdges">`, width, height, width, height)
	fmt.Fprintf(&b, `<rect width="100%%" height="100%%" fill="%s"/>`, html.EscapeString(bg))
	fmt.Fprintf(&b, `<path fill="%s" d="`, html.EscapeString(fg))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if matrix.Get(x, y) {
				fmt.Fprintf(&b, "M%d %dh1v1h-1z", x, y)
			}
		}
	}
	b.WriteString(`"/>`)
	if cfg.logo != nil {
		logo, left, top, logoWidth, logoHeight := prepareBarcodeLogo(cfg, image.Rect(0, 0, width, height))
		if logoWidth > 0 && logoHeight > 0 {
			data, err := pngDataURI(logo)
			if err != nil {
				return "", err
			}
			padding := max(2, min(logoWidth, logoHeight)/12)
			bgRect := image.Rect(left-padding, top-padding, left+logoWidth+padding, top+logoHeight+padding).
				Intersect(image.Rect(0, 0, width, height))
			fmt.Fprintf(
				&b,
				`<rect x="%d" y="%d" width="%d" height="%d" fill="%s"/>`,
				bgRect.Min.X,
				bgRect.Min.Y,
				bgRect.Dx(),
				bgRect.Dy(),
				html.EscapeString(bg),
			)
			fmt.Fprintf(
				&b,
				`<image x="%d" y="%d" width="%d" height="%d" href="%s"/>`,
				left,
				top,
				logoWidth,
				logoHeight,
				html.EscapeString(data),
			)
		}
	}
	b.WriteString(`</svg>`)
	return b.String(), nil
}

func svgColor(c color.Color) string {
	r, g, b, a := c.RGBA()
	if a == 0xffff {
		return fmt.Sprintf("#%02x%02x%02x", colorComponent8(r), colorComponent8(g), colorComponent8(b))
	}
	return fmt.Sprintf("rgba(%d,%d,%d,%.3f)", colorComponent8(r), colorComponent8(g), colorComponent8(b), float64(a)/65535)
}

func renderBarcodeASCII(matrix *gozxing.BitMatrix, setString, unsetString string) string {
	return matrix.ToString(setString, unsetString)
}

func barcodeBitmaps(img image.Image) ([]*gozxing.BinaryBitmap, error) {
	source := gozxing.NewLuminanceSourceFromImage(img)
	hybrid, err := gozxing.NewBinaryBitmap(gozxing.NewHybridBinarizer(source))
	if err != nil {
		return nil, err
	}
	global, err := gozxing.NewBinaryBitmap(gozxing.NewGlobalHistgramBinarizer(source))
	if err != nil {
		return nil, err
	}
	return []*gozxing.BinaryBitmap{hybrid, global}, nil
}

func withDecodeBoolHint(hint gozxing.DecodeHintType, enabled bool) DecodeOption {
	return func(c *decodeConfig) error {
		if enabled {
			setDecodeHint(c, hint, true)
		}
		return nil
	}
}

func setDecodeHint(c *decodeConfig, hint gozxing.DecodeHintType, value any) {
	if c.hints == nil {
		c.hints = make(map[gozxing.DecodeHintType]any)
	}
	c.hints[hint] = value
}

func newDecodeResult(result *gozxing.Result) *DecodeResult {
	raw := result.GetRawBytes()
	metadata := result.GetResultMetadata()
	out := &DecodeResult{
		Text:     result.GetText(),
		Format:   fromGozxingBarcodeFormat(result.GetBarcodeFormat()),
		RawBytes: slices.Clone(raw),
	}
	if len(metadata) > 0 {
		out.Metadata = make(map[string]any, len(metadata))
		for k, v := range metadata {
			out.Metadata[k.String()] = v
		}
	}
	return out
}

func toGozxingBarcodeFormat(format BarcodeFormat) (gozxing.BarcodeFormat, error) {
	switch format {
	case BarcodeFormatAztec:
		return gozxing.BarcodeFormat_AZTEC, nil
	case BarcodeFormatCodabar:
		return gozxing.BarcodeFormat_CODABAR, nil
	case BarcodeFormatCode39:
		return gozxing.BarcodeFormat_CODE_39, nil
	case BarcodeFormatCode93:
		return gozxing.BarcodeFormat_CODE_93, nil
	case BarcodeFormatCode128:
		return gozxing.BarcodeFormat_CODE_128, nil
	case BarcodeFormatDataMatrix:
		return gozxing.BarcodeFormat_DATA_MATRIX, nil
	case BarcodeFormatEAN8:
		return gozxing.BarcodeFormat_EAN_8, nil
	case BarcodeFormatEAN13:
		return gozxing.BarcodeFormat_EAN_13, nil
	case BarcodeFormatITF:
		return gozxing.BarcodeFormat_ITF, nil
	case BarcodeFormatMaxiCode:
		return gozxing.BarcodeFormat_MAXICODE, nil
	case BarcodeFormatPDF417:
		return gozxing.BarcodeFormat_PDF_417, nil
	case BarcodeFormatQRCode:
		return gozxing.BarcodeFormat_QR_CODE, nil
	case BarcodeFormatRSS14:
		return gozxing.BarcodeFormat_RSS_14, nil
	case BarcodeFormatRSSExpanded:
		return gozxing.BarcodeFormat_RSS_EXPANDED, nil
	case BarcodeFormatUPCA:
		return gozxing.BarcodeFormat_UPC_A, nil
	case BarcodeFormatUPCE:
		return gozxing.BarcodeFormat_UPC_E, nil
	default:
		return 0, &knifer.Error{Code: knifer.ErrCodeUnsupported, Message: fmt.Sprintf("image: unsupported barcode format %s", format)}
	}
}

func fromGozxingBarcodeFormat(format gozxing.BarcodeFormat) BarcodeFormat {
	switch format {
	case gozxing.BarcodeFormat_AZTEC:
		return BarcodeFormatAztec
	case gozxing.BarcodeFormat_CODABAR:
		return BarcodeFormatCodabar
	case gozxing.BarcodeFormat_CODE_39:
		return BarcodeFormatCode39
	case gozxing.BarcodeFormat_CODE_93:
		return BarcodeFormatCode93
	case gozxing.BarcodeFormat_CODE_128:
		return BarcodeFormatCode128
	case gozxing.BarcodeFormat_DATA_MATRIX:
		return BarcodeFormatDataMatrix
	case gozxing.BarcodeFormat_EAN_8:
		return BarcodeFormatEAN8
	case gozxing.BarcodeFormat_EAN_13:
		return BarcodeFormatEAN13
	case gozxing.BarcodeFormat_ITF:
		return BarcodeFormatITF
	case gozxing.BarcodeFormat_MAXICODE:
		return BarcodeFormatMaxiCode
	case gozxing.BarcodeFormat_PDF_417:
		return BarcodeFormatPDF417
	case gozxing.BarcodeFormat_QR_CODE:
		return BarcodeFormatQRCode
	case gozxing.BarcodeFormat_RSS_14:
		return BarcodeFormatRSS14
	case gozxing.BarcodeFormat_RSS_EXPANDED:
		return BarcodeFormatRSSExpanded
	case gozxing.BarcodeFormat_UPC_A:
		return BarcodeFormatUPCA
	case gozxing.BarcodeFormat_UPC_E:
		return BarcodeFormatUPCE
	default:
		return BarcodeFormatUnknown
	}
}

func toGozxingQRECLevel(level QRErrorCorrectionLevel) (decoder.ErrorCorrectionLevel, error) {
	switch level {
	case QRErrorCorrectionLow:
		return decoder.ErrorCorrectionLevel_L, nil
	case QRErrorCorrectionMedium:
		return decoder.ErrorCorrectionLevel_M, nil
	case QRErrorCorrectionQuartile:
		return decoder.ErrorCorrectionLevel_Q, nil
	case QRErrorCorrectionHigh:
		return decoder.ErrorCorrectionLevel_H, nil
	default:
		return decoder.ErrorCorrectionLevel_L, &knifer.Error{Code: knifer.ErrCodeInvalidInput, Message: "image: invalid qr error correction level"}
	}
}

func isOneDimensionalFormat(format BarcodeFormat) bool {
	switch format {
	case BarcodeFormatCodabar, BarcodeFormatCode39, BarcodeFormatCode93, BarcodeFormatCode128,
		BarcodeFormatEAN8, BarcodeFormatEAN13, BarcodeFormatITF, BarcodeFormatRSS14,
		BarcodeFormatRSSExpanded, BarcodeFormatUPCA, BarcodeFormatUPCE:
		return true
	default:
		return false
	}
}
