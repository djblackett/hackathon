package app

import (
	"time"

	"github.com/djblackett/bootdev-hackathon/internal/tika"
)

type ScanConfig struct {
	Root             string
	OutPath          string
	TikaURL          string
	TikaClient       *tika.Client
	NoTika           bool
	RequireTika      bool
	TikaTimeout      time.Duration
	UseSiegfried     bool
	SiegfriedTimeout time.Duration
	Hash             bool
	MaxTextPreview   int
	NoTimestamp      bool
}
