// Package vimg exposes a thin facade over the internal image helpers.
//
// The public surface intentionally mirrors internal/imgx so callers never
// depend on implementation details directly. Import this package from
// application code; import internal/imgx only from tests within the
// repository that need access to internal helpers.
package vimg

import (
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
