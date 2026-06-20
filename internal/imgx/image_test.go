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

	knifer "github.com/imajinyun/go-knifer"
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
