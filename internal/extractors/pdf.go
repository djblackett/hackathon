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

func init() { Register(pdfExtractor{}) }
