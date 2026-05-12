package ai

import (
	"strings"
	"testing"
)

func TestBuildEvidencePromptUsesEvidenceNotFullDocumentLanguage(t *testing.T) {
	evidence := "detected_type: pdf\ntop_samples:\n- source: pdf-first-text\n  text: Quarterly Revenue Review"

	got := buildEvidencePrompt(evidence)

	if !strings.Contains(got, evidence) {
		t.Fatalf("prompt missing evidence: %q", got)
	}
	if strings.Contains(got, "Here is the document") {
		t.Fatalf("evidence prompt should not use full-document wording: %q", got)
	}
	if !strings.Contains(got, "respond with the filename only") {
		t.Fatalf("prompt missing filename-only constraint: %q", got)
	}
}

func TestBuildEvidencePromptKeepsFilenameConstraints(t *testing.T) {
	got := buildEvidencePrompt("detected_type: csv")

	for _, want := range []string{"lowercase", "no file extension", "5-8 meaningful words"} {
		if !strings.Contains(got, want) {
			t.Fatalf("prompt missing %q: %q", want, got)
		}
	}
}
