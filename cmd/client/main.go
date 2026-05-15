package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/joho/godotenv"
	"github.com/urfave/cli/v2"

	"github.com/djblackett/bootdev-hackathon/internal/ai"
	"github.com/djblackett/bootdev-hackathon/internal/analysis"
	"github.com/djblackett/bootdev-hackathon/internal/config"
	"github.com/djblackett/bootdev-hackathon/internal/extractors"
	"github.com/djblackett/bootdev-hackathon/internal/report"
	"github.com/djblackett/bootdev-hackathon/internal/utils"
)

var newAIClient = ai.NewClient

func main() {
	if err := runApp(os.Args); err != nil {
		log.Fatal(err)
	}
}

func runApp(args []string) error {
	// Load environment variables from .env (useful during local dev).
	_ = godotenv.Load() // Ignore errors if .env is not present - for docker

	// Build a typed config object that holds OpenAI/Ollama creds, etc.
	cfg := config.FromEnv()

	// Reasonable default extensions; can be overridden with --types "csv,html,…".
	defaultFileTypes := []string{"txt", "text", "md", "markdown", "rtf", "csv", "pdf", "json", "ipynb", "notebook", "epub", "odt", "ods", "odp", "opendocument", "zip", "tar", "tgz", "archive", "html", "xml", "musicxml", "log", "cfg", "ini", "docx", "xlsx", "pptx", "office", "eml", "email", "image", "media"}

	// Define CLI application.
	app := &cli.App{
		Name:  "ai-file-renamer",
		Usage: "rename recovered docs via AI",
		Flags: []cli.Flag{
			&cli.StringFlag{ // where to start scanning; required.
				Name:  "input",
				Value: "files/input", // default input directory
				Usage: "input directory to scan for files",
			},
			&cli.StringFlag{ // where to write output files.
				Name:  "output",
				Value: "files/output", // default output directory
				Usage: "output directory for processed files",
			},
			&cli.BoolFlag{ // switch between OpenAI API and local Ollama.
				Name:  "local",
				Usage: "use local Ollama LLM",
			},
			&cli.StringFlag{ // e.g. "mistral", "gpt-4o".
				Name:  "model",
				Usage: "model name (ollama or openai)",
			},
			&cli.BoolFlag{ // dry‑run means log only.
				Name:  "dry-run",
				Usage: "preview changes only",
			},
			&cli.BoolFlag{ // rename instead of copy; copy is safer default.
				Name:  "rename",
				Usage: "rename files in place instead of copying to output directory",
			},
			&cli.BoolFlag{
				Name:  "debug",
				Usage: "return all errors joined together",
			},
			&cli.BoolFlag{
				Name:  "flatten",
				Value: false,
				Usage: "flatten output directory structure",
			},
			&cli.StringFlag{
				Name:  "strategy",
				Value: "auto",
				Usage: "rename strategy: auto, metadata-only, or ai-only",
			},
			&cli.Float64Flag{
				Name:  "confidence-threshold",
				Value: 0.75,
				Usage: "minimum local confidence before auto mode skips AI fallback",
			},
			&cli.IntFlag{
				Name:  "max-ai-chars",
				Value: 2000,
				Usage: "maximum compact evidence characters sent to AI in auto mode",
			},
			&cli.Float64Flag{
				Name:  "min-confidence-to-copy",
				Value: 0,
				Usage: "minimum confidence required before copying files; 0 disables copy skipping",
			},
			&cli.StringFlag{
				Name:  "report",
				Usage: "write a JSON report of processed files",
			},
			&cli.StringFlag{
				Name:  "apply-report",
				Usage: "copy files using destinations from a previously generated JSON report",
			},
			&cli.StringFlag{
				Name:  "apply-accepted",
				Usage: "copy planned files and accepted skipped entries from a JSON report",
			},
			&cli.StringFlag{
				Name:  "list-pending",
				Usage: "print pending review entries from a JSON report",
			},
			&cli.StringFlag{
				Name:  "set-review-status",
				Usage: "update review status values in a JSON report",
			},
			&cli.StringSliceFlag{
				Name:  "review-entry",
				Usage: "review update in source=status form; status may be accepted, rejected, or pending",
			},
			&cli.StringSliceFlag{
				Name:  "review-note",
				Usage: "review note update in source=note form",
			},
			&cli.StringFlag{
				Name:  "explain",
				Usage: "explain the metadata filename suggestion for one file",
			},
			&cli.BoolFlag{
				Name:  "include-skipped",
				Usage: "when applying a report, also copy skipped entries marked review_status=accepted",
			},
			&cli.StringFlag{
				Name:  "review-report",
				Usage: "write a Markdown review file for skipped or reviewed report entries",
			},
			&cli.StringSliceFlag{ // allowed extensions. Overrides defaultFileTypes.
				Name:  "types",
				Value: cli.NewStringSlice(defaultFileTypes...),
				Usage: "file types to process (comma-separated, e.g. txt,pdf,md)",
			},
		},
		Action: func(c *cli.Context) error {
			// Harvest flag values.
			input := c.String("input")
			output := c.String("output")
			local := c.Bool("local")
			model := c.String("model")
			dry := c.Bool("dry-run")
			renameMode := c.Bool("rename")
			debug := c.Bool("debug")
			flatten := c.Bool("flatten")
			strategy := c.String("strategy")
			confidenceThreshold := c.Float64("confidence-threshold")
			maxAIChars := c.Int("max-ai-chars")
			minConfidenceToCopy := c.Float64("min-confidence-to-copy")
			reportPath := c.String("report")
			applyReportPath := c.String("apply-report")
			applyAcceptedPath := c.String("apply-accepted")
			listPendingPath := c.String("list-pending")
			setReviewStatusPath := c.String("set-review-status")
			reviewEntries := c.StringSlice("review-entry")
			reviewNotes := c.StringSlice("review-note")
			explainPath := c.String("explain")
			includeSkipped := c.Bool("include-skipped")
			reviewReportPath := c.String("review-report")

			if explainPath != "" {
				return explainFile(explainPath, os.Stdout)
			}
			if setReviewStatusPath != "" {
				return updateReviewStatus(setReviewStatusPath, reviewEntries, reviewNotes)
			}
			if listPendingPath != "" {
				return listPendingReport(listPendingPath, os.Stdout)
			}
			if applyAcceptedPath != "" {
				return applyReport(applyAcceptedPath, dry, true, reviewReportPath)
			}
			if applyReportPath != "" {
				return applyReport(applyReportPath, dry, includeSkipped, reviewReportPath)
			}

			switch strategy {
			case "auto", "metadata-only", "ai-only":
			default:
				return fmt.Errorf("invalid strategy %q: use auto, metadata-only, or ai-only", strategy)
			}

			// Build a set[string]struct{} for O(1) membership tests during walk.
			types := make(map[string]struct{})
			for _, t := range c.StringSlice("types") {
				types[t] = struct{}{}
			}

			defaultModel := model
			// set different default models based on local flag.
			if model == "" && local {
				defaultModel = "mistral"
			} else if model == "" && !local {
				defaultModel = "gpt-3.5-turbo"
			}

			var (
				client     ai.Client
				clientErr  error
				clientOnce sync.Once
			)
			getAIClient := func() (ai.Client, error) {
				clientOnce.Do(func() {
					// Spin up LLM client once; reused by all goroutines.
					client, clientErr = newAIClient(cfg, local, defaultModel)
				})
				return client, clientErr
			}

			// Concurrency helpers.
			var wg sync.WaitGroup
			errChan := make(chan error, 100) // buffered to avoid blocking goroutines.
			var outputMu sync.Mutex
			reservedPaths := map[string]struct{}{}
			var reportMu sync.Mutex
			reportEntries := []report.Entry{}

			// processFile runs in its own goroutine per file.
			processFile := func(info extractors.ExtractedFileInfo) {
				defer wg.Done()

				path := info.Path
				suggestion := analysis.GenerateFilename(info)
				suggested := suggestion.Filename
				method := suggestion.Method
				confidence := suggestion.Confidence
				evidence := append([]string(nil), suggestion.Evidence...)

				if strategy == "ai-only" || (strategy == "auto" && confidence < confidenceThreshold) {
					client, err := getAIClient()
					if err != nil {
						errChan <- err
						return
					}

					content := info.RawContent
					if strategy == "auto" {
						content = analysis.CompactEvidence(info, maxAIChars)
						method = "ai-fallback"
						log.Printf("[AI] %s local confidence %.2f below threshold %.2f; sending %d compact evidence chars\n", path, confidence, confidenceThreshold, len(content))
					} else {
						method = "ai-only"
					}

					var aiSuggested string
					if strategy == "auto" {
						aiSuggested, err = client.SuggestFilenameFromEvidence(content)
					} else {
						aiSuggested, err = client.SuggestFilename(content)
					}
					if err != nil {
						errChan <- err
						return
					}
					suggested = aiSuggested
					confidence = 1
				}

				// Sanitize to avoid invalid characters.
				sanitized := utils.Sanitize(suggested)

				ext := filepath.Ext(path)
				if info.SuggestedExtension != "" {
					ext = "." + info.SuggestedExtension
				}

				var destPath string
				if renameMode {
					planned := filepath.Join(filepath.Dir(path), sanitized+ext)
					outputMu.Lock()
					if dry {
						destPath = utils.UniquePlannedPath(planned, reservedPaths)
					} else {
						destPath = utils.UniquePath(planned, reservedPaths)
					}
					outputMu.Unlock()
				} else {
					planned, err := utils.DestinationPath(input, path, output, sanitized, ext, flatten)
					if err != nil {
						errChan <- err
						return
					}
					outputMu.Lock()
					if dry {
						destPath = utils.UniquePlannedPath(planned, reservedPaths)
					} else {
						destPath = utils.UniquePath(planned, reservedPaths)
					}
					outputMu.Unlock()
				}

				skipped := false
				skipReason := ""
				reviewStatus := ""
				if !renameMode && minConfidenceToCopy > 0 && confidence < minConfidenceToCopy {
					skipped = true
					skipReason = fmt.Sprintf("confidence %.2f below copy threshold %.2f", confidence, minConfidenceToCopy)
					reviewStatus = "pending"
				}

				reportMu.Lock()
				reportEntries = append(reportEntries, report.Entry{
					SourcePath:      path,
					DestinationPath: destPath,
					SuggestedName:   sanitized + ext,
					Method:          method,
					Confidence:      confidence,
					Evidence:        evidence,
					Warnings:        append([]string(nil), info.Warnings...),
					DryRun:          dry,
					Skipped:         skipped,
					SkipReason:      skipReason,
					ReviewStatus:    reviewStatus,
				})
				reportMu.Unlock()

				// Depending on flags, perform or log the operation.
				switch {
				case dry && !renameMode:
					log.Printf("[DRY] %s  →  %s method=%s confidence=%.2f\n", path, destPath, method, confidence)
				case dry && renameMode:
					log.Printf("[DRY] %s  →  %s method=%s confidence=%.2f\n", path, destPath, method, confidence)
				case skipped:
					log.Printf("[SKIP] %s  →  %s method=%s confidence=%.2f reason=%s\n", path, destPath, method, confidence, skipReason)
				case !renameMode:
					if err := utils.CopyFileToPath(path, destPath); err != nil {
						errChan <- err
					}
				default: // rename in place
					if err := utils.RenameFileWithExtension(input, path, sanitized, ext); err != nil {
						errChan <- err
					}
				}
			}

			// Walk the directory tree and dispatch work.
			if err := extractors.WalkInfo(input, types, func(info extractors.ExtractedFileInfo) error {
				wg.Add(1)
				go processFile(info)
				return nil // continue walking
			}); err != nil {
				return err
			}

			// Wait for all processing to complete, then close channel so range terminates.
			wg.Wait()
			close(errChan)
			sort.Slice(reportEntries, func(i, j int) bool {
				return reportEntries[i].SourcePath < reportEntries[j].SourcePath
			})

			if err := report.Write(reportPath, reportEntries); err != nil {
				return err
			}
			if reviewReportPath != "" {
				if err := writeReviewFromEntries(reviewReportPath, reportEntries); err != nil {
					return err
				}
			}
			fmt.Fprintln(os.Stdout, formatSummary(report.BuildSummary(reportEntries)))

			var (
				firstErr error   // keeps behaviour for non‑debug
				allErrs  []error // collects when debug
			)

			for e := range errChan {
				if firstErr == nil {
					firstErr = e
				}
				if debug {
					allErrs = append(allErrs, e)
				}
			}

			switch {
			case debug && len(allErrs) > 0:
				return errors.Join(allErrs...)
			default:
				return firstErr
			}
		},
	}

	// Kick everything off.
	return app.Run(args)
}

