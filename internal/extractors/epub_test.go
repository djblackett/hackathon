package extractors

import (
	"path/filepath"
	"testing"
)

func TestEPUBExtractorUsesMetadataTitle(t *testing.T) {
	path := filepath.Join(t.TempDir(), "book.epub")
	writeZip(t, path, map[string]string{
		"mimetype":               "application/epub+zip",
		"META-INF/container.xml": `<container><rootfiles><rootfile full-path="OEBPS/content.opf"/></rootfiles></container>`,
		"OEBPS/content.opf": `<package><metadata>
			<dc:title>Practical Canal Engineering</dc:title>
			<dc:creator>Alex Rivera</dc:creator>
		</metadata></package>`,
		"OEBPS/chapter1.xhtml": `<html><body><h1>First Chapter</h1></body></html>`,
	})

	info, err := epubExtractor{}.ExtractInfo(path)
	if err != nil {
		t.Fatal(err)
	}

	if info.SuggestedExtension != "epub" {
		t.Fatalf("SuggestedExtension = %q, want epub", info.SuggestedExtension)
	}
	if !hasSample(info, "epub-title", "Practical Canal Engineering") {
		t.Fatalf("missing epub title sample: %+v", info.TextSamples)
	}
	if !hasSample(info, "epub-heading", "First Chapter") {
		t.Fatalf("missing epub heading sample: %+v", info.TextSamples)
	}
}
