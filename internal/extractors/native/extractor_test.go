package native

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestExtractorUsesCSVHeadersAsSignal(t *testing.T) {
	path := writeFile(t, "customers.csv", "name,email,status\nAda Lovelace,ada@example.com,active\n")

	got, err := Extractor{}.Extract(context.Background(), path)
	if err != nil {
		t.Fatal(err)
	}
	signals := got.Evidence.TextSignals
	if len(signals) != 1 || signals[0] != "name email status" {
		t.Fatalf("TextSignals = %+v, want CSV headers only", signals)
	}
}

func TestExtractorUsesJSONKeysAsSignal(t *testing.T) {
	path := writeFile(t, "file0007", `{"invoice":{"id":123,"customer":{"name":"Ada"},"total_due":42}}`)

	got, err := Extractor{}.Extract(context.Background(), path)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Evidence.TextSignals) != 1 {
		t.Fatalf("TextSignals = %+v", got.Evidence.TextSignals)
	}
	want := "customer id invoice name total_due"
	if got.Evidence.TextSignals[0] != want {
		t.Fatalf("TextSignals[0] = %q, want %q", got.Evidence.TextSignals[0], want)
	}
}

func TestExtractorUsesMarkdownHeadingAsSignal(t *testing.T) {
	path := writeFile(t, "note", "# Incident Response Runbook\n\nBody")

	got, err := Extractor{}.Extract(context.Background(), path)
	if err != nil {
		t.Fatal(err)
	}
	if got.Evidence.TextSignals[0] != "Incident Response Runbook" {
		t.Fatalf("TextSignals = %+v", got.Evidence.TextSignals)
	}
}

func TestExtractorUsesHTMLTitleAsSignal(t *testing.T) {
	path := writeFile(t, "recovered", `<!doctype html><html><head><title>Basics of Photosynthesis</title></head><body></body></html>`)

	got, err := Extractor{}.Extract(context.Background(), path)
	if err != nil {
		t.Fatal(err)
	}
	if got.Evidence.TextSignals[0] != "Basics of Photosynthesis" {
		t.Fatalf("TextSignals = %+v", got.Evidence.TextSignals)
	}
}

func TestExtractorUsesEmailSubjectAsSignal(t *testing.T) {
	path := writeFile(t, "message", "Subject: Customer Onboarding Checklist\nFrom: ops@example.com\n\nBody")

	got, err := Extractor{}.Extract(context.Background(), path)
	if err != nil {
		t.Fatal(err)
	}
	if got.Evidence.TextSignals[0] != "Customer Onboarding Checklist" {
		t.Fatalf("TextSignals = %+v", got.Evidence.TextSignals)
	}
}

func writeFile(t *testing.T, name, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}
