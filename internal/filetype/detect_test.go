package filetype

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
)

func TestDetectPDFWithoutPDFExtension(t *testing.T) {
	path := writeTestFile(t, "recovered.bin", "%PDF-1.7\nbody")

	got := Detect(path)

	if got.Type != "pdf" {
		t.Fatalf("Type = %q, want pdf", got.Type)
	}
	if got.Extension != "bin" {
		t.Fatalf("Extension = %q, want bin", got.Extension)
	}
}

func TestDetectJSONWithoutExtension(t *testing.T) {
	path := writeTestFile(t, "recovered", `{"title":"Quarterly Revenue Review","items":[1,2]}`)

	got := Detect(path)

	if got.Type != "json" {
		t.Fatalf("Type = %q, want json", got.Type)
	}
}

func TestDetectOfficeDocxContainer(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "recovered.bin")

	out, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(out)
	w, err := zw.Create("word/document.xml")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write([]byte("<w:document/>")); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := out.Close(); err != nil {
		t.Fatal(err)
	}

	got := Detect(path)

	if got.Type != "office" || got.Subtype != "docx" {
		t.Fatalf("got type=%q subtype=%q, want office/docx", got.Type, got.Subtype)
	}
}

func TestDetectUnknownBinary(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "blob")
	if err := os.WriteFile(path, []byte{0x00, 0x01, 0x02, 0x03}, 0644); err != nil {
		t.Fatal(err)
	}

	got := Detect(path)

	if got.Type != "unknown" {
		t.Fatalf("Type = %q, want unknown", got.Type)
	}
}

func TestDetectEmail(t *testing.T) {
	path := writeTestFile(t, "message", "Subject: Recovery Plan\nFrom: ops@example.com\nTo: team@example.com\n\nBody")

	got := Detect(path)

	if got.Type != "email" {
		t.Fatalf("Type = %q, want email", got.Type)
	}
}

func TestDetectHTMLWithoutHTMLExtension(t *testing.T) {
	path := writeTestFile(t, "recovered.bin", `<!doctype html><html><head><title>Basics of Photosynthesis</title></head><body></body></html>`)

	got := Detect(path)

	if got.Type != "html" {
		t.Fatalf("Type = %q, want html", got.Type)
	}
	if got.CanonicalExtension != "html" {
		t.Fatalf("CanonicalExtension = %q, want html", got.CanonicalExtension)
	}
}

func TestDetectXMLWithoutXMLExtension(t *testing.T) {
	path := writeTestFile(t, "recovered.bin", `<?xml version="1.0"?><archive-record><title>Quarterly Safety Inspection Log</title></archive-record>`)

	got := Detect(path)

	if got.Type != "xml" {
		t.Fatalf("Type = %q, want xml", got.Type)
	}
	if got.CanonicalExtension != "xml" {
		t.Fatalf("CanonicalExtension = %q, want xml", got.CanonicalExtension)
	}
}

func TestDetectMusicXMLWithoutExtension(t *testing.T) {
	path := writeTestFile(t, "recovered", `<?xml version="1.0"?><score-partwise version="3.0"><work><work-title>You Are My Sunshine</work-title></work></score-partwise>`)

	got := Detect(path)

	if got.Type != "xml" || got.Subtype != "musicxml" {
		t.Fatalf("got type=%q subtype=%q, want xml/musicxml", got.Type, got.Subtype)
	}
	if got.CanonicalExtension != "musicxml" {
		t.Fatalf("CanonicalExtension = %q, want musicxml", got.CanonicalExtension)
	}
}

func TestDetectMediaByExtension(t *testing.T) {
	path := filepath.Join(t.TempDir(), "clip.mp3")
	if err := os.WriteFile(path, []byte{0x00, 0x01, 0x02, 0x03}, 0644); err != nil {
		t.Fatal(err)
	}

	got := Detect(path)

	if got.Type != "media" {
		t.Fatalf("Type = %q, want media", got.Type)
	}
}

func TestDetectRTF(t *testing.T) {
	path := writeTestFile(t, "recovered.bin", `{\rtf1\ansi Project Notes}`)

	got := Detect(path)

	if got.Type != "rtf" || got.CanonicalExtension != "rtf" {
		t.Fatalf("got type=%q canonical=%q, want rtf/rtf", got.Type, got.CanonicalExtension)
	}
}

func TestDetectIPYNB(t *testing.T) {
	path := writeTestFile(t, "analysis.ipynb", `{"cells":[],"metadata":{},"nbformat":4,"nbformat_minor":5}`)

	got := Detect(path)

	if got.Type != "notebook" || got.Subtype != "ipynb" || got.CanonicalExtension != "ipynb" {
		t.Fatalf("got type=%q subtype=%q canonical=%q, want notebook/ipynb/ipynb", got.Type, got.Subtype, got.CanonicalExtension)
	}
}

func TestDetectEPUBContainer(t *testing.T) {
	path := writeZipFile(t, "book.epub", map[string]string{
		"mimetype":               "application/epub+zip",
		"META-INF/container.xml": `<container/>`,
	})

	got := Detect(path)

	if got.Type != "epub" || got.CanonicalExtension != "epub" {
		t.Fatalf("got type=%q canonical=%q, want epub/epub", got.Type, got.CanonicalExtension)
	}
}

func TestDetectOpenDocumentContainer(t *testing.T) {
	path := writeZipFile(t, "sheet.ods", map[string]string{
		"mimetype":    "application/vnd.oasis.opendocument.spreadsheet",
		"content.xml": `<office:document-content/>`,
	})

	got := Detect(path)

	if got.Type != "opendocument" || got.Subtype != "ods" || got.CanonicalExtension != "ods" {
		t.Fatalf("got type=%q subtype=%q canonical=%q, want opendocument/ods/ods", got.Type, got.Subtype, got.CanonicalExtension)
	}
}

func TestDetectGenericZipArchive(t *testing.T) {
	path := writeZipFile(t, "bundle.zip", map[string]string{
		"project/readme.txt": "hello",
	})

	got := Detect(path)

	if got.Type != "archive" || got.Subtype != "zip" || got.CanonicalExtension != "zip" {
		t.Fatalf("got type=%q subtype=%q canonical=%q, want archive/zip/zip", got.Type, got.Subtype, got.CanonicalExtension)
	}
}

func writeTestFile(t *testing.T, name, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func writeZipFile(t *testing.T, name string, files map[string]string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), name)
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
	return path
}
