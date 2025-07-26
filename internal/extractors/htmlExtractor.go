// extractors/html.go
package extractors

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type htmlExtractor struct{}

func (htmlExtractor) CanHandle(p string) bool {
	ext := strings.ToLower(filepath.Ext(p))
	return ext == ".html" || ext == ".htm"
}

func (htmlExtractor) Extract(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(b))
	if err != nil {
		return "", err
	}

	text := strings.TrimSpace(doc.Text())
	return text, nil
}

func init() { Register(htmlExtractor{}) }
