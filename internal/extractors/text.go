package extractors

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type textExtractor struct{}

func (textExtractor) CanHandle(path string) bool { return true } // fallback
func (textExtractor) CanHandleType(detectedType string) bool {
	switch detectedType {
	case "text", "markdown", "log", "cfg", "ini":
		return true
	default:
		return false
	}
}

func (textExtractor) Extract(path string) (string, error) {
	b, err := os.ReadFile(path)
	return string(b), err
}

func (textExtractor) ExtractInfo(path string) (ExtractedFileInfo, error) {
	content, err := textExtractor{}.Extract(path)
	if err != nil {
		return ExtractedFileInfo{}, err
	}

	detectedType := strings.TrimPrefix(strings.ToLower(filepath.Ext(path)), ".")
	if detectedType == "" {
		detectedType = "text"
	}
	info := NewExtractedFileInfo(path, detectedType, content)

	lines := nonEmptyTextLines(content)
	for _, line := range lines {
		line = strings.TrimSpace(line)
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

	if summary := firstSubstantiveTextLine(lines); summary != "" {
		info.TextSamples = append(info.TextSamples, TextSample{
			Source: "text-summary",
			Text:   summary,
			Score:  0.72,
		})
	} else if note := shortTextNote(content); note != "" {
		info.TextSamples = append(info.TextSamples, TextSample{
			Source: "short-text-note",
			Text:   note,
			Score:  0.58,
		})
	}

	return info, nil
}

func nonEmptyTextLines(content string) []string {
	var lines []string
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}

func firstSubstantiveTextLine(lines []string) string {
	for _, line := range lines {
		if strings.HasPrefix(line, "#") {
			continue
		}
		if isLowSignalTextLine(line) {
			continue
		}
		return firstSentence(line)
	}
	return ""
}

func firstSentence(line string) string {
	for i, r := range line {
		switch r {
		case '.', '!', '?':
			return cleanSummaryLine(line[:i])
		}
	}
	return cleanSummaryLine(line)
}

func cleanSummaryLine(line string) string {
	replacer := strings.NewReplacer(
		"for some reason", "",
		"For some reason", "",
	)
	return strings.TrimSpace(replacer.Replace(line))
}

var wordTokenPattern = regexp.MustCompile(`[A-Za-z0-9]+`)

func shortTextNote(content string) string {
	text := strings.Join(strings.Fields(content), " ")
	if text == "" || len(text) > 160 || looksRandomMediaName(text) {
		return ""
	}
	words := wordTokenPattern.FindAllString(text, -1)
	if len(words) < 3 {
		return ""
	}
	return text
}

func isLowSignalTextLine(line string) bool {
	trimmed := strings.TrimSpace(line)
	lower := strings.ToLower(strings.Trim(trimmed, " \t,.!?:;-"))
	words := wordTokenPattern.FindAllString(lower, -1)
	if len(words) == 0 {
		return true
	}

	switch lower {
	case "thanks", "thank you", "best", "regards", "sincerely":
		return true
	}
	if len(words) <= 3 {
		switch words[0] {
		case "hi", "hey", "hello", "dear":
			return true
		}
	}
	return len(trimmed) < 24 && len(words) < 5
}

func init() { Register(textExtractor{}) }
