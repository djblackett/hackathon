package extractors

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

type officeExtractor struct{}

func (officeExtractor) CanHandle(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".docx" || ext == ".xlsx" || ext == ".pptx"
}

func (officeExtractor) CanHandleType(detectedType string) bool {
	return detectedType == "office" || detectedType == "docx" || detectedType == "xlsx" || detectedType == "pptx"
}

func (officeExtractor) Extract(path string) (string, error) {
	info, err := officeExtractor{}.ExtractInfo(path)
	if err != nil {
		return "", err
	}
	return info.RawContent, nil
}

func (officeExtractor) ExtractInfo(path string) (ExtractedFileInfo, error) {
	zr, err := zip.OpenReader(path)
	if err != nil {
		return ExtractedFileInfo{}, err
	}
	defer zr.Close()

	subtype := officeSubtype(zr.File)
	content := officeText(zr.File, subtype)
	info := NewExtractedFileInfo(path, "office", content)
	info.SuggestedExtension = subtype
	if subtype != "" {
		info.Metadata["detected_subtype"] = subtype
	}

	if meta := officeCoreMetadata(zr.File); len(meta) > 0 {
		for k, v := range meta {
			info.Metadata[k] = v
		}
		if title := meta["title"]; title != "" {
			info.TextSamples = append([]TextSample{{
				Source: "office-title",
				Text:   title,
				Score:  0.95,
			}}, info.TextSamples...)
		}
		if subject := meta["subject"]; subject != "" {
			info.TextSamples = append(info.TextSamples, TextSample{
				Source: "office-subject",
				Text:   subject,
				Score:  0.88,
			})
		}
	}

	if strings.TrimSpace(content) != "" {
		info.TextSamples = append(info.TextSamples, TextSample{
			Source: "office-text",
			Text:   content,
			Score:  0.65,
		})
	}

	if len(info.TextSamples) == 0 {
		info.Warnings = append(info.Warnings, "no office metadata or text extracted")
	}
	return info, nil
}

func officeSubtype(files []*zip.File) string {
	for _, file := range files {
		name := strings.ToLower(file.Name)
		switch {
		case name == "word/document.xml":
			return "docx"
		case name == "xl/workbook.xml":
			return "xlsx"
		case name == "ppt/presentation.xml":
			return "pptx"
		}
	}
	return "office"
}

func officeCoreMetadata(files []*zip.File) map[string]string {
	meta := map[string]string{}
	data := readZipFile(files, "docProps/core.xml")
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
		case xml.CharData:
			value := strings.TrimSpace(string(t))
			if value != "" {
				switch current {
				case "title", "subject", "creator", "description", "keywords":
					meta[current] = value
				}
			}
		}
	}
	return meta
}

func officeText(files []*zip.File, subtype string) string {
	var paths []string
	switch subtype {
	case "docx":
		paths = []string{"word/document.xml"}
	case "xlsx":
		paths = []string{"xl/sharedStrings.xml", "xl/workbook.xml"}
	case "pptx":
		for _, file := range files {
			if strings.HasPrefix(file.Name, "ppt/slides/slide") && strings.HasSuffix(file.Name, ".xml") {
				paths = append(paths, file.Name)
			}
		}
	default:
		return ""
	}

	var parts []string
	for _, path := range paths {
		if text := xmlText(readZipFile(files, path)); text != "" {
			parts = append(parts, text)
		}
	}
	return strings.Join(parts, " ")
}

func readZipFile(files []*zip.File, name string) string {
	for _, file := range files {
		if file.Name != name {
			continue
		}
		rc, err := file.Open()
		if err != nil {
			return ""
		}
		defer rc.Close()
		b, err := io.ReadAll(rc)
		if err != nil {
			return ""
		}
		return string(b)
	}
	return ""
}

func xmlText(data string) string {
	if data == "" {
		return ""
	}
	decoder := xml.NewDecoder(strings.NewReader(data))
	var parts []string
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return strings.Join(parts, " ")
		}
		if t, ok := token.(xml.CharData); ok {
			value := strings.TrimSpace(string(t))
			if value != "" {
				parts = append(parts, value)
			}
		}
	}
	if len(parts) > 200 {
		parts = parts[:200]
	}
	return fmt.Sprint(strings.Join(parts, " "))
}

func init() { Register(officeExtractor{}) }
