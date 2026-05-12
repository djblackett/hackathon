package extractors

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestImageExtractorUsesPropertiesWithoutExiftool(t *testing.T) {
	t.Setenv("PATH", "")

	path := filepath.Join(t.TempDir(), "photo.png")
	out, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	img := image.NewRGBA(image.Rect(0, 0, 2, 3))
	img.Set(0, 0, color.RGBA{R: 255, A: 255})
	if err := png.Encode(out, img); err != nil {
		t.Fatal(err)
	}
	if err := out.Close(); err != nil {
		t.Fatal(err)
	}

	info, err := imageExtractor{}.ExtractInfo(path)
	if err != nil {
		t.Fatal(err)
	}

	if info.SuggestedExtension != "png" {
		t.Fatalf("SuggestedExtension = %q, want png", info.SuggestedExtension)
	}
	if !hasSample(info, "image-properties", "png image 2x3") {
		t.Fatalf("missing image properties sample: %+v", info.TextSamples)
	}
	if len(info.Warnings) == 0 {
		t.Fatal("expected exiftool warning when PATH is empty")
	}
}
