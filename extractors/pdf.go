package extractors

import (
	"bytes"
	"os/exec"
	"strings"
)

type pdfExtractor struct{}

func (pdfExtractor) CanHandle(path string) bool { return strings.HasSuffix(strings.ToLower(path), ".pdf") }

func (pdfExtractor) Extract(path string) (string, error) {
	cmd := exec.Command("pdftotext", "-layout", path, "-") // output to stdout
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	return out.String(), err
}

func init() { Register(pdfExtractor{}) }
