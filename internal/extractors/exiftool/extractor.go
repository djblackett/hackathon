package exiftool

import (
	"context"

	"github.com/djblackett/bootdev-hackathon/internal/evidence"
	"github.com/djblackett/bootdev-hackathon/internal/tools"
)

type Extractor struct{}

func (Extractor) Name() evidence.EvidenceSource { return evidence.SourceExifTool }

func (Extractor) Available(ctx context.Context) bool {
	return tools.Available("exiftool")
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
