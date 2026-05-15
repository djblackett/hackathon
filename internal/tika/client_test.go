package tika

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

func TestClientExtractFileParsesTextAndMetadata(t *testing.T) {
	client, err := NewClientWithHTTPClient("http://tika.test", &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		switch r.URL.Path {
		case "/tika":
			if r.Method != http.MethodPut {
				t.Fatalf("method = %s, want PUT", r.Method)
			}
			return response(http.StatusOK, "Quarterly Revenue Review\nDetailed body"), nil
		case "/meta":
			return response(http.StatusOK, `{"dc:title":"Revenue Review","Author":["Ada","Grace"],"xmpTPg:NPages":3}`), nil
		default:
			return response(http.StatusNotFound, "not found"), nil
		}
	})})
	if err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(t.TempDir(), "doc.bin")
	if err := os.WriteFile(path, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	got, err := client.ExtractFile(context.Background(), path)
	if err != nil {
		t.Fatal(err)
	}

	if got.Text != "Quarterly Revenue Review\nDetailed body" {
		t.Fatalf("Text = %q", got.Text)
	}
	if got.Metadata["dc:title"] != "Revenue Review" {
		t.Fatalf("metadata title = %q", got.Metadata["dc:title"])
	}
	if got.Metadata["Author"] != "Ada, Grace" {
		t.Fatalf("metadata author = %q", got.Metadata["Author"])
	}
	if got.Metadata["xmpTPg:NPages"] != "3" {
		t.Fatalf("metadata pages = %q", got.Metadata["xmpTPg:NPages"])
	}
}

func TestClientExtractFileReturnsPartialTextWhenMetadataFails(t *testing.T) {
	client, err := NewClientWithHTTPClient("http://tika.test", &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		switch r.URL.Path {
		case "/tika":
			return response(http.StatusOK, "Readable content"), nil
		case "/meta":
			return response(http.StatusInternalServerError, "no metadata"), nil
		default:
			return response(http.StatusNotFound, "not found"), nil
		}
	})})
	if err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(t.TempDir(), "doc.bin")
	if err := os.WriteFile(path, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	got, err := client.ExtractFile(context.Background(), path)
	if err != nil {
		t.Fatal(err)
	}
	if got.Text != "Readable content" {
		t.Fatalf("Text = %q", got.Text)
	}
	if len(got.Warnings) != 1 {
		t.Fatalf("Warnings = %+v, want metadata warning", got.Warnings)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func response(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Status:     http.StatusText(status),
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(bytes.NewBufferString(body)),
	}
}
