package naming

import "testing"

func TestSanitizeBase(t *testing.T) {
	tests := map[string]string{
		"Monthly Statement.pdf": "monthly-statement.pdf",
		"CON":                   "con-file",
		"scan///2024":           "scan-2024",
		"bad\x00 name...":       "bad-name",
	}
	for input, want := range tests {
		if got := SanitizeBase(input); got != want {
			t.Fatalf("SanitizeBase(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestWithExtensionPreservesExtension(t *testing.T) {
	got := WithExtension("Monthly Statement.pdf", ".PDF")
	if got != "monthly-statement.pdf.pdf" {
		t.Fatalf("WithExtension = %q", got)
	}
}
