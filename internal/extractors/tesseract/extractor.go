package tesseract

import (
	"context"
	"os/exec"

	"github.com/djblackett/bootdev-hackathon/internal/evidence"
)

type Extractor struct{}

func (Extractor) Name() evidence.EvidenceSource { return evidence.SourceTesseract }

func (Extractor) Available(ctx context.Context) bool {
	_, err := exec.LookPath("tesseract")
	return err == nil
}

func (Extractor) Extract(ctx context.Context, path string) (evidence.PartialEvidence, error) {
	return evidence.PartialEvidence{
		Source: evidence.SourceTesseract,
		Evidence: evidence.FileEvidence{
			Path:     path,
			Sources:  []evidence.EvidenceSource{evidence.SourceTesseract},
			Warnings: []string{"tesseract OCR is planned for a later milestone"},
		},
	}, nil
}
