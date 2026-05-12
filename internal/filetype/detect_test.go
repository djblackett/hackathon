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

func writeTestFile(t *testing.T, name, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}
