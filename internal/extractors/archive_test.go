package extractors

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"
)

func TestArchiveExtractorUsesTopLevelDirectory(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bundle.zip")
	writeZip(t, path, map[string]string{
		"customer-export/accounts.csv": "id,name",
		"customer-export/orders.csv":   "id,total",
	})

	info, err := archiveExtractor{}.ExtractInfo(path)
	if err != nil {
		t.Fatal(err)
	}

	if info.SuggestedExtension != "zip" {
		t.Fatalf("SuggestedExtension = %q, want zip", info.SuggestedExtension)
	}
	if !hasSample(info, "archive-contents", "customer-export") {
		t.Fatalf("missing archive contents sample: %+v", info.TextSamples)
	}
}

func TestArchiveExtractorReadsTarGz(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bundle.tar.gz")
	if err := writeTarGz(path, map[string]string{
		"release-notes/README.md": "hello",
		"release-notes/app.log":   "log",
	}); err != nil {
		t.Fatal(err)
	}

	info, err := archiveExtractor{}.ExtractInfo(path)
	if err != nil {
		t.Fatal(err)
	}

	if info.SuggestedExtension != "tar.gz" {
		t.Fatalf("SuggestedExtension = %q, want tar.gz", info.SuggestedExtension)
	}
	if !hasSample(info, "archive-contents", "release-notes") {
		t.Fatalf("missing archive contents sample: %+v", info.TextSamples)
	}
}

func writeTarGz(path string, files map[string]string) error {
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)
	for name, content := range files {
		if err := tw.WriteHeader(&tar.Header{Name: name, Mode: 0644, Size: int64(len(content))}); err != nil {
			return err
		}
		if _, err := tw.Write([]byte(content)); err != nil {
			return err
		}
	}
	if err := tw.Close(); err != nil {
		return err
	}
	if err := gw.Close(); err != nil {
		return err
	}
	return os.WriteFile(path, buf.Bytes(), 0644)
}
