package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanAcceptsFlagsAfterDirectory(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "file000421"), []byte("%PDF-1.7\nbody"), 0644); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(t.TempDir(), "plan.json")

	if err := runApp([]string{"recovername", "scan", root, "--out", out, "--no-timestamp"}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(out); err != nil {
		t.Fatalf("expected plan at trailing --out path: %v", err)
	}
}
