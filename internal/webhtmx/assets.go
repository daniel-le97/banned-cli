package webhtmx

import (
	"embed"
	"io/fs"
)

//go:embed web-htmx/*
var htmxWebFS embed.FS

// HTMXWebAssets holds the embedded HTMX web UI files
type HTMXWebAssets struct {
	fs embed.FS
}

// NewHTMXWebAssets creates a new HTMXWebAssets instance with the provided embed.FS
func NewHTMXWebAssets(embedFS embed.FS) *HTMXWebAssets {
	return &HTMXWebAssets{fs: embedFS}
}

// GetHTMXWebAssets returns the embedded HTMX web UI filesystem
func GetHTMXWebAssets() *HTMXWebAssets {
	return NewHTMXWebAssets(htmxWebFS)
}

// FS returns the embedded filesystem
func (w *HTMXWebAssets) FS() embed.FS {
	return w.fs
}

// SubFS returns a sub-filesystem rooted at the specified directory
func (w *HTMXWebAssets) SubFS(dir string) (fs.FS, error) {
	return fs.Sub(w.fs, dir)
}
