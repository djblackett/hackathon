package extractors

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/djblackett/bootdev-hackathon/internal/tika"
)

func TestWalkInfoUsesTikaForUnknownSupportedType(t *testing.T) {
	defer ConfigureTika("")

	client := newTestTikaClient(t, func(r *http.Request) (*http.Response, error) {
		switch r.URL.Path {
		case "/tika":
			return tikaResponse(http.StatusOK, "Fallback Document Title\nBody text"), nil
		case "/meta":
			return tikaResponse(http.StatusOK, `{"dc:title":"Fallback Metadata Title"}`), nil
		default:
			return tikaResponse(http.StatusNotFound, "not found"), nil
		}
	})
	configureTikaClientForTest(client)

	dir := t.TempDir()
	path := filepath.Join(dir, "mystery.bin")
	if err := os.WriteFile(path, []byte{0, 1, 2, 3}, 0644); err != nil {
		t.Fatal(err)
	}

	var infos []ExtractedFileInfo
	err := WalkInfo(dir, map[string]struct{}{"unknown": {}}, func(info ExtractedFileInfo) error {
		infos = append(infos, info)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(infos) != 1 {
		t.Fatalf("infos = %d, want 1", len(infos))
	}
	if !hasSample(infos[0], "tika-title", "Fallback Metadata Title") {
		t.Fatalf("missing tika title sample: %+v", infos[0].TextSamples)
	}
	if !hasSample(infos[0], "tika-first-text", "Fallback Document Title") {
		t.Fatalf("missing tika first text sample: %+v", infos[0].TextSamples)
	}
}

func TestWalkInfoDoesNotCallTikaWhenExtractorHasEvidence(t *testing.T) {
	defer ConfigureTika("")

	called := false
	client := newTestTikaClient(t, func(r *http.Request) (*http.Response, error) {
		called = true
		return tikaResponse(http.StatusInternalServerError, "unexpected"), nil
	})
	configureTikaClientForTest(client)

	dir := t.TempDir()
	path := filepath.Join(dir, "note.txt")
	if err := os.WriteFile(path, []byte("hello there\ngeneral kenobi\n"), 0644); err != nil {
		t.Fatal(err)
	}

	err := WalkInfo(dir, map[string]struct{}{"text": {}}, func(info ExtractedFileInfo) error {
		if !hasSample(info, "short-text-note", "hello there general kenobi") {
			t.Fatalf("missing text extractor evidence: %+v", info.TextSamples)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if called {
		t.Fatal("tika should not be called when local extractor has evidence")
	}
}

func TestWalkInfoRecordsTikaFailureAsWarning(t *testing.T) {
	defer ConfigureTika("")

	client := newTestTikaClient(t, func(r *http.Request) (*http.Response, error) {
		return tikaResponse(http.StatusInternalServerError, "failed"), nil
	})
	configureTikaClientForTest(client)

	dir := t.TempDir()
	path := filepath.Join(dir, "mystery.bin")
	if err := os.WriteFile(path, []byte{0, 1, 2, 3}, 0644); err != nil {
		t.Fatal(err)
	}

	var info ExtractedFileInfo
	err := WalkInfo(dir, map[string]struct{}{"unknown": {}}, func(got ExtractedFileInfo) error {
		info = got
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if !containsWarning(info.Warnings, "tika extraction failed") {
		t.Fatalf("warnings = %+v, want tika failure", info.Warnings)
	}
}

func configureTikaClientForTest(client *tika.Client) {
	tikaClient = client
}

func newTestTikaClient(t *testing.T, fn func(*http.Request) (*http.Response, error)) *tika.Client {
	t.Helper()
	client, err := tika.NewClientWithHTTPClient("http://tika.test", &http.Client{Transport: testRoundTripFunc(fn)})
	if err != nil {
		t.Fatal(err)
	}
	return client
}

type testRoundTripFunc func(*http.Request) (*http.Response, error)

func (f testRoundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func tikaResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Status:     http.StatusText(status),
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(bytes.NewBufferString(body)),
	}
}