func applyReport(path string, dry bool, includeSkipped bool, reviewReportPath string) error {
	planned, err := report.Read(path)
	if err != nil {
		return err
	}
	if reviewReportPath != "" {
		if err := report.WriteReviewMarkdown(reviewReportPath, planned); err != nil {
			return err
		}
	}

	applied := 0
	dryRunCount := 0
	skippedCount := 0
	for i, entry := range planned.Entries {
		if entry.SourcePath == "" && entry.DestinationPath == "" {
			continue
		}
		if entry.SourcePath == "" {
			return fmt.Errorf("report entry %d has empty source path", i)
		}
		if entry.DestinationPath == "" {
			return fmt.Errorf("report entry %d has empty destination path", i)
		}
		if _, err := os.Stat(entry.SourcePath); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("report entry %d source file does not exist: %s", i, entry.SourcePath)
			}
			return fmt.Errorf("report entry %d source file could not be checked: %s: %w", i, entry.SourcePath, err)
		}
		if entry.Skipped {
			status := report.NormalizeReviewStatus(entry.ReviewStatus)
			if !includeSkipped || status != "accepted" {
				skippedCount++
				log.Printf("[SKIP] %s  →  %s method=apply-report status=%s reason=%s\n", entry.SourcePath, entry.DestinationPath, status, entry.SkipReason)
				continue
			}
		}
		if dry {
			dryRunCount++
			log.Printf("[DRY] %s  →  %s method=apply-report confidence=%.2f\n", entry.SourcePath, entry.DestinationPath, entry.Confidence)
			continue
		}
		if err := utils.CopyFileToPath(entry.SourcePath, entry.DestinationPath); err != nil {
			return err
		}
		applied++
		log.Printf("[APPLY] %s  →  %s\n", entry.SourcePath, entry.DestinationPath)
	}
	fmt.Fprintf(os.Stdout, "apply_summary: total=%d applied=%d dry_run=%d skipped=%d\n", len(planned.Entries), applied, dryRunCount, skippedCount)
	return nil
}

