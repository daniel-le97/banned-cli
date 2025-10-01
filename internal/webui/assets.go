package webui

import (
	"embed"
	"io/fs"
)

//go:embed web/*
var webFS embed.FS

// WebAssets holds the embedded web UI files
type WebAssets struct {
	fs embed.FS
}

// NewWebAssets creates a new WebAssets instance with the provided embed.FS
func NewWebAssets(embedFS embed.FS) *WebAssets {
	return &WebAssets{fs: embedFS}
}

// GetWebAssets returns the default WebAssets instance with embedded files
func GetWebAssets() *WebAssets {
	return &WebAssets{fs: webFS}
}

// FS returns the embedded filesystem
func (w *WebAssets) FS() embed.FS {
	return w.fs
}

// Sub returns a sub-filesystem rooted at dir
func (w *WebAssets) Sub(dir string) (fs.FS, error) {
	return fs.Sub(w.fs, dir)
}

// ReadFile reads and returns the content of the named file
func (w *WebAssets) ReadFile(name string) ([]byte, error) {
	return w.fs.ReadFile(name)
}

// Open opens the named file for reading
func (w *WebAssets) Open(name string) (fs.File, error) {
	return w.fs.Open(name)
}
