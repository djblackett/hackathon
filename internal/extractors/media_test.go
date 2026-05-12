package extractors

import "testing"

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
