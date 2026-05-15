package extractors

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNotebookExtractorUsesMarkdownHeading(t *testing.T) {
	path := filepath.Join(t.TempDir(), "analysis.ipynb")
	content := `{
  "metadata": {"title": "Fallback Notebook Title"},
  "nbformat": 4,
  "nbformat_minor": 5,
  "cells": [
    {"cell_type": "code", "source": "print(1)"},
    {"cell_type": "markdown", "source": ["# Revenue Forecast Analysis\n", "Model notes"]}
  ]
}`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	info, err := notebookExtractor{}.ExtractInfo(path)
	if err != nil {
		t.Fatal(err)
	}

	if info.SuggestedExtension != "ipynb" {
		t.Fatalf("SuggestedExtension = %q, want ipynb", info.SuggestedExtension)
	}
	if !hasSample(info, "notebook-heading", "Revenue Forecast Analysis") {
		t.Fatalf("missing notebook heading sample: %+v", info.TextSamples)
	}
	if !hasSample(info, "notebook-title", "Fallback Notebook Title") {
		t.Fatalf("missing notebook title sample: %+v", info.TextSamples)
	}
}

func TestNotebookExtractorWarnsOnMalformedJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "broken.ipynb")
	if err := os.WriteFile(path, []byte(`not-json`), 0644); err != nil {
		t.Fatal(err)
	}

	info, err := notebookExtractor{}.ExtractInfo(path)
	if err != nil {
		t.Fatal(err)
	}

	if !containsWarning(info.Warnings, "could not parse notebook json") {
		t.Fatalf("warnings = %+v", info.Warnings)
	}
}
