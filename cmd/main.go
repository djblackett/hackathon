package main

import (
	"log"
	"os"
	"sync"

	"github.com/joho/godotenv"
	"github.com/urfave/cli/v2"

	"github.com/djblackett/bootdev-hackathon/ai"
	"github.com/djblackett/bootdev-hackathon/config"
	"github.com/djblackett/bootdev-hackathon/extractors"
	"github.com/djblackett/bootdev-hackathon/utils"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	cfg := config.FromEnv()
	var defaultFileTypes = []string{"txt", "md"}
	app := &cli.App{
		Name:  "ai-file-renamer",
		Usage: "rename recovered docs via AI",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "dir", Required: true, Usage: "directory to scan"},
			&cli.BoolFlag{Name: "local", Usage: "use local Ollama LLM"},
			&cli.StringFlag{Name: "model", Value: "mistral", Usage: "model name (ollama or openai)"},
			&cli.BoolFlag{Name: "dry-run", Usage: "preview changes only"},
			&cli.BoolFlag{Name: "copy", Value: true, Usage: "copy files to output directory instead of renaming originals"},
			&cli.StringSliceFlag{Name: "types", Value: cli.NewStringSlice(defaultFileTypes...)},
		},
		Action: func(c *cli.Context) error {
			dir := c.String("dir")
			local := c.Bool("local")
			model := c.String("model")
			dry := c.Bool("dry-run")
			copy := c.Bool("copy")
			types := map[string]struct{}{}
			for _, t := range c.StringSlice("types") {
				types[t] = struct{}{}
			}

			client, err := ai.NewClient(cfg, local, model)
			if err != nil {
				return err
			}

			var wg sync.WaitGroup
			var walkErr error
			var mu sync.Mutex

			errChan := make(chan error, 100)

			processFile := func(path string, content string) {
				defer wg.Done()
				suggested, err := client.SuggestFilename(content)
				if err != nil {
					errChan <- err
					return
				}
				sanitizedFilename := utils.Sanitize(suggested)

				if dry {
					if copy {
						// Use a fixed output directory relative to the project root
						outputDir := "files/output"
						log.Printf("[DRY] %s  →  %s/%s\n", path, outputDir, sanitizedFilename)
					} else {
						log.Printf("[DRY] %s  →  %s\n", path, sanitizedFilename)
					}
					return
				}

				if copy {
					// Copy file to output directory
					outputDir := "files/output"
					if err := utils.CopyFile(path, outputDir, sanitizedFilename); err != nil {
						errChan <- err
					}
				} else {
					// Rename file in place
					if err := utils.RenameFile(path, sanitizedFilename); err != nil {
						errChan <- err
					}
				}
			}

			err = extractors.Walk(dir, types, func(path string, content string) error {
				wg.Add(1)
				go processFile(path, content)
				return nil
			})
			if err != nil {
				return err
			}

			wg.Wait()
			close(errChan)
			for e := range errChan {
				mu.Lock()
				if walkErr == nil {
					walkErr = e
				}
				mu.Unlock()
			}
			return walkErr
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
