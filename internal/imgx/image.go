// Package imgx provides image utilities on top of the standard image packages
// plus optional ZXing-backed QR code and barcode helpers.
//
// It intentionally keeps its helper set small: image metadata inspection,
// lossless format conversion between PNG/JPEG/GIF, and a simple proportional
// downscaling helper. QR code and barcode generation/decoding is implemented
// with gozxing to align with Hutool-style QR utilities.
package imgx

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"strings"

	knifer "github.com/imajinyun/knifer-go"
)

// supportedFormats enumerates the output formats accepted by Thumbnail and
// ConvertFormat. The same set is also used by Info to identify the source
// stream's format after decoding.
var supportedFormats = map[string]bool{
	"png":  true,
	"jpeg": true,
	"gif":  true,
}

// normalizeFormat normalizes a caller-supplied format string.
func normalizeFormat(format string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(format))
	if normalized == "jpg" {
		normalized = "jpeg"
	}
	if !supportedFormats[normalized] {
		return "", &knifer.Error{
			Code:    knifer.ErrCodeUnsupported,
			Message: fmt.Sprintf("image: unsupported format %q", format),
		}
	}
	return normalized, nil
}

// Thumbnail decodes a raster image from r and writes a downscaled copy to w.
//
// The output is resized proportionally so that its longest edge is at most
// maxEdge pixels. Images that are already smaller than maxEdge on both edges
// are re-encoded unchanged. If maxEdge is zero or negative the function
// returns ErrCodeInvalidInput.
//
// The resulting image is encoded using format, which must be one of "png",
// "jpeg"/"jpg" or "gif".
func Thumbnail(w io.Writer, r io.Reader, maxEdge int, format string) error {
	if w == nil {
		return &knifer.Error{Code: knifer.ErrCodeInvalidInput, Message: "image: nil writer"}
	}
	if r == nil {
		return &knifer.Error{Code: knifer.ErrCodeInvalidInput, Message: "image: nil reader"}
	}
	if maxEdge <= 0 {
		return &knifer.Error{Code: knifer.ErrCodeInvalidInput, Message: "image: maxEdge must be positive"}
	}
	normalized, err := normalizeFormat(format)
	if err != nil {
		return err
	}

	src, err := decodeAny(r)
	if err != nil {
		return err
	}

	bounds := src.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	if width == 0 || height == 0 {
		return &knifer.Error{Code: knifer.ErrCodeInvalidInput, Message: "image: empty source image"}
	}

	resized := src
	if width > maxEdge || height > maxEdge {
		newW, newH := fitLongEdge(width, height, maxEdge)
		resized = downsample(src, bounds, newW, newH)
	}

	return encodeAny(w, resized, normalized)
}

// ConvertFormat decodes r and re-encodes it into the target format.
//
// Source and target format may differ; the pixel payload is preserved. If r
// cannot be decoded as one of the supported formats the returned error
// carries ErrCodeInvalidInput. Invalid format names carry ErrCodeUnsupported.
func ConvertFormat(w io.Writer, r io.Reader, format string) error {
	if w == nil {
		return &knifer.Error{Code: knifer.ErrCodeInvalidInput, Message: "image: nil writer"}
	}
	if r == nil {
		return &knifer.Error{Code: knifer.ErrCodeInvalidInput, Message: "image: nil reader"}
	}
	normalized, err := normalizeFormat(format)
	if err != nil {
		return err
	}

	src, err := decodeAny(r)
	if err != nil {
		return err
	}

	return encodeAny(w, src, normalized)
}

// Info returns the width, height and detected format of the raster image
// available from r. It reads only the leading bytes required by the standard
// library decoders, so it remains cheap for large inputs.
//
// The format name is one of "png", "jpeg" or "gif". Unknown formats produce
// ErrCodeInvalidInput.
func Info(r io.Reader) (width int, height int, format string, err error) {
	if r == nil {
		return 0, 0, "", &knifer.Error{Code: knifer.ErrCodeInvalidInput, Message: "image: nil reader"}
	}

	cfg, name, err := decodeConfigAny(r)
	if err != nil {
		return 0, 0, "", err
	}
	return cfg.Width, cfg.Height, name, nil
}

// Resize returns img scaled to width x height using nearest-neighbor sampling.
func Resize(img image.Image, width, height int) (image.Image, error) {
	if img == nil {
		return nil, &knifer.Error{Code: knifer.ErrCodeInvalidInput, Message: "image: nil image"}
	}
	if width <= 0 || height <= 0 {
		return nil, &knifer.Error{Code: knifer.ErrCodeInvalidInput, Message: "image: resize dimensions must be positive"}
	}
	srcBounds := img.Bounds()
	if srcBounds.Dx() == 0 || srcBounds.Dy() == 0 {
		return nil, &knifer.Error{Code: knifer.ErrCodeInvalidInput, Message: "image: empty source image"}
	}
	return resizeNearest(img, width, height), nil
}

