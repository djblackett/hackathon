package main

import (
	"log"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/djblackett/bootdev-hackathon/ai"
	"github.com/djblackett/bootdev-hackathon/extractors"
	"github.com/djblackett/bootdev-hackathon/utils"
)

func main() {
	app := &cli.App{
		Name:  "ai-file-renamer",
		Usage: "rename recovered docs via AI",

		Flags: []cli.Flag{
			&cli.StringFlag{Name: "dir", Required: true, Usage: "directory to scan"},
			&cli.BoolFlag{Name: "local", Usage: "use local Ollama LLM"},
			&cli.StringFlag{Name: "model", Value: "mistral", Usage: "model name (ollama or openai)"},
			&cli.BoolFlag{Name: "dry-run", Usage: "preview changes only"},
			&cli.StringSliceFlag{Name: "types", Value: cli.NewStringSlice("txt", "pdf", "md")},
		},

		Action: func(c *cli.Context) error {
			dir := c.String("dir")
			local := c.Bool("local")
			model := c.String("model")
			dry := c.Bool("dry-run")
			types := map[string]struct{}{}
			for _, t := range c.StringSlice("types") {
				types[t] = struct{}{}
			}

			client, err := ai.NewClient(local, model)
			if err != nil {
				return err
			}

			return extractors.Walk(dir, types, func(path string, content string) error {
				suggested, err := client.SuggestFilename(content)
				if err != nil {
					return err
				}
				clean := utils.Sanitize(suggested)

				if dry {
					log.Printf("[DRY] %s  →  %s\n", path, clean)
					return nil
				}
				return utils.RenameFile(path, clean)
			})
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
