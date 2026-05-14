package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestUniquePathUsesCounterForReservedCollision(t *testing.T) {
	reserved := map[string]struct{}{}
	path := filepath.Join(t.TempDir(), "customer-list.csv")

	first := UniquePath(path, reserved)
	second := UniquePath(path, reserved)

	if first != path {
		t.Fatalf("first path = %q, want %q", first, path)
	}
	want := filepath.Join(filepath.Dir(path), "customer-list-2.csv")
	if second != want {
		t.Fatalf("second path = %q, want %q", second, want)
	}
}

func TestUniquePathSkipsExistingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "customer-list.csv")
	if err := os.WriteFile(path, []byte("existing"), 0644); err != nil {
		t.Fatal(err)
	}

	got := UniquePath(path, map[string]struct{}{})

	want := filepath.Join(dir, "customer-list-2.csv")
	if got != want {
		t.Fatalf("path = %q, want %q", got, want)
	}
}

func TestUniquePlannedPathIgnoresExistingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "customer-list.csv")
	if err := os.WriteFile(path, []byte("existing"), 0644); err != nil {
		t.Fatal(err)
	}

	reserved := map[string]struct{}{}
	first := UniquePlannedPath(path, reserved)
	second := UniquePlannedPath(path, reserved)

	if first != path {
		t.Fatalf("first path = %q, want %q", first, path)
	}
	want := filepath.Join(dir, "customer-list-2.csv")
	if second != want {
		t.Fatalf("second path = %q, want %q", second, want)
	}
}
