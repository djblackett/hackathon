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
