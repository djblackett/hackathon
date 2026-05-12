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
func (htmlExtractor) CanHandleType(detectedType string) bool { return detectedType == "html" }

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
	info := NewExtractedFileInfo(path, "html", text)
	info.SuggestedExtension = "html"

	if title := cleanHTMLText(doc.Find("title").First().Text()); title != "" {
		info.Metadata["title"] = title
		info.TextSamples = append(info.TextSamples, TextSample{
			Source: "html-title",
			Text:   title,
			Score:  0.95,
		})
	}

	if ogTitle := htmlMetaContent(doc, `meta[property="og:title"], meta[name="og:title"]`); ogTitle != "" {
		info.Metadata["og_title"] = ogTitle
		info.TextSamples = append(info.TextSamples, TextSample{
			Source: "html-og-title",
			Text:   ogTitle,
			Score:  0.96,
		})
	}

	if heading := cleanHTMLText(doc.Find("h1").First().Text()); heading != "" {
		info.Metadata["heading"] = heading
		info.TextSamples = append(info.TextSamples, TextSample{
			Source: "html-h1",
			Text:   heading,
			Score:  0.9,
		})
	}

	if description := htmlMetaContent(doc, `meta[name="description"]`); description != "" {
		info.Metadata["description"] = description
		info.TextSamples = append(info.TextSamples, TextSample{
			Source: "html-meta-description",
			Text:   description,
			Score:  0.8,
		})
	}

	if ogDescription := htmlMetaContent(doc, `meta[property="og:description"], meta[name="og:description"]`); ogDescription != "" {
		info.Metadata["og_description"] = ogDescription
		info.TextSamples = append(info.TextSamples, TextSample{
			Source: "html-og-description",
			Text:   ogDescription,
			Score:  0.83,
		})
	}

	return info, nil
}

func htmlMetaContent(doc *goquery.Document, selector string) string {
	content, exists := doc.Find(selector).First().Attr("content")
	if !exists {
		return ""
	}
	return cleanHTMLText(content)
}

func cleanHTMLText(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

func init() { Register(htmlExtractor{}) }
