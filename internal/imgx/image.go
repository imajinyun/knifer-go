// Package imgx provides a tiny, stdlib-only surface on top of the
// standard image, image/jpeg, image/png and image/gif packages.
//
// It intentionally keeps its helper set small: image metadata inspection,
// lossless format conversion between PNG/JPEG/GIF, and a simple proportional
// downscaling helper. Nothing is drawn on top of third-party libraries so
// callers only pay for the pieces they actually use.
package imgx

import (
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"strings"

	knifer "github.com/imajinyun/go-knifer"
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

// decodeAny decodes r using the registered image formats, translating the
// generic image.ErrFormat into a go-knifer classified error.
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
	if avg > 255 {
		return 255
	}
	return uint8(avg)
}
