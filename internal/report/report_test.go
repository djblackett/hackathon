package report

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteReport(t *testing.T) {
	path := filepath.Join(t.TempDir(), "report.json")
	entries := []Entry{{
		SourcePath:      "input/a.txt",
		DestinationPath: "output/a.txt",
		SuggestedName:   "a.txt",
		Method:          "metadata",
		Confidence:      0.8500000000000001,
		Evidence:        []string{"markdown-heading"},
		Reason:          "generated from local evidence source: markdown-heading",
		DryRun:          true,
	}}

	if err := Write(path, entries); err != nil {
		t.Fatal(err)
	}

	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	var got Report
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatal(err)
	}
	if len(got.Entries) != 1 || got.Entries[0].Method != "metadata" {
		t.Fatalf("unexpected report: %+v", got)
	}
	if got.Summary.TotalFiles != 1 || got.Summary.PlannedCount != 1 {
		t.Fatalf("unexpected summary: %+v", got.Summary)
	}
	if got.Entries[0].Confidence != 0.85 {
		t.Fatalf("confidence = %v, want rounded 0.85", got.Entries[0].Confidence)
	}
	if got.Entries[0].ConfidenceBand != "high" {
		t.Fatalf("confidence band = %q, want high", got.Entries[0].ConfidenceBand)
	}
	if got.Entries[0].Reason != "generated from local evidence source: markdown-heading" {
		t.Fatalf("reason = %q", got.Entries[0].Reason)
	}
}

func TestReadReport(t *testing.T) {
	path := filepath.Join(t.TempDir(), "report.json")
	if err := Write(path, []Entry{{
		SourcePath:      "input/a.txt",
		DestinationPath: "output/a.txt",
		SuggestedName:   "a.txt",
		Method:          "metadata",
	}}); err != nil {
		t.Fatal(err)
	}

	got, err := Read(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Entries) != 1 || got.Entries[0].DestinationPath != "output/a.txt" {
		t.Fatalf("unexpected report: %+v", got)
	}
}

func TestBuildSummary(t *testing.T) {
	entries := []Entry{
		{Method: "metadata", Confidence: 0.95, DryRun: true},
		{Method: "ai-fallback", Confidence: 1, DryRun: false},
		{Method: "metadata", Confidence: 0.1, DryRun: false, Skipped: true, SkipReason: "low confidence"},
		{Method: "metadata", Confidence: 0.6, DryRun: false, Skipped: true, ReviewStatus: "accepted"},
		{Method: "metadata", Confidence: 0.7, DryRun: false, Skipped: true, ReviewStatus: "rejected"},
		{Method: "metadata", Confidence: 0.8, DryRun: false, Warnings: []string{"warning"}},
	}

	got := BuildSummary(entries)

	if got.TotalFiles != 6 {
		t.Fatalf("TotalFiles = %d, want 6", got.TotalFiles)
	}
	if got.PlannedCount != 3 {
		t.Fatalf("PlannedCount = %d, want 3", got.PlannedCount)
	}
	if got.CopiedCount != 2 {
		t.Fatalf("CopiedCount = %d, want 2", got.CopiedCount)
	}
	if got.SkippedCount != 3 {
		t.Fatalf("SkippedCount = %d, want 3", got.SkippedCount)
	}
	if got.LowConfidenceCount != 1 {
		t.Fatalf("LowConfidenceCount = %d, want 1", got.LowConfidenceCount)
	}
	if got.AIFallbackCount != 1 {
		t.Fatalf("AIFallbackCount = %d, want 1", got.AIFallbackCount)
	}
	if got.WarningsCount != 1 {
		t.Fatalf("WarningsCount = %d, want 1", got.WarningsCount)
	}
	if got.PendingReviewCount != 1 {
		t.Fatalf("PendingReviewCount = %d, want 1", got.PendingReviewCount)
	}
	if got.AcceptedCount != 1 {
		t.Fatalf("AcceptedCount = %d, want 1", got.AcceptedCount)
	}
	if got.RejectedCount != 1 {
		t.Fatalf("RejectedCount = %d, want 1", got.RejectedCount)
	}
}

func TestWriteReviewMarkdown(t *testing.T) {
	path := filepath.Join(t.TempDir(), "review.md")
	r := Report{
		Entries: []Entry{
			{
				SourcePath:      "files/input/random.txt",
				DestinationPath: "files/output/random-text-fragment.txt",
				Method:          "metadata",
				Confidence:      0.1,
				Reason:          "no strong local evidence found",
				Skipped:         true,
				SkipReason:      "confidence 0.10 below copy threshold 0.75",
				ReviewStatus:    "pending",
			},
			{
				SourcePath:      "files/input/note.md",
				DestinationPath: "files/output/incidents.md",
				Method:          "metadata",
				Confidence:      0.9,
				Evidence:        []string{"markdown-heading"},
				Reason:          "generated from local evidence source: markdown-heading",
			},
		},
	}
	r.Summary = BuildSummary(r.Entries)

	if err := WriteReviewMarkdown(path, r); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	got := string(data)
	for _, want := range []string{"# File Rename Review", "pending", "planned", "low", "high", "files/input/random.txt", "files/input/note.md", "markdown-heading", "generated from local evidence source: markdown-heading", "review_status"} {
		if !strings.Contains(got, want) {
			t.Fatalf("review markdown missing %q:\n%s", want, got)
		}
	}
}

func TestConfidenceBand(t *testing.T) {
	cases := map[float64]string{
		0.95: "high",
		0.85: "high",
		0.84: "medium",
		0.75: "medium",
		0.74: "review",
		0.4:  "review",
		0.39: "low",
	}
	for in, want := range cases {
		if got := ConfidenceBand(in); got != want {
			t.Fatalf("ConfidenceBand(%v) = %q, want %q", in, got, want)
		}
	}
}

func TestNormalizeReviewStatus(t *testing.T) {
	cases := map[string]string{
		"":         "pending",
		"APPROVE":  "accepted",
		"accepted": "accepted",
		"deny":     "rejected",
		"wat":      "pending",
	}
	for in, want := range cases {
		if got := NormalizeReviewStatus(in); got != want {
			t.Fatalf("NormalizeReviewStatus(%q) = %q, want %q", in, got, want)
		}
	}
}
