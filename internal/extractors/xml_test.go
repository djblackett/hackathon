package extractors

import (
	"os"
	"path/filepath"
	"testing"
)

func TestXMLExtractorUsesGenericTitle(t *testing.T) {
	path := filepath.Join(t.TempDir(), "recovered.bin")
	content := `<?xml version="1.0"?>
<archive-record>
  <title>Quarterly Safety Inspection Log</title>
  <name>Warehouse North Wing</name>
  <author>Facilities Team</author>
</archive-record>`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	info, err := xmlExtractor{}.ExtractInfo(path)
	if err != nil {
		t.Fatal(err)
	}

	if info.SuggestedExtension != "xml" {
		t.Fatalf("SuggestedExtension = %q, want xml", info.SuggestedExtension)
	}
	if info.Metadata["root"] != "archive-record" {
		t.Fatalf("root = %q, want archive-record", info.Metadata["root"])
	}
	if !hasSample(info, "xml-title", "Quarterly Safety Inspection Log") {
		t.Fatalf("missing xml title sample: %+v", info.TextSamples)
	}
}

func TestXMLExtractorUsesMusicXMLFields(t *testing.T) {
	path := filepath.Join(t.TempDir(), "recovered")
	content := `<?xml version="1.0"?>
<score-partwise version="3.0">
  <work><work-title>You Are My Sunshine</work-title></work>
  <movement-title>Beginner Band</movement-title>
  <identification><creator type="composer">6 Note Songs</creator></identification>
  <part-list>
    <score-part id="P1"><part-name>Flute</part-name></score-part>
    <score-part id="P2"><part-name>Clarinet</part-name></score-part>
  </part-list>
</score-partwise>`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	info, err := xmlExtractor{}.ExtractInfo(path)
	if err != nil {
		t.Fatal(err)
	}

	if info.SuggestedExtension != "musicxml" {
		t.Fatalf("SuggestedExtension = %q, want musicxml", info.SuggestedExtension)
	}
	if info.Metadata["detected_subtype"] != "musicxml" {
		t.Fatalf("detected_subtype = %q, want musicxml", info.Metadata["detected_subtype"])
	}
	if !hasSample(info, "musicxml-work-title", "You Are My Sunshine") {
		t.Fatalf("missing MusicXML work title sample: %+v", info.TextSamples)
	}
	if !hasSample(info, "musicxml-creator", "6 Note Songs") {
		t.Fatalf("missing MusicXML creator sample: %+v", info.TextSamples)
	}
	if !hasSample(info, "musicxml-parts", "Flute Clarinet") {
		t.Fatalf("missing MusicXML parts sample: %+v", info.TextSamples)
	}
}
