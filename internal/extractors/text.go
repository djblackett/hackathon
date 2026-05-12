package extractors

import (
	"os"
	"path/filepath"
	"strings"
)

type textExtractor struct{}

func (textExtractor) CanHandle(path string) bool { return true } // fallback
func (textExtractor) Extract(path string) (string, error) {
	b, err := os.ReadFile(path)
	return string(b), err
}

func (textExtractor) ExtractInfo(path string) (ExtractedFileInfo, error) {
	content, err := textExtractor{}.Extract(path)
	if err != nil {
		return ExtractedFileInfo{}, err
	}

	info := NewExtractedFileInfo(path, content)
	info.DetectedType = strings.TrimPrefix(strings.ToLower(filepath.Ext(path)), ".")
	if info.DetectedType == "" {
		info.DetectedType = "text"
	}

	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		source := "first-meaningful-line"
		score := 0.55
		if strings.HasPrefix(line, "#") {
			source = "markdown-heading"
			score = 0.9
			line = strings.TrimSpace(strings.TrimLeft(line, "#"))
		}
		info.TextSamples = append([]TextSample{{
			Source: source,
			Text:   line,
			Score:  score,
		}}, info.TextSamples...)
		break
	}

	return info, nil
}

func init() { Register(textExtractor{}) }
