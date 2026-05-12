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
		score -= evidencePenalty(text)
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
	case "html-title", "json-title-field":
		return 0.95
	case "markdown-heading", "html-h1":
		return 0.9
	case "csv-headers":
		return 0.86
	case "html-meta-description":
		return 0.82
	case "json-keys":
		return 0.78
	case "pdf-first-text":
		return 0.72
	case "first-meaningful-line":
		return 0.58
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

func evidencePenalty(text string) float64 {
	lower := strings.ToLower(text)
	penalty := 0.0

	if len(meaningfulWords(text, 12)) < 2 {
		penalty += 0.25
	}
	if containsBoilerplate(lower) {
		penalty += 0.5
	}
	if looksRandom(text) {
		penalty += 0.25
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
	longestRun := 0
	currentRun := 0

	for _, r := range strings.ToLower(s) {
		switch {
		case unicode.IsDigit(r):
			digits++
			currentRun++
		case unicode.IsLetter(r):
			letters++
			currentRun++
			if strings.ContainsRune("aeiou", r) {
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
	return longestRun >= 18
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
