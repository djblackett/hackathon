package extractors

import (
	"bytes"
	"strings"

	"github.com/ledongthuc/pdf"
)

type pdfExtractor struct{}

func (pdfExtractor) CanHandle(path string) bool {
	return strings.HasSuffix(strings.ToLower(path), ".pdf")
}
func (pdfExtractor) CanHandleType(detectedType string) bool { return detectedType == "pdf" }

func (pdfExtractor) Extract(path string) (string, error) {
	pdf.DebugOn = true

	f, r, err := pdf.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	var buf bytes.Buffer
	b, err := r.GetPlainText()
	if err != nil {
		return "", err
	}
	buf.ReadFrom(b)
	content := buf.String()
	return content, nil
}

func (pdfExtractor) ExtractInfo(path string) (ExtractedFileInfo, error) {
	content, err := pdfExtractor{}.Extract(path)
	if err != nil {
		return ExtractedFileInfo{}, err
	}

	info := NewExtractedFileInfo(path, "pdf", content)

	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		info.TextSamples = append([]TextSample{{
			Source: "pdf-first-text",
			Text:   line,
			Score:  0.7,
		}}, info.TextSamples...)
		break
	}

	return info, nil
}

func init() { Register(pdfExtractor{}) }
