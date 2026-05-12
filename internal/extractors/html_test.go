package extractors

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHTMLExtractorUsesOpenGraphTitle(t *testing.T) {
	path := filepath.Join(t.TempDir(), "recovered")
	content := `<!doctype html>
<html>
<head>
  <title>Noisy Export Page</title>
  <meta property="og:title" content="Network Migration Cutover Plan">
  <meta property="og:description" content="Detailed weekend migration checklist for the core network.">
</head>
<body><h1>Fallback Heading</h1><p>Copyright footer and repeated body text.</p></body>
</html>`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	info, err := htmlExtractor{}.ExtractInfo(path)
	if err != nil {
		t.Fatal(err)
	}

	if info.SuggestedExtension != "html" {
		t.Fatalf("SuggestedExtension = %q, want html", info.SuggestedExtension)
	}
	if !hasSample(info, "html-og-title", "Network Migration Cutover Plan") {
		t.Fatalf("missing OpenGraph title sample: %+v", info.TextSamples)
	}
	if !hasSample(info, "html-og-description", "Detailed weekend migration checklist for the core network.") {
		t.Fatalf("missing OpenGraph description sample: %+v", info.TextSamples)
	}
}

func TestHTMLExtractorFallsBackToHeading(t *testing.T) {
	path := filepath.Join(t.TempDir(), "recovered.bin")
	content := `<html><head></head><body><h1>Incident Response Runbook</h1><p>Escalation and recovery notes.</p></body></html>`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	info, err := htmlExtractor{}.ExtractInfo(path)
	if err != nil {
		t.Fatal(err)
	}

	if !hasSample(info, "html-h1", "Incident Response Runbook") {
		t.Fatalf("missing h1 sample: %+v", info.TextSamples)
	}
}
