package siegfried

import (
	"context"

	"github.com/djblackett/bootdev-hackathon/internal/evidence"
	"github.com/djblackett/bootdev-hackathon/internal/tools"
)

type Extractor struct{}

func (Extractor) Name() evidence.EvidenceSource { return evidence.SourceSiegfried }

func (Extractor) Available(ctx context.Context) bool {
	return tools.Available("sf")
}

func (Extractor) Extract(ctx context.Context, path string) (evidence.PartialEvidence, error) {
	return evidence.PartialEvidence{
		Source: evidence.SourceSiegfried,
		Evidence: evidence.FileEvidence{
			Path:     path,
			Sources:  []evidence.EvidenceSource{evidence.SourceSiegfried},
			Warnings: []string{"siegfried extractor is planned for a later milestone"},
		},
	}, nil
}
