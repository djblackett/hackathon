package extractors

import (
	"archive/zip"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOfficeExtractorUsesXLSXHeaders(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sheet.xlsx")
	writeZip(t, path, map[string]string{
		"xl/workbook.xml": `<workbook><sheets><sheet name="Customers"/></sheets></workbook>`,
		"xl/sharedStrings.xml": `<sst>
			<si><t>Customer Name</t></si>
			<si><t>Email</t></si>
			<si><t>Status</t></si>
		</sst>`,
		"xl/worksheets/sheet1.xml": `<worksheet><sheetData>
			<row r="1"><c t="s"><v>0</v></c><c t="s"><v>1</v></c><c t="s"><v>2</v></c></row>
		</sheetData></worksheet>`,
	})

	info, err := officeExtractor{}.ExtractInfo(path)
	if err != nil {
		t.Fatal(err)
	}

	if !hasSample(info, "office-headers", "Customer Name Email Status") {
		t.Fatalf("missing xlsx header sample: %+v", info.TextSamples)
	}
	if !hasSample(info, "office-sheet-name", "Customers") {
		t.Fatalf("missing xlsx sheet name sample: %+v", info.TextSamples)
	}
}

func TestOfficeExtractorUsesPPTXSlideTitle(t *testing.T) {
	path := filepath.Join(t.TempDir(), "slides.pptx")
	writeZip(t, path, map[string]string{
		"ppt/presentation.xml":  `<p:presentation/>`,
		"ppt/slides/slide1.xml": `<p:sld><a:t>Quarterly Planning Review</a:t><a:t>Budget and staffing</a:t></p:sld>`,
	})

	info, err := officeExtractor{}.ExtractInfo(path)
	if err != nil {
		t.Fatal(err)
	}

	if !hasSample(info, "office-slide-title", "Quarterly Planning Review Budget and staffing") {
		t.Fatalf("missing pptx slide title sample: %+v", info.TextSamples)
	}
}

func TestOfficeExtractorUsesRecoveredDocText(t *testing.T) {
	path := filepath.Join("..", "..", "testdata", "recovered", "recovered-doc")

	info, err := officeExtractor{}.ExtractInfo(path)
	if err != nil {
		t.Fatal(err)
	}

	if !hasSamplePrefix(info, "office-text", "The Monumental Construction of the Panama Canal") {
		t.Fatalf("missing recovered doc heading sample: %+v", info.TextSamples)
	}
	if !hasSample(info, "office-heading", "The Monumental Construction of the Panama Canal") {
		t.Fatalf("missing recovered doc heading sample: %+v", info.TextSamples)
	}
	if !hasSamplePrefix(info, "office-first-paragraph", "The Panama Canal, completed in 1914") {
		t.Fatalf("missing recovered doc first paragraph sample: %+v", info.TextSamples)
	}
}

func writeZip(t *testing.T, path string, files map[string]string) {
	t.Helper()

	out, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(out)
	for name, content := range files {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := out.Close(); err != nil {
		t.Fatal(err)
	}
}

func hasSample(info ExtractedFileInfo, source, text string) bool {
	for _, sample := range info.TextSamples {
		if sample.Source == source && sample.Text == text {
			return true
		}
	}
	return false
}

func hasSamplePrefix(info ExtractedFileInfo, source, prefix string) bool {
	for _, sample := range info.TextSamples {
		if sample.Source == source && strings.HasPrefix(sample.Text, prefix) {
			return true
		}
	}
	return false
}
