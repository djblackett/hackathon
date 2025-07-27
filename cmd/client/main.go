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
	// 1. Load environment variables from .env (useful during local dev).
	_ = godotenv.Load() // Ignore errors if .env is not present - for docker

	// 2. Build a typed config object that holds OpenAI/Ollama creds, etc.
	cfg := config.FromEnv()

	// Reasonable default extensions; can be overridden with --types "csv,html,…".
	defaultFileTypes := []string{"txt", "md", "log", "cfg", "ini", "pdf", "json"}

	// 3. Define CLI application.
	app := &cli.App{
		Name:  "ai-file-renamer",
		Usage: "rename recovered docs via AI",
		Flags: []cli.Flag{
			&cli.StringFlag{ // where to start scanning; required.
				Name:     "dir",
				Required: true,
				Usage:    "directory to scan",
			},
			&cli.BoolFlag{ // switch between OpenAI API and local Ollama.
				Name:  "local",
				Usage: "use local Ollama LLM",
			},
			&cli.StringFlag{ // e.g. "mistral", "gpt-4o".
				Name:  "model",
				Value: "mistral",
				Usage: "model name (ollama or openai)",
			},
			&cli.BoolFlag{ // dry‑run means log only.
				Name:  "dry-run",
				Usage: "preview changes only",
			},
			&cli.BoolFlag{ // copy instead of rename; safer default (true).
				Name:  "copy",
				Value: true,
				Usage: "copy files to output directory instead of renaming originals",
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
			&cli.StringSliceFlag{ // allowed extensions.
				Name:  "types",
				Value: cli.NewStringSlice(defaultFileTypes...),
			},
		},
		Action: func(c *cli.Context) error {
			// Harvest flag values.
			dir := c.String("dir")
			local := c.Bool("local")
			model := c.String("model")
			dry := c.Bool("dry-run")
			copyMode := c.Bool("copy")
			debug := c.Bool("debug")
			flatten := c.Bool("flatten")

			// Build a set[string]struct{} for O(1) membership tests during walk.
			types := make(map[string]struct{})
			for _, t := range c.StringSlice("types") {
				types[t] = struct{}{}
			}

			// 4. Spin up LLM client once; reused by all goroutines.
			client, err := ai.NewClient(cfg, local, model)
			if err != nil {
				return err
			}

			// Concurrency helpers.
			var wg sync.WaitGroup
			errChan := make(chan error, 100) // buffered to avoid blocking goroutines.

			// processFile runs in its own goroutine per file.
			processFile := func(path, content string) {
				defer wg.Done()

				// 5a. Ask the model for a descriptive filename.
				suggested, err := client.SuggestFilename(content)
				if err != nil {
					errChan <- err
					return
				}

				// Sanitize to avoid invalid characters.
				sanitized := utils.Sanitize(suggested)

				ext := filepath.Ext(path)

				// 5b. Depending on flags, perform or log the operation.
				switch {
				case dry && copyMode:
					log.Printf("[DRY] %s  →  files/output/%s\n", path, sanitized+ext)
				case dry && !copyMode:
					log.Printf("[DRY] %s  →  %s\n", path, sanitized+ext)
				case copyMode:
					if err := utils.CopyFile(dir, path, "files/output", sanitized, flatten); err != nil {
						errChan <- err
					}
				default: // rename in place
					if err := utils.RenameFile(dir, path, sanitized); err != nil {
						errChan <- err
					}
				}
			}

			// 6. Walk the directory tree and dispatch work.
			if err := extractors.Walk(dir, types, func(path, content string) error {
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
