package analysis

import (
	"strings"
	"testing"

	"github.com/djblackett/bootdev-hackathon/internal/extractors"
)

func TestRankEvidenceHeadingOutranksBodyText(t *testing.T) {
	info := extractors.ExtractedFileInfo{
		TextSamples: []extractors.TextSample{
			{Source: "content", Text: "Revenue was discussed in the meeting, along with staffing and logistics.", Score: 0.35},
			{Source: "markdown-heading", Text: "Incident Response Runbook", Score: 0.7},
		},
	}

	got := RankEvidence(info)

	if got[0].Source != "markdown-heading" {
		t.Fatalf("top source = %q, want markdown-heading", got[0].Source)
	}
}

func TestRankEvidencePenalizesBoilerplate(t *testing.T) {
	info := extractors.ExtractedFileInfo{
		TextSamples: []extractors.TextSample{
			{Source: "first-meaningful-line", Text: "Copyright 2025 All rights reserved", Score: 0.8},
			{Source: "content", Text: "Customer onboarding checklist account setup billing support", Score: 0.35},
		},
	}

	got := RankEvidence(info)

	if got[0].Text == "Copyright 2025 All rights reserved" {
		t.Fatalf("boilerplate was ranked first: %#v", got)
	}
}

func TestGenerateFilenameFromCSVHeaders(t *testing.T) {
	info := extractors.ExtractedFileInfo{
		DetectedType: "csv",
		TextSamples: []extractors.TextSample{
			{Source: "csv-headers", Text: "customer_name customer_email contact_phone account_status", Score: 0.85},
		},
	}

	got := GenerateFilename(info)

	if got.Filename != "customer-name-email-contact-phone-account-status" {
		t.Fatalf("filename = %q", got.Filename)
	}
}

func TestGenerateFilenameFromJSONKeys(t *testing.T) {
	info := extractors.ExtractedFileInfo{
		DetectedType: "json",
		TextSamples: []extractors.TextSample{
			{Source: "json-keys", Text: "invoice_id export_date customer_name total_due records", Score: 0.75},
		},
	}

	got := GenerateFilename(info)

	if got.Filename != "invoice-export-date-customer-name-total-due-records" {
		t.Fatalf("filename = %q", got.Filename)
	}
}

func TestGenerateFilenameFromJSONStructuredEvidence(t *testing.T) {
	info := extractors.ExtractedFileInfo{
		DetectedType: "json",
		TextSamples: []extractors.TextSample{
			{Source: "content", Text: "Which one is correct team name in NBA?", Score: 0.35},
			{Source: "json-structured", Text: "quiz sport q1 question Which one is correct team name in NBA?", Score: 0.84},
		},
	}

	got := GenerateFilename(info)

	if got.Filename != "quiz-sport-q1-question-which-one-correct-team" {
		t.Fatalf("filename = %q", got.Filename)
	}
	if got.Confidence < 0.75 {
		t.Fatalf("confidence = %.2f, want high confidence", got.Confidence)
	}
	if got.Evidence[0] != "json-structured" {
		t.Fatalf("evidence = %+v, want json-structured first", got.Evidence)
	}
}

func TestCompactEvidenceUsesRankedSamples(t *testing.T) {
	info := extractors.ExtractedFileInfo{
		DetectedType: "html",
		TextSamples: []extractors.TextSample{
			{Source: "content", Text: strings.Repeat("footer copyright ", 20), Score: 0.6},
			{Source: "html-title", Text: "Network Migration Plan", Score: 0.7},
		},
	}

	got := CompactEvidence(info, 500)

	titleIdx := strings.Index(got, "html-title")
	contentIdx := strings.Index(got, "content")
	if titleIdx == -1 {
		t.Fatalf("compact evidence missing html-title: %q", got)
	}
	if contentIdx != -1 && contentIdx < titleIdx {
		t.Fatalf("content ranked before title: %q", got)
	}
}

func TestGenerateFilenameFromHTMLPrefersOpenGraphTitle(t *testing.T) {
	info := extractors.ExtractedFileInfo{
		DetectedType: "html",
		TextSamples: []extractors.TextSample{
			{Source: "content", Text: "Footer privacy policy copyright and repeated navigation text", Score: 0.35},
			{Source: "html-title", Text: "Generic Export Page", Score: 0.95},
			{Source: "html-og-title", Text: "Network Migration Cutover Plan", Score: 0.96},
		},
	}

	got := GenerateFilename(info)

	if got.Filename != "network-migration-cutover-plan" {
		t.Fatalf("filename = %q", got.Filename)
	}
	if got.Evidence[0] != "html-og-title" {
		t.Fatalf("evidence = %+v, want html-og-title first", got.Evidence)
	}
}

