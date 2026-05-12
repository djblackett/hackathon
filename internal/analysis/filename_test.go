package analysis

import (
	"strings"
	"testing"

	"github.com/djblackett/bootdev-hackathon/internal/extractors"
)

func TestGenerateFilenamePrefersHighScoringEvidence(t *testing.T) {
	info := extractors.ExtractedFileInfo{
		DetectedType: "markdown",
		TextSamples: []extractors.TextSample{
			{Source: "content", Text: "a bunch of generic filler text", Score: 0.2},
			{Source: "markdown-heading", Text: "Quarterly Revenue Review Q4 2025", Score: 0.9},
		},
	}

	got := GenerateFilename(info)

	if got.Filename != "quarterly-revenue-review-q4-2025" {
		t.Fatalf("filename = %q, want %q", got.Filename, "quarterly-revenue-review-q4-2025")
	}
	if got.Confidence < 0.9 {
		t.Fatalf("confidence = %.2f, want at least 0.90", got.Confidence)
	}
}

func TestCompactEvidenceLimitsOutput(t *testing.T) {
	info := extractors.ExtractedFileInfo{
		DetectedType: "text",
		Metadata: map[string]string{
			"title": strings.Repeat("important ", 20),
		},
		TextSamples: []extractors.TextSample{
			{Source: "content", Text: strings.Repeat("sample ", 100), Score: 0.5},
		},
	}

	got := CompactEvidence(info, 120)

	if len(got) > 120 {
		t.Fatalf("compact evidence length = %d, want <= 120", len(got))
	}
	if !strings.Contains(got, "detected_type: text") {
		t.Fatalf("compact evidence missing detected type: %q", got)
	}
}
