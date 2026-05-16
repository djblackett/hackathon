package tesseract

import "testing"

func TestParseOCRText(t *testing.T) {
	got := Parse("recovered/scan.png", []byte("Nova Scotia Power Bill\nAmount due April 2022\n"), 20)

	if got.TextPreview != "Nova Scotia Power" {
		t.Fatalf("TextPreview = %q", got.TextPreview)
	}
	if len(got.TextSignals) != 1 || got.TextSignals[0] != "Nova Scotia Power Bill Amount due April 2022" {
		t.Fatalf("TextSignals = %+v", got.TextSignals)
	}
	if got.Metadata["ocr"] != "true" {
		t.Fatalf("Metadata = %+v", got.Metadata)
	}
	if len(got.Warnings) == 0 {
		t.Fatal("expected OCR caution warning")
	}
}

func TestParseEmptyOCRText(t *testing.T) {
	got := Parse("recovered/blank.png", nil, 2000)

	if got.TextPreview != "" || len(got.TextSignals) != 0 {
		t.Fatalf("unexpected text evidence: %+v", got)
	}
	if len(got.Warnings) < 2 {
		t.Fatalf("Warnings = %+v, want caution and empty text warnings", got.Warnings)
	}
}
