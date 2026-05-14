package extractors

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTextExtractorUsesShortTextNoteForTinyReadableFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "note.txt")
	if err := os.WriteFile(path, []byte("hello there\ngeneral kenobi\nroger roger\n"), 0644); err != nil {
		t.Fatal(err)
	}

	info, err := textExtractor{}.ExtractInfo(path)
	if err != nil {
		t.Fatal(err)
	}

	if !hasSample(info, "short-text-note", "hello there general kenobi roger roger") {
		t.Fatalf("missing short text note sample: %+v", info.TextSamples)
	}
}

func TestTextExtractorRejectsRandomShortTextNote(t *testing.T) {
	path := filepath.Join(t.TempDir(), "random.txt")
	if err := os.WriteFile(path, []byte("OeYV/jjq0pT9Jn4oiiJG\nUHmmYZszQjxHikWZF8lCoisYzBgiJEuZoRpmcYzMQ8RmMIivI5GYwhm44R8UvH42M2M5HhnoIOVa\n"), 0644); err != nil {
		t.Fatal(err)
	}

	info, err := textExtractor{}.ExtractInfo(path)
	if err != nil {
		t.Fatal(err)
	}

	if hasSamplePrefix(info, "short-text-note", "OeYV") {
		t.Fatalf("random text should not create short text note: %+v", info.TextSamples)
	}
}
