package vfile

import (
	"io"
	"io/fs"
	"time"
)

type fakeFacadeFileInfo struct {
	name string
}

func (f fakeFacadeFileInfo) Name() string       { return f.name }
func (f fakeFacadeFileInfo) Size() int64        { return 1 }
func (f fakeFacadeFileInfo) Mode() fs.FileMode  { return 0o644 }
func (f fakeFacadeFileInfo) ModTime() time.Time { return time.Unix(0, 0) }
func (f fakeFacadeFileInfo) IsDir() bool        { return false }
func (f fakeFacadeFileInfo) Sys() any           { return nil }

type fakeDirFacadeFileInfo struct{ name string }

func (f fakeDirFacadeFileInfo) Name() string       { return f.name }
func (f fakeDirFacadeFileInfo) Size() int64        { return 0 }
func (f fakeDirFacadeFileInfo) Mode() fs.FileMode  { return fs.ModeDir | 0o755 }
func (f fakeDirFacadeFileInfo) ModTime() time.Time { return time.Unix(0, 0) }
func (f fakeDirFacadeFileInfo) IsDir() bool        { return true }
func (f fakeDirFacadeFileInfo) Sys() any           { return nil }

type fakeSizedFacadeFileInfo struct {
	name string
	size int64
}

func (f fakeSizedFacadeFileInfo) Name() string       { return f.name }
func (f fakeSizedFacadeFileInfo) Size() int64        { return f.size }
func (f fakeSizedFacadeFileInfo) Mode() fs.FileMode  { return 0o644 }
func (f fakeSizedFacadeFileInfo) ModTime() time.Time { return time.Unix(0, 0) }
func (f fakeSizedFacadeFileInfo) IsDir() bool        { return false }
func (f fakeSizedFacadeFileInfo) Sys() any           { return nil }

type nopFacadeWriteCloser struct{ io.Writer }

func (w nopFacadeWriteCloser) Close() error { return nil }
