package extractors

import (
	"archive/zip"
	"encoding/xml"
	"io"
	"path/filepath"
	"strings"
)

type epubExtractor struct{}

func (epubExtractor) CanHandle(path string) bool {
	return strings.EqualFold(filepath.Ext(path), ".epub")
}

func (epubExtractor) CanHandleType(detectedType string) bool { return detectedType == "epub" }

func (epubExtractor) Extract(path string) (string, error) {
	info, err := epubExtractor{}.ExtractInfo(path)
	if err != nil {
		return "", err
	}
	return info.RawContent, nil
}

func (epubExtractor) ExtractInfo(path string) (ExtractedFileInfo, error) {
	zr, err := zip.OpenReader(path)
	if err != nil {
		return ExtractedFileInfo{}, err
	}
	defer zr.Close()

	info := NewExtractedFileInfo(path, "epub", "")
	info.SuggestedExtension = "epub"

	opfPath := epubOPFPath(zr.File)
	meta := epubMetadata(readZipFile(zr.File, opfPath))
	if meta.title != "" {
		info.Metadata["title"] = meta.title
		info.TextSamples = append(info.TextSamples, TextSample{
			Source: "epub-title",
			Text:   meta.title,
			Score:  0.94,
		})
	}
	if meta.creator != "" {
		info.Metadata["creator"] = meta.creator
		info.TextSamples = append(info.TextSamples, TextSample{
			Source: "epub-creator",
			Text:   meta.creator,
			Score:  0.7,
		})
	}
	if heading := firstEPUBHeading(zr.File); heading != "" {
		info.TextSamples = append(info.TextSamples, TextSample{
			Source: "epub-heading",
			Text:   heading,
			Score:  0.86,
		})
	}
	if len(info.TextSamples) == 0 {
		info.Warnings = append(info.Warnings, "no epub metadata or headings extracted")
	}
	return info, nil
}

func epubOPFPath(files []*zip.File) string {
	container := readZipFile(files, "META-INF/container.xml")
	if container == "" {
		return ""
	}
	decoder := xml.NewDecoder(strings.NewReader(container))
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return ""
		}
		start, ok := token.(xml.StartElement)
		if !ok || strings.ToLower(start.Name.Local) != "rootfile" {
			continue
		}
		for _, attr := range start.Attr {
			if strings.ToLower(attr.Name.Local) == "full-path" {
				return attr.Value
			}
		}
	}
	return ""
}

type epubMeta struct {
	title   string
	creator string
}

func epubMetadata(data string) epubMeta {
	var meta epubMeta
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
			case "creator":
				if meta.creator == "" {
					meta.creator = text
				}
			}
		}
	}
	return meta
}

func firstEPUBHeading(files []*zip.File) string {
	for _, file := range files {
		name := strings.ToLower(file.Name)
		if !strings.HasSuffix(name, ".xhtml") && !strings.HasSuffix(name, ".html") {
			continue
		}
		if heading := firstHTMLHeading(readZipFile(files, file.Name)); heading != "" {
			return heading
		}
	}
	return ""
}

func firstHTMLHeading(data string) string {
	decoder := xml.NewDecoder(strings.NewReader(data))
	decoder.Strict = false
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
			if name == "h1" || name == "h2" || name == "title" {
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

func init() { Register(epubExtractor{}) }
