package extractors

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEmailExtractorUsesSubject(t *testing.T) {
	path := filepath.Join(t.TempDir(), "message")
	if err := os.WriteFile(path, []byte("Subject: Customer Onboarding Checklist\nFrom: ops@example.com\nTo: support@example.com\n\nBody"), 0644); err != nil {
		t.Fatal(err)
	}

	info, err := emailExtractor{}.ExtractInfo(path)
	if err != nil {
		t.Fatal(err)
	}

	if info.SuggestedExtension != "eml" {
		t.Fatalf("SuggestedExtension = %q, want eml", info.SuggestedExtension)
	}
	if !hasSample(info, "email-subject", "Customer Onboarding Checklist") {
		t.Fatalf("missing email subject sample: %+v", info.TextSamples)
	}
}
