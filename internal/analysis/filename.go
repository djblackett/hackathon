package analysis

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/djblackett/bootdev-hackathon/internal/extractors"
	"github.com/djblackett/bootdev-hackathon/internal/utils"
)

type FilenameSuggestion struct {
	Filename   string
	Confidence float64
	Method     string
	Evidence   []string
	Reason     string
}

const minimumLocalEvidenceScore = 0.4

var wordPattern = regexp.MustCompile(`[a-zA-Z0-9]+`)

var stopWords = map[string]struct{}{
	"a": {}, "an": {}, "and": {}, "are": {}, "as": {}, "at": {}, "be": {}, "by": {},
	"audio": {},
	"for":   {}, "from": {}, "in": {}, "is": {}, "it": {}, "of": {}, "on": {}, "or": {},
	"the": {}, "this": {}, "to": {}, "video": {}, "with": {},
}

var genericWords = map[string]struct{}{
	"data": {}, "document": {}, "file": {}, "final": {}, "id": {}, "index": {}, "new": {},
	"note": {}, "notes": {}, "scan": {}, "text": {}, "untitled": {},
}

func GenerateFilename(info extractors.ExtractedFileInfo) FilenameSuggestion {
	samples := RankEvidence(info)

	for _, sample := range samples {
		if sample.Score < minimumLocalEvidenceScore {
			continue
		}
		words := meaningfulWords(sample.Text, 8)
		if len(words) < 2 && !allowSingleWordEvidence(sample.Source, words) {
			continue
		}

		filename := utils.Sanitize(strings.Join(words, "-"))
		if filename == "" {
			continue
		}

		confidence := sample.Score
		if metadataSource(sample.Source) {
			confidence += 0.05
		}
		if len(words) >= 3 {
			confidence += 0.05
		}
		if confidence > 0.95 {
			confidence = 0.95
		}

		return FilenameSuggestion{
			Filename:   filename,
			Confidence: confidence,
			Method:     "metadata",
			Evidence:   []string{sample.Source},
			Reason:     "generated from local file evidence",
		}
	}

	return FilenameSuggestion{
		Filename:   unidentifiedFallback(info),
		Confidence: 0.1,
		Method:     "metadata",
		Reason:     "no strong local evidence found",
	}
}

func unidentifiedFallback(info extractors.ExtractedFileInfo) string {
	if info.DetectedType != "media" {
		return "unidentified-content"
	}
	switch strings.ToLower(info.SuggestedExtension) {
	case "mp3", "m4a", "wav", "flac":
		return "unidentified-audio"
	case "mp4", "mov", "mkv", "avi":
		return "unidentified-video"
	default:
		return "unidentified-media"
	}
}

func allowSingleWordEvidence(source string, words []string) bool {
	return source == "media-filename" && len(words) == 1 && len(words[0]) >= 4
}

func CompactEvidence(info extractors.ExtractedFileInfo, maxChars int) string {
	if maxChars <= 0 {
		maxChars = 2000
	}

	samples := RankEvidence(info)
	if len(samples) > 5 {
		samples = samples[:5]
	}

	var b strings.Builder
	fmt.Fprintf(&b, "detected_type: %s\n", info.DetectedType)
	if info.Extension != "" {
		fmt.Fprintf(&b, "extension: %s\n", info.Extension)
	}
	if len(info.Metadata) > 0 {
		keys := make([]string, 0, len(info.Metadata))
		for key := range info.Metadata {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		b.WriteString("metadata:\n")
		for _, key := range keys {
			fmt.Fprintf(&b, "- %s: %s\n", key, trimForEvidence(info.Metadata[key], 240))
		}
	}
	if len(samples) > 0 {
		b.WriteString("top_samples:\n")
		for _, sample := range samples {
			fmt.Fprintf(&b, "- source: %s\n  text: %s\n", sample.Source, trimForEvidence(sample.Text, 360))
		}
	}
	b.WriteString("constraints: lowercase kebab-case, no extension, 5-8 meaningful words, filename only\n")

	out := b.String()
	if len(out) > maxChars {
		out = out[:maxChars]
	}
	return out
}

func meaningfulWords(text string, limit int) []string {
	normalized := splitCompoundWords(text)
	raw := wordPattern.FindAllString(normalized, -1)
	words := make([]string, 0, limit)
	seen := map[string]struct{}{}

	for _, word := range raw {
		word = strings.ToLower(strings.TrimSpace(word))
		if len(word) < 2 {
			continue
		}
		if _, ok := stopWords[word]; ok {
			continue
		}
		if _, ok := genericWords[word]; ok {
			continue
		}
		if _, ok := seen[word]; ok {
			continue
		}
		seen[word] = struct{}{}
		words = append(words, word)
		if len(words) >= limit {
			break
		}
	}

	return words
}

func splitCompoundWords(s string) string {
	replacer := strings.NewReplacer("_", " ", "-", " ", ".", " ", "/", " ")
	return replacer.Replace(s)
}

func metadataSource(source string) bool {
	return strings.Contains(source, "title") ||
		strings.Contains(source, "subject") ||
		strings.Contains(source, "heading") ||
		strings.Contains(source, "headers") ||
		strings.Contains(source, "keys") ||
		strings.Contains(source, "tags")
}

func trimForEvidence(s string, max int) string {
	s = strings.Join(strings.Fields(s), " ")
	if len(s) <= max {
		return s
	}
	return strings.TrimSpace(s[:max])
}
