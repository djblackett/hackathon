package jhove

import (
	"testing"
)

func TestParseValidStatus(t *testing.T) {
	got, err := Parse("recovered/file.pdf", []byte(`<?xml version="1.0"?>
<jhove>
  <version>1.28.0</version>
  <date>2026-05-15T12:00:00Z</date>
  <repInfo uri="recovered/file.pdf">
    <status>Well-Formed and valid</status>
  </repInfo>
</jhove>`))
	if err != nil {
		t.Fatal(err)
	}
	if got.Validation == nil {
		t.Fatal("Validation missing")
	}
	if got.Validation.Valid == nil || !*got.Validation.Valid {
		t.Fatalf("Valid = %+v, want true", got.Validation.Valid)
	}
	if got.Validation.Status != "Well-Formed and valid" {
		t.Fatalf("Status = %q", got.Validation.Status)
	}
	if got.Metadata["jhove_version"] != "1.28.0" {
		t.Fatalf("Metadata = %+v", got.Metadata)
	}
}

func TestParseInvalidStatusAndMessages(t *testing.T) {
	got, err := Parse("recovered/bad.pdf", []byte(`<?xml version="1.0"?>
<jhove>
  <repInfo uri="recovered/bad.pdf">
    <status>Not well-formed</status>
    <messages>
      <message>PDF-HUL-1: Header missing</message>
      <message>Trailing bytes found</message>
    </messages>
  </repInfo>
</jhove>`))
	if err != nil {
		t.Fatal(err)
	}
	if got.Validation == nil || got.Validation.Valid == nil || *got.Validation.Valid {
		t.Fatalf("Validation = %+v, want invalid", got.Validation)
	}
	if len(got.Validation.Warnings) != 2 {
		t.Fatalf("Validation warnings = %+v, want 2", got.Validation.Warnings)
	}
	if len(got.Warnings) != 2 {
		t.Fatalf("Evidence warnings = %+v, want 2", got.Warnings)
	}
}
