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
	"strings"
	"time"

	"github.com/djblackett/bootdev-hackathon/internal/evidence"
	"github.com/djblackett/bootdev-hackathon/internal/extractors/exiftool"
	"github.com/djblackett/bootdev-hackathon/internal/extractors/ffprobe"
	"github.com/djblackett/bootdev-hackathon/internal/extractors/jhove"
	"github.com/djblackett/bootdev-hackathon/internal/extractors/native"
	"github.com/djblackett/bootdev-hackathon/internal/extractors/siegfried"
	"github.com/djblackett/bootdev-hackathon/internal/extractors/tesseract"
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
	if cfg.ExifToolTimeout <= 0 {
		cfg.ExifToolTimeout = 15 * time.Second
	}
	if cfg.FFProbeTimeout <= 0 {
		cfg.FFProbeTimeout = 15 * time.Second
	}
	if cfg.JHOVETimeout <= 0 {
		cfg.JHOVETimeout = 30 * time.Second
	}
	if cfg.OCRTimeout <= 0 {
		cfg.OCRTimeout = 60 * time.Second
	}
	if strings.TrimSpace(cfg.OCRLang) == "" {
		cfg.OCRLang = "eng"
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
	exifToolExtractor := exiftool.Extractor{Timeout: cfg.ExifToolTimeout}
	exifToolUnavailable := cfg.UseExifTool && !exifToolExtractor.Available(ctx)
	ffprobeExtractor := ffprobe.Extractor{Timeout: cfg.FFProbeTimeout}
	ffprobeUnavailable := cfg.UseFFProbe && !ffprobeExtractor.Available(ctx)
	jhoveExtractor := jhove.Extractor{Timeout: cfg.JHOVETimeout}
	jhoveUnavailableWarning := ""
	if cfg.Validate && !jhoveExtractor.Available(ctx) {
		jhoveUnavailableWarning = "jhove unavailable; validation skipped"
	}
	tesseractExtractor := tesseract.Extractor{Timeout: cfg.OCRTimeout, Lang: cfg.OCRLang, MaxTextPreview: cfg.MaxTextPreview}
	tesseractUnavailable := cfg.UseOCR && !tesseractExtractor.Available(ctx)
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
		if jhoveUnavailableWarning != "" {
			ev.Warnings = append(ev.Warnings, jhoveUnavailableWarning)
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

		if cfg.UseExifTool && shouldRunExifTool(ev) {
			if exifToolUnavailable {
				ev.Warnings = append(ev.Warnings, "exiftool unavailable; extraction skipped")
			} else {
				extractCtx, cancel := context.WithTimeout(ctx, cfg.ExifToolTimeout)
				partial, err := exifToolExtractor.Extract(extractCtx, path)
				cancel()
				if err != nil {
					ev.Errors = append(ev.Errors, evidence.ToolError{Source: evidence.SourceExifTool, Message: err.Error()})
				} else {
					ev = evidence.Merge(ev, partial)
				}
			}
		}

		if cfg.UseFFProbe && shouldRunFFProbe(ev) {
			if ffprobeUnavailable {
				ev.Warnings = append(ev.Warnings, "ffprobe unavailable; extraction skipped")
			} else {
				extractCtx, cancel := context.WithTimeout(ctx, cfg.FFProbeTimeout)
				partial, err := ffprobeExtractor.Extract(extractCtx, path)
				cancel()
				if err != nil {
					ev.Errors = append(ev.Errors, evidence.ToolError{Source: evidence.SourceFFProbe, Message: err.Error()})
				} else {
					ev = evidence.Merge(ev, partial)
				}
			}
		}

		if cfg.UseOCR && shouldRunOCR(ev) {
			if tesseractUnavailable {
				ev.Warnings = append(ev.Warnings, "tesseract unavailable; OCR skipped")
			} else {
				extractCtx, cancel := context.WithTimeout(ctx, cfg.OCRTimeout)
				partial, err := tesseractExtractor.Extract(extractCtx, path)
				cancel()
				if err != nil {
					ev.Errors = append(ev.Errors, evidence.ToolError{Source: evidence.SourceTesseract, Message: err.Error()})
				} else {
					ev = evidence.Merge(ev, partial)
				}
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

		if cfg.Validate && jhoveUnavailableWarning == "" {
			extractCtx, cancel := context.WithTimeout(ctx, cfg.JHOVETimeout)
			partial, err := jhoveExtractor.Extract(extractCtx, path)
			cancel()
			if err != nil {
				ev.Errors = append(ev.Errors, evidence.ToolError{Source: evidence.SourceJHOVE, Message: err.Error()})
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

func shouldRunOCR(ev evidence.FileEvidence) bool {
	mime := strings.ToLower(ev.DetectedMIME)
	ext := strings.ToLower(ev.Extension)
	if strings.HasPrefix(mime, "image/") {
		return true
	}
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".tif", ".tiff", ".heic":
		return true
	default:
		return false
	}
}

func shouldRunFFProbe(ev evidence.FileEvidence) bool {
	mime := strings.ToLower(ev.DetectedMIME)
	ext := strings.ToLower(ev.Extension)
	if strings.HasPrefix(mime, "audio/") || strings.HasPrefix(mime, "video/") {
		return true
	}
	switch ext {
	case ".mp3", ".m4a", ".wav", ".flac", ".mp4", ".mov", ".mkv", ".avi":
		return true
	default:
		return false
	}
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

func shouldRunExifTool(ev evidence.FileEvidence) bool {
	mime := strings.ToLower(ev.DetectedMIME)
	ext := strings.ToLower(ev.Extension)
	if strings.HasPrefix(mime, "image/") || strings.HasPrefix(mime, "audio/") || strings.HasPrefix(mime, "video/") {
		return true
	}
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".tif", ".tiff", ".heic", ".mp3", ".m4a", ".wav", ".flac", ".mp4", ".mov", ".mkv", ".avi":
		return true
	default:
		return false
	}
}
