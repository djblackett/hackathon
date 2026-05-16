package jhove

import (
	"context"
	"encoding/xml"
	"fmt"
	"strings"
	"time"

	"github.com/djblackett/bootdev-hackathon/internal/evidence"
	"github.com/djblackett/bootdev-hackathon/internal/tools"
)

type Extractor struct {
	Timeout time.Duration
}

func (Extractor) Name() evidence.EvidenceSource { return evidence.SourceJHOVE }

func (Extractor) Available(ctx context.Context) bool {
	return tools.Available("jhove")
}

func (e Extractor) Extract(ctx context.Context, path string) (evidence.PartialEvidence, error) {
	result, err := tools.Run(ctx, e.Timeout, "jhove", "-h", "xml", path)
	if err != nil {
		message := strings.TrimSpace(string(result.Stderr))
		if message == "" {
			message = err.Error()
		}
		return evidence.PartialEvidence{}, fmt.Errorf("jhove failed: %s", message)
	}
	ev, err := Parse(path, result.Stdout)
	if err != nil {
		return evidence.PartialEvidence{}, err
	}
	return evidence.PartialEvidence{Source: evidence.SourceJHOVE, Evidence: ev}, nil
}

func Parse(path string, data []byte) (evidence.FileEvidence, error) {
	var doc jhoveDoc
	if err := xml.Unmarshal(data, &doc); err != nil {
		return evidence.FileEvidence{}, fmt.Errorf("parse jhove xml: %w", err)
	}
	ev := evidence.FileEvidence{
		Path:     path,
		Metadata: map[string]string{},
		Sources:  []evidence.EvidenceSource{evidence.SourceJHOVE},
	}
	rep := firstRepInfo(doc.RepInfo, path)
	status := strings.TrimSpace(rep.Status)
	warnings := cleanMessages(rep.Messages.Messages)
	valid := validFromStatus(status)
	ev.Validation = &evidence.ValidationResult{
		Source:   evidence.SourceJHOVE,
		Valid:    valid,
		Status:   status,
		Warnings: warnings,
	}
	ev.Warnings = append(ev.Warnings, warnings...)
	if doc.Date != "" {
		ev.Metadata["jhove_date"] = strings.TrimSpace(doc.Date)
	}
	if doc.Version != "" {
		ev.Metadata["jhove_version"] = strings.TrimSpace(doc.Version)
	}
	if rep.URI != "" {
		ev.Metadata["jhove_uri"] = strings.TrimSpace(rep.URI)
	}
	if status == "" {
		ev.Warnings = append(ev.Warnings, "jhove returned no validation status")
	}
	return ev, nil
}

type jhoveDoc struct {
	XMLName xml.Name       `xml:"jhove"`
	Version string         `xml:"version"`
	Date    string         `xml:"date"`
	RepInfo []jhoveRepInfo `xml:"repInfo"`
}

type jhoveRepInfo struct {
	URI      string        `xml:"uri,attr"`
	Status   string        `xml:"status"`
	Messages jhoveMessages `xml:"messages"`
}

type jhoveMessages struct {
	Messages []string `xml:"message"`
}

func firstRepInfo(values []jhoveRepInfo, path string) jhoveRepInfo {
	if len(values) == 0 {
		return jhoveRepInfo{}
	}
	for _, value := range values {
		if value.URI == path {
			return value
		}
	}
	return values[0]
}

func validFromStatus(status string) *bool {
	normalized := strings.ToLower(strings.TrimSpace(status))
	if normalized == "" {
		return nil
	}
	valid := normalized == "well-formed and valid" || normalized == "valid"
	if strings.Contains(normalized, "not well-formed") || strings.Contains(normalized, "not valid") {
		valid = false
	}
	return &valid
}

func cleanMessages(values []string) []string {
	out := []string{}
	for _, value := range values {
		value = strings.Join(strings.Fields(value), " ")
		if value != "" {
			out = append(out, value)
		}
	}
	return out
}
