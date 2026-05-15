package extractors

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

type notebookExtractor struct{}

func (notebookExtractor) CanHandle(path string) bool {
	return strings.EqualFold(filepath.Ext(path), ".ipynb")
}

func (notebookExtractor) CanHandleType(detectedType string) bool { return detectedType == "notebook" }

func (notebookExtractor) Extract(path string) (string, error) {
	info, err := notebookExtractor{}.ExtractInfo(path)
	if err != nil {
		return "", err
	}
	return info.RawContent, nil
}

func (notebookExtractor) ExtractInfo(path string) (ExtractedFileInfo, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return ExtractedFileInfo{}, err
	}

	info := NewExtractedFileInfo(path, "notebook", string(b))
	info.SuggestedExtension = "ipynb"

	var payload notebookPayload
	if err := json.Unmarshal(b, &payload); err != nil {
		info.Warnings = append(info.Warnings, "could not parse notebook json")
		return info, nil
	}

	if title := notebookMetadataTitle(payload.Metadata); title != "" {
		info.Metadata["title"] = title
		info.TextSamples = append(info.TextSamples, TextSample{
			Source: "notebook-title",
			Text:   title,
			Score:  0.92,
		})
	}
	if heading := firstNotebookHeading(payload.Cells); heading != "" {
		info.TextSamples = append([]TextSample{{
			Source: "notebook-heading",
			Text:   heading,
			Score:  0.9,
		}}, info.TextSamples...)
	}
	if summary := firstNotebookMarkdown(payload.Cells); summary != "" {
		info.TextSamples = append(info.TextSamples, TextSample{
			Source: "notebook-markdown",
			Text:   summary,
			Score:  0.72,
		})
	}
	return info, nil
}

type notebookPayload struct {
	Metadata map[string]any  `json:"metadata"`
	Cells    []notebookCell  `json:"cells"`
	NBFormat int             `json:"nbformat"`
	Raw      json.RawMessage `json:"-"`
}

type notebookCell struct {
	CellType string `json:"cell_type"`
	Source   any    `json:"source"`
}

func notebookMetadataTitle(meta map[string]any) string {
	for _, key := range []string{"title", "name"} {
		if value, ok := meta[key].(string); ok && strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	if kernelspec, ok := meta["kernelspec"].(map[string]any); ok {
		if value, ok := kernelspec["display_name"].(string); ok && strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func firstNotebookHeading(cells []notebookCell) string {
	for _, cell := range cells {
		if cell.CellType != "markdown" {
			continue
		}
		for _, line := range strings.Split(notebookSourceText(cell.Source), "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "#") {
				return strings.TrimSpace(strings.TrimLeft(line, "#"))
			}
		}
	}
	return ""
}

func firstNotebookMarkdown(cells []notebookCell) string {
	for _, cell := range cells {
		if cell.CellType != "markdown" {
			continue
		}
		for _, line := range strings.Split(notebookSourceText(cell.Source), "\n") {
			line = strings.TrimSpace(strings.TrimLeft(strings.TrimSpace(line), "#"))
			if line != "" {
				return line
			}
		}
	}
	return ""
}

func notebookSourceText(source any) string {
	switch typed := source.(type) {
	case string:
		return typed
	case []any:
		var parts []string
		for _, part := range typed {
			if s, ok := part.(string); ok {
				parts = append(parts, s)
			}
		}
		return strings.Join(parts, "")
	default:
		return ""
	}
}

func init() { Register(notebookExtractor{}) }
