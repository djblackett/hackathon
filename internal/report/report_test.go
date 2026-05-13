package report

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestWriteReport(t *testing.T) {
	path := filepath.Join(t.TempDir(), "report.json")
	entries := []Entry{{
		SourcePath:      "input/a.txt",
		DestinationPath: "output/a.txt",
		SuggestedName:   "a.txt",
		Method:          "metadata",
		Confidence:      0.8,
		Evidence:        []string{"markdown-heading"},
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
		{Method: "metadata", Confidence: 0.8, DryRun: false, Warnings: []string{"warning"}},
	}

	got := BuildSummary(entries)

	if got.TotalFiles != 4 {
		t.Fatalf("TotalFiles = %d, want 4", got.TotalFiles)
	}
	if got.PlannedCount != 3 {
		t.Fatalf("PlannedCount = %d, want 3", got.PlannedCount)
	}
	if got.CopiedCount != 2 {
		t.Fatalf("CopiedCount = %d, want 2", got.CopiedCount)
	}
	if got.SkippedCount != 1 {
		t.Fatalf("SkippedCount = %d, want 1", got.SkippedCount)
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
}
