package app

import (
	"context"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/djblackett/bootdev-hackathon/internal/plan"
)

func TestScanWritesDeterministicPlanWithoutTimestamp(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "file000421"), []byte("%PDF-1.7\nbody"), 0644); err != nil {
		t.Fatal(err)
	}
	out1 := filepath.Join(t.TempDir(), "plan1.json")
	out2 := filepath.Join(t.TempDir(), "plan2.json")

	cfg := ScanConfig{Root: root, OutPath: out1, Hash: true, NoTimestamp: true}
	if _, err := Scan(context.Background(), cfg); err != nil {
		t.Fatal(err)
	}
	cfg.OutPath = out2
	if _, err := Scan(context.Background(), cfg); err != nil {
		t.Fatal(err)
	}

	a, err := os.ReadFile(out1)
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(out2)
	if err != nil {
		t.Fatal(err)
	}
	if string(a) != string(b) {
		t.Fatalf("plans differ\n%s\n---\n%s", a, b)
	}
	if strings.Contains(string(a), "generatedAt") {
		t.Fatalf("plan should omit generatedAt with --no-timestamp:\n%s", a)
	}
}

func TestScanExtensionFallbacks(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "pdf-file"), []byte("%PDF-1.7\nbody"), 0644); err != nil {
		t.Fatal(err)
	}
	writePNG(t, filepath.Join(root, "png-file"))
	out := filepath.Join(t.TempDir(), "plan.json")

	got, err := Scan(context.Background(), ScanConfig{Root: root, OutPath: out, Hash: false, NoTimestamp: true})
	if err != nil {
		t.Fatal(err)
	}
	byBase := map[string]plan.Item{}
	for _, item := range got.Items {
		byBase[filepath.Base(item.OldPath)] = item
	}
	if byBase["pdf-file"].Evidence.Extension != ".pdf" {
		t.Fatalf("pdf extension = %q", byBase["pdf-file"].Evidence.Extension)
	}
	if byBase["png-file"].Evidence.Extension != ".png" {
		t.Fatalf("png extension = %q", byBase["png-file"].Evidence.Extension)
	}
}

func TestScanMissingTikaServerAddsWarning(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "note"), []byte("Quarterly revenue review"), 0644); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(t.TempDir(), "plan.json")

	got, err := Scan(context.Background(), ScanConfig{
		Root:        root,
		OutPath:     out,
		TikaURL:     "http://127.0.0.1:1",
		TikaTimeout: 10 * time.Millisecond,
		Hash:        false,
		NoTimestamp: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	var written plan.Plan
	if err := json.Unmarshal(data, &written); err != nil {
		t.Fatal(err)
	}
	if len(got.Items) != 1 || len(written.Items) != 1 {
		t.Fatalf("got %d/%d items", len(got.Items), len(written.Items))
	}
	if len(got.Items[0].Warnings) == 0 || !strings.Contains(got.Items[0].Warnings[0], "tika unavailable") {
		t.Fatalf("missing tika warning: %+v", got.Items[0].Warnings)
	}
}

func writePNG(t *testing.T, path string) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.RGBA{R: 255, A: 255})
	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}
}
