package extractors

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestJSONExtractorWarnsOnInvalidJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "broken.json")
	if err := os.WriteFile(path, []byte(`{"missing":`), 0644); err != nil {
		t.Fatal(err)
	}

	info, err := jsonExtractor{}.ExtractInfo(path)
	if err != nil {
		t.Fatal(err)
	}
	if !containsWarning(info.Warnings, "could not parse json") {
		t.Fatalf("warnings = %+v, want parse warning", info.Warnings)
	}
}

func TestCSVExtractorWarnsOnInvalidCSV(t *testing.T) {
	path := filepath.Join(t.TempDir(), "broken.csv")
	if err := os.WriteFile(path, []byte("name,\"email\n"), 0644); err != nil {
		t.Fatal(err)
	}

	info, err := csvExtractor{}.ExtractInfo(path)
	if err != nil {
		t.Fatal(err)
	}
	if !containsWarning(info.Warnings, "could not parse csv headers") {
		t.Fatalf("warnings = %+v, want csv parse warning", info.Warnings)
	}
}

func TestPDFExtractorErrorsOnBrokenPDF(t *testing.T) {
	path := filepath.Join(t.TempDir(), "broken.pdf")
	if err := os.WriteFile(path, []byte("%PDF-1.7\nnot really a pdf"), 0644); err != nil {
		t.Fatal(err)
	}

	if _, err := (pdfExtractor{}).ExtractInfo(path); err == nil {
		t.Fatal("expected broken pdf error")
	}
}

func TestOfficeExtractorErrorsOnTruncatedOfficeFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "broken.docx")
	if err := os.WriteFile(path, []byte("PK\x03\x04truncated"), 0644); err != nil {
		t.Fatal(err)
	}

	if _, err := (officeExtractor{}).ExtractInfo(path); err == nil {
		t.Fatal("expected broken office error")
	}
}

func containsWarning(warnings []string, want string) bool {
	for _, warning := range warnings {
		if strings.Contains(warning, want) {
			return true
		}
	}
	return false
}
