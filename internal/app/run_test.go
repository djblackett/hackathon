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
