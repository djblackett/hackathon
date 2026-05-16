package app

import (
	"bytes"
	"context"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/djblackett/bootdev-hackathon/internal/evidence"
	"github.com/djblackett/bootdev-hackathon/internal/plan"
	"github.com/djblackett/bootdev-hackathon/internal/tika"
	"github.com/djblackett/bootdev-hackathon/internal/tools"
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

	var parsed struct {
		Version int    `json:"version"`
		Root    string `json:"root"`
		Items   []struct {
			OldPath       string          `json:"oldPath"`
			SuggestedPath string          `json:"suggestedPath"`
			Confidence    string          `json:"confidence"`
			Score         float64         `json:"score"`
			Evidence      json.RawMessage `json:"evidence"`
			Reasons       []string        `json:"reasons"`
		} `json:"items"`
	}
	if err := json.Unmarshal(a, &parsed); err != nil {
		t.Fatal(err)
	}
	if parsed.Version != 1 || parsed.Root != root || len(parsed.Items) != 1 {
		t.Fatalf("unexpected plan envelope: %+v", parsed)
	}
	if parsed.Items[0].OldPath == "" || parsed.Items[0].SuggestedPath == "" || parsed.Items[0].Confidence == "" || len(parsed.Items[0].Reasons) == 0 {
		t.Fatalf("missing required item fields: %+v", parsed.Items[0])
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

func TestScanAppliesMaxTextPreviewToNativeEvidence(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "long.txt"), []byte(strings.Repeat("word ", 100)), 0644); err != nil {
		t.Fatal(err)
	}

	got, err := Scan(context.Background(), ScanConfig{
		Root:           root,
		OutPath:        filepath.Join(t.TempDir(), "plan.json"),
		Hash:           false,
		MaxTextPreview: 25,
		NoTimestamp:    true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Items) != 1 {
		t.Fatalf("items = %d, want 1", len(got.Items))
	}
	if len(got.Items[0].Evidence.TextPreview) > 25 {
		t.Fatalf("text preview length = %d, want <= 25: %q", len(got.Items[0].Evidence.TextPreview), got.Items[0].Evidence.TextPreview)
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

func TestScanMissingSiegfriedAddsWarning(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "note.txt"), []byte("Quarterly revenue review"), 0644); err != nil {
		t.Fatal(err)
	}

	got, err := Scan(context.Background(), ScanConfig{
		Root:         root,
		OutPath:      filepath.Join(t.TempDir(), "plan.json"),
		UseSiegfried: true,
		Hash:         false,
		NoTimestamp:  true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Items) != 1 {
		t.Fatalf("items = %d, want 1", len(got.Items))
	}
	found := false
	for _, warning := range got.Items[0].Warnings {
		if strings.Contains(warning, "siegfried unavailable") {
			found = true
		}
	}
	if !found && !siegfriedInstalled() {
		t.Fatalf("missing siegfried unavailable warning: %+v", got.Items[0].Warnings)
	}
}

func TestScanMissingExifToolAddsWarningForImage(t *testing.T) {
	root := t.TempDir()
	writePNG(t, filepath.Join(root, "photo"))

	got, err := Scan(context.Background(), ScanConfig{
		Root:        root,
		OutPath:     filepath.Join(t.TempDir(), "plan.json"),
		UseExifTool: true,
		Hash:        false,
		NoTimestamp: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Items) != 1 {
		t.Fatalf("items = %d, want 1", len(got.Items))
	}
	found := false
	for _, warning := range got.Items[0].Warnings {
		if strings.Contains(warning, "exiftool unavailable") {
			found = true
		}
	}
	if !found && !exifToolInstalled() {
		t.Fatalf("missing exiftool unavailable warning: %+v", got.Items[0].Warnings)
	}
}

func TestScanMissingFFProbeAddsWarningForMedia(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "clip.mp4"), []byte{0x00, 0x01, 0x02, 0x03}, 0644); err != nil {
		t.Fatal(err)
	}

	got, err := Scan(context.Background(), ScanConfig{
		Root:        root,
		OutPath:     filepath.Join(t.TempDir(), "plan.json"),
		UseFFProbe:  true,
		Hash:        false,
		NoTimestamp: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Items) != 1 {
		t.Fatalf("items = %d, want 1", len(got.Items))
	}
	foundUnavailable := false
	foundError := false
	for _, warning := range got.Items[0].Warnings {
		if strings.Contains(warning, "ffprobe unavailable") {
			foundUnavailable = true
		}
		if strings.Contains(warning, "ffprobe failed") {
			foundError = true
		}
	}
	if !ffprobeInstalled() && !foundUnavailable {
		t.Fatalf("missing ffprobe unavailable warning: %+v", got.Items[0].Warnings)
	}
	if ffprobeInstalled() && !foundError && len(got.Items[0].Evidence.Errors) == 0 {
		t.Fatalf("expected ffprobe failure/error for invalid media fixture: %+v", got.Items[0])
	}
}

func TestScanMissingJHOVEAddsValidationWarning(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "note.txt"), []byte("Quarterly revenue review"), 0644); err != nil {
		t.Fatal(err)
	}

	got, err := Scan(context.Background(), ScanConfig{
		Root:        root,
		OutPath:     filepath.Join(t.TempDir(), "plan.json"),
		Validate:    true,
		Hash:        false,
		NoTimestamp: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Items) != 1 {
		t.Fatalf("items = %d, want 1", len(got.Items))
	}
	found := false
	for _, warning := range got.Items[0].Warnings {
		if strings.Contains(warning, "jhove unavailable") {
			found = true
		}
	}
	if !found && !jhoveInstalled() {
		t.Fatalf("missing jhove unavailable warning: %+v", got.Items[0].Warnings)
	}
}

func TestScanMissingTesseractAddsOCRWarningForImage(t *testing.T) {
	root := t.TempDir()
	writePNG(t, filepath.Join(root, "scan"))

	got, err := Scan(context.Background(), ScanConfig{
		Root:        root,
		OutPath:     filepath.Join(t.TempDir(), "plan.json"),
		UseOCR:      true,
		OCRLang:     "eng",
		Hash:        false,
		NoTimestamp: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Items) != 1 {
		t.Fatalf("items = %d, want 1", len(got.Items))
	}
	found := false
	for _, warning := range got.Items[0].Warnings {
		if strings.Contains(warning, "tesseract unavailable") {
			found = true
		}
	}
	if !found && !tesseractInstalled() {
		t.Fatalf("missing tesseract unavailable warning: %+v", got.Items[0].Warnings)
	}
}

func TestScanRecordsBadFileWithoutFailingBatch(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "good.txt"), []byte("Quarterly revenue review"), 0644); err != nil {
		t.Fatal(err)
	}
	badPath := filepath.Join(root, "broken-link")
	if err := os.Symlink(filepath.Join(root, "missing-target"), badPath); err != nil {
		t.Skipf("symlink not supported in this environment: %v", err)
	}

	got, err := Scan(context.Background(), ScanConfig{
		Root:        root,
		OutPath:     filepath.Join(t.TempDir(), "plan.json"),
		Hash:        true,
		NoTimestamp: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Items) != 2 {
		t.Fatalf("items = %d, want 2", len(got.Items))
	}
	byBase := map[string]plan.Item{}
	for _, item := range got.Items {
		byBase[filepath.Base(item.OldPath)] = item
	}
	bad := byBase["broken-link"]
	if len(bad.Evidence.Errors) == 0 {
		t.Fatalf("bad file should include evidence errors: %+v", bad.Evidence)
	}
	if len(bad.Warnings) == 0 {
		t.Fatalf("bad file should include item warnings: %+v", bad)
	}
	if _, ok := byBase["good.txt"]; !ok {
		t.Fatalf("good file missing from plan: %+v", byBase)
	}
}

func TestScanUsesTikaMetadataAndTextEvidence(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "file000421")
	if err := os.WriteFile(path, []byte("%PDF-1.7\nbody"), 0644); err != nil {
		t.Fatal(err)
	}

	client, err := tika.NewClientWithHTTPClient("http://tika.test", &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		switch r.URL.Path {
		case "/tika":
			if r.Method == http.MethodGet {
				return testResponse(http.StatusOK, "ok"), nil
			}
			if r.Method != http.MethodPut {
				t.Fatalf("unexpected method for /tika: %s", r.Method)
			}
			return testResponse(http.StatusOK, "Monthly statement\nAccount summary for April 2022"), nil
		case "/meta":
			if r.Method != http.MethodPut {
				t.Fatalf("unexpected method for /meta: %s", r.Method)
			}
			return testResponse(http.StatusOK, `{"Content-Type":"application/pdf","dc:title":"Statement April 2022","dcterms:created":"2022-04-30T12:00:00Z"}`), nil
		default:
			return testResponse(http.StatusNotFound, "not found"), nil
		}
	})})
	if err != nil {
		t.Fatal(err)
	}

	got, err := Scan(context.Background(), ScanConfig{
		Root:        root,
		OutPath:     filepath.Join(t.TempDir(), "plan.json"),
		TikaClient:  client,
		Hash:        false,
		NoTimestamp: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Items) != 1 {
		t.Fatalf("items = %d, want 1", len(got.Items))
	}
	item := got.Items[0]
	if filepath.Base(item.SuggestedPath) != "2022-04-30_statement-april-2022.pdf" {
		t.Fatalf("suggested path = %q", item.SuggestedPath)
	}
	if item.Confidence != "high" {
		t.Fatalf("confidence = %q, want high", item.Confidence)
	}
	if !containsSource(item.Evidence.Sources, "tika") {
		t.Fatalf("sources = %+v, want tika", item.Evidence.Sources)
	}
	if item.Evidence.Metadata["dc:title"] != "Statement April 2022" {
		t.Fatalf("missing tika metadata: %+v", item.Evidence.Metadata)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func testResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Status:     http.StatusText(status),
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(bytes.NewBufferString(body)),
	}
}

func containsSource(sources []evidence.EvidenceSource, want string) bool {
	for _, source := range sources {
		if string(source) == want {
			return true
		}
	}
	return false
}

func siegfriedInstalled() bool {
	return tools.Available("sf")
}

func exifToolInstalled() bool {
	return tools.Available("exiftool")
}

func ffprobeInstalled() bool {
	return tools.Available("ffprobe")
}

func jhoveInstalled() bool {
	return tools.Available("jhove")
}

func tesseractInstalled() bool {
	return tools.Available("tesseract")
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
