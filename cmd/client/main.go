package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
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
	defaultFileTypes := []string{"txt", "text", "md", "markdown", "csv", "pdf", "json", "html", "log", "cfg", "ini", "docx", "xlsx", "pptx", "office", "eml", "email", "image", "media"}

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
			&cli.StringFlag{
				Name:  "report",
				Usage: "write a JSON report of processed files",
			},
			&cli.StringFlag{
				Name:  "apply-report",
				Usage: "copy files using destinations from a previously generated JSON report",
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
			reportPath := c.String("report")
			applyReportPath := c.String("apply-report")

			if applyReportPath != "" {
				return applyReport(applyReportPath, dry)
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
					destPath = utils.UniquePath(planned, reservedPaths)
					outputMu.Unlock()
				} else {
					planned, err := utils.DestinationPath(input, path, output, sanitized, ext, flatten)
					if err != nil {
						errChan <- err
						return
					}
					outputMu.Lock()
					destPath = utils.UniquePath(planned, reservedPaths)
					outputMu.Unlock()
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
				})
				reportMu.Unlock()

				// Depending on flags, perform or log the operation.
				switch {
				case dry && !renameMode:
					log.Printf("[DRY] %s  →  %s method=%s confidence=%.2f\n", path, destPath, method, confidence)
				case dry && renameMode:
					log.Printf("[DRY] %s  →  %s method=%s confidence=%.2f\n", path, destPath, method, confidence)
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

			if err := report.Write(reportPath, reportEntries); err != nil {
				return err
			}

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

func applyReport(path string, dry bool) error {
	planned, err := report.Read(path)
	if err != nil {
		return err
	}

	for _, entry := range planned.Entries {
		if entry.SourcePath == "" || entry.DestinationPath == "" {
			continue
		}
		if dry {
			log.Printf("[DRY] %s  →  %s method=apply-report confidence=%.2f\n", entry.SourcePath, entry.DestinationPath, entry.Confidence)
			continue
		}
		if err := utils.CopyFileToPath(entry.SourcePath, entry.DestinationPath); err != nil {
			return err
		}
		log.Printf("[APPLY] %s  →  %s\n", entry.SourcePath, entry.DestinationPath)
	}
	return nil
}
