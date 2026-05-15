package app

import "time"

type ScanConfig struct {
	Root           string
	OutPath        string
	TikaURL        string
	NoTika         bool
	RequireTika    bool
	TikaTimeout    time.Duration
	Hash           bool
	MaxTextPreview int
	NoTimestamp    bool
}
