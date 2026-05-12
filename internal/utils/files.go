package utils

import (
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func RenameFile(baseDir, oldPath, newName string) error {
	return RenameFileWithExtension(baseDir, oldPath, newName, filepath.Ext(oldPath))
}

func RenameFileWithExtension(baseDir, oldPath, newName, ext string) error {
	dir := filepath.Dir(oldPath)
	newPath := filepath.Join(dir, newName+ext)
	return os.Rename(oldPath, newPath)
}

func CopyFile(baseDir, srcPath, destDir, newName string, flatten bool) error {
	return CopyFileWithExtension(baseDir, srcPath, destDir, newName, filepath.Ext(srcPath), flatten)
}

func CopyFileWithExtension(baseDir, srcPath, destDir, newName, ext string, flatten bool) error {
	destPath, err := DestinationPath(baseDir, srcPath, destDir, newName, ext, flatten)
	if err != nil {
		return err
	}
	return CopyFileToPath(srcPath, destPath)
}

func DestinationPath(baseDir, srcPath, destDir, newName, ext string, flatten bool) (string, error) {
	// Create destination directory if it doesn't exist
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return "", err
	}
	ext = normalizeExt(ext)

	// If flattening, use the new name directly in the destination directory.
	if flatten {
		return filepath.Join(destDir, newName+ext), nil
	}

	// If not flattening, preserve the directory structure.
	relPath, err := filepath.Rel(baseDir, srcPath)
	if err != nil {
		return "", err
	}
	destDir = filepath.Join(destDir, filepath.Dir(relPath))
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return "", err
	}
	return filepath.Join(destDir, newName+ext), nil
}

func CopyFileToPath(srcPath, destPath string) error {
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return err
	}

	src, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer src.Close()

	dest, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer dest.Close()

	// Copy file contents
	_, err = io.Copy(dest, src)
	return err
}

func UniquePath(path string, reserved map[string]struct{}) string {
	if _, ok := reserved[path]; !ok && !fileExists(path) {
		reserved[path] = struct{}{}
		return path
	}

	ext := filepath.Ext(path)
	base := strings.TrimSuffix(path, ext)
	for i := 2; ; i++ {
		candidate := base + "-" + strconv.Itoa(i) + ext
		if _, ok := reserved[candidate]; ok {
			continue
		}
		if fileExists(candidate) {
			continue
		}
		reserved[candidate] = struct{}{}
		return candidate
	}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func normalizeExt(ext string) string {
	if ext == "" || strings.HasPrefix(ext, ".") {
		return ext
	}
	return "." + ext
}