func writeReviewFromEntries(path string, entries []report.Entry) error {
	return report.WriteReviewMarkdown(path, report.Report{
		Summary: report.BuildSummary(entries),
		Entries: entries,
	})
}

func listPendingReport(path string, out io.Writer) error {
	planned, err := report.Read(path)
	if err != nil {
		return err
	}

	pending := pendingEntries(planned.Entries)
	fmt.Fprintf(out, "Pending review: %d\n", len(pending))
	for _, entry := range pending {
		fmt.Fprintf(out, "\n%s\n", entry.SourcePath)
		fmt.Fprintf(out, "  destination: %s\n", entry.DestinationPath)
		fmt.Fprintf(out, "  confidence: %.2f\n", entry.Confidence)
		if len(entry.Evidence) > 0 {
			fmt.Fprintf(out, "  evidence: %s\n", strings.Join(entry.Evidence, ", "))
		}
		if len(entry.Warnings) > 0 {
			fmt.Fprintf(out, "  warnings: %s\n", strings.Join(entry.Warnings, "; "))
		}
		if entry.SkipReason != "" {
			fmt.Fprintf(out, "  reason: %s\n", entry.SkipReason)
		}
	}
	return nil
}

func pendingEntries(entries []report.Entry) []report.Entry {
	pending := []report.Entry{}
	for _, entry := range entries {
		status := report.NormalizeReviewStatus(entry.ReviewStatus)
		if entry.Skipped && status == "pending" {
			pending = append(pending, entry)
		}
	}
	sort.Slice(pending, func(i, j int) bool {
		return pending[i].SourcePath < pending[j].SourcePath
	})
	return pending
}

