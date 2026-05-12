package extractors

import (
	"encoding/xml"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type xmlExtractor struct{}

func (xmlExtractor) CanHandle(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".xml" || ext == ".musicxml"
}

func (xmlExtractor) CanHandleType(detectedType string) bool {
	return detectedType == "xml" || detectedType == "musicxml"
}

func (xmlExtractor) Extract(path string) (string, error) {
	info, err := xmlExtractor{}.ExtractInfo(path)
	if err != nil {
		return "", err
	}
	return info.RawContent, nil
}

func (xmlExtractor) ExtractInfo(path string) (ExtractedFileInfo, error) {
	f, err := os.Open(path)
	if err != nil {
		return ExtractedFileInfo{}, err
	}
	defer f.Close()

	parsed, err := parseXMLMetadata(f)
	if err != nil {
		return ExtractedFileInfo{}, err
	}

	content := strings.Join(parsed.evidenceText(), " ")
	info := NewExtractedFileInfo(path, "xml", content)
	info.SuggestedExtension = "xml"
	info.Metadata["root"] = parsed.root

	if parsed.musicXML {
		info.SuggestedExtension = "musicxml"
		info.Metadata["detected_subtype"] = "musicxml"
		if parsed.workTitle != "" {
			info.Metadata["work_title"] = parsed.workTitle
			info.TextSamples = append(info.TextSamples, TextSample{Source: "musicxml-work-title", Text: parsed.workTitle, Score: 0.95})
		}
		if parsed.movementTitle != "" {
			info.Metadata["movement_title"] = parsed.movementTitle
			info.TextSamples = append(info.TextSamples, TextSample{Source: "musicxml-movement-title", Text: parsed.movementTitle, Score: 0.9})
		}
		if parsed.creator != "" {
			info.Metadata["creator"] = parsed.creator
			info.TextSamples = append(info.TextSamples, TextSample{Source: "musicxml-creator", Text: parsed.creator, Score: 0.78})
		}
		if len(parsed.partNames) > 0 {
			parts := strings.Join(parsed.partNames, " ")
			info.Metadata["parts"] = strings.Join(parsed.partNames, ", ")
			info.TextSamples = append(info.TextSamples, TextSample{Source: "musicxml-parts", Text: parts, Score: 0.62})
		}
		return info, nil
	}

	if parsed.title != "" {
		info.Metadata["title"] = parsed.title
		info.TextSamples = append(info.TextSamples, TextSample{Source: "xml-title", Text: parsed.title, Score: 0.9})
	}
	if parsed.name != "" {
		info.Metadata["name"] = parsed.name
		info.TextSamples = append(info.TextSamples, TextSample{Source: "xml-name", Text: parsed.name, Score: 0.82})
	}
	if parsed.creator != "" {
		info.Metadata["creator"] = parsed.creator
		info.TextSamples = append(info.TextSamples, TextSample{Source: "xml-creator", Text: parsed.creator, Score: 0.72})
	}
	if parsed.root != "" {
		info.TextSamples = append(info.TextSamples, TextSample{Source: "xml-root", Text: parsed.root, Score: 0.48})
	}

	return info, nil
}

type parsedXML struct {
	root          string
	musicXML      bool
	title         string
	name          string
	workTitle     string
	movementTitle string
	creator       string
	partNames     []string
}

func (p parsedXML) evidenceText() []string {
	values := []string{p.workTitle, p.movementTitle, p.title, p.name, p.creator}
	values = append(values, p.partNames...)
	if p.root != "" {
		values = append(values, p.root)
	}
	out := make([]string, 0, len(values))
	for _, value := range values {
		if value = cleanXMLText(value); value != "" {
			out = append(out, value)
		}
	}
	return out
}

func parseXMLMetadata(r io.Reader) (parsedXML, error) {
	decoder := xml.NewDecoder(r)
	decoder.Strict = false
	var parsed parsedXML
	var current string

	for {
		tok, err := decoder.Token()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return parsed, err
		}

		switch token := tok.(type) {
		case xml.StartElement:
			name := strings.ToLower(token.Name.Local)
			if parsed.root == "" {
				parsed.root = name
				parsed.musicXML = name == "score-partwise" || name == "score-timewise"
			}
			current = name
		case xml.CharData:
			text := cleanXMLText(string(token))
			if text == "" {
				continue
			}
			switch current {
			case "title":
				if parsed.title == "" {
					parsed.title = text
				}
			case "name":
				if parsed.name == "" {
					parsed.name = text
				}
			case "work-title":
				if parsed.workTitle == "" {
					parsed.workTitle = text
				}
			case "movement-title":
				if parsed.movementTitle == "" {
					parsed.movementTitle = text
				}
			case "creator", "composer", "author":
				if parsed.creator == "" {
					parsed.creator = text
				}
			case "part-name":
				if len(parsed.partNames) < 8 && !containsString(parsed.partNames, text) {
					parsed.partNames = append(parsed.partNames, text)
				}
			}
		case xml.EndElement:
			current = ""
		}
	}

	return parsed, nil
}

func cleanXMLText(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

func containsString(values []string, needle string) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}

func init() { Register(xmlExtractor{}) }
