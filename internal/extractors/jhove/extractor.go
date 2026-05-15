package jhove

import (
	"context"

	"github.com/djblackett/bootdev-hackathon/internal/evidence"
	"github.com/djblackett/bootdev-hackathon/internal/tools"
)

type Extractor struct{}

func (Extractor) Name() evidence.EvidenceSource { return evidence.SourceJHOVE }

func (Extractor) Available(ctx context.Context) bool {
	return tools.Available("jhove")
}

func (Extractor) Extract(ctx context.Context, path string) (evidence.PartialEvidence, error) {
	return evidence.PartialEvidence{
		Source: evidence.SourceJHOVE,
		Evidence: evidence.FileEvidence{
			Path:     path,
			Sources:  []evidence.EvidenceSource{evidence.SourceJHOVE},
			Warnings: []string{"jhove validation is planned for a later milestone"},
		},
	}, nil
}
