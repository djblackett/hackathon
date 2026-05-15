package siegfried

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/djblackett/bootdev-hackathon/internal/evidence"
	"github.com/djblackett/bootdev-hackathon/internal/tools"
)

type Extractor struct {
	Timeout time.Duration
}

func (Extractor) Name() evidence.EvidenceSource { return evidence.SourceSiegfried }

func (Extractor) Available(ctx context.Context) bool {
	return tools.Available("sf")
}

func (e Extractor) Extract(ctx context.Context, path string) (evidence.PartialEvidence, error) {
	result, err := tools.Run(ctx, e.Timeout, "sf", "-json", path)
	if err != nil {
		message := strings.TrimSpace(string(result.Stderr))
		if message == "" {
			message = err.Error()
		}
		return evidence.PartialEvidence{}, fmt.Errorf("siegfried failed: %s", message)
	}
	ev, err := Parse(path, result.Stdout)
	if err != nil {
		return evidence.PartialEvidence{}, err
	}
	return evidence.PartialEvidence{Source: evidence.SourceSiegfried, Evidence: ev}, nil
}

func Parse(path string, data []byte) (evidence.FileEvidence, error) {
	var payload sfPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return evidence.FileEvidence{}, fmt.Errorf("parse siegfried json: %w", err)
	}

	ev := evidence.FileEvidence{
		Path:     path,
		Metadata: map[string]string{},
		Sources:  []evidence.EvidenceSource{evidence.SourceSiegfried},
	}
	if payload.Siegfried != "" {
		ev.Metadata["siegfried_version"] = payload.Siegfried
	}
	if payload.Signature != "" {
		ev.Metadata["siegfried_signature"] = payload.Signature
	}

	file := firstFile(payload.Files, path)
	if file == nil {
		ev.Warnings = append(ev.Warnings, "siegfried returned no file results")
		return ev, nil
	}
	if strings.TrimSpace(file.Errors) != "" {
		ev.Warnings = append(ev.Warnings, "siegfried file error: "+strings.TrimSpace(file.Errors))
	}

	for _, match := range file.Matches {
		if strings.EqualFold(match.ID, "UNKNOWN") || strings.EqualFold(match.Format, "UNKNOWN") {
			if strings.TrimSpace(match.Warning) != "" {
				ev.Warnings = append(ev.Warnings, "siegfried: "+strings.TrimSpace(match.Warning))
			}
			continue
		}
		formatID := evidence.FormatID{
			Source:     evidence.SourceSiegfried,
			ID:         strings.TrimSpace(match.ID),
			Name:       strings.TrimSpace(match.Format),
			Version:    strings.TrimSpace(match.Version),
			MIME:       firstNonEmpty(match.MIME),
			Extension:  dotExt(firstExtension(match)),
			Confidence: 0.9,
		}
		ev.FormatIDs = append(ev.FormatIDs, formatID)
		if ev.DetectedMIME == "" && formatID.MIME != "" {
			ev.DetectedMIME = formatID.MIME
		}
		if ev.Extension == "" && formatID.Extension != "" {
			ev.Extension = formatID.Extension
		}
		if strings.TrimSpace(match.Basis) != "" {
			ev.Metadata["siegfried_basis"] = strings.TrimSpace(match.Basis)
		}
		if strings.TrimSpace(match.Warning) != "" {
			ev.Warnings = append(ev.Warnings, "siegfried: "+strings.TrimSpace(match.Warning))
		}
	}
	if len(ev.FormatIDs) == 0 {
		ev.Warnings = append(ev.Warnings, "siegfried returned no positive format matches")
	}
	return ev, nil
}

type sfPayload struct {
	Siegfried string   `json:"siegfried"`
	Signature string   `json:"signature"`
	Files     []sfFile `json:"files"`
}

type sfFile struct {
	Filename string    `json:"filename"`
	Errors   string    `json:"errors"`
	Matches  []sfMatch `json:"matches"`
}

type sfMatch struct {
	Namespace  string `json:"ns"`
	ID         string `json:"id"`
	Format     string `json:"format"`
	Version    string `json:"version"`
	MIME       string `json:"mime"`
	Basis      string `json:"basis"`
	Warning    string `json:"warning"`
	Extension  string `json:"extension"`
	Extensions any    `json:"extensions"`
}

func firstFile(files []sfFile, path string) *sfFile {
	if len(files) == 0 {
		return nil
	}
	for i := range files {
		if files[i].Filename == path {
			return &files[i]
		}
	}
	return &files[0]
}

func firstExtension(match sfMatch) string {
	if strings.TrimSpace(match.Extension) != "" {
		return match.Extension
	}
	switch v := match.Extensions.(type) {
	case string:
		return firstDelimited(v)
	case []any:
		for _, item := range v {
			if s, ok := item.(string); ok && strings.TrimSpace(s) != "" {
				return s
			}
		}
	}
	return ""
}

func firstDelimited(value string) string {
	value = strings.NewReplacer(",", " ", ";", " ", "|", " ").Replace(value)
	for _, part := range strings.Fields(value) {
		return part
	}
	return ""
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func dotExt(ext string) string {
	ext = strings.TrimSpace(strings.ToLower(ext))
	if ext == "" {
		return ""
	}
	if strings.HasPrefix(ext, ".") {
		return ext
	}
	return "." + ext
}
