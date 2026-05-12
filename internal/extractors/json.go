package extractors

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
)

type jsonExtractor struct{}

func (jsonExtractor) CanHandle(path string) bool {
	return strings.HasSuffix(strings.ToLower(path), ".json")
}
func (jsonExtractor) CanHandleType(detectedType string) bool { return detectedType == "json" }

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

func (jsonExtractor) ExtractInfo(path string) (ExtractedFileInfo, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return ExtractedFileInfo{}, err
	}

	info := NewExtractedFileInfo(path, "json", string(b))

	var value any
	if err := json.Unmarshal(b, &value); err != nil {
		info.Warnings = append(info.Warnings, "could not parse json")
		return info, nil
	}

	subject, keys := jsonEvidence(value)
	if subject != "" {
		info.Metadata["subject"] = subject
		info.TextSamples = append([]TextSample{{
			Source: "json-title-field",
			Text:   subject,
			Score:  0.9,
		}}, info.TextSamples...)
	}
	if len(keys) > 0 {
		keyText := strings.Join(keys, " ")
		info.Metadata["keys"] = keyText
		info.TextSamples = append(info.TextSamples, TextSample{
			Source: "json-keys",
			Text:   keyText,
			Score:  0.75,
		})
	}

	return info, nil
}

func jsonEvidence(value any) (string, []string) {
	if arr, ok := value.([]any); ok && len(arr) > 0 {
		value = arr[0]
	}

	obj, ok := value.(map[string]any)
	if !ok {
		return "", nil
	}

	for _, key := range []string{"title", "name", "subject", "description", "summary"} {
		if raw, ok := obj[key]; ok {
			if s, ok := raw.(string); ok && strings.TrimSpace(s) != "" {
				return strings.TrimSpace(s), nil
			}
		}
	}

	keys := make([]string, 0, len(obj))
	for key := range obj {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	if len(keys) > 12 {
		keys = keys[:12]
	}
	return "", keys
}

func init() { Register(jsonExtractor{}) }
