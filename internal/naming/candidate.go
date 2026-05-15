package naming

import (
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/djblackett/bootdev-hackathon/internal/evidence"
)

type Confidence string

const (
	ConfidenceHigh   Confidence = "high"
	ConfidenceMedium Confidence = "medium"
	ConfidenceLow    Confidence = "low"
	ConfidenceNone   Confidence = "none"
)

type Suggestion struct {
	Filename   string
	Confidence Confidence
	Score      float64
	Reasons    []string
	Warnings   []string
}

type scoredCandidate struct {
	Base    string
	Score   float64
	Reasons []string
}

var datePattern = regexp.MustCompile(`\b(\d{4})[-:/](\d{2})[-:/](\d{2})(?:[ T_](\d{2})[-:](\d{2})(?:[-:](\d{2}))?)?\b`)
var wordPattern = regexp.MustCompile(`[a-zA-Z0-9]+`)

func Generate(ev evidence.FileEvidence, sequence int) Suggestion {
	ext := NormalizeExtension(ev.Extension)
	candidates := candidatesForEvidence(ev)
	sort.SliceStable(candidates, func(i, j int) bool {
		return candidates[i].Score > candidates[j].Score
	})

	if len(candidates) > 0 {
		best := candidates[0]
		filename := WithExtension(best.Base, ext)
		return Suggestion{
			Filename:   filename,
			Confidence: confidenceForScore(best.Score),
			Score:      best.Score,
			Reasons:    best.Reasons,
			Warnings:   append([]string(nil), ev.Warnings...),
		}
	}

	base := fallbackBase(ev, sequence)
	score := 0.18
	if ev.DetectedMIME != "" && ev.DetectedMIME != "application/octet-stream" {
		score = 0.28
	}
	return Suggestion{
		Filename:   WithExtension(base, ext),
		Confidence: confidenceForScore(score),
		Score:      score,
		Reasons:    []string{"only file type evidence was available"},
		Warnings:   append([]string(nil), ev.Warnings...),
	}
}

func candidatesForEvidence(ev evidence.FileEvidence) []scoredCandidate {
	out := []scoredCandidate{}
	date := strongestDate(ev)
	title, titleReason := strongestTitle(ev)
	textSignal := strongestTextSignal(ev)

	if date != "" && title != "" {
		out = append(out, scoredCandidate{
			Base:    date + "_" + title,
			Score:   0.88,
			Reasons: []string{"found useful date evidence", titleReason},
		})
	}
	if date != "" && textSignal != "" && textSignal != title {
		out = append(out, scoredCandidate{
			Base:    date + "_" + textSignal,
			Score:   0.78,
			Reasons: []string{"found useful date evidence", "extracted useful text signal"},
		})
	}
	if title != "" {
		out = append(out, scoredCandidate{
			Base:    title,
			Score:   0.66,
			Reasons: []string{titleReason},
		})
	}
	if textSignal != "" {
		out = append(out, scoredCandidate{
			Base:    textSignal,
			Score:   0.56,
			Reasons: []string{"extracted useful text signal"},
		})
	}
	if ev.Image != nil {
		if taken := firstNonEmpty(ev.Image.TakenAt, ev.Image.GPSDate, date); taken != "" {
			device := SanitizeBase(strings.TrimSpace(ev.Image.CameraMake + " " + ev.Image.CameraModel))
			if device == "unidentified" {
				device = "image"
			}
			out = append(out, scoredCandidate{
				Base:    normalizeDate(taken) + "_" + device,
				Score:   0.62,
				Reasons: []string{"found image timestamp or camera evidence"},
			})
		}
	}
	return out
}

func strongestTitle(ev evidence.FileEvidence) (string, string) {
	keys := []string{
		"title", "dc:title", "pdf:docinfo:title", "subject", "dc:subject",
		"description", "dc:description", "resourceName", "tika:title",
	}
	for _, key := range keys {
		for gotKey, value := range ev.Metadata {
			if !strings.EqualFold(gotKey, key) {
				continue
			}
			value = strings.TrimSpace(value)
			if value == "" || IsGenericTitle(value) {
				continue
			}
			if base := wordsBase(value, 8); base != "" {
				return base, "metadata provided a useful title"
			}
		}
	}
	return "", ""
}

func strongestTextSignal(ev evidence.FileEvidence) string {
	for _, signal := range ev.TextSignals {
		if base := wordsBase(signal, 8); base != "" && !IsGenericTitle(base) {
			return base
		}
	}
	if classified := classifyText(ev.TextPreview); classified != "" {
		return classified
	}
	if base := wordsBase(ev.TextPreview, 7); base != "" && !IsGenericTitle(base) {
		return base
	}
	return ""
}

