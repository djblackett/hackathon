package naming

import (
	"strings"
	"testing"

	"github.com/djblackett/bootdev-hackathon/internal/evidence"
)

func TestGeneratePenalizesGenericMetadataTitle(t *testing.T) {
	got := Generate(evidence.FileEvidence{
		Path:         "recovered/file0007",
		DetectedMIME: "application/pdf",
		Extension:    ".pdf",
		Metadata:     map[string]string{"title": "Document1"},
		TextPreview:  "2022-03-15 T4 tax slip employment income",
		TextSignals:  []string{"T4 tax slip employment income"},
		Sources:      []evidence.EvidenceSource{evidence.SourceNativeMIME},
	}, 7)

	if strings.Contains(got.Filename, "document1") {
		t.Fatalf("generic title dominated filename: %+v", got)
	}
	if !strings.Contains(got.Filename, "tax-slip") {
		t.Fatalf("filename = %q, want tax-slip signal", got.Filename)
	}
}

func TestGenerateUsesExtensionFallback(t *testing.T) {
	got := Generate(evidence.FileEvidence{
		Path:         "recovered/file000421",
		DetectedMIME: "application/pdf",
		Extension:    ".pdf",
		Sources:      []evidence.EvidenceSource{evidence.SourceNativeMIME},
	}, 1)

	if got.Filename != "unknown-pdf_000421.pdf" {
		t.Fatalf("Filename = %q, want unknown-pdf_000421.pdf", got.Filename)
	}
	if got.Confidence != ConfidenceLow {
		t.Fatalf("Confidence = %q, want low", got.Confidence)
	}
}

func TestIsGenericTitle(t *testing.T) {
	for _, value := range []string{"untitled", "Document1", "Microsoft Word - Document", "scan 0001"} {
		if !IsGenericTitle(value) {
			t.Fatalf("%q should be generic", value)
		}
	}
	if IsGenericTitle("Monthly Statement April") {
		t.Fatal("useful title classified as generic")
	}
}
