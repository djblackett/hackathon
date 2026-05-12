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
}

type Report struct {
	Entries []Entry `json:"entries"`
}

func Write(path string, entries []Entry) error {
	if path == "" {
		return nil
	}

	data, err := json.MarshalIndent(Report{Entries: entries}, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
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
