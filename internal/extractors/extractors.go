package extractors

import (
	"context"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/djblackett/bootdev-hackathon/internal/filetype"
	"github.com/djblackett/bootdev-hackathon/internal/tika"
)

type Extractor interface {
	CanHandle(path string) bool
	Extract(path string) (string, error)
}

var registered []Extractor
var tikaClient *tika.Client

func Register(e Extractor) { registered = append(registered, e) }

func ConfigureTika(baseURL string) error {
	if strings.TrimSpace(baseURL) == "" {
		tikaClient = nil
		return nil
	}
	client, err := tika.NewClient(baseURL)
	if err != nil {
		return err
	}
	tikaClient = client
	return nil
}

// Walk over dir; for each supported file call fn(path, content)
func Walk(dir string, types map[string]struct{}, fn func(string, string) error) error {
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
				if shouldTryTikaFallback(info) {
					info = mergeTikaFallback(info)
				}
				return fn(info)
			}

			content, err := ex.Extract(path)
			if err != nil {
				return err
			}
			info := NewExtractedFileInfo(path, detection.Type, content)
			applyDetection(&info, detection)
			if shouldTryTikaFallback(info) {
				info = mergeTikaFallback(info)
			}
			return fn(info)
		}
		if tikaClient != nil {
			info := NewExtractedFileInfo(path, detection.Type, "")
			applyDetection(&info, detection)
			return fn(mergeTikaFallback(info))
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
			if shouldTryTikaFallback(info) {
				info = mergeTikaFallback(info)
			}
			return info, nil
		}
		content, err := ex.Extract(path)
		if err != nil {
			return ExtractedFileInfo{}, err
		}
		info := NewExtractedFileInfo(path, detection.Type, content)
		applyDetection(&info, detection)
		if shouldTryTikaFallback(info) {
			info = mergeTikaFallback(info)
		}
		return info, nil
	}
	if tikaClient != nil {
		info := NewExtractedFileInfo(path, detection.Type, "")
		applyDetection(&info, detection)
		return mergeTikaFallback(info), nil
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

func shouldTryTikaFallback(info ExtractedFileInfo) bool {
	return tikaClient != nil && strings.TrimSpace(info.RawContent) == "" && len(info.TextSamples) == 0
}

func mergeTikaFallback(info ExtractedFileInfo) ExtractedFileInfo {
	if tikaClient == nil {
		return info
	}

	extracted, err := tikaClient.ExtractFile(context.Background(), info.Path)
	if err != nil {
		info.Warnings = append(info.Warnings, "tika extraction failed: "+err.Error())
		return info
	}
	info.Warnings = append(info.Warnings, extracted.Warnings...)

	text := strings.TrimSpace(extracted.Text)
	if info.RawContent == "" {
		info.RawContent = text
	}
	for key, value := range extracted.Metadata {
		if strings.TrimSpace(value) == "" {
			continue
		}
		info.Metadata["tika:"+key] = value
	}
	appendTikaMetadataSample := func(keys []string, source string, score float64) {
		for _, key := range keys {
			if value := firstMetadataValue(extracted.Metadata, key); value != "" {
				info.TextSamples = append(info.TextSamples, TextSample{Source: source, Text: value, Score: score})
				return
			}
		}
	}

	appendTikaMetadataSample([]string{"title", "dc:title", "pdf:docinfo:title", "resourceName"}, "tika-title", 0.86)
	appendTikaMetadataSample([]string{"subject", "dc:subject", "description", "dc:description"}, "tika-subject", 0.78)
	appendTikaMetadataSample([]string{"creator", "dc:creator", "author", "meta:author"}, "tika-author", 0.62)

	if text != "" {
		if first := firstMeaningfulLine(text); first != "" {
			info.TextSamples = append(info.TextSamples, TextSample{
				Source: "tika-first-text",
				Text:   first,
				Score:  0.62,
			})
		}
	}
	if len(info.TextSamples) == 0 {
		info.Warnings = append(info.Warnings, "tika returned no text or metadata")
	}
	return info
}

func firstMetadataValue(metadata map[string]string, keys ...string) string {
	for _, want := range keys {
		for key, value := range metadata {
			if strings.EqualFold(key, want) && strings.TrimSpace(value) != "" {
				return strings.TrimSpace(value)
			}
		}
	}
	return ""
}

func firstMeaningfulLine(text string) string {
	for _, line := range strings.Split(text, "\n") {
		line = strings.Join(strings.Fields(line), " ")
		if line != "" {
			return line
		}
	}
	return ""
}
