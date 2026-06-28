package file

import (
	"bytes"
	"io"
	"os"
	"strings"
)

// FileCategory identifies a broad file family.
type FileCategory string

const (
	// FileCategoryUnknown means the file type could not be identified.
	FileCategoryUnknown FileCategory = "unknown"
	// FileCategoryImage identifies image files.
	FileCategoryImage FileCategory = "image"
	// FileCategoryVideo identifies video files.
	FileCategoryVideo FileCategory = "video"
	// FileCategoryAudio identifies audio files.
	FileCategoryAudio FileCategory = "audio"
	// FileCategoryArchive identifies archive or compressed files.
	FileCategoryArchive FileCategory = "archive"
	// FileCategoryDocument identifies document files.
	FileCategoryDocument FileCategory = "document"
	// FileCategoryFont identifies font files.
	FileCategoryFont FileCategory = "font"
	// FileCategoryExecutable identifies executable or object files.
	FileCategoryExecutable FileCategory = "executable"
)

// FileType contains magic-number detection metadata.
type FileType struct {
	MIME      string
	Extension string
	Category  FileCategory
}

var UnknownFileType = FileType{
	MIME:      "application/octet-stream",
	Extension: "",
	Category:  FileCategoryUnknown,
}

type fileSignature struct {
	prefix    []byte
	offset    int
	fileType  FileType
	matchFunc func([]byte) bool
}

