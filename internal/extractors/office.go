package extractors

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
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

	switch subtype {
	case "docx":
		for _, sample := range docxParagraphSamples(zr.File) {
			info.TextSamples = append(info.TextSamples, sample)
		}
	case "xlsx":
		for _, sheet := range xlsxSheetNames(zr.File) {
			info.TextSamples = append(info.TextSamples, TextSample{
				Source: "office-sheet-name",
				Text:   sheet,
				Score:  0.78,
			})
		}
		if headers := xlsxFirstRowHeaders(zr.File); headers != "" {
			info.TextSamples = append(info.TextSamples, TextSample{
				Source: "office-headers",
				Text:   headers,
				Score:  0.86,
			})
		}
	case "pptx":
		for _, title := range pptxSlideTitles(zr.File) {
			info.TextSamples = append(info.TextSamples, TextSample{
				Source: "office-slide-title",
				Text:   title,
				Score:  0.88,
			})
		}
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

func docxParagraphSamples(files []*zip.File) []TextSample {
	paragraphs := docxParagraphs(readZipFile(files, "word/document.xml"))
	var samples []TextSample
	for _, paragraph := range paragraphs {
		if paragraph.text == "" {
			continue
		}
		if paragraph.heading {
			samples = append(samples, TextSample{
				Source: "office-heading",
				Text:   paragraph.text,
				Score:  0.9,
			})
			break
		}
	}
	for _, paragraph := range paragraphs {
		if paragraph.text == "" || paragraph.heading {
			continue
		}
		samples = append(samples, TextSample{
			Source: "office-first-paragraph",
			Text:   firstOfficeSentence(paragraph.text),
			Score:  0.74,
		})
		break
	}
	return samples
}

type docxParagraph struct {
	text    string
	heading bool
}

func docxParagraphs(data string) []docxParagraph {
	if data == "" {
		return nil
	}

	decoder := xml.NewDecoder(strings.NewReader(data))
	var (
		inParagraph bool
		inText      bool
		textParts   []string
		heading     bool
		paragraphs  []docxParagraph
	)
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return paragraphs
		}
		switch t := token.(type) {
		case xml.StartElement:
			switch strings.ToLower(t.Name.Local) {
			case "p":
				inParagraph = true
				heading = false
				textParts = nil
			case "pstyle":
				if inParagraph && hasHeadingStyle(t) {
					heading = true
				}
			case "t":
				if inParagraph {
					inText = true
				}
			}
		case xml.EndElement:
			switch strings.ToLower(t.Name.Local) {
			case "p":
				text := strings.Join(strings.Fields(strings.Join(textParts, " ")), " ")
				if text != "" {
					paragraphs = append(paragraphs, docxParagraph{text: text, heading: heading})
				}
				inParagraph = false
				inText = false
				textParts = nil
				heading = false
			case "t":
				inText = false
			}
		case xml.CharData:
			if inParagraph && inText {
				value := strings.TrimSpace(string(t))
				if value != "" {
					textParts = append(textParts, value)
				}
			}
		}
	}
	return paragraphs
}

func hasHeadingStyle(el xml.StartElement) bool {
	for _, attr := range el.Attr {
		if strings.ToLower(attr.Name.Local) != "val" {
			continue
		}
		return strings.HasPrefix(strings.ToLower(attr.Value), "heading")
	}
	return false
}

func firstOfficeSentence(text string) string {
	for i, r := range text {
		switch r {
		case '.', '!', '?':
			return strings.TrimSpace(text[:i])
		}
	}
	return text
}

func xlsxSheetNames(files []*zip.File) []string {
	data := readZipFile(files, "xl/workbook.xml")
	if data == "" {
		return nil
	}

	decoder := xml.NewDecoder(strings.NewReader(data))
	var names []string
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return names
		}
		start, ok := token.(xml.StartElement)
		if !ok || strings.ToLower(start.Name.Local) != "sheet" {
			continue
		}
		for _, attr := range start.Attr {
			if strings.ToLower(attr.Name.Local) == "name" && strings.TrimSpace(attr.Value) != "" {
				names = append(names, strings.TrimSpace(attr.Value))
			}
		}
	}
	return names
}

