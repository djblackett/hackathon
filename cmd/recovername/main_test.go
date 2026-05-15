package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/djblackett/bootdev-hackathon/internal/app"
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

func TestApplyTrailingScanFlagsAcceptsSiegfried(t *testing.T) {
	cfg := app.ScanConfig{}

	err := applyTrailingScanFlags([]string{"--siegfried", "--siegfried-timeout", "3s", "--exiftool", "--exiftool-timeout=4s", "--ffprobe", "--ffprobe-timeout", "5s", "--hash=false"}, &cfg)
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.UseSiegfried {
		t.Fatal("UseSiegfried = false, want true")
	}
	if cfg.SiegfriedTimeout != 3*time.Second {
		t.Fatalf("SiegfriedTimeout = %s, want 3s", cfg.SiegfriedTimeout)
	}
	if !cfg.UseExifTool {
		t.Fatal("UseExifTool = false, want true")
	}
	if cfg.ExifToolTimeout != 4*time.Second {
		t.Fatalf("ExifToolTimeout = %s, want 4s", cfg.ExifToolTimeout)
	}
	if !cfg.UseFFProbe {
		t.Fatal("UseFFProbe = false, want true")
	}
	if cfg.FFProbeTimeout != 5*time.Second {
		t.Fatalf("FFProbeTimeout = %s, want 5s", cfg.FFProbeTimeout)
	}
	if cfg.Hash {
		t.Fatal("Hash = true, want false")
	}
}