var fileSignatures = []fileSignature{
	{prefix: []byte{0xFF, 0xD8, 0xFF}, fileType: FileType{MIME: "image/jpeg", Extension: ".jpg", Category: FileCategoryImage}},
	{prefix: []byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A}, fileType: FileType{MIME: "image/png", Extension: ".png", Category: FileCategoryImage}},
	{prefix: []byte("GIF87a"), fileType: FileType{MIME: "image/gif", Extension: ".gif", Category: FileCategoryImage}},
	{prefix: []byte("GIF89a"), fileType: FileType{MIME: "image/gif", Extension: ".gif", Category: FileCategoryImage}},
	{prefix: []byte("RIFF"), fileType: FileType{MIME: "image/webp", Extension: ".webp", Category: FileCategoryImage}, matchFunc: matchWebP},
	{prefix: []byte{'B', 'M'}, fileType: FileType{MIME: "image/bmp", Extension: ".bmp", Category: FileCategoryImage}},
	{prefix: []byte{0x49, 0x49, 0x2A, 0x00}, fileType: FileType{MIME: "image/tiff", Extension: ".tif", Category: FileCategoryImage}},
	{prefix: []byte{0x4D, 0x4D, 0x00, 0x2A}, fileType: FileType{MIME: "image/tiff", Extension: ".tif", Category: FileCategoryImage}},
	{prefix: []byte("%PDF-"), fileType: FileType{MIME: "application/pdf", Extension: ".pdf", Category: FileCategoryDocument}},
	{prefix: []byte("PK\x03\x04"), fileType: FileType{MIME: "application/zip", Extension: ".zip", Category: FileCategoryArchive}, matchFunc: matchZipFamily},
	{prefix: []byte{0x50, 0x4B, 0x03, 0x04}, fileType: FileType{MIME: "application/vnd.openxmlformats-officedocument.wordprocessingml.document", Extension: ".docx", Category: FileCategoryDocument}, matchFunc: matchOOXMLWord},
	{prefix: []byte{0x50, 0x4B, 0x03, 0x04}, fileType: FileType{MIME: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", Extension: ".xlsx", Category: FileCategoryDocument}, matchFunc: matchOOXMLSpreadsheet},
	{prefix: []byte{0x50, 0x4B, 0x03, 0x04}, fileType: FileType{MIME: "application/vnd.openxmlformats-officedocument.presentationml.presentation", Extension: ".pptx", Category: FileCategoryDocument}, matchFunc: matchOOXMLPresentation},
	{prefix: []byte("Rar!\x1A\x07\x00"), fileType: FileType{MIME: "application/vnd.rar", Extension: ".rar", Category: FileCategoryArchive}},
	{prefix: []byte("7z\xBC\xAF\x27\x1C"), fileType: FileType{MIME: "application/x-7z-compressed", Extension: ".7z", Category: FileCategoryArchive}},
	{prefix: []byte{0x1F, 0x8B, 0x08}, fileType: FileType{MIME: "application/gzip", Extension: ".gz", Category: FileCategoryArchive}},
	{prefix: []byte("ustar"), offset: 257, fileType: FileType{MIME: "application/x-tar", Extension: ".tar", Category: FileCategoryArchive}},
	{prefix: []byte("ID3"), fileType: FileType{MIME: "audio/mpeg", Extension: ".mp3", Category: FileCategoryAudio}},
	{prefix: []byte{0xFF, 0xFB}, fileType: FileType{MIME: "audio/mpeg", Extension: ".mp3", Category: FileCategoryAudio}},
	{prefix: []byte("RIFF"), fileType: FileType{MIME: "audio/wav", Extension: ".wav", Category: FileCategoryAudio}, matchFunc: matchWAV},
	{prefix: []byte("OggS"), fileType: FileType{MIME: "audio/ogg", Extension: ".ogg", Category: FileCategoryAudio}},
	{prefix: []byte("fLaC"), fileType: FileType{MIME: "audio/flac", Extension: ".flac", Category: FileCategoryAudio}},
	{prefix: []byte("ftyp"), offset: 4, fileType: FileType{MIME: "video/mp4", Extension: ".mp4", Category: FileCategoryVideo}, matchFunc: matchMP4},
	{prefix: []byte{0x1A, 0x45, 0xDF, 0xA3}, fileType: FileType{MIME: "video/webm", Extension: ".webm", Category: FileCategoryVideo}},
	{prefix: []byte{0x00, 0x00, 0x01, 0xBA}, fileType: FileType{MIME: "video/mpeg", Extension: ".mpg", Category: FileCategoryVideo}},
	{prefix: []byte{0xD0, 0xCF, 0x11, 0xE0, 0xA1, 0xB1, 0x1A, 0xE1}, fileType: FileType{MIME: "application/vnd.ms-office", Extension: ".doc", Category: FileCategoryDocument}},
	{prefix: []byte{0x00, 0x01, 0x00, 0x00}, fileType: FileType{MIME: "font/ttf", Extension: ".ttf", Category: FileCategoryFont}},
	{prefix: []byte("OTTO"), fileType: FileType{MIME: "font/otf", Extension: ".otf", Category: FileCategoryFont}},
	{prefix: []byte{0x7F, 'E', 'L', 'F'}, fileType: FileType{MIME: "application/x-elf", Extension: ".elf", Category: FileCategoryExecutable}},
	{prefix: []byte{0xFE, 0xED, 0xFA, 0xCE}, fileType: FileType{MIME: "application/x-mach-binary", Extension: ".macho", Category: FileCategoryExecutable}},
	{prefix: []byte{0xFE, 0xED, 0xFA, 0xCF}, fileType: FileType{MIME: "application/x-mach-binary", Extension: ".macho", Category: FileCategoryExecutable}},
	{prefix: []byte{0xCF, 0xFA, 0xED, 0xFE}, fileType: FileType{MIME: "application/x-mach-binary", Extension: ".macho", Category: FileCategoryExecutable}},
	{prefix: []byte{0xCA, 0xFE, 0xBA, 0xBE}, fileType: FileType{MIME: "application/x-mach-binary", Extension: ".macho", Category: FileCategoryExecutable}},
	{prefix: []byte{'M', 'Z'}, fileType: FileType{MIME: "application/vnd.microsoft.portable-executable", Extension: ".exe", Category: FileCategoryExecutable}},
}

// DetectFileType detects a file type from leading magic-number bytes.
func DetectFileType(r io.Reader) (FileType, error) {
	if r == nil {
		return UnknownFileType, invalidInputf("reader is nil")
	}
	head := make([]byte, 560)
	n, err := io.ReadFull(r, head)
	if err != nil && err != io.ErrUnexpectedEOF && err != io.EOF {
		return UnknownFileType, wrapFileIO("detect file type", err)
	}
	return DetectFileTypeBytes(head[:n]), nil
}

// DetectFileTypeBytes detects a file type from bytes already in memory.
func DetectFileTypeBytes(data []byte) FileType {
	for _, sig := range fileSignatures {
		end := sig.offset + len(sig.prefix)
		if len(data) < end || !bytes.Equal(data[sig.offset:end], sig.prefix) {
			continue
		}
		if sig.matchFunc != nil && !sig.matchFunc(data) {
			continue
		}
		return sig.fileType
	}
	return UnknownFileType
}

// DetectFileTypeFromPath detects a file type by opening path and reading its header.
func DetectFileTypeFromPath(path string) (FileType, error) {
	// #nosec G304 -- file helpers intentionally read caller-provided paths.
	f, err := os.Open(path)
	if err != nil {
		return UnknownFileType, wrapFileIO("open file "+path, err)
	}
	defer CloseQuietly(f)
	return DetectFileType(f)
}

func IsImage(ft FileType) bool    { return ft.Category == FileCategoryImage }
func IsVideo(ft FileType) bool    { return ft.Category == FileCategoryVideo }
func IsAudio(ft FileType) bool    { return ft.Category == FileCategoryAudio }
func IsArchive(ft FileType) bool  { return ft.Category == FileCategoryArchive }
func IsDocument(ft FileType) bool { return ft.Category == FileCategoryDocument }

func matchWebP(data []byte) bool {
	return len(data) >= 12 && string(data[8:12]) == "WEBP"
}

func matchWAV(data []byte) bool {
	return len(data) >= 12 && string(data[8:12]) == "WAVE"
}

func matchMP4(data []byte) bool {
	if len(data) < 12 {
		return false
	}
	brand := string(data[8:12])
	return strings.HasPrefix(brand, "mp4") ||
		strings.HasPrefix(brand, "isom") ||
		strings.HasPrefix(brand, "iso2") ||
		strings.HasPrefix(brand, "avc1") ||
		strings.HasPrefix(brand, "M4")
}

func matchZipFamily(data []byte) bool {
	return !matchOOXML(data)
}

func matchOOXML(data []byte) bool {
	return matchOOXMLWord(data) || matchOOXMLSpreadsheet(data) || matchOOXMLPresentation(data)
}

func matchOOXMLWord(data []byte) bool {
	return bytes.Contains(data, []byte("word/"))
}

func matchOOXMLSpreadsheet(data []byte) bool {
	return bytes.Contains(data, []byte("xl/"))
}

func matchOOXMLPresentation(data []byte) bool {
	return bytes.Contains(data, []byte("ppt/"))
}