func updateReviewStatus(path string, updates []string, notes []string) error {
	if len(updates) == 0 && len(notes) == 0 {
		return fmt.Errorf("at least one --review-entry or --review-note is required")
	}
	planned, err := report.Read(path)
	if err != nil {
		return err
	}
	changed := 0
	for _, raw := range updates {
		key, value, ok := strings.Cut(raw, "=")
		if !ok || strings.TrimSpace(key) == "" {
			return fmt.Errorf("invalid --review-entry %q: use source=status", raw)
		}
		status := report.NormalizeReviewStatus(value)
		n, err := updateMatchingEntries(planned.Entries, key, func(entry *report.Entry) {
			entry.ReviewStatus = status
			if entry.Skipped && status == "accepted" {
				entry.SkipReason = ""
			}
		})
		if err != nil {
			return err
		}
		planned.Entries = n.entries
		changed += n.count
	}
	for _, raw := range notes {
		key, value, ok := strings.Cut(raw, "=")
		if !ok || strings.TrimSpace(key) == "" {
			return fmt.Errorf("invalid --review-note %q: use source=note", raw)
		}
		n, err := updateMatchingEntries(planned.Entries, key, func(entry *report.Entry) {
			entry.ReviewNote = value
		})
		if err != nil {
			return err
		}
		planned.Entries = n.entries
		changed += n.count
	}
	if err := report.Write(path, planned.Entries); err != nil {
		return err
	}
	fmt.Fprintf(os.Stdout, "Updated review entries: %d\n", changed)
	fmt.Fprintln(os.Stdout, formatSummary(report.BuildSummary(planned.Entries)))
	return nil
}

