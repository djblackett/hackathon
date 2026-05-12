package extractors

import (
	"path/filepath"
	"strings"
)

type TextSample struct {
	Source string
	Text   string
	Score  float64
}

type ExtractedFileInfo struct {
	Path               string
	Extension          string
	SuggestedExtension string
	DetectedType       string
	Metadata           map[string]string
	TextSamples        []TextSample
	RawContent         string
	Warnings           []string
}

type InfoExtractor interface {
	Extractor
	ExtractInfo(path string) (ExtractedFileInfo, error)
}

type TypeExtractor interface {
	CanHandleType(detectedType string) bool
}

func NewExtractedFileInfo(path, detectedType, content string) ExtractedFileInfo {
	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(path)), ".")
	if detectedType == "" {
		detectedType = ext
	}
	info := ExtractedFileInfo{
		Path:               path,
		Extension:          ext,
		SuggestedExtension: ext,
		DetectedType:       detectedType,
		Metadata:           map[string]string{},
		RawContent:         content,
	}

	if strings.TrimSpace(content) != "" {
		info.TextSamples = append(info.TextSamples, TextSample{
			Source: "content",
			Text:   content,
			Score:  0.35,
		})
	}

	return info
}
