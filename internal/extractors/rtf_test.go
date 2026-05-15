package extractors

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRTFExtractorUsesHeading(t *testing.T) {
	path := filepath.Join(t.TempDir(), "notes.rtf")
	content := `{\\rtf1\\ansi Project Launch Notes\\par First paragraph with details.}`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	info, err := rtfExtractor{}.ExtractInfo(path)
	if err != nil {
		t.Fatal(err)
	}

	if info.SuggestedExtension != "rtf" {
		t.Fatalf("SuggestedExtension = %q, want rtf", info.SuggestedExtension)
	}
	if !hasSample(info, "rtf-heading", "Project Launch Notes") {
		t.Fatalf("missing rtf heading sample: %+v", info.TextSamples)
	}
}

func TestRTFExtractorWarnsOnEmptyText(t *testing.T) {
	path := filepath.Join(t.TempDir(), "empty.rtf")
	if err := os.WriteFile(path, []byte(`{\\rtf1\\ansi}`), 0644); err != nil {
		t.Fatal(err)
	}

	info, err := rtfExtractor{}.ExtractInfo(path)
	if err != nil {
		t.Fatal(err)
	}

	if !containsWarning(info.Warnings, "no rtf text extracted") {
		t.Fatalf("warnings = %+v", info.Warnings)
	}
}
