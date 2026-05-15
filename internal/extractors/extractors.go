package extractors

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/djblackett/bootdev-hackathon/internal/filetype"
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

// WalkInfo walks over dir and returns structured extraction output for each
// supported file. Existing extractors can opt into richer metadata by
// implementing InfoExtractor; otherwise the plain extracted content is wrapped.
func WalkInfo(dir string, types map[string]struct{}, fn func(ExtractedFileInfo) error) error {
	fmt.Printf("Registered extractors: %d\n", len(registered))
	return filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}

		detection := filetype.Detect(path)
		if !allowedType(types, detection.Extension, detection.Type, detection.Subtype) {
			return nil
		}

		for _, ex := range registered {
			if !extractorCanHandle(ex, path, detection.Type, detection.Subtype) {
				continue
			}

			if infoEx, ok := ex.(InfoExtractor); ok {
				info, err := infoEx.ExtractInfo(path)
				if err != nil {
					return err
				}
				applyDetection(&info, detection)
				return fn(info)
			}

			content, err := ex.Extract(path)
			if err != nil {
				return err
			}
			info := NewExtractedFileInfo(path, detection.Type, content)
			applyDetection(&info, detection)
			return fn(info)
		}
		return nil
	})
}

func ExtractInfoForPath(path string) (ExtractedFileInfo, error) {
	detection := filetype.Detect(path)
	for _, ex := range registered {
		if !extractorCanHandle(ex, path, detection.Type, detection.Subtype) {
			continue
		}
		if infoEx, ok := ex.(InfoExtractor); ok {
			info, err := infoEx.ExtractInfo(path)
			if err != nil {
				return ExtractedFileInfo{}, err
			}
			applyDetection(&info, detection)
			return info, nil
		}
		content, err := ex.Extract(path)
		if err != nil {
			return ExtractedFileInfo{}, err
		}
		info := NewExtractedFileInfo(path, detection.Type, content)
		applyDetection(&info, detection)
		return info, nil
	}
	return ExtractedFileInfo{}, fmt.Errorf("no extractor for %s", path)
}

func allowedType(types map[string]struct{}, values ...string) bool {
	for _, value := range values {
		if value == "" {
			continue
		}
		if _, ok := types[value]; ok {
			return true
		}
	}
	return false
}

func extractorCanHandle(ex Extractor, path, detectedType, detectedSubtype string) bool {
	if typed, ok := ex.(TypeExtractor); ok {
		return typed.CanHandleType(detectedType) || typed.CanHandleType(detectedSubtype)
	}
	return ex.CanHandle(path)
}

func applyDetection(info *ExtractedFileInfo, detection filetype.Detection) {
	info.DetectedType = detection.Type
	if info.Extension == "" {
		info.Extension = detection.Extension
	}
	if info.SuggestedExtension == "" || info.SuggestedExtension == info.Extension {
		info.SuggestedExtension = detection.CanonicalExtension
	}
	if info.Metadata == nil {
		info.Metadata = map[string]string{}
	}
	if detection.Subtype != "" {
		info.Metadata["detected_subtype"] = detection.Subtype
	}
	if detection.Warning != "" {
		info.Warnings = append(info.Warnings, detection.Warning)
	}
}
