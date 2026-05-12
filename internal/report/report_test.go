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
}
