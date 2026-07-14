// Package filesystem provides a single abstraction over all filesystem
// operations used by the application. No other package should call the
// os package directly for path/file manipulation; they should depend on
// the FileSystem interface instead, so behavior stays testable and
// centrally controlled.
package filesystem

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// FileSystem abstracts filesystem operations needed by builders and
// generators. A real implementation (OSFileSystem) wraps the os package;
// tests can supply an in-memory fake.
type FileSystem interface {
	// MkdirAll creates a directory and any necessary parents.
	MkdirAll(path string) error

	// WriteFile writes data to a file, creating it if necessary.
	WriteFile(path string, data []byte) error

	// CopyFile copies a single file from src to dst.
	CopyFile(src string, dst string) error

	// Exists reports whether a path exists (file or directory).
	Exists(path string) bool

	// IsDir reports whether a path exists and is a directory.
	IsDir(path string) bool

	// Join joins path elements using the OS-appropriate separator.
	Join(elem ...string) string

	// Abs returns an absolute representation of path.
	Abs(path string) (string, error)
}

// OSFileSystem is the production FileSystem implementation backed by the
// standard library.
type OSFileSystem struct{}

// NewOSFileSystem constructs an OSFileSystem.
func NewOSFileSystem() *OSFileSystem {
	return &OSFileSystem{}
}

func (fs *OSFileSystem) MkdirAll(path string) error {
	if err := os.MkdirAll(path, 0o755); err != nil {
		return fmt.Errorf("filesystem: failed to create directory %q: %w", path, err)
	}
	return nil
}

func (fs *OSFileSystem) WriteFile(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("filesystem: failed to create parent directory for %q: %w", path, err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("filesystem: failed to write file %q: %w", path, err)
	}
	return nil
}

func (fs *OSFileSystem) CopyFile(src string, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("filesystem: failed to open source file %q: %w", src, err)
	}
	defer source.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return fmt.Errorf("filesystem: failed to create parent directory for %q: %w", dst, err)
	}

	destination, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("filesystem: failed to create destination file %q: %w", dst, err)
	}
	defer destination.Close()

	if _, err := io.Copy(destination, source); err != nil {
		return fmt.Errorf("filesystem: failed to copy %q to %q: %w", src, dst, err)
	}

	return nil
}

func (fs *OSFileSystem) Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func (fs *OSFileSystem) IsDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func (fs *OSFileSystem) Join(elem ...string) string {
	return filepath.Join(elem...)
}

func (fs *OSFileSystem) Abs(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("filesystem: failed to resolve absolute path for %q: %w", path, err)
	}
	return abs, nil
}