// Crop returns the rectangular region of img starting at x,y with width,height.
func Crop(img image.Image, x, y, width, height int) (image.Image, error) {
	if img == nil {
		return nil, &knifer.Error{Code: knifer.ErrCodeInvalidInput, Message: "image: nil image"}
	}
	if width <= 0 || height <= 0 {
		return nil, &knifer.Error{Code: knifer.ErrCodeInvalidInput, Message: "image: crop dimensions must be positive"}
	}
	rect := image.Rect(x, y, x+width, y+height)
	if !rect.In(img.Bounds()) {
		return nil, &knifer.Error{Code: knifer.ErrCodeInvalidInput, Message: "image: crop rectangle outside image bounds"}
	}
	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(dst, dst.Bounds(), img, rect.Min, draw.Src)
	return dst, nil
}

// CropCenter returns the centered width x height region of img.
func CropCenter(img image.Image, width, height int) (image.Image, error) {
	if img == nil {
		return nil, &knifer.Error{Code: knifer.ErrCodeInvalidInput, Message: "image: nil image"}
	}
	bounds := img.Bounds()
	x := bounds.Min.X + (bounds.Dx()-width)/2
	y := bounds.Min.Y + (bounds.Dy()-height)/2
	return Crop(img, x, y, width, height)
}

// FlipHorizontal mirrors img left-to-right.
func FlipHorizontal(img image.Image) (image.Image, error) {
	if img == nil {
		return nil, &knifer.Error{Code: knifer.ErrCodeInvalidInput, Message: "image: nil image"}
	}
	bounds := img.Bounds()
	dst := image.NewRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			dst.Set(bounds.Dx()-1-x, y, img.At(bounds.Min.X+x, bounds.Min.Y+y))
		}
	}
	return dst, nil
}

// FlipVertical mirrors img top-to-bottom.
func FlipVertical(img image.Image) (image.Image, error) {
	if img == nil {
		return nil, &knifer.Error{Code: knifer.ErrCodeInvalidInput, Message: "image: nil image"}
	}
	bounds := img.Bounds()
	dst := image.NewRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			dst.Set(x, bounds.Dy()-1-y, img.At(bounds.Min.X+x, bounds.Min.Y+y))
		}
	}
	return dst, nil
}

// Rotate90 rotates img 90 degrees clockwise.
func Rotate90(img image.Image) (image.Image, error) {
	if img == nil {
		return nil, &knifer.Error{Code: knifer.ErrCodeInvalidInput, Message: "image: nil image"}
	}
	bounds := img.Bounds()
	dst := image.NewRGBA(image.Rect(0, 0, bounds.Dy(), bounds.Dx()))
	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			dst.Set(bounds.Dy()-1-y, x, img.At(bounds.Min.X+x, bounds.Min.Y+y))
		}
	}
	return dst, nil
}

// Rotate180 rotates img 180 degrees.
func Rotate180(img image.Image) (image.Image, error) {
	if img == nil {
		return nil, &knifer.Error{Code: knifer.ErrCodeInvalidInput, Message: "image: nil image"}
	}
	bounds := img.Bounds()
	dst := image.NewRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			dst.Set(bounds.Dx()-1-x, bounds.Dy()-1-y, img.At(bounds.Min.X+x, bounds.Min.Y+y))
		}
	}
	return dst, nil
}

// Rotate270 rotates img 270 degrees clockwise.
func Rotate270(img image.Image) (image.Image, error) {
	if img == nil {
		return nil, &knifer.Error{Code: knifer.ErrCodeInvalidInput, Message: "image: nil image"}
	}
	bounds := img.Bounds()
	dst := image.NewRGBA(image.Rect(0, 0, bounds.Dy(), bounds.Dx()))
	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			dst.Set(y, bounds.Dx()-1-x, img.At(bounds.Min.X+x, bounds.Min.Y+y))
		}
	}
	return dst, nil
}

// Grayscale returns a grayscale copy of img while preserving alpha.
func Grayscale(img image.Image) (image.Image, error) {
	if img == nil {
		return nil, &knifer.Error{Code: knifer.ErrCodeInvalidInput, Message: "image: nil image"}
	}
	bounds := img.Bounds()
	dst := image.NewRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			r, g, b, a := img.At(bounds.Min.X+x, bounds.Min.Y+y).RGBA()
			gray := component8From32((299*(r>>8) + 587*(g>>8) + 114*(b>>8)) / 1000)
			dst.SetRGBA(x, y, color.RGBA{R: gray, G: gray, B: gray, A: component8From32(a >> 8)})
		}
	}
	return dst, nil
}

// CompressJPEG encodes img as JPEG with quality in [1,100].
func CompressJPEG(w io.Writer, img image.Image, quality int) error {
	if w == nil {
		return &knifer.Error{Code: knifer.ErrCodeInvalidInput, Message: "image: nil writer"}
	}
	if img == nil {
		return &knifer.Error{Code: knifer.ErrCodeInvalidInput, Message: "image: nil image"}
	}
	if quality < 1 || quality > 100 {
		return &knifer.Error{Code: knifer.ErrCodeInvalidInput, Message: "image: jpeg quality must be between 1 and 100"}
	}
	if err := jpeg.Encode(w, img, &jpeg.Options{Quality: quality}); err != nil {
		return &knifer.Error{Code: knifer.ErrCodeInternal, Message: "image: jpeg encode failed", Cause: err}
	}
	return nil
}

