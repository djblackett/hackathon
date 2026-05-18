package plan

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/djblackett/bootdev-hackathon/internal/evidence"
	"github.com/djblackett/bootdev-hackathon/internal/naming"
)

type Plan struct {
	Version     int    `json:"version"`
	Root        string `json:"root"`
	GeneratedAt string `json:"generatedAt,omitempty"`
	Items       []Item `json:"items"`
}

type Item struct {
	OldPath       string                `json:"oldPath"`
	SuggestedPath string                `json:"suggestedPath"`
	Confidence    naming.Confidence     `json:"confidence"`
	Score         float64               `json:"score"`
	Evidence      evidence.FileEvidence `json:"evidence"`
	Reasons       []string              `json:"reasons"`
	Warnings      []string              `json:"warnings,omitempty"`
	Conflict      *ConflictResolution   `json:"conflict,omitempty"`
}

type ConflictResolution struct {
	OriginalSuggestedPath string `json:"originalSuggestedPath"`
	Reason                string `json:"reason"`
}

func Build(root string, files []evidence.FileEvidence, generatedAt time.Time) Plan {
	p := Plan{
		Version: 1,
		Root:    root,
		Items:   make([]Item, 0, len(files)),
	}
	if !generatedAt.IsZero() {
		p.GeneratedAt = generatedAt.UTC().Format(time.RFC3339)
	}

	reserved := map[string]struct{}{}
	for i, ev := range files {
		suggestion := naming.Generate(ev, i+1)
		suggested := filepath.Join(filepath.Dir(ev.Path), suggestion.Filename)
		resolved, conflict := reserveSuggestedPath(ev.Path, suggested, reserved)
		warnings := append([]string(nil), suggestion.Warnings...)
		for _, toolErr := range ev.Errors {
			warnings = append(warnings, string(toolErr.Source)+": "+toolErr.Message)
		}
		if conflict != nil {
			warnings = append(warnings, conflict.Reason)
		}

		p.Items = append(p.Items, Item{
			OldPath:       ev.Path,
			SuggestedPath: resolved,
			Confidence:    suggestion.Confidence,
			Score:         suggestion.Score,
			Evidence:      ev,
			Reasons:       suggestion.Reasons,
			Warnings:      warnings,
			Conflict:      conflict,
		})
	}
	return p
}

func Write(path string, p Plan) error {
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0644)
}

func WriteReviewMarkdown(path string, p Plan) error {
	if path == "" {
		return nil
	}

	var b strings.Builder
	b.WriteString("# Rename Plan Review\n\n")
	fmt.Fprintf(&b, "- Root: `%s`\n", escapeMarkdownCode(p.Root))
	fmt.Fprintf(&b, "- Total files: %d\n", len(p.Items))
	fmt.Fprintf(&b, "- High confidence: %d\n", countPlanConfidence(p, naming.ConfidenceHigh))
	fmt.Fprintf(&b, "- Medium confidence: %d\n", countPlanConfidence(p, naming.ConfidenceMedium))
	fmt.Fprintf(&b, "- Low confidence: %d\n", countPlanConfidence(p, naming.ConfidenceLow))
	fmt.Fprintf(&b, "- No confidence: %d\n", countPlanConfidence(p, naming.ConfidenceNone))
	fmt.Fprintf(&b, "- Warnings: %d\n\n", countPlanWarnings(p))
	b.WriteString("Review low-confidence rows and any rows with warnings before applying or scripting changes from this plan.\n\n")
	b.WriteString("| Confidence | Score | Old Path | Suggested Path | Reasons | Warnings |\n")
	b.WriteString("|---|---:|---|---|---|---|\n")

	for _, item := range p.Items {
		fmt.Fprintf(
			&b,
			"| %s | %.2f | `%s` | `%s` | %s | %s |\n",
			escapeMarkdownTable(string(item.Confidence)),
			item.Score,
			escapeMarkdownCode(item.OldPath),
			escapeMarkdownCode(item.SuggestedPath),
			escapeMarkdownTable(strings.Join(item.Reasons, "; ")),
			escapeMarkdownTable(strings.Join(item.Warnings, "; ")),
		)
	}

	return os.WriteFile(path, []byte(b.String()), 0644)
}

func countPlanConfidence(p Plan, confidence naming.Confidence) int {
	count := 0
	for _, item := range p.Items {
		if item.Confidence == confidence {
			count++
		}
	}
	return count
}

func countPlanWarnings(p Plan) int {
	count := 0
	for _, item := range p.Items {
		if len(item.Warnings) > 0 {
			count++
		}
	}
	return count
}

func reserveSuggestedPath(oldPath, suggested string, reserved map[string]struct{}) (string, *ConflictResolution) {
	if pathAvailable(oldPath, suggested, reserved) {
		reserved[suggested] = struct{}{}
		return suggested, nil
	}

	original := suggested
	ext := filepath.Ext(suggested)
	base := strings.TrimSuffix(suggested, ext)
	for i := 2; ; i++ {
		candidate := base + "_" + leftPad3(i) + ext
		if pathAvailable(oldPath, candidate, reserved) {
			reserved[candidate] = struct{}{}
			return candidate, &ConflictResolution{
				OriginalSuggestedPath: original,
				Reason:                "suggested path conflicted with another plan item or existing file",
			}
		}
	}
}

func pathAvailable(oldPath, candidate string, reserved map[string]struct{}) bool {
	if _, ok := reserved[candidate]; ok {
		return false
	}
	if sameCleanPath(oldPath, candidate) {
		return true
	}
	_, err := os.Stat(candidate)
	return os.IsNotExist(err)
}

func sameCleanPath(a, b string) bool {
	return filepath.Clean(a) == filepath.Clean(b)
}

func leftPad3(value int) string {
	return fmt.Sprintf("%03d", value)
}

func escapeMarkdownTable(s string) string {
	s = strings.ReplaceAll(s, "|", "\\|")
	s = strings.ReplaceAll(s, "\n", " ")
	return s
}

func escapeMarkdownCode(s string) string {
	return strings.ReplaceAll(s, "`", "'")
}