func TestGenerateFilenameFromHTMLHeadingFallback(t *testing.T) {
	info := extractors.ExtractedFileInfo{
		DetectedType: "html",
		TextSamples: []extractors.TextSample{
			{Source: "content", Text: "Footer privacy policy copyright and repeated navigation text", Score: 0.35},
			{Source: "html-h1", Text: "Incident Response Runbook", Score: 0.9},
		},
	}

	got := GenerateFilename(info)

	if got.Filename != "incident-response-runbook" {
		t.Fatalf("filename = %q", got.Filename)
	}
}

func TestGenerateFilenameFromMusicXMLWorkTitle(t *testing.T) {
	info := extractors.ExtractedFileInfo{
		DetectedType: "xml",
		Metadata:     map[string]string{"detected_subtype": "musicxml"},
		TextSamples: []extractors.TextSample{
			{Source: "musicxml-parts", Text: "Flute Clarinet Alto Sax", Score: 0.62},
			{Source: "musicxml-work-title", Text: "You Are My Sunshine", Score: 0.95},
		},
	}

	got := GenerateFilename(info)

	if got.Filename != "you-my-sunshine" {
		t.Fatalf("filename = %q", got.Filename)
	}
	if got.Evidence[0] != "musicxml-work-title" {
		t.Fatalf("evidence = %+v, want musicxml-work-title first", got.Evidence)
	}
}

func TestGenerateFilenameFromGenericXMLTitle(t *testing.T) {
	info := extractors.ExtractedFileInfo{
		DetectedType: "xml",
		TextSamples: []extractors.TextSample{
			{Source: "xml-root", Text: "archive-record", Score: 0.42},
			{Source: "xml-title", Text: "Quarterly Safety Inspection Log", Score: 0.9},
		},
	}

	got := GenerateFilename(info)

	if got.Filename != "quarterly-safety-inspection-log" {
		t.Fatalf("filename = %q", got.Filename)
	}
}

func TestGenerateFilenameRejectsRandomText(t *testing.T) {
	info := extractors.ExtractedFileInfo{
		DetectedType: "text",
		TextSamples: []extractors.TextSample{
			{Source: "first-meaningful-line", Text: "OeYV/jjq0pT9Jn4oiiJG UHmmYZszQjxHikWZF8lCoisYzBgiJEuZoRpmcYzMQ8RmMIivI5GYwhm44R8UvH42M2M5HhnoIOVa", Score: 0.58},
		},
	}

	got := GenerateFilename(info)

	if got.Filename != "unidentified-content" {
		t.Fatalf("filename = %q, want unidentified-content", got.Filename)
	}
	if got.Confidence >= 0.4 {
		t.Fatalf("confidence = %.2f, want low confidence", got.Confidence)
	}
}

func TestGenerateFilenamePrefersTextSummaryOverGreeting(t *testing.T) {
	info := extractors.ExtractedFileInfo{
		DetectedType: "text",
		TextSamples: []extractors.TextSample{
			{Source: "first-meaningful-line", Text: "Hi Jenna,", Score: 0.55},
			{Source: "text-summary", Text: "Please find attached the April 2025 invoice for the Woodbridge account.", Score: 0.72},
		},
	}

	got := GenerateFilename(info)

	if got.Filename != "april-2025-invoice-woodbridge-account" {
		t.Fatalf("filename = %q", got.Filename)
	}
	if got.Confidence < 0.75 {
		t.Fatalf("confidence = %.2f, want copy-threshold friendly confidence", got.Confidence)
	}
	if got.Evidence[0] != "text-summary" {
		t.Fatalf("evidence = %+v, want text-summary first", got.Evidence)
	}
}

func TestGenerateFilenameAllowsMeaningfulMediaBasename(t *testing.T) {
	info := extractors.ExtractedFileInfo{
		DetectedType:       "media",
		SuggestedExtension: "mp4",
		TextSamples: []extractors.TextSample{
			{Source: "media-filename", Text: "alice", Score: 0.72},
		},
	}

	got := GenerateFilename(info)

	if got.Filename != "alice" {
		t.Fatalf("filename = %q, want alice", got.Filename)
	}
	if got.Confidence < 0.7 {
		t.Fatalf("confidence = %.2f, want basename confidence", got.Confidence)
	}
}

func TestGenerateFilenameUsesMediaTagsBeforeProperties(t *testing.T) {
	info := extractors.ExtractedFileInfo{
		DetectedType:       "media",
		SuggestedExtension: "mp3",
		TextSamples: []extractors.TextSample{
			{Source: "media-properties", Text: "audio MP3 audio 123.45", Score: 0.35},
			{Source: "media-tags", Text: "Late Night Drive Retro Metro City Lights", Score: 0.9},
		},
	}

	got := GenerateFilename(info)

	if got.Filename != "late-night-drive-retro-metro-city-lights" {
		t.Fatalf("filename = %q", got.Filename)
	}
	if got.Evidence[0] != "media-tags" {
		t.Fatalf("evidence = %+v, want media-tags first", got.Evidence)
	}
}

