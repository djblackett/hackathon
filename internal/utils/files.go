package utils

import (
	"io"
	"os"
	"path/filepath"
)

func RenameFile(baseDir, oldPath, newName string) error {
	dir := filepath.Dir(oldPath)
	ext := filepath.Ext(oldPath)
	newPath := filepath.Join(dir, newName+ext)
	return os.Rename(oldPath, newPath)
}

func CopyFile(baseDir, srcPath, destDir, newName string, flatten bool) error {
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

	// If flattening, use the new name directly in the destination directory.
	if flatten {

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

	// If not flattening, preserve the directory structure.
	relPath, err := filepath.Rel(baseDir, srcPath)
	if err != nil {
		return err
	}
	destDir = filepath.Join(destDir, filepath.Dir(relPath))
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}
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
