package report

import (
	"encoding/json"
	"os"
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