func strongestDate(ev evidence.FileEvidence) string {
	keys := make([]string, 0, len(ev.Metadata))
	for key := range ev.Metadata {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		lower := strings.ToLower(key)
		if !strings.Contains(lower, "date") && !strings.Contains(lower, "created") && !strings.Contains(lower, "modified") {
			continue
		}
		if date := normalizeDate(ev.Metadata[key]); date != "" {
			return date
		}
	}
	if ev.Image != nil {
		if date := normalizeDate(firstNonEmpty(ev.Image.TakenAt, ev.Image.GPSDate)); date != "" {
			return date
		}
	}
	if date := normalizeDate(ev.TextPreview); date != "" {
		return date
	}
	return ""
}

func normalizeDate(value string) string {
	match := datePattern.FindStringSubmatch(value)
	if len(match) == 0 {
		return ""
	}
	if match[4] != "" && match[5] != "" {
		sec := match[6]
		if sec == "" {
			sec = "00"
		}
		return fmt.Sprintf("%s-%s-%s_%s-%s-%s", match[1], match[2], match[3], match[4], match[5], sec)
	}
	return fmt.Sprintf("%s-%s-%s", match[1], match[2], match[3])
}

func classifyText(text string) string {
	lower := strings.ToLower(text)
	switch {
	case strings.Contains(lower, "monthly statement") || strings.Contains(lower, "account summary"):
		return "monthly-statement"
	case strings.Contains(lower, "invoice"):
		return "invoice"
	case strings.Contains(lower, "receipt"):
		return "receipt"
	case strings.Contains(lower, "tax slip") || strings.Contains(lower, "t4"):
		return "possible-tax-slip"
	default:
		return ""
	}
}

func wordsBase(text string, limit int) string {
	words := wordPattern.FindAllString(text, -1)
	out := []string{}
	seen := map[string]bool{}
	for _, word := range words {
		if randomToken(word) {
			continue
		}
		word = strings.ToLower(strings.TrimSpace(word))
		if len(word) < 2 || stopWord(word) || seen[word] {
			continue
		}
		seen[word] = true
		out = append(out, word)
		if len(out) >= limit {
			break
		}
	}
	if len(out) < 2 {
		return ""
	}
	return SanitizeBase(strings.Join(out, "-"))
}

func randomToken(word string) bool {
	if len(word) < 10 {
		return false
	}
	letters := 0
	vowels := 0
	digits := 0
	hasUpper := false
	hasLower := false
	for _, r := range word {
		switch {
		case r >= '0' && r <= '9':
			digits++
		case r >= 'a' && r <= 'z':
			letters++
			hasLower = true
			if strings.ContainsRune("aeiou", r) {
				vowels++
			}
		case r >= 'A' && r <= 'Z':
			letters++
			hasUpper = true
			if strings.ContainsRune("AEIOU", r) {
				vowels++
			}
		}
	}
	if letters >= 8 && float64(vowels)/float64(letters) < 0.2 {
		return true
	}
	if digits >= 2 && hasUpper && hasLower {
		return true
	}
	return digits >= 4 && letters >= 4
}

func stopWord(word string) bool {
	switch word {
	case "a", "an", "and", "are", "as", "at", "be", "by", "for", "from", "in", "is", "it", "of", "on", "or", "the", "to", "with":
		return true
	case "file", "document", "untitled", "scan", "image", "photo":
		return true
	default:
		return false
	}
}

func fallbackBase(ev evidence.FileEvidence, sequence int) string {
	kind := kindFromEvidence(ev)
	token := sequenceToken(ev.Path, sequence)
	if kind == "unidentified" {
		return "unidentified_" + token
	}
	return "unknown-" + kind + "_" + token
}

func kindFromEvidence(ev evidence.FileEvidence) string {
	switch {
	case strings.Contains(ev.DetectedMIME, "pdf"):
		return "pdf"
	case strings.Contains(ev.DetectedMIME, "word"):
		return "docx"
	case strings.Contains(ev.DetectedMIME, "spreadsheet"):
		return "xlsx"
	case strings.Contains(ev.DetectedMIME, "presentation"):
		return "pptx"
	case strings.Contains(ev.DetectedMIME, "jpeg"):
		return "jpeg"
	case strings.Contains(ev.DetectedMIME, "png"):
		return "png"
	case strings.HasPrefix(ev.DetectedMIME, "text/"):
		return "text"
	case ev.Extension != "":
		return strings.TrimPrefix(ev.Extension, ".")
	default:
		return "unidentified"
	}
}

func sequenceToken(path string, sequence int) string {
	base := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	digits := regexp.MustCompile(`\d+`).FindAllString(base, -1)
	if len(digits) > 0 {
		last := digits[len(digits)-1]
		if len(last) >= 3 {
			return last
		}
	}
	if sequence < 1 {
		sequence = 1
	}
	return fmt.Sprintf("%06d", sequence)
}

func confidenceForScore(score float64) Confidence {
	switch {
	case score >= 0.85:
		return ConfidenceHigh
	case score >= 0.55:
		return ConfidenceMedium
	case score > 0:
		return ConfidenceLow
	default:
		return ConfidenceNone
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
