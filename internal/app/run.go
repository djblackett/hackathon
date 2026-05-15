package app

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/djblackett/bootdev-hackathon/internal/evidence"
	"github.com/djblackett/bootdev-hackathon/internal/extractors/native"
	"github.com/djblackett/bootdev-hackathon/internal/extractors/siegfried"
	"github.com/djblackett/bootdev-hackathon/internal/extractors/tikaextractor"
	"github.com/djblackett/bootdev-hackathon/internal/plan"
	"github.com/djblackett/bootdev-hackathon/internal/tika"
	"github.com/djblackett/bootdev-hackathon/internal/walk"
)

func Scan(ctx context.Context, cfg ScanConfig) (plan.Plan, error) {
	if cfg.Root == "" {
		return plan.Plan{}, fmt.Errorf("scan root is required")
	}
	if cfg.OutPath == "" {
		cfg.OutPath = "rename-plan.json"
	}
	if cfg.TikaTimeout <= 0 {
		cfg.TikaTimeout = 30 * time.Second
	}
	if cfg.SiegfriedTimeout <= 0 {
		cfg.SiegfriedTimeout = 10 * time.Second
	}
	if cfg.MaxTextPreview <= 0 {
		cfg.MaxTextPreview = 2000
	}

	paths, err := walk.Files(cfg.Root)
	if err != nil {
		return plan.Plan{}, err
	}

	nativeExtractor := native.Extractor{MaxTextPreview: cfg.MaxTextPreview}
	siegfriedExtractor := siegfried.Extractor{Timeout: cfg.SiegfriedTimeout}
	siegfriedUnavailableWarning := ""
	if cfg.UseSiegfried && !siegfriedExtractor.Available(ctx) {
		siegfriedUnavailableWarning = "siegfried unavailable; extraction skipped"
	}
	var tikaExtractor *tikaextractor.Extractor
	tikaUnavailableWarning := ""
	if !cfg.NoTika && (cfg.TikaURL != "" || cfg.TikaClient != nil) {
		client := cfg.TikaClient
		var err error
		if client == nil {
			client, err = tika.NewClientWithHTTPClient(cfg.TikaURL, &http.Client{Timeout: cfg.TikaTimeout})
		}
		if err != nil {
			if cfg.RequireTika {
				return plan.Plan{}, err
			}
			tikaUnavailableWarning = "tika unavailable: " + err.Error()
		} else {
			tikaExtractor = &tikaextractor.Extractor{Client: client, MaxTextPreview: cfg.MaxTextPreview}
			checkCtx, cancel := context.WithTimeout(ctx, cfg.TikaTimeout)
			available := tikaExtractor.Available(checkCtx)
			cancel()
			if !available {
				if cfg.RequireTika {
					return plan.Plan{}, fmt.Errorf("tika server is not available at %s", cfg.TikaURL)
				}
				tikaUnavailableWarning = "tika unavailable; extraction skipped"
				tikaExtractor = nil
			}
		}
	}

	files := make([]evidence.FileEvidence, 0, len(paths))
	for _, path := range paths {
		ev := baseEvidence(path)
		if cfg.Hash {
			if hash, err := sha256File(path); err != nil {
				ev.Errors = append(ev.Errors, evidence.ToolError{Source: evidence.SourceNativeMIME, Message: "sha256 failed: " + err.Error()})
			} else {
				ev.SHA256 = hash
			}
		}
		if tikaUnavailableWarning != "" {
			ev.Warnings = append(ev.Warnings, tikaUnavailableWarning)
		}
		if siegfriedUnavailableWarning != "" {
			ev.Warnings = append(ev.Warnings, siegfriedUnavailableWarning)
		}

		if nativeExtractor.Available(ctx) {
			partial, err := nativeExtractor.Extract(ctx, path)
			if err != nil {
				ev.Errors = append(ev.Errors, evidence.ToolError{Source: evidence.SourceNativeMIME, Message: err.Error()})
			} else {
				ev = evidence.Merge(ev, partial)
			}
		}

		if cfg.UseSiegfried && siegfriedUnavailableWarning == "" {
			extractCtx, cancel := context.WithTimeout(ctx, cfg.SiegfriedTimeout)
			partial, err := siegfriedExtractor.Extract(extractCtx, path)
			cancel()
			if err != nil {
				ev.Errors = append(ev.Errors, evidence.ToolError{Source: evidence.SourceSiegfried, Message: err.Error()})
			} else {
				ev = evidence.Merge(ev, partial)
			}
		}

		if tikaExtractor != nil {
			extractCtx, cancel := context.WithTimeout(ctx, cfg.TikaTimeout)
			partial, err := tikaExtractor.Extract(extractCtx, path)
			cancel()
			if err != nil {
				ev.Errors = append(ev.Errors, evidence.ToolError{Source: evidence.SourceTika, Message: err.Error()})
			} else {
				ev = evidence.Merge(ev, partial)
			}
		}

		files = append(files, ev)
	}

	generatedAt := time.Now().UTC()
	if cfg.NoTimestamp {
		generatedAt = time.Time{}
	}
	p := plan.Build(cfg.Root, files, generatedAt)
	if err := os.MkdirAll(filepath.Dir(cfg.OutPath), 0755); err != nil && filepath.Dir(cfg.OutPath) != "." {
		return plan.Plan{}, err
	}
	if err := plan.Write(cfg.OutPath, p); err != nil {
		return plan.Plan{}, err
	}
	return p, nil
}

func baseEvidence(path string) evidence.FileEvidence {
	ev := evidence.FileEvidence{
		Path:     path,
		Metadata: map[string]string{},
	}
	if st, err := os.Stat(path); err == nil {
		ev.SizeBytes = st.Size()
	} else {
		ev.Errors = append(ev.Errors, evidence.ToolError{Source: evidence.SourceNativeMIME, Message: "stat failed: " + err.Error()})
	}
	return ev
}

func sha256File(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
