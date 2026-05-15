package siegfried

import (
	"context"
	"os/exec"

	"github.com/djblackett/bootdev-hackathon/internal/evidence"
)

type Extractor struct{}

func (Extractor) Name() evidence.EvidenceSource { return evidence.SourceSiegfried }

func (Extractor) Available(ctx context.Context) bool {
	_, err := exec.LookPath("sf")
	return err == nil
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
