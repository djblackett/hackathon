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

func TestCompactEvidenceIncludesConstraintsWhenRoomAllows(t *testing.T) {
	info := extractors.ExtractedFileInfo{
		DetectedType: "csv",
		TextSamples: []extractors.TextSample{
			{Source: "csv-headers", Text: "customer_name customer_email account_status", Score: 0.85},
		},
	}

	got := CompactEvidence(info, 1000)

	if !strings.Contains(got, "constraints:") {
		t.Fatalf("compact evidence missing constraints: %q", got)
	}
	if !strings.Contains(got, "csv-headers") {
		t.Fatalf("compact evidence missing ranked source: %q", got)
	}
}

func TestCompactEvidenceSnapshot(t *testing.T) {
	info := extractors.ExtractedFileInfo{
		DetectedType: "json",
		Extension:    "bin",
		Metadata: map[string]string{
			"keys": "invoice_id customer_name total_due",
		},
		TextSamples: []extractors.TextSample{
			{Source: "content", Text: strings.Repeat("full body should not dominate ", 20), Score: 0.35},
			{Source: "json-keys", Text: "invoice_id customer_name total_due", Score: 0.75},
		},
		RawContent: strings.Repeat("sensitive raw content ", 50),
	}

	got := CompactEvidence(info, 1000)
	want := `detected_type: json
extension: bin
metadata:
- keys: invoice_id customer_name total_due
top_samples:
- source: json-keys
  text: invoice_id customer_name total_due
- source: content
  text: ` + strings.TrimSpace(strings.Repeat("full body should not dominate ", 12)) + `
constraints: lowercase kebab-case, no extension, 5-8 meaningful words, filename only
`
	if got != want {
		t.Fatalf("compact evidence mismatch\n got:\n%q\nwant:\n%q", got, want)
	}
	if strings.Contains(got, "sensitive raw content") {
		t.Fatal("compact evidence included raw content")
	}
}
