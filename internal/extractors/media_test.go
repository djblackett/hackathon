package extractors

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestMeaningfulMediaBasenameRejectsTimestamp(t *testing.T) {
	if got := meaningfulMediaBasename("2026-01-22_00-59-57.mkv"); got != "" {
		t.Fatalf("basename = %q, want empty", got)
	}
}

func TestMeaningfulMediaBasenameAllowsShortName(t *testing.T) {
	if got := meaningfulMediaBasename("alice.mp4"); got != "alice" {
		t.Fatalf("basename = %q, want alice", got)
	}
}

func TestMediaFilenameScorePrefersDescriptiveName(t *testing.T) {
	if mediaFilenameScore("Retro Metro mix 2 Dec 28 01 Start") <= mediaFilenameScore("alice") {
		t.Fatal("descriptive media filename should score higher than short basename")
	}
}

func TestFFProbeMetadataUnavailable(t *testing.T) {
	t.Setenv("PATH", "")

	_, warning := ffprobeMetadata("anything.mp3")

	if warning != "ffprobe not available; media metadata skipped" {
		t.Fatalf("warning = %q", warning)
	}
}

func TestFFProbeMetadataMalformedOutput(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell script fake ffprobe is POSIX-only")
	}

	dir := t.TempDir()
	ffprobe := filepath.Join(dir, "ffprobe")
	if err := os.WriteFile(ffprobe, []byte("#!/bin/sh\nprintf 'not-json'\n"), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir)

	_, warning := ffprobeMetadata("anything.mp3")

	if warning != "ffprobe output could not be parsed" {
		t.Fatalf("warning = %q", warning)
	}
}
