package extractors

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
)

type Extractor interface {
	CanHandle(path string) bool
	Extract(path string) (string, error)
}

var registered []Extractor

func Register(e Extractor) { registered = append(registered, e) }

// Walk over dir; for each supported file call fn(path, content)
func Walk(dir string, types map[string]struct{}, fn func(string, string) error) error {
	fmt.Printf("Registered extractors: %d\n", len(registered)) // Add this debug line
	return filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}

		ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(path)), ".")
		// fmt.Printf("Processing file: %s, extension: %s\n", path, ext) // Add this debug line
		if _, ok := types[ext]; !ok {
			return nil // skip unsupported filetypes
		}
		for _, ex := range registered {
			if ex.CanHandle(path) {
				content, err := ex.Extract(path)
				if err != nil {
					return err
				}
				return fn(path, content)
			}
		}
		return nil
	})
}
