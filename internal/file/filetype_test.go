package file

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	knifer "github.com/imajinyun/knifer-go"
)

func TestDetectFileTypeBytes(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		mime     string
		category FileCategory
		ext      string
	}{
		{name: "jpeg", data: []byte{0xFF, 0xD8, 0xFF, 0xE0}, mime: "image/jpeg", category: FileCategoryImage, ext: ".jpg"},
		{name: "png", data: []byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A}, mime: "image/png", category: FileCategoryImage, ext: ".png"},
		{name: "gif87a", data: []byte("GIF87a"), mime: "image/gif", category: FileCategoryImage, ext: ".gif"},
		{name: "gif89a", data: []byte("GIF89a"), mime: "image/gif", category: FileCategoryImage, ext: ".gif"},
		{name: "webp", data: []byte("RIFFxxxxWEBPVP8 "), mime: "image/webp", category: FileCategoryImage, ext: ".webp"},
		{name: "bmp", data: []byte("BMxxxx"), mime: "image/bmp", category: FileCategoryImage, ext: ".bmp"},
		{name: "tiff little endian", data: []byte{0x49, 0x49, 0x2A, 0x00}, mime: "image/tiff", category: FileCategoryImage, ext: ".tif"},
		{name: "tiff big endian", data: []byte{0x4D, 0x4D, 0x00, 0x2A}, mime: "image/tiff", category: FileCategoryImage, ext: ".tif"},
		{name: "pdf", data: []byte("%PDF-1.7"), mime: "application/pdf", category: FileCategoryDocument, ext: ".pdf"},
		{name: "zip", data: []byte("PK\x03\x04plain"), mime: "application/zip", category: FileCategoryArchive, ext: ".zip"},
		{name: "docx", data: []byte("PK\x03\x04word/document.xml"), mime: "application/vnd.openxmlformats-officedocument.wordprocessingml.document", category: FileCategoryDocument, ext: ".docx"},
		{name: "xlsx", data: []byte("PK\x03\x04xl/workbook.xml"), mime: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", category: FileCategoryDocument, ext: ".xlsx"},
		{name: "pptx", data: []byte("PK\x03\x04ppt/presentation.xml"), mime: "application/vnd.openxmlformats-officedocument.presentationml.presentation", category: FileCategoryDocument, ext: ".pptx"},
		{name: "rar", data: []byte("Rar!\x1A\x07\x00"), mime: "application/vnd.rar", category: FileCategoryArchive, ext: ".rar"},
		{name: "7z", data: []byte("7z\xBC\xAF\x27\x1C"), mime: "application/x-7z-compressed", category: FileCategoryArchive, ext: ".7z"},
		{name: "gzip", data: []byte{0x1F, 0x8B, 0x08}, mime: "application/gzip", category: FileCategoryArchive, ext: ".gz"},
		{name: "tar", data: tarHeader(), mime: "application/x-tar", category: FileCategoryArchive, ext: ".tar"},
		{name: "mp3 id3", data: []byte("ID3"), mime: "audio/mpeg", category: FileCategoryAudio, ext: ".mp3"},
		{name: "mp3 frame", data: []byte{0xFF, 0xFB}, mime: "audio/mpeg", category: FileCategoryAudio, ext: ".mp3"},
		{name: "wav", data: []byte("RIFFxxxxWAVEfmt "), mime: "audio/wav", category: FileCategoryAudio, ext: ".wav"},
		{name: "ogg", data: []byte("OggS"), mime: "audio/ogg", category: FileCategoryAudio, ext: ".ogg"},
		{name: "flac", data: []byte("fLaC"), mime: "audio/flac", category: FileCategoryAudio, ext: ".flac"},
		{name: "mp4", data: []byte{0x00, 0x00, 0x00, 0x18, 'f', 't', 'y', 'p', 'm', 'p', '4', '2'}, mime: "video/mp4", category: FileCategoryVideo, ext: ".mp4"},
		{name: "webm", data: []byte{0x1A, 0x45, 0xDF, 0xA3}, mime: "video/webm", category: FileCategoryVideo, ext: ".webm"},
		{name: "mpeg", data: []byte{0x00, 0x00, 0x01, 0xBA}, mime: "video/mpeg", category: FileCategoryVideo, ext: ".mpg"},
		{name: "ms office", data: []byte{0xD0, 0xCF, 0x11, 0xE0, 0xA1, 0xB1, 0x1A, 0xE1}, mime: "application/vnd.ms-office", category: FileCategoryDocument, ext: ".doc"},
		{name: "ttf", data: []byte{0x00, 0x01, 0x00, 0x00}, mime: "font/ttf", category: FileCategoryFont, ext: ".ttf"},
		{name: "otf", data: []byte("OTTO"), mime: "font/otf", category: FileCategoryFont, ext: ".otf"},
		{name: "elf", data: []byte{0x7F, 'E', 'L', 'F'}, mime: "application/x-elf", category: FileCategoryExecutable, ext: ".elf"},
		{name: "mach-o 32 big", data: []byte{0xFE, 0xED, 0xFA, 0xCE}, mime: "application/x-mach-binary", category: FileCategoryExecutable, ext: ".macho"},
		{name: "mach-o 64 big", data: []byte{0xFE, 0xED, 0xFA, 0xCF}, mime: "application/x-mach-binary", category: FileCategoryExecutable, ext: ".macho"},
		{name: "mach-o 32 little", data: []byte{0xCF, 0xFA, 0xED, 0xFE}, mime: "application/x-mach-binary", category: FileCategoryExecutable, ext: ".macho"},
		{name: "fat mach-o", data: []byte{0xCA, 0xFE, 0xBA, 0xBE}, mime: "application/x-mach-binary", category: FileCategoryExecutable, ext: ".macho"},
		{name: "pe", data: []byte("MZ"), mime: "application/vnd.microsoft.portable-executable", category: FileCategoryExecutable, ext: ".exe"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetectFileTypeBytes(tt.data)
			if got.MIME != tt.mime || got.Category != tt.category || got.Extension != tt.ext {
				t.Fatalf("DetectFileTypeBytes = %#v", got)
			}
		})
	}
}

func TestDetectFileTypeUnknownNilAndTruncated(t *testing.T) {
	if got := DetectFileTypeBytes([]byte("hello")); got != UnknownFileType {
		t.Fatalf("unknown = %#v", got)
	}
	if got := DetectFileTypeBytes([]byte{0x89, 'P', 'N'}); got != UnknownFileType {
		t.Fatalf("truncated png = %#v, want unknown", got)
	}
	if got := DetectFileTypeBytes([]byte("RIFFxxxxAVI ")); got != UnknownFileType {
		t.Fatalf("unsupported riff family = %#v, want unknown", got)
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

func TestDetectFileTypeFromPath(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sample.gz")
	if err := os.WriteFile(path, []byte{0x1F, 0x8B, 0x08}, 0o600); err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}
	ft, err := DetectFileTypeFromPath(path)
	if err != nil {
		t.Fatalf("DetectFileTypeFromPath error = %v", err)
	}
	if !IsArchive(ft) || ft.Extension != ".gz" {
		t.Fatalf("DetectFileTypeFromPath = %#v", ft)
	}
}

func tarHeader() []byte {
	data := make([]byte, 265)
	copy(data[257:], "ustar")
	return data
}
