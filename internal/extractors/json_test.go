package extractors

import (
	"os"
	"path/filepath"
	"testing"
)

func TestJSONExtractorUsesNestedStructuredEvidence(t *testing.T) {
	path := filepath.Join(t.TempDir(), "quiz.bin")
	content := `{
  "quiz": {
    "sport": {
      "q1": {
        "question": "Which one is correct team name in NBA?",
        "options": ["New York Bulls", "Los Angeles Kings"],
        "answer": "Huston Rocket"
      }
    }
  }
}`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	info, err := jsonExtractor{}.ExtractInfo(path)
	if err != nil {
		t.Fatal(err)
	}

	if !hasSample(info, "json-structured", "quiz sport q1 question Which one is correct team name in NBA?") {
		t.Fatalf("missing structured JSON sample: %+v", info.TextSamples)
	}
	if info.Metadata["structured"] == "" {
		t.Fatalf("missing structured metadata: %+v", info.Metadata)
	}
}
