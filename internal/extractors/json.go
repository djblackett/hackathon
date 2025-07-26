package extractors

import (
	"fmt"
	"os/exec"
	"strings"
)

type jsonExtractor struct{}

func (jsonExtractor) CanHandle(path string) bool {
	return strings.HasSuffix(strings.ToLower(path), ".json")
}

func (jsonExtractor) Extract(path string) (string, error) {
	cmd := exec.Command("sh", "-c", fmt.Sprintf(`jq -r '
		(if type == "array" then .[0] else . end)
		| paths(scalars)
		| map(tostring)
		| join(".")
	' %q | paste -sd ' ' -`, path)) // %q safely escapes path

	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to run jq: %w", err)
	}

	return string(out), nil
}

func init() { Register(jsonExtractor{}) }
