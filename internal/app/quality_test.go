package app

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestScanUsesFocusedNativeSignalsForNames(t *testing.T) {
	root := t.TempDir()
	files := map[string]string{
		"customer-a.csv": "name,email,status\nAda Lovelace,ada@example.com,active\n",
		"customer-b.csv": "name,email,status\nGrace Hopper,grace@example.com,active\n",
		"markdown-note":  "# Incident Response Runbook\n\nSteps for triage.",
		"recovered-html": `<!doctype html><html><head><title>Basics of Photosynthesis</title></head><body></body></html>`,
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(root, name), []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	got, err := Scan(context.Background(), ScanConfig{Root: root, OutPath: filepath.Join(t.TempDir(), "plan.json"), Hash: false, NoTimestamp: true})
	if err != nil {
		t.Fatal(err)
	}

	byBase := map[string]string{}
	for _, item := range got.Items {
		byBase[filepath.Base(item.OldPath)] = filepath.Base(item.SuggestedPath)
	}
	if byBase["customer-a.csv"] != "name-email-status.csv" {
		t.Fatalf("customer-a = %q", byBase["customer-a.csv"])
	}
	if byBase["customer-b.csv"] != "name-email-status_002.csv" {
		t.Fatalf("customer-b = %q", byBase["customer-b.csv"])
	}
	if byBase["markdown-note"] != "incident-response-runbook.md" {
		t.Fatalf("markdown-note = %q", byBase["markdown-note"])
	}
	if byBase["recovered-html"] != "basics-photosynthesis.html" {
		t.Fatalf("recovered-html = %q", byBase["recovered-html"])
	}
}
