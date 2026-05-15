package plan

import (
	"testing"
	"time"

	"github.com/djblackett/bootdev-hackathon/internal/evidence"
)

func TestBuildResolvesDuplicateSuggestedNamesDeterministically(t *testing.T) {
	files := []evidence.FileEvidence{
		testEvidence("recovered/a", "Statement April"),
		testEvidence("recovered/b", "Statement April"),
		testEvidence("recovered/c", "Statement April"),
	}

	got := Build("recovered", files, zeroTime())

	want := []string{
		"recovered/statement-april.pdf",
		"recovered/statement-april_002.pdf",
		"recovered/statement-april_003.pdf",
	}
	for i := range want {
		if got.Items[i].SuggestedPath != want[i] {
			t.Fatalf("item %d suggested = %q, want %q", i, got.Items[i].SuggestedPath, want[i])
		}
	}
	if got.Items[1].Conflict == nil || got.Items[2].Conflict == nil {
		t.Fatalf("duplicate items should include conflict metadata: %+v", got.Items)
	}
}

func testEvidence(path, title string) evidence.FileEvidence {
	return evidence.FileEvidence{
		Path:         path,
		DetectedMIME: "application/pdf",
		Extension:    ".pdf",
		Metadata:     map[string]string{"title": title},
		Sources:      []evidence.EvidenceSource{evidence.SourceNativeMIME},
	}
}

func zeroTime() time.Time {
	return time.Time{}
}
