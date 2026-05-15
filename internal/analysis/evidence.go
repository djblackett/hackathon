package analysis

import (
	"sort"
	"strings"
	"unicode"

	"github.com/djblackett/bootdev-hackathon/internal/extractors"
)

type RankedEvidence struct {
	Source string
	Text   string
	Score  float64
}

func RankEvidence(info extractors.ExtractedFileInfo) []RankedEvidence {
	ranked := make([]RankedEvidence, 0, len(info.TextSamples))

	for _, sample := range info.TextSamples {
		text := strings.Join(strings.Fields(sample.Text), " ")
		if text == "" {
			continue
		}

		score := sourceWeight(sample.Source)
		if sample.Score > score {
			score = sample.Score
		}
		score += qualityBoost(text)
		score -= evidencePenalty(sample.Source, text)
		score = clamp(score, 0, 0.98)

		ranked = append(ranked, RankedEvidence{
			Source: sample.Source,
			Text:   text,
			Score:  score,
		})
	}

	sort.SliceStable(ranked, func(i, j int) bool {
		if ranked[i].Score == ranked[j].Score {
			return sourceWeight(ranked[i].Source) > sourceWeight(ranked[j].Source)
		}
		return ranked[i].Score > ranked[j].Score
	})

	return ranked
}

func sourceWeight(source string) float64 {
	switch source {
	case "html-og-title":
		return 0.97
	case "html-title", "json-title-field", "office-title", "email-subject":
		return 0.95
	case "notebook-title":
		return 0.88
	case "musicxml-work-title":
		return 0.94
	case "notebook-heading":
		return 0.94
	case "musicxml-movement-title", "xml-title":
		return 0.9
	case "media-tags":
		return 0.9
	case "media-filename":
		return 0.7
	case "image-filename":
		return 0.75
	case "media-timestamp":
		return 0.74
	case "media-date":
		return 0.56
	case "media-properties":
		return 0.35
	case "office-subject":
		return 0.88
	case "office-heading":
		return 0.9
	case "office-slide-title":
		return 0.88
	case "office-headers":
		return 0.86
	case "office-sheet-name":
		return 0.78
	case "musicxml-creator", "xml-name":
		return 0.76
	case "markdown-heading", "html-h1":
		return 0.9
	case "image-exif-title", "image-exif-description", "image-exif-imagedescription", "image-exif-objectname":
		return 0.88
	case "csv-headers":
		return 0.86
	case "html-og-description", "html-meta-description":
		return 0.82
	case "json-keys":
		return 0.78
	case "json-structured":
		return 0.84
	case "xml-creator":
		return 0.68
	case "musicxml-parts":
		return 0.62
	case "xml-root":
		return 0.42
	case "pdf-first-text":
		return 0.72
	case "notebook-markdown":
		return 0.72
	case "office-first-paragraph":
		return 0.74
	case "office-text":
		return 0.68
	case "rtf-heading":
		return 0.86
	case "text-summary":
		return 0.72
	case "short-text-note":
		return 0.58
	case "first-meaningful-line":
		return 0.58
	case "image-properties":
		return 0.42
	case "content":
		return 0.35
	default:
		return 0.45
	}
}

func qualityBoost(text string) float64 {
	words := meaningfulWords(text, 12)
	switch {
	case len(words) >= 4:
		return 0.06
	case len(words) >= 2:
		return 0.03
	default:
		return 0
	}
}

func evidencePenalty(source, text string) float64 {
	lower := strings.ToLower(text)
	penalty := 0.0

	words := meaningfulWords(text, 12)
	if len(words) < 2 && !allowSingleWordEvidence(source, words) {
		penalty += 0.25
	}
	if containsBoilerplate(lower) {
		penalty += 0.5
	}
	if source != "media-timestamp" && looksRandom(text) {
		penalty += 0.55
	}
	if len(text) > 1200 {
		penalty += 0.1
	}

	return penalty
}

func containsBoilerplate(s string) bool {
	phrases := []string{
		"all rights reserved",
		"copyright",
		"lorem ipsum",
		"privacy policy",
		"terms and conditions",
		"unsubscribe",
	}
	for _, phrase := range phrases {
		if strings.Contains(s, phrase) {
			return true
		}
	}
	return false
}

func looksRandom(s string) bool {
	letters := 0
	vowels := 0
	digits := 0
	upper := 0
	lower := 0
	longestRun := 0
	currentRun := 0

	for _, r := range s {
		switch {
		case unicode.IsDigit(r):
			digits++
			currentRun++
		case unicode.IsLetter(r):
			letters++
			currentRun++
			if unicode.IsUpper(r) {
				upper++
			}
			if unicode.IsLower(r) {
				lower++
			}
			if strings.ContainsRune("aeiou", unicode.ToLower(r)) {
				vowels++
			}
		default:
			currentRun = 0
		}
		if currentRun > longestRun {
			longestRun = currentRun
		}
	}

	if letters >= 8 && float64(vowels)/float64(letters) < 0.18 {
		return true
	}
	if digits >= 4 && digits >= letters {
		return true
	}
	if hasRandomMixedToken(s) {
		return true
	}
	return longestRun >= 18
}

func hasRandomMixedToken(s string) bool {
	for _, token := range strings.FieldsFunc(s, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	}) {
		if len(token) < 12 {
			continue
		}
		hasUpper := false
		hasLower := false
		hasDigit := false
		for _, r := range token {
			hasUpper = hasUpper || unicode.IsUpper(r)
			hasLower = hasLower || unicode.IsLower(r)
			hasDigit = hasDigit || unicode.IsDigit(r)
		}
		if hasUpper && hasLower && hasDigit {
			return true
		}
	}
	return false
}

func clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
