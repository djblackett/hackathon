package tikaextractor

import (
	"context"
	"strings"

	"github.com/djblackett/bootdev-hackathon/internal/evidence"
	"github.com/djblackett/bootdev-hackathon/internal/tika"
)

type Extractor struct {
	Client         *tika.Client
	MaxTextPreview int
}

func (Extractor) Name() evidence.EvidenceSource { return evidence.SourceTika }

func (e Extractor) Available(ctx context.Context) bool {
	if e.Client == nil {
		return false
	}
	return e.Client.Health(ctx) == nil
}

func (e Extractor) Extract(ctx context.Context, path string) (evidence.PartialEvidence, error) {
	ev := evidence.FileEvidence{
		Path:     path,
		Metadata: map[string]string{},
		Sources:  []evidence.EvidenceSource{evidence.SourceTika},
	}
	if e.Client == nil {
		ev.Warnings = append(ev.Warnings, "tika client is not configured")
		return evidence.PartialEvidence{Source: evidence.SourceTika, Evidence: ev}, nil
	}

	extracted, err := e.Client.ExtractFile(ctx, path)
	if err != nil {
		return evidence.PartialEvidence{}, err
	}
	for key, value := range extracted.Metadata {
		if strings.TrimSpace(value) == "" {
			continue
		}
		ev.Metadata[key] = value
	}
	ev.Warnings = append(ev.Warnings, extracted.Warnings...)

	if mime := firstMetadataValue(extracted.Metadata, "Content-Type", "content-type", "dc:format"); mime != "" {
		ev.DetectedMIME = mime
	}
	if ext := extensionFromMetadata(extracted.Metadata); ext != "" {
		ev.Extension = ext
	}
	text := strings.TrimSpace(extracted.Text)
	if text != "" {
		limit := e.MaxTextPreview
		if limit <= 0 {
			limit = 2000
		}
		ev.TextPreview = trimPreview(text, limit)
		if first := firstMeaningfulLine(text); first != "" {
			ev.TextSignals = append(ev.TextSignals, first)
		}
	}
	if title := firstMetadataValue(extracted.Metadata, "title", "dc:title", "pdf:docinfo:title", "resourceName"); title != "" {
		ev.TextSignals = append([]string{title}, ev.TextSignals...)
	}
	return evidence.PartialEvidence{Source: evidence.SourceTika, Evidence: ev}, nil
}

func firstMetadataValue(metadata map[string]string, keys ...string) string {
	for _, want := range keys {
		for key, value := range metadata {
			if strings.EqualFold(key, want) && strings.TrimSpace(value) != "" {
				return strings.TrimSpace(value)
			}
		}
	}
	return ""
}

func extensionFromMetadata(metadata map[string]string) string {
	mime := strings.ToLower(firstMetadataValue(metadata, "Content-Type", "content-type", "dc:format"))
	switch {
	case strings.Contains(mime, "pdf"):
		return ".pdf"
	case strings.Contains(mime, "wordprocessingml"):
		return ".docx"
	case strings.Contains(mime, "spreadsheetml"):
		return ".xlsx"
	case strings.Contains(mime, "presentationml"):
		return ".pptx"
	case strings.Contains(mime, "jpeg"):
		return ".jpg"
	case strings.Contains(mime, "png"):
		return ".png"
	case strings.Contains(mime, "plain"):
		return ".txt"
	default:
		return ""
	}
}

func trimPreview(text string, max int) string {
	text = strings.Join(strings.Fields(text), " ")
	if len(text) <= max {
		return text
	}
	return strings.TrimSpace(text[:max])
}

func firstMeaningfulLine(text string) string {
	for _, line := range strings.Split(text, "\n") {
		line = strings.Join(strings.Fields(line), " ")
		if line != "" {
			return line
		}
	}
	return ""
}
