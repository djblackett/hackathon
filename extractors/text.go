package extractors

import (
	"os"
)

type textExtractor struct{}

func (textExtractor) CanHandle(path string) bool { return true } // fallback
func (textExtractor) Extract(path string) (string, error) {
	b, err := os.ReadFile(path)
	return string(b), err
}

func init() { Register(textExtractor{}) }
