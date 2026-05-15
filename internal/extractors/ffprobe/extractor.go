package ffprobe

import (
	"context"
	"os/exec"

	"github.com/djblackett/bootdev-hackathon/internal/evidence"
)

type Extractor struct{}

func (Extractor) Name() evidence.EvidenceSource { return evidence.SourceFFProbe }

func (Extractor) Available(ctx context.Context) bool {
	_, err := exec.LookPath("ffprobe")
	return err == nil
}

func (Extractor) Extract(ctx context.Context, path string) (evidence.PartialEvidence, error) {
	return evidence.PartialEvidence{
		Source: evidence.SourceFFProbe,
		Evidence: evidence.FileEvidence{
			Path:     path,
			Sources:  []evidence.EvidenceSource{evidence.SourceFFProbe},
			Warnings: []string{"ffprobe extractor is planned for a later milestone"},
		},
	}, nil
}