// decodeAny decodes r using the registered image formats, translating the
// generic image.ErrFormat into a knifer-go classified error.
func decodeAny(r io.Reader) (image.Image, error) {
	img, _, err := image.Decode(r)
	if err != nil {
		return nil, &knifer.Error{
			Code:    knifer.ErrCodeInvalidInput,
			Message: "image: decode failed",
			Cause:   err,
		}
	}
	return img, nil
}

// decodeConfigAny returns the configuration (bounds) of r without fully
// decoding the pixel data.
func decodeConfigAny(r io.Reader) (image.Config, string, error) {
	cfg, name, err := image.DecodeConfig(r)
	if err != nil {
		return image.Config{}, "", &knifer.Error{
			Code:    knifer.ErrCodeInvalidInput,
			Message: "image: decode config failed",
			Cause:   err,
		}
	}
	return cfg, name, nil
}

// encodeAny writes img to w using the named encoder.
func encodeAny(w io.Writer, img image.Image, format string) error {
	switch format {
	case "png":
		if err := png.Encode(w, img); err != nil {
			return &knifer.Error{Code: knifer.ErrCodeInternal, Message: "image: png encode failed", Cause: err}
		}
	case "jpeg":
		opts := &jpeg.Options{Quality: jpeg.DefaultQuality}
		if err := jpeg.Encode(w, img, opts); err != nil {
			return &knifer.Error{Code: knifer.ErrCodeInternal, Message: "image: jpeg encode failed", Cause: err}
		}
	case "gif":
		opts := &gif.Options{NumColors: 256}
		if err := gif.Encode(w, img, opts); err != nil {
			return &knifer.Error{Code: knifer.ErrCodeInternal, Message: "image: gif encode failed", Cause: err}
		}
	default:
		return &knifer.Error{
			Code:    knifer.ErrCodeUnsupported,
			Message: fmt.Sprintf("image: unsupported format %q", format),
		}
	}
	return nil
}

// fitLongEdge returns the (width, height) that fits within maxEdge while
// keeping the original aspect ratio. Both dimensions are clamped to at least
// one pixel so the output is never degenerate.
func fitLongEdge(width, height, maxEdge int) (int, int) {
	if width >= height {
		newW := maxEdge
		newH := (height * newW) / width
		if newH == 0 {
			newH = 1
		}
		return newW, newH
	}
	newH := maxEdge
	newW := (width * newH) / height
	if newW == 0 {
		newW = 1
	}
	return newW, newH
}

// downsample builds a newWidth x newHeight image by averaging the pixels in
// each source cell. It avoids visible aliasing for simple thumbnails while
// remaining a pure stdlib implementation.
func downsample(src image.Image, srcBounds image.Rectangle, newWidth, newHeight int) image.Image {
	dst := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))

	srcWidth := srcBounds.Dx()
	srcHeight := srcBounds.Dy()

	for dy := 0; dy < newHeight; dy++ {
		for dx := 0; dx < newWidth; dx++ {
			sxMin := (dx * srcWidth) / newWidth
			syMin := (dy * srcHeight) / newHeight
			sxMax := ((dx + 1) * srcWidth) / newWidth
			syMax := ((dy + 1) * srcHeight) / newHeight
			if sxMax == sxMin {
				sxMax = sxMin + 1
			}
			if syMax == syMin {
				syMax = syMin + 1
			}
			if sxMax > srcWidth {
				sxMax = srcWidth
			}
			if syMax > srcHeight {
				syMax = srcHeight
			}

			var r, g, b, a uint64
			count := uint64(0)
			for sy := syMin; sy < syMax; sy++ {
				for sx := sxMin; sx < sxMax; sx++ {
					cr, cg, cb, ca := src.At(srcBounds.Min.X+sx, srcBounds.Min.Y+sy).RGBA()
					r += uint64(cr >> 8)
					g += uint64(cg >> 8)
					b += uint64(cb >> 8)
					a += uint64(ca >> 8)
					count++
				}
			}
			if count == 0 {
				count = 1
			}
			dst.SetRGBA(dx, dy, color.RGBA{
				R: averageComponent8(r, count),
				G: averageComponent8(g, count),
				B: averageComponent8(b, count),
				A: averageComponent8(a, count),
			})
		}
	}
	return dst
}

func averageComponent8(total, count uint64) uint8 {
	if count == 0 {
		return 0
	}
	avg := total / count
	return component8From64(avg)
}

func component8From32(v uint32) uint8 {
	if v > 255 {
		return 255
	}
	return uint8(v)
}

func component8From64(v uint64) uint8 {
	if v > 255 {
		return 255
	}
	return uint8(v)
}