type updateResult struct {
	entries []report.Entry
	count   int
}

func updateMatchingEntries(entries []report.Entry, selector string, update func(*report.Entry)) (updateResult, error) {
	selector = strings.TrimSpace(selector)
	matches := []int{}
	for i, entry := range entries {
		if reportEntryMatches(entry, selector) {
			matches = append(matches, i)
		}
	}
	if len(matches) == 0 {
		return updateResult{}, fmt.Errorf("no report entry matches %q", selector)
	}
	if len(matches) > 1 {
		return updateResult{}, fmt.Errorf("multiple report entries match %q; use the full source path", selector)
	}
	update(&entries[matches[0]])
	return updateResult{entries: entries, count: 1}, nil
}

func reportEntryMatches(entry report.Entry, selector string) bool {
	return entry.SourcePath == selector ||
		entry.DestinationPath == selector ||
		entry.SuggestedName == selector ||
		filepath.Base(entry.SourcePath) == selector
}

func explainFile(path string, out io.Writer) error {
	info, err := extractors.ExtractInfoForPath(path)
	if err != nil {
		return err
	}
	suggestion := analysis.GenerateFilename(info)
	ext := filepath.Ext(path)
	if info.SuggestedExtension != "" {
		ext = "." + info.SuggestedExtension
	}
	fmt.Fprintf(out, "source: %s\n", path)
	fmt.Fprintf(out, "detected_type: %s\n", info.DetectedType)
	if info.SuggestedExtension != "" {
		fmt.Fprintf(out, "suggested_extension: %s\n", info.SuggestedExtension)
	}
	fmt.Fprintf(out, "suggested_name: %s%s\n", suggestion.Filename, ext)
	fmt.Fprintf(out, "method: %s\n", suggestion.Method)
	fmt.Fprintf(out, "confidence: %.2f\n", suggestion.Confidence)
	if len(suggestion.Evidence) > 0 {
		fmt.Fprintf(out, "evidence: %s\n", strings.Join(suggestion.Evidence, ", "))
	}
	if len(info.Warnings) > 0 {
		fmt.Fprintf(out, "warnings: %s\n", strings.Join(info.Warnings, "; "))
	}
	samples := analysis.RankEvidence(info)
	if len(samples) > 0 {
		fmt.Fprintln(out, "top_evidence:")
		for i, sample := range samples {
			if i >= 5 {
				break
			}
			fmt.Fprintf(out, "- %s %.2f %s\n", sample.Source, sample.Score, trimExplain(sample.Text, 120))
		}
	}
	return nil
}

func trimExplain(s string, max int) string {
	s = strings.Join(strings.Fields(s), " ")
	if len(s) <= max {
		return s
	}
	return strings.TrimSpace(s[:max])
}

func formatSummary(summary report.Summary) string {
	return fmt.Sprintf(
		"summary: total=%d planned=%d copied=%d skipped=%d pending_review=%d warnings=%d ai_fallback=%d",
		summary.TotalFiles,
		summary.PlannedCount,
		summary.CopiedCount,
		summary.SkippedCount,
		summary.PendingReviewCount,
		summary.WarningsCount,
		summary.AIFallbackCount,
	)
}
