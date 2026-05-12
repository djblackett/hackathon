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

func (htmlExtractor) ExtractInfo(path string) (ExtractedFileInfo, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return ExtractedFileInfo{}, err
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(b))
	if err != nil {
		return ExtractedFileInfo{}, err
	}

	text := strings.TrimSpace(doc.Text())
	info := NewExtractedFileInfo(path, text)
	info.DetectedType = "html"

	title := strings.TrimSpace(doc.Find("title").First().Text())
	if title != "" {
		info.Metadata["title"] = title
		info.TextSamples = append([]TextSample{{
			Source: "html-title",
			Text:   title,
			Score:  0.95,
		}}, info.TextSamples...)
	}

	heading := strings.TrimSpace(doc.Find("h1").First().Text())
	if heading != "" {
		info.Metadata["heading"] = heading
		info.TextSamples = append(info.TextSamples, TextSample{
			Source: "html-h1",
			Text:   heading,
			Score:  0.9,
		})
	}

	description, exists := doc.Find(`meta[name="description"]`).Attr("content")
	description = strings.TrimSpace(description)
	if exists && description != "" {
		info.Metadata["description"] = description
		info.TextSamples = append(info.TextSamples, TextSample{
			Source: "html-meta-description",
			Text:   description,
			Score:  0.8,
		})
	}

	return info, nil
}

func init() { Register(htmlExtractor{}) }
