package utils

import (
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var invalid = regexp.MustCompile(`[^a-z0-9\-_]+`)

func Sanitize(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, " ", "-")
	s = invalid.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-_")
	if len(s) > 60 {
		s = s[:60]
	}

	banned := []string{"document", "file", "note", "notes", "txt"}
	for _, b := range banned {
		s = strings.TrimSuffix(s, "-"+b)
		s = strings.TrimSuffix(s, "_"+b)
	}
	return s
}

func RenameFile(oldPath, newName string) error {
	dir := filepath.Dir(oldPath)
	ext := filepath.Ext(oldPath)
	newPath := filepath.Join(dir, newName+ext)
	return os.Rename(oldPath, newPath)
}

func CopyFile(srcPath, destDir, newName string) error {
	// Create destination directory if it doesn't exist
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}

	// Open source file
	src, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer src.Close()

	// Create destination file
	ext := filepath.Ext(srcPath)
	destPath := filepath.Join(destDir, newName+ext)
	dest, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer dest.Close()

	// Copy file contents
	_, err = io.Copy(dest, src)
	return err
}
