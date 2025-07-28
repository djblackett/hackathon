package main

import (
	"errors"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/joho/godotenv"
	"github.com/urfave/cli/v2"

	"github.com/djblackett/bootdev-hackathon/internal/ai"
	"github.com/djblackett/bootdev-hackathon/internal/config"
	"github.com/djblackett/bootdev-hackathon/internal/extractors"
	"github.com/djblackett/bootdev-hackathon/internal/utils"
)

func main() {
	// Load environment variables from .env (useful during local dev).
	_ = godotenv.Load() // Ignore errors if .env is not present - for docker

	// Build a typed config object that holds OpenAI/Ollama creds, etc.
	cfg := config.FromEnv()

	// Reasonable default extensions; can be overridden with --types "csv,html,…".
	defaultFileTypes := []string{"txt", "md", "csv", "pdf", "json", "html", "log", "cfg", "ini"}

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

			// Spin up LLM client once; reused by all goroutines.
			client, err := ai.NewClient(cfg, local, defaultModel)
			if err != nil {
				return err
			}

			// Concurrency helpers.
			var wg sync.WaitGroup
			errChan := make(chan error, 100) // buffered to avoid blocking goroutines.

			// processFile runs in its own goroutine per file.
			processFile := func(path, content string) {
				defer wg.Done()

				// Ask the model for a descriptive filename.
				suggested, err := client.SuggestFilename(content)
				if err != nil {
					errChan <- err
					return
				}

				// Sanitize to avoid invalid characters.
				sanitized := utils.Sanitize(suggested)

				ext := filepath.Ext(path)

				// Depending on flags, perform or log the operation.
				switch {
				case dry && !renameMode:
					log.Printf("[DRY] %s  →  %s/%s\n", path, output, sanitized+ext)
				case dry && renameMode:
					log.Printf("[DRY] %s  →  %s\n", path, sanitized+ext)
				case !renameMode:
					if err := utils.CopyFile(input, path, output, sanitized, flatten); err != nil {
						errChan <- err
					}
				default: // rename in place
					if err := utils.RenameFile(input, path, sanitized); err != nil {
						errChan <- err
					}
				}
			}

			// Walk the directory tree and dispatch work.
			if err := extractors.Walk(input, types, func(path, content string) error {
				wg.Add(1)
				go processFile(path, content)
				return nil // continue walking
			}); err != nil {
				return err
			}

			// Wait for all processing to complete, then close channel so range terminates.
			wg.Wait()
			close(errChan)

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
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