func xlsxFirstRowHeaders(files []*zip.File) string {
	shared := xlsxSharedStrings(files)
	data := readZipFile(files, "xl/worksheets/sheet1.xml")
	if data == "" {
		return ""
	}

	decoder := xml.NewDecoder(strings.NewReader(data))
	var (
		inFirstRow bool
		cellType   string
		inValue    bool
		headers    []string
	)
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return strings.Join(headers, " ")
		}
		switch t := token.(type) {
		case xml.StartElement:
			switch strings.ToLower(t.Name.Local) {
			case "row":
				inFirstRow = false
				for _, attr := range t.Attr {
					if strings.ToLower(attr.Name.Local) == "r" && attr.Value == "1" {
						inFirstRow = true
					}
				}
			case "c":
				cellType = ""
				for _, attr := range t.Attr {
					if strings.ToLower(attr.Name.Local) == "t" {
						cellType = attr.Value
					}
				}
			case "v", "t":
				inValue = inFirstRow
			}
		case xml.EndElement:
			switch strings.ToLower(t.Name.Local) {
			case "row":
				if inFirstRow {
					return strings.Join(headers, " ")
				}
				inFirstRow = false
			case "v", "t":
				inValue = false
			}
		case xml.CharData:
			if !inValue {
				continue
			}
			value := strings.TrimSpace(string(t))
			if value == "" {
				continue
			}
			if cellType == "s" {
				if idx, ok := parseSharedStringIndex(value, len(shared)); ok {
					value = shared[idx]
				}
			}
			if value != "" {
				headers = append(headers, value)
			}
		}
	}
	return strings.Join(headers, " ")
}

func xlsxSharedStrings(files []*zip.File) []string {
	data := readZipFile(files, "xl/sharedStrings.xml")
	if data == "" {
		return nil
	}

	decoder := xml.NewDecoder(strings.NewReader(data))
	var (
		inText bool
		items  []string
	)
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return items
		}
		switch t := token.(type) {
		case xml.StartElement:
			if strings.ToLower(t.Name.Local) == "t" {
				inText = true
			}
		case xml.EndElement:
			if strings.ToLower(t.Name.Local) == "t" {
				inText = false
			}
		case xml.CharData:
			if inText {
				items = append(items, strings.TrimSpace(string(t)))
			}
		}
	}
	return items
}

func parseSharedStringIndex(value string, max int) (int, bool) {
	var idx int
	if _, err := fmt.Sscanf(value, "%d", &idx); err != nil {
		return 0, false
	}
	return idx, idx >= 0 && idx < max
}

func pptxSlideTitles(files []*zip.File) []string {
	var titles []string
	for _, file := range files {
		if !strings.HasPrefix(file.Name, "ppt/slides/slide") || !strings.HasSuffix(file.Name, ".xml") {
			continue
		}
		text := xmlText(readZipFile(files, file.Name))
		if text == "" {
			continue
		}
		words := strings.Fields(text)
		if len(words) > 12 {
			words = words[:12]
		}
		titles = append(titles, strings.Join(words, " "))
	}
	return titles
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
			if len(parts) > 0 {
				return strings.Join(parts, " ")
			}
			return xmlTextByTagStripping(data)
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
	text := fmt.Sprint(strings.Join(parts, " "))
	if strings.TrimSpace(text) != "" {
		return text
	}
	return xmlTextByTagStripping(data)
}

func xmlTextByTagStripping(data string) string {
	tag := regexp.MustCompile(`<[^>]+>`)
	text := tag.ReplaceAllString(data, " ")
	text = strings.Join(strings.Fields(text), " ")
	return text
}

func init() { Register(officeExtractor{}) }
