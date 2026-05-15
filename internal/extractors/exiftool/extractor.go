package exiftool

import (
	"context"
	"os/exec"

	"github.com/djblackett/bootdev-hackathon/internal/evidence"
)

type Extractor struct{}

func (Extractor) Name() evidence.EvidenceSource { return evidence.SourceExifTool }

func (Extractor) Available(ctx context.Context) bool {
	_, err := exec.LookPath("exiftool")
	return err == nil
}

func (Extractor) Extract(ctx context.Context, path string) (evidence.PartialEvidence, error) {
	return evidence.PartialEvidence{
		Source: evidence.SourceExifTool,
		Evidence: evidence.FileEvidence{
			Path:     path,
			Sources:  []evidence.EvidenceSource{evidence.SourceExifTool},
			Warnings: []string{"exiftool extractor is planned for a later milestone"},
		},
	}, nil
}
