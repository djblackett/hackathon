package extractors

import (
	"encoding/csv"
	"os"
	"strings"
)

type csvExtractor struct{}

func (csvExtractor) CanHandle(path string) bool {
	return strings.HasSuffix(strings.ToLower(path), ".csv")
}

func (csvExtractor) Extract(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	content := string(b)

	firstLine := strings.SplitN(content, "\n", 2)[0]
	if len(firstLine) > 60 {
		firstLine = firstLine[:60]
	}
	return firstLine, nil
}

func (csvExtractor) ExtractInfo(path string) (ExtractedFileInfo, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return ExtractedFileInfo{}, err
	}

	content := string(b)
	info := NewExtractedFileInfo(path, content)
	info.DetectedType = "csv"

	reader := csv.NewReader(strings.NewReader(content))
	headers, err := reader.Read()
	if err != nil {
		info.Warnings = append(info.Warnings, "could not parse csv headers")
		return info, nil
	}

	for i := range headers {
		headers[i] = strings.TrimSpace(headers[i])
	}
	headerText := strings.Join(headers, " ")
	info.Metadata["headers"] = headerText
	info.TextSamples = append([]TextSample{{
		Source: "csv-headers",
		Text:   headerText,
		Score:  0.85,
	}}, info.TextSamples...)

	return info, nil
}

func init() { Register(csvExtractor{}) }
