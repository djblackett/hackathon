package extractors

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type rtfExtractor struct{}

func (rtfExtractor) CanHandle(path string) bool {
	return strings.EqualFold(filepath.Ext(path), ".rtf")
}

func (rtfExtractor) CanHandleType(detectedType string) bool { return detectedType == "rtf" }

func (rtfExtractor) Extract(path string) (string, error) {
	info, err := rtfExtractor{}.ExtractInfo(path)
	if err != nil {
		return "", err
	}
	return info.RawContent, nil
}

func (rtfExtractor) ExtractInfo(path string) (ExtractedFileInfo, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return ExtractedFileInfo{}, err
	}

	text := rtfPlainText(string(b))
	info := NewExtractedFileInfo(path, "rtf", text)
	info.SuggestedExtension = "rtf"
	if text == "" {
		info.Warnings = append(info.Warnings, "no rtf text extracted")
		return info, nil
	}

	if title := firstRTFHeading(text); title != "" {
		info.TextSamples = append([]TextSample{{
			Source: "rtf-heading",
			Text:   title,
			Score:  0.86,
		}}, info.TextSamples...)
	}
	return info, nil
}

func rtfPlainText(s string) string {
	s = strings.ReplaceAll(s, `\par`, "\n")
	s = regexp.MustCompile(`\\'[0-9a-fA-F]{2}`).ReplaceAllString(s, " ")
	s = regexp.MustCompile(`\\[a-zA-Z]+-?\d* ?`).ReplaceAllString(s, " ")
	s = strings.NewReplacer("{", " ", "}", " ", "\\", " ").Replace(s)

	var lines []string
	for _, line := range strings.Split(s, "\n") {
		line = strings.Join(strings.Fields(line), " ")
		if line != "" {
			lines = append(lines, line)
		}
	}
	return strings.Join(lines, "\n")
}

func firstRTFHeading(text string) string {
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if idx := strings.IndexAny(line, ".!?"); idx > 0 {
			line = strings.TrimSpace(line[:idx])
		}
		return line
	}
	return ""
}

func init() { Register(rtfExtractor{}) }
