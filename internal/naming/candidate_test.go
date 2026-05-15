package naming

import (
	"strings"
	"testing"

	"github.com/djblackett/bootdev-hackathon/internal/evidence"
)

func TestGeneratePenalizesGenericMetadataTitle(t *testing.T) {
	got := Generate(evidence.FileEvidence{
		Path:         "recovered/file0007",
		DetectedMIME: "application/pdf",
		Extension:    ".pdf",
		Metadata:     map[string]string{"title": "Document1"},
		TextPreview:  "2022-03-15 T4 tax slip employment income",
		TextSignals:  []string{"T4 tax slip employment income"},
		Sources:      []evidence.EvidenceSource{evidence.SourceNativeMIME},
	}, 7)

	if strings.Contains(got.Filename, "document1") {
		t.Fatalf("generic title dominated filename: %+v", got)
	}
	if !strings.Contains(got.Filename, "tax-slip") {
		t.Fatalf("filename = %q, want tax-slip signal", got.Filename)
	}
}

func TestGenerateUsesExtensionFallback(t *testing.T) {
	got := Generate(evidence.FileEvidence{
		Path:         "recovered/file000421",
		DetectedMIME: "application/pdf",
		Extension:    ".pdf",
		Sources:      []evidence.EvidenceSource{evidence.SourceNativeMIME},
	}, 1)

	if got.Filename != "unknown-pdf_000421.pdf" {
		t.Fatalf("Filename = %q, want unknown-pdf_000421.pdf", got.Filename)
	}
	if got.Confidence != ConfidenceLow {
		t.Fatalf("Confidence = %q, want low", got.Confidence)
	}
}

func TestGenerateRejectsRandomTextSignal(t *testing.T) {
	got := Generate(evidence.FileEvidence{
		Path:         "recovered/random.txt",
		DetectedMIME: "text/plain",
		Extension:    ".txt",
		TextPreview:  "OeYV/jjq0pT9Jn4oiiJG\nUHmmYZszQjxHikWZF8lCoisYzBgiJEuZoRpmcYzMQ8R",
		TextSignals:  []string{"OeYV/jjq0pT9Jn4oiiJG"},
		Sources:      []evidence.EvidenceSource{evidence.SourceNativeMIME},
	}, 12)

	if got.Filename != "unknown-text_000012.txt" {
		t.Fatalf("Filename = %q, want unknown-text_000012.txt", got.Filename)
	}
	if got.Confidence != ConfidenceLow {
		t.Fatalf("Confidence = %q, want low", got.Confidence)
	}
}

func TestGenerateUsesImageTimestampAndCameraEvidence(t *testing.T) {
	got := Generate(evidence.FileEvidence{
		Path:         "recovered/photo",
		DetectedMIME: "image/jpeg",
		Extension:    ".jpg",
		Image: &evidence.ImageEvidence{
			TakenAt:     "2021:08:14 16:22:09",
			CameraMake:  "Apple",
			CameraModel: "iPhone 12",
		},
		Sources: []evidence.EvidenceSource{evidence.SourceNativeMIME, evidence.SourceExifTool},
	}, 3)

	if got.Filename != "2021-08-14_16-22-09_apple-iphone-12.jpg" {
		t.Fatalf("Filename = %q", got.Filename)
	}
	if got.Confidence != ConfidenceMedium {
		t.Fatalf("Confidence = %q, want medium", got.Confidence)
	}
}

func TestGenerateUsesMediaTitleAndCreationDate(t *testing.T) {
	got := Generate(evidence.FileEvidence{
		Path:         "recovered/video",
		DetectedMIME: "video/mp4",
		Extension:    ".mp4",
		Metadata: map[string]string{
			"title":         "Family Video",
			"creation_time": "2020-10-03T18:12:01.000000Z",
		},
		Media:   &evidence.MediaEvidence{DurationSeconds: 12.5, Codec: "h264", Width: 1920, Height: 1080},
		Sources: []evidence.EvidenceSource{evidence.SourceNativeMIME, evidence.SourceFFProbe},
	}, 4)

	if got.Filename != "2020-10-03_family-video.mp4" {
		t.Fatalf("Filename = %q", got.Filename)
	}
	if got.Confidence != ConfidenceHigh {
		t.Fatalf("Confidence = %q, want high", got.Confidence)
	}
}

func TestIsGenericTitle(t *testing.T) {
	for _, value := range []string{"untitled", "Document1", "Microsoft Word - Document", "scan 0001"} {
		if !IsGenericTitle(value) {
			t.Fatalf("%q should be generic", value)
		}
	}
	if IsGenericTitle("Monthly Statement April") {
		t.Fatal("useful title classified as generic")
	}
}
