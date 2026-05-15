package jhove

import (
	"context"
	"os/exec"

	"github.com/djblackett/bootdev-hackathon/internal/evidence"
)

type Extractor struct{}

func (Extractor) Name() evidence.EvidenceSource { return evidence.SourceJHOVE }

func (Extractor) Available(ctx context.Context) bool {
	_, err := exec.LookPath("jhove")
	return err == nil
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
