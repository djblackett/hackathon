package evidence

import "strings"

func Merge(base FileEvidence, partials ...PartialEvidence) FileEvidence {
	if base.Metadata == nil {
		base.Metadata = map[string]string{}
	}
	seenSources := map[EvidenceSource]bool{}
	for _, source := range base.Sources {
		seenSources[source] = true
	}

	for _, partial := range partials {
		p := partial.Evidence
		source := partial.Source
		if source == "" {
			source = p.firstSource()
		}
		if source != "" && !seenSources[source] {
			base.Sources = append(base.Sources, source)
			seenSources[source] = true
		}
		if base.Path == "" {
			base.Path = p.Path
		}
		if base.SizeBytes == 0 {
			base.SizeBytes = p.SizeBytes
		}
		if base.SHA256 == "" {
			base.SHA256 = p.SHA256
		}
		if p.DetectedMIME != "" {
			base.DetectedMIME = p.DetectedMIME
		}
		if p.Extension != "" {
			base.Extension = p.Extension
		}
		for _, id := range p.FormatIDs {
			base.FormatIDs = append(base.FormatIDs, id)
		}
		for key, value := range p.Metadata {
			if strings.TrimSpace(value) == "" {
				continue
			}
			base.Metadata[key] = value
		}
		if strings.TrimSpace(p.TextPreview) != "" {
			if base.TextPreview == "" || len(p.TextPreview) > len(base.TextPreview) {
				base.TextPreview = p.TextPreview
			}
		}
		base.TextSignals = appendUniqueStrings(base.TextSignals, p.TextSignals...)
		if p.Media != nil {
			base.Media = p.Media
		}
		if p.Image != nil {
			base.Image = p.Image
		}
		if p.Validation != nil {
			base.Validation = p.Validation
		}
		base.Warnings = appendUniqueStrings(base.Warnings, p.Warnings...)
		base.Errors = append(base.Errors, p.Errors...)
	}

	return base
}

func (e FileEvidence) firstSource() EvidenceSource {
	if len(e.Sources) == 0 {
		return ""
	}
	return e.Sources[0]
}

func appendUniqueStrings(values []string, add ...string) []string {
	seen := map[string]bool{}
	for _, value := range values {
		seen[value] = true
	}
	for _, value := range add {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		values = append(values, value)
		seen[value] = true
	}
	return values
}
