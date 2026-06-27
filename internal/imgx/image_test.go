package imgx

import (
	"bytes"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"testing"

	knifer "github.com/imajinyun/knifer-go"
)

func makePNG(width, height int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.SetRGBA(x, y, color.RGBA{R: uint8(x % 256), G: uint8(y % 256), B: 128, A: 255})
		}
	}
	buf := &bytes.Buffer{}
	_ = png.Encode(buf, img)
	return buf.Bytes()
}

func makeJPEG(width, height int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.SetRGBA(x, y, color.RGBA{R: uint8(x % 256), G: uint8(y % 256), B: 32, A: 255})
		}
	}
	buf := &bytes.Buffer{}
	_ = jpeg.Encode(buf, img, &jpeg.Options{Quality: 90})
	return buf.Bytes()
}

func makeGIF(width, height int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.SetRGBA(x, y, color.RGBA{R: uint8(x % 256), G: uint8(y % 256), B: 0, A: 255})
		}
	}
	buf := &bytes.Buffer{}
	_ = gif.Encode(buf, img, &gif.Options{NumColors: 256})
	return buf.Bytes()
}

func TestInfoPNG(t *testing.T) {
	src := bytes.NewReader(makePNG(200, 120))
	w, h, f, err := Info(src)
	if err != nil {
		t.Fatalf("Info: %v", err)
	}
	if w != 200 || h != 120 {
		t.Errorf("Info: got %dx%d, want 200x120", w, h)
	}
	if f != "png" {
		t.Errorf("Info format: got %q, want png", f)
	}
}

func TestInfoJPEG(t *testing.T) {
	src := bytes.NewReader(makeJPEG(80, 60))
	w, h, f, err := Info(src)
	if err != nil {
		t.Fatalf("Info: %v", err)
	}
	if w != 80 || h != 60 {
		t.Errorf("Info: got %dx%d, want 80x60", w, h)
	}
	if f != "jpeg" {
		t.Errorf("Info format: got %q, want jpeg", f)
	}
}

func TestInfoGIF(t *testing.T) {
	src := bytes.NewReader(makeGIF(32, 48))
	_, _, f, err := Info(src)
	if err != nil {
		t.Fatalf("Info: %v", err)
	}
	if f != "gif" {
		t.Errorf("Info format: got %q, want gif", f)
	}
}

func TestInfoNilReader(t *testing.T) {
	_, _, _, err := Info(nil)
	if err == nil {
		t.Fatal("Info(nil): expected error, got nil")
	}
}

func TestInfoInvalid(t *testing.T) {
	src := bytes.NewReader([]byte("this is not an image"))
	_, _, _, err := Info(src)
	if err == nil {
		t.Fatal("Info(invalid): expected error")
	}
}

func TestConvertFormatPNGToJPEG(t *testing.T) {
	src := bytes.NewReader(makePNG(100, 80))
	out := &bytes.Buffer{}
	if err := ConvertFormat(out, src, "jpeg"); err != nil {
		t.Fatalf("ConvertFormat: %v", err)
	}
	if out.Len() == 0 {
		t.Fatal("ConvertFormat: empty output")
	}
	_, err := jpeg.DecodeConfig(out)
	if err != nil {
		t.Errorf("output not a jpeg: %v", err)
	}
}

func TestConvertFormatJPEGToPNG(t *testing.T) {
	src := bytes.NewReader(makeJPEG(64, 64))
	out := &bytes.Buffer{}
	if err := ConvertFormat(out, src, "png"); err != nil {
		t.Fatalf("ConvertFormat: %v", err)
	}
	_, err := png.DecodeConfig(out)
	if err != nil {
		t.Errorf("output not a png: %v", err)
	}
}

func TestConvertFormatUnsupported(t *testing.T) {
	src := bytes.NewReader(makePNG(10, 10))
	err := ConvertFormat(&bytes.Buffer{}, src, "bmp")
	if err == nil {
		t.Fatal("ConvertFormat(bmp): expected error")
	}
}

func TestConvertFormatBadInput(t *testing.T) {
	src := bytes.NewReader([]byte("not an image"))
	err := ConvertFormat(&bytes.Buffer{}, src, "png")
	if err == nil {
		t.Fatal("ConvertFormat(bad): expected error")
	}
}

func TestThumbnailShrinkPNG(t *testing.T) {
	src := bytes.NewReader(makePNG(400, 300))
	out := &bytes.Buffer{}
	if err := Thumbnail(out, src, 64, "png"); err != nil {
		t.Fatalf("Thumbnail: %v", err)
	}
	if out.Len() == 0 {
		t.Fatal("Thumbnail: empty output")
	}
	cfg, err := png.DecodeConfig(out)
	if err != nil {
		t.Fatalf("thumbnail output not a png: %v", err)
	}
	if cfg.Width > 64 || cfg.Height > 64 {
		t.Errorf("thumbnail too large: %dx%d", cfg.Width, cfg.Height)
	}
	if cfg.Width == 0 || cfg.Height == 0 {
		t.Errorf("thumbnail degenerate: %dx%d", cfg.Width, cfg.Height)
	}
}

