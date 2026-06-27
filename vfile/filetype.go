package vfile

import (
	"io"

	fileimpl "github.com/imajinyun/knifer-go/internal/file"
)

// FileCategory identifies a broad file family.
type FileCategory = fileimpl.FileCategory

const (
	// FileCategoryUnknown means the file type could not be identified.
	FileCategoryUnknown FileCategory = fileimpl.FileCategoryUnknown
	// FileCategoryImage identifies image files.
	FileCategoryImage FileCategory = fileimpl.FileCategoryImage
	// FileCategoryVideo identifies video files.
	FileCategoryVideo FileCategory = fileimpl.FileCategoryVideo
	// FileCategoryAudio identifies audio files.
	FileCategoryAudio FileCategory = fileimpl.FileCategoryAudio
	// FileCategoryArchive identifies archive or compressed files.
	FileCategoryArchive FileCategory = fileimpl.FileCategoryArchive
	// FileCategoryDocument identifies document files.
	FileCategoryDocument FileCategory = fileimpl.FileCategoryDocument
	// FileCategoryFont identifies font files.
	FileCategoryFont FileCategory = fileimpl.FileCategoryFont
	// FileCategoryExecutable identifies executable or object files.
	FileCategoryExecutable FileCategory = fileimpl.FileCategoryExecutable
)

// FileType contains magic-number detection metadata.
type FileType = fileimpl.FileType

var UnknownFileType = fileimpl.UnknownFileType

// DetectFileType detects a file type from leading magic-number bytes.
func DetectFileType(r io.Reader) (FileType, error) { return fileimpl.DetectFileType(r) }

// DetectFileTypeBytes detects a file type from bytes already in memory.
func DetectFileTypeBytes(data []byte) FileType { return fileimpl.DetectFileTypeBytes(data) }

// DetectFileTypeFromPath detects a file type by opening path and reading its header.
func DetectFileTypeFromPath(path string) (FileType, error) {
	return fileimpl.DetectFileTypeFromPath(path)
}

// IsImage reports whether ft is an image file type.
func IsImage(ft FileType) bool { return fileimpl.IsImage(ft) }

// IsVideo reports whether ft is a video file type.
func IsVideo(ft FileType) bool { return fileimpl.IsVideo(ft) }

// IsAudio reports whether ft is an audio file type.
func IsAudio(ft FileType) bool { return fileimpl.IsAudio(ft) }

// IsArchive reports whether ft is an archive file type.
func IsArchive(ft FileType) bool { return fileimpl.IsArchive(ft) }

// IsDocument reports whether ft is a document file type.
func IsDocument(ft FileType) bool { return fileimpl.IsDocument(ft) }
