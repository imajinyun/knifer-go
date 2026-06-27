// Package vimg exposes a thin facade over the internal image helpers.
//
// The public surface intentionally mirrors internal/imgx so callers never
// depend on implementation details directly. Import this package from
// application code; import internal/imgx only from tests within the
// repository that need access to internal helpers.
package vimg

import (
	"image"
	"io"

	"github.com/imajinyun/knifer-go/internal/imgx"
)

// Thumbnail decodes a raster image from r and writes a downscaled copy to w.
//
// The image is resized proportionally so that its longest edge is at most
// maxEdge pixels; smaller images are re-encoded unchanged. The output is
// written in the requested format (png, jpeg/jpg or gif).
func Thumbnail(w io.Writer, r io.Reader, maxEdge int, format string) error {
	return imgx.Thumbnail(w, r, maxEdge, format)
}

// ConvertFormat decodes r and re-encodes it into the target format.
func ConvertFormat(w io.Writer, r io.Reader, format string) error {
	return imgx.ConvertFormat(w, r, format)
}

// Info returns the width, height and detected format of the raster image
// available from r.
func Info(r io.Reader) (width, height int, format string, err error) {
	return imgx.Info(r)
}

// Resize returns img scaled to width x height using nearest-neighbor sampling.
func Resize(img image.Image, width, height int) (image.Image, error) {
	return imgx.Resize(img, width, height)
}

// Crop returns the rectangular region of img starting at x,y with width,height.
func Crop(img image.Image, x, y, width, height int) (image.Image, error) {
	return imgx.Crop(img, x, y, width, height)
}

// CropCenter returns the centered width x height region of img.
func CropCenter(img image.Image, width, height int) (image.Image, error) {
	return imgx.CropCenter(img, width, height)
}

// FlipHorizontal mirrors img left-to-right.
func FlipHorizontal(img image.Image) (image.Image, error) {
	return imgx.FlipHorizontal(img)
}

// FlipVertical mirrors img top-to-bottom.
func FlipVertical(img image.Image) (image.Image, error) {
	return imgx.FlipVertical(img)
}

// Rotate90 rotates img 90 degrees clockwise.
func Rotate90(img image.Image) (image.Image, error) {
	return imgx.Rotate90(img)
}

// Rotate180 rotates img 180 degrees.
func Rotate180(img image.Image) (image.Image, error) {
	return imgx.Rotate180(img)
}

// Rotate270 rotates img 270 degrees clockwise.
func Rotate270(img image.Image) (image.Image, error) {
	return imgx.Rotate270(img)
}

// Grayscale returns a grayscale copy of img while preserving alpha.
func Grayscale(img image.Image) (image.Image, error) {
	return imgx.Grayscale(img)
}

// CompressJPEG encodes img as JPEG with quality in [1,100].
func CompressJPEG(w io.Writer, img image.Image, quality int) error {
	return imgx.CompressJPEG(w, img, quality)
}