func TestThumbnailSmallerThanMax(t *testing.T) {
	src := bytes.NewReader(makePNG(20, 20))
	out := &bytes.Buffer{}
	if err := Thumbnail(out, src, 64, "png"); err != nil {
		t.Fatalf("Thumbnail: %v", err)
	}
	cfg, err := png.DecodeConfig(out)
	if err != nil {
		t.Fatalf("thumbnail output not a png: %v", err)
	}
	if cfg.Width != 20 || cfg.Height != 20 {
		t.Errorf("small image was resized: got %dx%d want 20x20", cfg.Width, cfg.Height)
	}
}

func TestThumbnailInvalidArgs(t *testing.T) {
	src := makePNG(10, 10)
	cases := []struct {
		name    string
		w       io.Writer
		r       io.Reader
		maxEdge int
		format  string
	}{
		{"nil writer", nil, bytes.NewReader(src), 10, "png"},
		{"nil reader", io.Discard, nil, 10, "png"},
		{"maxEdge zero", io.Discard, bytes.NewReader(src), 0, "png"},
		{"maxEdge negative", io.Discard, bytes.NewReader(src), -1, "png"},
		{"bad format", io.Discard, bytes.NewReader(src), 10, "bmp"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := Thumbnail(tc.w, tc.r, tc.maxEdge, tc.format)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}

func TestThumbnailErrClassified(t *testing.T) {
	src := bytes.NewReader(makePNG(10, 10))
	err := Thumbnail(&bytes.Buffer{}, src, -5, "png")
	code, ok := knifer.CodeOf(err)
	if !ok {
		t.Fatalf("expected CodeCarrier; got %v", err)
	}
	if code != knifer.ErrCodeInvalidInput {
		t.Errorf("got code %v, want %v", code, knifer.ErrCodeInvalidInput)
	}
}

func TestThumbnailRoundTripJPEG(t *testing.T) {
	src := bytes.NewReader(makeJPEG(512, 256))
	out := &bytes.Buffer{}
	if err := Thumbnail(out, src, 128, "jpeg"); err != nil {
		t.Fatalf("Thumbnail: %v", err)
	}
	_, err := jpeg.DecodeConfig(out)
	if err != nil {
		t.Errorf("thumbnail output not a jpeg: %v", err)
	}
}

func TestThumbnailPortrait(t *testing.T) {
	src := bytes.NewReader(makePNG(80, 320))
	out := &bytes.Buffer{}
	if err := Thumbnail(out, src, 64, "png"); err != nil {
		t.Fatalf("Thumbnail: %v", err)
	}
	cfg, err := png.DecodeConfig(out)
	if err != nil {
		t.Fatalf("thumbnail output not a png: %v", err)
	}
	if cfg.Height > 64 {
		t.Errorf("portrait thumbnail too tall: %dx%d", cfg.Width, cfg.Height)
	}
}

func TestImageOperations(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 3, 2))
	red := color.RGBA{R: 255, A: 255}
	green := color.RGBA{G: 255, A: 255}
	blue := color.RGBA{B: 255, A: 255}
	yellow := color.RGBA{R: 255, G: 255, A: 255}
	src.Set(0, 0, red)
	src.Set(1, 0, green)
	src.Set(2, 0, blue)
	src.Set(0, 1, yellow)

	cropped, err := Crop(src, 1, 0, 2, 1)
	if err != nil {
		t.Fatalf("Crop error = %v", err)
	}
	if cropped.Bounds().Dx() != 2 || cropped.Bounds().Dy() != 1 {
		t.Fatalf("Crop bounds = %v, want 2x1", cropped.Bounds())
	}
	assertSameColor(t, cropped.At(0, 0), green)

	center, err := CropCenter(src, 1, 2)
	if err != nil {
		t.Fatalf("CropCenter error = %v", err)
	}
	if center.Bounds().Dx() != 1 || center.Bounds().Dy() != 2 {
		t.Fatalf("CropCenter bounds = %v, want 1x2", center.Bounds())
	}
	assertSameColor(t, center.At(0, 0), green)

	horizontal, err := FlipHorizontal(src)
	if err != nil {
		t.Fatalf("FlipHorizontal error = %v", err)
	}
	assertSameColor(t, horizontal.At(2, 0), red)

	vertical, err := FlipVertical(src)
	if err != nil {
		t.Fatalf("FlipVertical error = %v", err)
	}
	assertSameColor(t, vertical.At(0, 1), red)
}