func TestGenerateFilenameRejectsSparseMediaProperties(t *testing.T) {
	info := extractors.ExtractedFileInfo{
		DetectedType:       "media",
		SuggestedExtension: "mkv",
		TextSamples: []extractors.TextSample{
			{Source: "media-properties", Text: "video Matroska WebM 12.0", Score: 0.35},
		},
	}

	got := GenerateFilename(info)

	if got.Filename != "unidentified-video" {
		t.Fatalf("filename = %q, want unidentified-video", got.Filename)
	}
	if got.Confidence >= 0.4 {
		t.Fatalf("confidence = %.2f, want low confidence", got.Confidence)
	}
}

func TestGenerateFilenameRejectsSparseImageProperties(t *testing.T) {
	info := extractors.ExtractedFileInfo{
		DetectedType:       "image",
		SuggestedExtension: "png",
		TextSamples: []extractors.TextSample{
			{Source: "image-properties", Text: "png image 1024x1024", Score: 0.45},
		},
	}

	got := GenerateFilename(info)

	if got.Filename != "unidentified-image" {
		t.Fatalf("filename = %q, want unidentified-image", got.Filename)
	}
	if got.Confidence >= 0.4 {
		t.Fatalf("confidence = %.2f, want low confidence", got.Confidence)
	}
}

func TestGenerateFilenameUsesImageExifBeforeProperties(t *testing.T) {
	info := extractors.ExtractedFileInfo{
		DetectedType:       "image",
		SuggestedExtension: "jpg",
		TextSamples: []extractors.TextSample{
			{Source: "image-properties", Text: "jpeg image 4000x3000", Score: 0.45},
			{Source: "image-exif-title", Text: "Cliffs at Western Brook Pond", Score: 0.9},
		},
	}

	got := GenerateFilename(info)

	if got.Filename != "cliffs-western-brook-pond" {
		t.Fatalf("filename = %q", got.Filename)
	}
	if got.Evidence[0] != "image-exif-title" {
		t.Fatalf("evidence = %+v, want image-exif-title first", got.Evidence)
	}
}

func TestGenerateFilenameUsesImageFilename(t *testing.T) {
	info := extractors.ExtractedFileInfo{
		DetectedType:       "image",
		SuggestedExtension: "png",
		TextSamples: []extractors.TextSample{
			{Source: "image-properties", Text: "png image 1024x768", Score: 0.45},
			{Source: "image-filename", Text: "cliffs", Score: 0.75},
		},
	}

	got := GenerateFilename(info)

	if got.Filename != "cliffs" {
		t.Fatalf("filename = %q, want cliffs", got.Filename)
	}
	if got.Confidence < 0.75 {
		t.Fatalf("confidence = %.2f, want copy-threshold friendly confidence", got.Confidence)
	}
}

func TestGenerateFilenameUsesMediaTimestamp(t *testing.T) {
	info := extractors.ExtractedFileInfo{
		DetectedType:       "media",
		SuggestedExtension: "mkv",
		TextSamples: []extractors.TextSample{
			{Source: "media-timestamp", Text: "video 2026 01 22 00 59 57", Score: 0.74},
		},
	}

	got := GenerateFilename(info)

	if got.Filename != "video-2026-01-22-00-59-57" {
		t.Fatalf("filename = %q, want timestamp name", got.Filename)
	}
	if got.Confidence < 0.75 {
		t.Fatalf("confidence = %.2f, want copy-threshold friendly confidence", got.Confidence)
	}
}

func TestGenerateFilenameUsesShortTextNoteAtMediumConfidence(t *testing.T) {
	info := extractors.ExtractedFileInfo{
		DetectedType: "text",
		TextSamples: []extractors.TextSample{
			{Source: "short-text-note", Text: "hello there general kenobi roger roger", Score: 0.58},
		},
	}

	got := GenerateFilename(info)

	if got.Filename != "hello-there-general-kenobi-roger" {
		t.Fatalf("filename = %q", got.Filename)
	}
	if got.Confidence >= 0.75 {
		t.Fatalf("confidence = %.2f, want below automatic copy threshold", got.Confidence)
	}
}

func TestGenerateFilenameUsesMediaKindFallbacks(t *testing.T) {
	cases := []struct {
		ext  string
		want string
	}{
		{ext: "wav", want: "unidentified-audio"},
		{ext: "mp4", want: "unidentified-video"},
		{ext: "", want: "unidentified-media"},
	}

	for _, tc := range cases {
		info := extractors.ExtractedFileInfo{DetectedType: "media", SuggestedExtension: tc.ext}
		got := GenerateFilename(info)
		if got.Filename != tc.want {
			t.Fatalf("ext %q filename = %q, want %q", tc.ext, got.Filename, tc.want)
		}
	}
}
