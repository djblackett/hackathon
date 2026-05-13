package report

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type Entry struct {
	SourcePath      string   `json:"source_path"`
	DestinationPath string   `json:"destination_path"`
	SuggestedName   string   `json:"suggested_name"`
	Method          string   `json:"method"`
	Confidence      float64  `json:"confidence"`
	Evidence        []string `json:"evidence,omitempty"`
	Warnings        []string `json:"warnings,omitempty"`
	DryRun          bool     `json:"dry_run"`
	Skipped         bool     `json:"skipped,omitempty"`
	SkipReason      string   `json:"skip_reason,omitempty"`
	ReviewStatus    string   `json:"review_status,omitempty"`
	ReviewNote      string   `json:"review_note,omitempty"`
}

type Report struct {
	Summary Summary `json:"summary"`
	Entries []Entry `json:"entries"`
}

type Summary struct {
	TotalFiles         int `json:"total_files"`
	PlannedCount       int `json:"planned_count"`
	CopiedCount        int `json:"copied_count"`
	SkippedCount       int `json:"skipped_count"`
	LowConfidenceCount int `json:"low_confidence_count"`
	AIFallbackCount    int `json:"ai_fallback_count"`
	WarningsCount      int `json:"warnings_count"`
	PendingReviewCount int `json:"pending_review_count"`
	AcceptedCount      int `json:"accepted_count"`
	RejectedCount      int `json:"rejected_count"`
}

func Write(path string, entries []Entry) error {
	if path == "" {
		return nil
	}

	data, err := json.MarshalIndent(Report{Summary: BuildSummary(entries), Entries: entries}, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func BuildSummary(entries []Entry) Summary {
	summary := Summary{TotalFiles: len(entries)}
	for _, entry := range entries {
		if entry.Skipped {
			summary.SkippedCount++
		} else {
			summary.PlannedCount++
			if !entry.DryRun {
				summary.CopiedCount++
			}
		}
		if entry.Confidence < 0.4 {
			summary.LowConfidenceCount++
		}
		if entry.Method == "ai-fallback" || entry.Method == "ai-only" {
			summary.AIFallbackCount++
		}
		if len(entry.Warnings) > 0 {
			summary.WarningsCount++
		}
		status := strings.ToLower(strings.TrimSpace(entry.ReviewStatus))
		if status == "" && entry.Skipped {
			status = "pending"
		}
		switch status {
		case "pending":
			summary.PendingReviewCount++
		case "accepted":
			summary.AcceptedCount++
		case "rejected":
			summary.RejectedCount++
		}
	}
	return summary
}

func Read(path string) (Report, error) {
	var report Report
	data, err := os.ReadFile(path)
	if err != nil {
		return report, err
	}
	if err := json.Unmarshal(data, &report); err != nil {
		return report, err
	}
	return report, nil
}

func WriteReviewMarkdown(path string, report Report) error {
	if path == "" {
		return nil
	}

	var b strings.Builder
	b.WriteString("# File Rename Review\n\n")
	fmt.Fprintf(&b, "- Total files: %d\n", report.Summary.TotalFiles)
	fmt.Fprintf(&b, "- Planned/copied files: %d\n", report.Summary.PlannedCount)
	fmt.Fprintf(&b, "- Skipped files: %d\n", report.Summary.SkippedCount)
	fmt.Fprintf(&b, "- Pending review: %d\n", report.Summary.PendingReviewCount)
	fmt.Fprintf(&b, "- Accepted: %d\n", report.Summary.AcceptedCount)
	fmt.Fprintf(&b, "- Rejected: %d\n\n", report.Summary.RejectedCount)
	b.WriteString("To accept a skipped entry, edit the JSON report and set `review_status` to `accepted`. Then run `--apply-report report.json --include-skipped`.\n\n")
	b.WriteString("| Status | Confidence | Source | Destination | Reason | Notes |\n")
	b.WriteString("|---|---:|---|---|---|---|\n")

	for _, entry := range report.Entries {
		if !entry.Skipped && entry.ReviewStatus == "" {
			continue
		}
		status := entry.ReviewStatus
		if status == "" && entry.Skipped {
			status = "pending"
		}
		fmt.Fprintf(
			&b,
			"| %s | %.2f | `%s` | `%s` | %s | %s |\n",
			escapeMarkdownTable(status),
			entry.Confidence,
			escapeMarkdownCode(entry.SourcePath),
			escapeMarkdownCode(entry.DestinationPath),
			escapeMarkdownTable(entry.SkipReason),
			escapeMarkdownTable(entry.ReviewNote),
		)
	}

	return os.WriteFile(path, []byte(b.String()), 0644)
}

func NormalizeReviewStatus(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "", "pending":
		return "pending"
	case "accepted", "accept", "approved", "approve":
		return "accepted"
	case "rejected", "reject", "denied", "deny":
		return "rejected"
	default:
		return "pending"
	}
}

func escapeMarkdownTable(s string) string {
	s = strings.ReplaceAll(s, "|", "\\|")
	s = strings.ReplaceAll(s, "\n", " ")
	return s
}

func escapeMarkdownCode(s string) string {
	return strings.ReplaceAll(s, "`", "'")
}