func TestImageRotateResizeGrayscaleAndJPEG(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 2, 1))
	red := color.RGBA{R: 255, A: 255}
	blue := color.RGBA{B: 255, A: 255}
	src.Set(0, 0, red)
	src.Set(1, 0, blue)

	rot90, err := Rotate90(src)
	if err != nil {
		t.Fatalf("Rotate90 error = %v", err)
	}
	if rot90.Bounds().Dx() != 1 || rot90.Bounds().Dy() != 2 {
		t.Fatalf("Rotate90 bounds = %v, want 1x2", rot90.Bounds())
	}
	assertSameColor(t, rot90.At(0, 0), red)

	rot180, err := Rotate180(src)
	if err != nil {
		t.Fatalf("Rotate180 error = %v", err)
	}
	assertSameColor(t, rot180.At(1, 0), red)

	rot270, err := Rotate270(src)
	if err != nil {
		t.Fatalf("Rotate270 error = %v", err)
	}
	if rot270.Bounds().Dx() != 1 || rot270.Bounds().Dy() != 2 {
		t.Fatalf("Rotate270 bounds = %v, want 1x2", rot270.Bounds())
	}
	assertSameColor(t, rot270.At(0, 0), blue)

	resized, err := Resize(src, 4, 2)
	if err != nil {
		t.Fatalf("Resize error = %v", err)
	}
	if resized.Bounds().Dx() != 4 || resized.Bounds().Dy() != 2 {
		t.Fatalf("Resize bounds = %v, want 4x2", resized.Bounds())
	}
	assertSameColor(t, resized.At(0, 0), red)
	assertSameColor(t, resized.At(3, 1), blue)

	gray, err := Grayscale(src)
	if err != nil {
		t.Fatalf("Grayscale error = %v", err)
	}
	r, g, b, _ := gray.At(0, 0).RGBA()
	if r != g || g != b {
		t.Fatalf("Grayscale channels = %d/%d/%d, want equal", r, g, b)
	}

	out := &bytes.Buffer{}
	if err := CompressJPEG(out, src, 80); err != nil {
		t.Fatalf("CompressJPEG error = %v", err)
	}
	if _, err := jpeg.DecodeConfig(out); err != nil {
		t.Fatalf("CompressJPEG output is not JPEG: %v", err)
	}
}

func TestImageOperationInvalidInput(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	tests := []struct {
		name string
		err  error
	}{
		{name: "resize nil", err: imageErr(Resize(nil, 1, 1))},
		{name: "resize invalid", err: imageErr(Resize(img, 0, 1))},
		{name: "crop nil", err: imageErr(Crop(nil, 0, 0, 1, 1))},
		{name: "crop invalid size", err: imageErr(Crop(img, 0, 0, 0, 1))},
		{name: "crop outside", err: imageErr(Crop(img, 1, 1, 2, 2))},
		{name: "flip nil", err: imageErr(FlipHorizontal(nil))},
		{name: "rotate nil", err: imageErr(Rotate90(nil))},
		{name: "gray nil", err: imageErr(Grayscale(nil))},
		{name: "compress nil writer", err: CompressJPEG(nil, img, 80)},
		{name: "compress nil image", err: CompressJPEG(&bytes.Buffer{}, nil, 80)},
		{name: "compress bad quality", err: CompressJPEG(&bytes.Buffer{}, img, 0)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, ok := knifer.CodeOf(tt.err)
			if !ok {
				t.Fatalf("error = %v, want CodeCarrier", tt.err)
			}
			if code != knifer.ErrCodeInvalidInput {
				t.Fatalf("code = %v, want %v", code, knifer.ErrCodeInvalidInput)
			}
		})
	}
}

func assertSameColor(t *testing.T, got color.Color, want color.RGBA) {
	t.Helper()
	r, g, b, a := got.RGBA()
	if uint8(r>>8) != want.R || uint8(g>>8) != want.G || uint8(b>>8) != want.B || uint8(a>>8) != want.A {
		t.Fatalf("color = rgba(%d,%d,%d,%d), want %+v", uint8(r>>8), uint8(g>>8), uint8(b>>8), uint8(a>>8), want)
	}
}

func imageErr(_ image.Image, err error) error {
	return err
}

func BenchmarkThumbnailPNG(b *testing.B) {
	src := makePNG(320, 240)
	b.ReportAllocs()
	var sink int
	for b.Loop() {
		out := &bytes.Buffer{}
		if err := Thumbnail(out, bytes.NewReader(src), 80, "png"); err != nil {
			b.Fatal(err)
		}
		sink = out.Len()
	}
	_ = sink
}
