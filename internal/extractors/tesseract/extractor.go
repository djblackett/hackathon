package tesseract

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/djblackett/bootdev-hackathon/internal/evidence"
	"github.com/djblackett/bootdev-hackathon/internal/tools"
)

type Extractor struct {
	Timeout        time.Duration
	Lang           string
	MaxTextPreview int
}

func (Extractor) Name() evidence.EvidenceSource { return evidence.SourceTesseract }

func (Extractor) Available(ctx context.Context) bool {
	return tools.Available("tesseract")
}

func (e Extractor) Extract(ctx context.Context, path string) (evidence.PartialEvidence, error) {
	lang := strings.TrimSpace(e.Lang)
	if lang == "" {
		lang = "eng"
	}
	result, err := tools.Run(ctx, e.Timeout, "tesseract", path, "stdout", "-l", lang)
	if err != nil {
		message := strings.TrimSpace(string(result.Stderr))
		if message == "" {
			message = err.Error()
		}
		return evidence.PartialEvidence{}, fmt.Errorf("tesseract failed: %s", message)
	}
	ev := Parse(path, result.Stdout, e.MaxTextPreview)
	ev.Metadata["tesseract_lang"] = lang
	return evidence.PartialEvidence{Source: evidence.SourceTesseract, Evidence: ev}, nil
}

func Parse(path string, data []byte, maxPreview int) evidence.FileEvidence {
	if maxPreview <= 0 {
		maxPreview = 2000
	}
	text := strings.Join(strings.Fields(string(data)), " ")
	ev := evidence.FileEvidence{
		Path:     path,
		Metadata: map[string]string{"ocr": "true"},
		Sources:  []evidence.EvidenceSource{evidence.SourceTesseract},
		Warnings: []string{"OCR evidence can be noisy; review suggested names carefully"},
	}
	if text == "" {
		ev.Warnings = append(ev.Warnings, "tesseract returned no text")
		return ev
	}
	ev.TextPreview = trimText(text, maxPreview)
	if signal := firstSignal(text); signal != "" {
		ev.TextSignals = append(ev.TextSignals, signal)
	}
	return ev
}

func firstSignal(text string) string {
	if len(text) <= 240 {
		return strings.TrimSpace(text)
	}
	return strings.TrimSpace(text[:240])
}

func trimText(text string, max int) string {
	if max <= 0 || len(text) <= max {
		return text
	}
	trimmed := strings.TrimSpace(text[:max])
	if idx := strings.LastIndex(trimmed, " "); idx > 0 {
		return strings.TrimSpace(trimmed[:idx])
	}
	return trimmed
}
