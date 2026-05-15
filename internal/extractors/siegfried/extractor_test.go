package siegfried

import (
	"testing"

	"github.com/djblackett/bootdev-hackathon/internal/evidence"
)

func TestParsePositivePRONOMMatch(t *testing.T) {
	got, err := Parse("recovered/file000421", []byte(`{
	  "siegfried": "1.11.2",
	  "signature": "default.sig",
	  "files": [{
	    "filename": "recovered/file000421",
	    "matches": [{
	      "ns": "pronom",
	      "id": "fmt/18",
	      "format": "Acrobat PDF 1.4 - Portable Document Format",
	      "version": "1.4",
	      "mime": "application/pdf",
	      "extensions": "pdf",
	      "basis": "byte match at [[0 8]]"
	    }]
	  }]
	}`))
	if err != nil {
		t.Fatal(err)
	}
	if len(got.FormatIDs) != 1 {
		t.Fatalf("FormatIDs = %+v, want 1", got.FormatIDs)
	}
	id := got.FormatIDs[0]
	if id.Source != evidence.SourceSiegfried || id.ID != "fmt/18" || id.MIME != "application/pdf" || id.Extension != ".pdf" {
		t.Fatalf("unexpected format id: %+v", id)
	}
	if got.DetectedMIME != "application/pdf" || got.Extension != ".pdf" {
		t.Fatalf("mime/ext = %q/%q, want application/pdf/.pdf", got.DetectedMIME, got.Extension)
	}
	if got.Metadata["siegfried_version"] != "1.11.2" {
		t.Fatalf("missing siegfried version metadata: %+v", got.Metadata)
	}
}

func TestParseUnknownMatchAddsWarning(t *testing.T) {
	got, err := Parse("recovered/blob", []byte(`{
	  "files": [{
	    "filename": "recovered/blob",
	    "matches": [{
	      "ns": "pronom",
	      "id": "UNKNOWN",
	      "format": "UNKNOWN",
	      "warning": "no match"
	    }]
	  }]
	}`))
	if err != nil {
		t.Fatal(err)
	}
	if len(got.FormatIDs) != 0 {
		t.Fatalf("FormatIDs = %+v, want none", got.FormatIDs)
	}
	if len(got.Warnings) == 0 {
		t.Fatalf("expected warning for unknown match: %+v", got)
	}
}
