package extractors

import (
	"archive/zip"
	"encoding/xml"
	"io"
	"path/filepath"
	"strings"
)

type openDocumentExtractor struct{}

func (openDocumentExtractor) CanHandle(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".odt", ".ods", ".odp":
		return true
	default:
		return false
	}
}

func (openDocumentExtractor) CanHandleType(detectedType string) bool {
	return detectedType == "opendocument"
}

func (openDocumentExtractor) Extract(path string) (string, error) {
	info, err := openDocumentExtractor{}.ExtractInfo(path)
	if err != nil {
		return "", err
	}
	return info.RawContent, nil
}

func (openDocumentExtractor) ExtractInfo(path string) (ExtractedFileInfo, error) {
	zr, err := zip.OpenReader(path)
	if err != nil {
		return ExtractedFileInfo{}, err
	}
	defer zr.Close()

	subtype := openDocumentExtension(path)
	info := NewExtractedFileInfo(path, "opendocument", "")
	info.SuggestedExtension = subtype
	info.Metadata["detected_subtype"] = subtype

	meta := openDocumentMetadata(readZipFile(zr.File, "meta.xml"))
	if meta.title != "" {
		info.Metadata["title"] = meta.title
		info.TextSamples = append(info.TextSamples, TextSample{
			Source: "opendocument-title",
			Text:   meta.title,
			Score:  0.92,
		})
	}
	if meta.subject != "" {
		info.Metadata["subject"] = meta.subject
		info.TextSamples = append(info.TextSamples, TextSample{
			Source: "opendocument-subject",
			Text:   meta.subject,
			Score:  0.84,
		})
	}

	if heading := firstOpenDocumentHeading(readZipFile(zr.File, "content.xml")); heading != "" {
		info.TextSamples = append(info.TextSamples, TextSample{
			Source: "opendocument-heading",
			Text:   heading,
			Score:  0.88,
		})
	}
	if len(info.TextSamples) == 0 {
		info.Warnings = append(info.Warnings, "no opendocument metadata or headings extracted")
	}
	return info, nil
}

func openDocumentExtension(path string) string {
	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(path)), ".")
	switch ext {
	case "odt", "ods", "odp":
		return ext
	default:
		return "odt"
	}
}

type openDocumentMeta struct {
	title   string
	subject string
}

func openDocumentMetadata(data string) openDocumentMeta {
	var meta openDocumentMeta
	if data == "" {
		return meta
	}
	decoder := xml.NewDecoder(strings.NewReader(data))
	var current string
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return meta
		}
		switch t := token.(type) {
		case xml.StartElement:
			current = strings.ToLower(t.Name.Local)
		case xml.EndElement:
			current = ""
		case xml.CharData:
			text := strings.Join(strings.Fields(string(t)), " ")
			if text == "" {
				continue
			}
			switch current {
			case "title":
				if meta.title == "" {
					meta.title = text
				}
			case "subject":
				if meta.subject == "" {
					meta.subject = text
				}
			}
		}
	}
	return meta
}

func firstOpenDocumentHeading(data string) string {
	if data == "" {
		return ""
	}
	decoder := xml.NewDecoder(strings.NewReader(data))
	var current string
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return ""
		}
		switch t := token.(type) {
		case xml.StartElement:
			name := strings.ToLower(t.Name.Local)
			if name == "h" || name == "p" {
				current = name
			}
		case xml.EndElement:
			current = ""
		case xml.CharData:
			if current != "" {
				if text := strings.Join(strings.Fields(string(t)), " "); text != "" {
					return text
				}
			}
		}
	}
	return ""
}

func init() { Register(openDocumentExtractor{}) }
