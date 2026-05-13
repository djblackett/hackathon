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

func TestMeaningfulMediaBasenameRejectsCameraTimestamp(t *testing.T) {
	cases := []string{
		"VID_20240201_103312.mp4",
		"PXL_20240201_103312000.mp4",
		"IMG_0001.mov",
	}
	for _, path := range cases {
		if got := meaningfulMediaBasename(path); got != "" {
			t.Fatalf("%s basename = %q, want empty", path, got)
		}
	}
}

func TestMeaningfulMediaBasenameRejectsRandomName(t *testing.T) {
	if got := meaningfulMediaBasename("xqzprtnmlk884422.mp4"); got != "" {
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
	if got := mediaFilenameScore("alice"); got < 0.75 {
		t.Fatalf("alice score = %.2f, want copy-threshold friendly score", got)
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

func TestMediaExtractorUsesBasenameWhenFFProbeUnavailable(t *testing.T) {
	t.Setenv("PATH", "")
	path := filepath.Join(t.TempDir(), "alice.mp4")
	if err := os.WriteFile(path, []byte{0x00, 0x01}, 0644); err != nil {
		t.Fatal(err)
	}

	info, err := mediaExtractor{}.ExtractInfo(path)
	if err != nil {
		t.Fatal(err)
	}

	if !hasSample(info, "media-filename", "alice") {
		t.Fatalf("missing basename sample: %+v", info.TextSamples)
	}
	if len(info.Warnings) == 0 {
		t.Fatal("expected ffprobe warning")
	}
}

func TestMediaExtractorUsesTags(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell script fake ffprobe is POSIX-only")
	}

	dir := t.TempDir()
	ffprobe := filepath.Join(dir, "ffprobe")
	script := `#!/bin/sh
printf '%s\n' '{"format":{"format_name":"mp3","format_long_name":"MP3 audio","duration":"123.45","tags":{"title":"Late Night Drive","artist":"Retro Metro","album":"City Lights"}}}'
`
	if err := os.WriteFile(ffprobe, []byte(script), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir)

	path := filepath.Join(t.TempDir(), "track001.mp3")
	if err := os.WriteFile(path, []byte{0x00, 0x01}, 0644); err != nil {
		t.Fatal(err)
	}

	info, err := mediaExtractor{}.ExtractInfo(path)
	if err != nil {
		t.Fatal(err)
	}

	if !hasSample(info, "media-tags", "Late Night Drive Retro Metro City Lights") {
		t.Fatalf("missing tag sample: %+v", info.TextSamples)
	}
	if info.Metadata["artist"] != "Retro Metro" {
		t.Fatalf("artist metadata = %q", info.Metadata["artist"])
	}
}
