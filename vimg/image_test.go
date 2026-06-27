package vimg

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"
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

func TestInfoFacade(t *testing.T) {
	src := bytes.NewReader(makePNG(32, 16))
	w, h, f, err := Info(src)
	if err != nil {
		t.Fatalf("Info: %v", err)
	}
	if w != 32 || h != 16 {
		t.Errorf("Info: got %dx%d, want 32x16", w, h)
	}
	if f != "png" {
		t.Errorf("Info format: got %q, want png", f)
	}
}

func TestThumbnailFacade(t *testing.T) {
	src := bytes.NewReader(makePNG(200, 200))
	out := &bytes.Buffer{}
	if err := Thumbnail(out, src, 32, "png"); err != nil {
		t.Fatalf("Thumbnail: %v", err)
	}
	if out.Len() == 0 {
		t.Fatal("Thumbnail: empty output")
	}
	cfg, err := png.DecodeConfig(out)
	if err != nil {
		t.Fatalf("thumbnail output not a png: %v", err)
	}
	if cfg.Width > 32 || cfg.Height > 32 {
		t.Errorf("thumbnail too large: %dx%d", cfg.Width, cfg.Height)
	}
}

func TestConvertFormatFacade(t *testing.T) {
	src := bytes.NewReader(makePNG(16, 16))
	out := &bytes.Buffer{}
	if err := ConvertFormat(out, src, "jpeg"); err != nil {
		t.Fatalf("ConvertFormat: %v", err)
	}
	if out.Len() == 0 {
		t.Fatal("ConvertFormat: empty output")
	}
}

func TestImageOpsFacade(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 3, 2))
	src.Set(0, 0, color.RGBA{R: 255, A: 255})
	src.Set(1, 0, color.RGBA{G: 255, A: 255})
	src.Set(2, 0, color.RGBA{B: 255, A: 255})

	resized, err := Resize(src, 6, 4)
	if err != nil {
		t.Fatalf("Resize: %v", err)
	}
	if resized.Bounds().Dx() != 6 || resized.Bounds().Dy() != 4 {
		t.Fatalf("Resize bounds = %v, want 6x4", resized.Bounds())
	}

	cropped, err := Crop(src, 1, 0, 2, 1)
	if err != nil {
		t.Fatalf("Crop: %v", err)
	}
	if cropped.Bounds().Dx() != 2 || cropped.Bounds().Dy() != 1 {
		t.Fatalf("Crop bounds = %v, want 2x1", cropped.Bounds())
	}

	center, err := CropCenter(src, 1, 2)
	if err != nil {
		t.Fatalf("CropCenter: %v", err)
	}
	if center.Bounds().Dx() != 1 || center.Bounds().Dy() != 2 {
		t.Fatalf("CropCenter bounds = %v, want 1x2", center.Bounds())
	}

	if _, err := FlipHorizontal(src); err != nil {
		t.Fatalf("FlipHorizontal: %v", err)
	}
	if _, err := FlipVertical(src); err != nil {
		t.Fatalf("FlipVertical: %v", err)
	}
	if _, err := Rotate90(src); err != nil {
		t.Fatalf("Rotate90: %v", err)
	}
	if _, err := Rotate180(src); err != nil {
		t.Fatalf("Rotate180: %v", err)
	}
	if _, err := Rotate270(src); err != nil {
		t.Fatalf("Rotate270: %v", err)
	}
	if _, err := Grayscale(src); err != nil {
		t.Fatalf("Grayscale: %v", err)
	}

	out := &bytes.Buffer{}
	if err := CompressJPEG(out, src, 85); err != nil {
		t.Fatalf("CompressJPEG: %v", err)
	}
	if out.Len() == 0 {
		t.Fatal("CompressJPEG: empty output")
	}
}

func TestThumbnailBadArgs(t *testing.T) {
	if err := Thumbnail(nil, bytes.NewReader(makePNG(8, 8)), 8, "png"); err == nil {
		t.Fatal("expected error for nil writer")
	}
	if err := Thumbnail(&bytes.Buffer{}, nil, 8, "png"); err == nil {
		t.Fatal("expected error for nil reader")
	}
	if err := Thumbnail(&bytes.Buffer{}, bytes.NewReader(makePNG(8, 8)), 0, "png"); err == nil {
		t.Fatal("expected error for zero maxEdge")
	}
	if err := Thumbnail(&bytes.Buffer{}, bytes.NewReader(makePNG(8, 8)), 8, "bmp"); err == nil {
		t.Fatal("expected error for unsupported format")
	}
}

func BenchmarkThumbnailFacadePNG(b *testing.B) {
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
