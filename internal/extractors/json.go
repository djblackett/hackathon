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

	subject, keys, structured := jsonEvidence(value)
	if subject != "" {
		info.Metadata["subject"] = subject
		info.TextSamples = append([]TextSample{{
			Source: "json-title-field",
			Text:   subject,
			Score:  0.9,
		}}, info.TextSamples...)
	}
	if structured != "" {
		info.Metadata["structured"] = structured
		info.TextSamples = append(info.TextSamples, TextSample{
			Source: "json-structured",
			Text:   structured,
			Score:  0.84,
		})
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

func jsonEvidence(value any) (string, []string, string) {
	if arr, ok := value.([]any); ok && len(arr) > 0 {
		value = arr[0]
	}

	obj, ok := value.(map[string]any)
	if !ok {
		return "", nil, ""
	}

	for _, key := range []string{"title", "name", "subject", "description", "summary"} {
		if raw, ok := obj[key]; ok {
			if s, ok := raw.(string); ok && strings.TrimSpace(s) != "" {
				return strings.TrimSpace(s), nil, ""
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
	return "", keys, jsonStructuredEvidence(value)
}

func jsonStructuredEvidence(value any) string {
	var candidates []jsonEvidenceCandidate
	collectJSONEvidence(value, nil, &candidates)
	if len(candidates) == 0 {
		return ""
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].priority != candidates[j].priority {
			return candidates[i].priority < candidates[j].priority
		}
		if candidates[i].wordCount != candidates[j].wordCount {
			return candidates[i].wordCount > candidates[j].wordCount
		}
		return candidates[i].text < candidates[j].text
	})
	return candidates[0].text
}

type jsonEvidenceCandidate struct {
	text      string
	priority  int
	wordCount int
}

func collectJSONEvidence(value any, path []string, candidates *[]jsonEvidenceCandidate) {
	switch typed := value.(type) {
	case map[string]any:
		for _, key := range orderedJSONKeys(typed) {
			collectJSONEvidence(typed[key], append(path, key), candidates)
		}
	case []any:
		limit := len(typed)
		if limit > 8 {
			limit = 8
		}
		for i := 0; i < limit; i++ {
			collectJSONEvidence(typed[i], path, candidates)
		}
	case string:
		text := strings.Join(strings.Fields(typed), " ")
		if text == "" || !usefulJSONScalar(path, text) {
			return
		}
		*candidates = append(*candidates, jsonEvidenceCandidate{
			text:      strings.Join(append(path, text), " "),
			priority:  jsonScalarPriority(path),
			wordCount: len(strings.Fields(text)),
		})
	}
}

func orderedJSONKeys(obj map[string]any) []string {
	keys := make([]string, 0, len(obj))
	for key := range obj {
		keys = append(keys, key)
	}
	sort.SliceStable(keys, func(i, j int) bool {
		left := jsonKeyPriority(keys[i])
		right := jsonKeyPriority(keys[j])
		if left == right {
			return keys[i] < keys[j]
		}
		return left < right
	})
	return keys
}

func jsonKeyPriority(key string) int {
	switch strings.ToLower(key) {
	case "title", "name", "subject":
		return 0
	case "question", "prompt", "heading", "label", "description", "summary":
		return 1
	case "answer":
		return 2
	default:
		return 3
	}
}

func usefulJSONScalar(path []string, text string) bool {
	if len(path) == 0 {
		return false
	}
	last := strings.ToLower(path[len(path)-1])
	switch last {
	case "title", "name", "subject", "description", "summary", "question", "prompt", "heading", "label":
		return true
	case "answer":
		return len(strings.Fields(text)) > 1
	default:
		return false
	}
}

func jsonScalarPriority(path []string) int {
	if len(path) == 0 {
		return 9
	}
	return jsonKeyPriority(path[len(path)-1])
}

func init() { Register(jsonExtractor{}) }
