package extractors

import (
	"path/filepath"
	"testing"
)

func TestOpenDocumentExtractorUsesMetadataTitle(t *testing.T) {
	path := filepath.Join(t.TempDir(), "plan.odt")
	writeZip(t, path, map[string]string{
		"mimetype": "application/vnd.oasis.opendocument.text",
		"meta.xml": `<office:document-meta><office:meta>
			<dc:title>Quarterly Planning Notes</dc:title>
			<dc:subject>Roadmap review</dc:subject>
		</office:meta></office:document-meta>`,
		"content.xml": `<office:document-content><office:body><text:h>Fallback Heading</text:h></office:body></office:document-content>`,
	})

	info, err := openDocumentExtractor{}.ExtractInfo(path)
	if err != nil {
		t.Fatal(err)
	}

	if info.SuggestedExtension != "odt" {
		t.Fatalf("SuggestedExtension = %q, want odt", info.SuggestedExtension)
	}
	if !hasSample(info, "opendocument-title", "Quarterly Planning Notes") {
		t.Fatalf("missing opendocument title sample: %+v", info.TextSamples)
	}
	if !hasSample(info, "opendocument-heading", "Fallback Heading") {
		t.Fatalf("missing opendocument heading sample: %+v", info.TextSamples)
	}
}
