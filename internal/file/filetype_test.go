package file

import (
	"errors"
	"strings"
	"testing"

	knifer "github.com/imajinyun/knifer-go"
)

func TestDetectFileTypeBytes(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		category FileCategory
		ext      string
	}{
		{name: "jpeg", data: []byte{0xFF, 0xD8, 0xFF, 0xE0}, category: FileCategoryImage, ext: ".jpg"},
		{name: "png", data: []byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A}, category: FileCategoryImage, ext: ".png"},
		{name: "pdf", data: []byte("%PDF-1.7"), category: FileCategoryDocument, ext: ".pdf"},
		{name: "zip", data: []byte("PK\x03\x04plain"), category: FileCategoryArchive, ext: ".zip"},
		{name: "webp", data: []byte("RIFFxxxxWEBPVP8 "), category: FileCategoryImage, ext: ".webp"},
		{name: "wav", data: []byte("RIFFxxxxWAVEfmt "), category: FileCategoryAudio, ext: ".wav"},
		{name: "elf", data: []byte{0x7F, 'E', 'L', 'F'}, category: FileCategoryExecutable, ext: ".elf"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetectFileTypeBytes(tt.data)
			if got.Category != tt.category || got.Extension != tt.ext {
				t.Fatalf("DetectFileTypeBytes = %#v", got)
			}
		})
	}
}

func TestDetectFileTypeUnknownAndNil(t *testing.T) {
	if got := DetectFileTypeBytes([]byte("hello")); got != UnknownFileType {
		t.Fatalf("unknown = %#v", got)
	}
	_, err := DetectFileType(nil)
	if !errors.Is(err, knifer.ErrCodeInvalidInput) {
		t.Fatalf("nil reader err = %v", err)
	}
	ft, err := DetectFileType(strings.NewReader("\x89PNG\r\n\x1a\n"))
	if err != nil || !IsImage(ft) {
		t.Fatalf("DetectFileType reader = %#v, %v", ft, err)
	}
}
