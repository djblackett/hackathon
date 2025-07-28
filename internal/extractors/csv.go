package extractors

import (
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

func init() { Register(csvExtractor{}) }
