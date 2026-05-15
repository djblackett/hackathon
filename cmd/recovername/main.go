package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/djblackett/bootdev-hackathon/internal/app"
)

func main() {
	if err := runApp(os.Args); err != nil {
		log.Fatal(err)
	}
}

func runApp(args []string) error {
	cliApp := &cli.App{
		Name:  "recovername",
		Usage: "create safe rename plans for recovered files",
		Commands: []*cli.Command{
			{
				Name:      "scan",
				Usage:     "scan a directory and write rename-plan.json",
				ArgsUsage: "<directory>",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "out",
						Value: "rename-plan.json",
						Usage: "path to write the JSON rename plan",
					},
					&cli.StringFlag{
						Name:  "tika-url",
						Usage: "optional Apache Tika server URL",
					},
					&cli.BoolFlag{
						Name:  "no-tika",
						Usage: "disable Tika even when a URL is provided",
					},
					&cli.BoolFlag{
						Name:  "require-tika",
						Usage: "fail the scan when the configured Tika server is unavailable",
					},
					&cli.DurationFlag{
						Name:  "tika-timeout",
						Value: 30 * time.Second,
						Usage: "timeout for Tika health and extraction requests",
					},
					&cli.BoolFlag{
						Name:  "exiftool",
						Usage: "run ExifTool metadata extraction for image/media-like files when available",
					},
					&cli.DurationFlag{
						Name:  "exiftool-timeout",
						Value: 15 * time.Second,
						Usage: "timeout for ExifTool metadata extraction",
					},
					&cli.BoolFlag{
						Name:  "ffprobe",
						Usage: "run ffprobe metadata extraction for audio/video-like files when available",
					},
					&cli.DurationFlag{
						Name:  "ffprobe-timeout",
						Value: 15 * time.Second,
						Usage: "timeout for ffprobe metadata extraction",
					},
					&cli.BoolFlag{
						Name:  "siegfried",
						Usage: "run Siegfried format identification when sf is available",
					},
					&cli.DurationFlag{
						Name:  "siegfried-timeout",
						Value: 10 * time.Second,
						Usage: "timeout for Siegfried format identification",
					},
					&cli.BoolFlag{
						Name:  "hash",
						Value: true,
						Usage: "compute SHA-256 hashes",
					},
					&cli.IntFlag{
						Name:  "max-text-preview",
						Value: 2000,
						Usage: "maximum extracted text preview bytes stored in the plan",
					},
					&cli.BoolFlag{
						Name:  "no-timestamp",
						Usage: "omit generatedAt for byte-identical deterministic output",
					},
				},
				Action: func(c *cli.Context) error {
					root := c.Args().First()
					if root == "" {
						return fmt.Errorf("scan requires a directory")
					}
					cfg := app.ScanConfig{
						Root:             root,
						OutPath:          c.String("out"),
						TikaURL:          c.String("tika-url"),
						NoTika:           c.Bool("no-tika"),
						RequireTika:      c.Bool("require-tika"),
						TikaTimeout:      c.Duration("tika-timeout"),
						UseExifTool:      c.Bool("exiftool"),
						ExifToolTimeout:  c.Duration("exiftool-timeout"),
						UseFFProbe:       c.Bool("ffprobe"),
						FFProbeTimeout:   c.Duration("ffprobe-timeout"),
						UseSiegfried:     c.Bool("siegfried"),
						SiegfriedTimeout: c.Duration("siegfried-timeout"),
						Hash:             c.Bool("hash"),
						MaxTextPreview:   c.Int("max-text-preview"),
						NoTimestamp:      c.Bool("no-timestamp"),
					}
					if err := applyTrailingScanFlags(c.Args().Slice()[1:], &cfg); err != nil {
						return err
					}
					_, err := app.Scan(context.Background(), cfg)
					return err
				},
			},
		},
	}
	return cliApp.Run(args)
}

func applyTrailingScanFlags(args []string, cfg *app.ScanConfig) error {
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--out":
			i++
			if i >= len(args) {
				return fmt.Errorf("--out requires a value")
			}
			cfg.OutPath = args[i]
		case strings.HasPrefix(arg, "--out="):
			cfg.OutPath = strings.TrimPrefix(arg, "--out=")
		case arg == "--tika-url":
			i++
			if i >= len(args) {
				return fmt.Errorf("--tika-url requires a value")
			}
			cfg.TikaURL = args[i]
		case strings.HasPrefix(arg, "--tika-url="):
			cfg.TikaURL = strings.TrimPrefix(arg, "--tika-url=")
		case arg == "--tika-timeout":
			i++
			if i >= len(args) {
				return fmt.Errorf("--tika-timeout requires a value")
			}
			duration, err := time.ParseDuration(args[i])
			if err != nil {
				return err
			}
			cfg.TikaTimeout = duration
		case strings.HasPrefix(arg, "--tika-timeout="):
			duration, err := time.ParseDuration(strings.TrimPrefix(arg, "--tika-timeout="))
			if err != nil {
				return err
			}
			cfg.TikaTimeout = duration
		case arg == "--exiftool-timeout":
			i++
			if i >= len(args) {
				return fmt.Errorf("--exiftool-timeout requires a value")
			}
			duration, err := time.ParseDuration(args[i])
			if err != nil {
				return err
			}
			cfg.ExifToolTimeout = duration
		case strings.HasPrefix(arg, "--exiftool-timeout="):
			duration, err := time.ParseDuration(strings.TrimPrefix(arg, "--exiftool-timeout="))
			if err != nil {
				return err
			}
			cfg.ExifToolTimeout = duration
		case arg == "--ffprobe-timeout":
			i++
			if i >= len(args) {
				return fmt.Errorf("--ffprobe-timeout requires a value")
			}
			duration, err := time.ParseDuration(args[i])
			if err != nil {
				return err
			}
			cfg.FFProbeTimeout = duration
		case strings.HasPrefix(arg, "--ffprobe-timeout="):
			duration, err := time.ParseDuration(strings.TrimPrefix(arg, "--ffprobe-timeout="))
			if err != nil {
				return err
			}
			cfg.FFProbeTimeout = duration
		case arg == "--siegfried-timeout":
			i++
			if i >= len(args) {
				return fmt.Errorf("--siegfried-timeout requires a value")
			}
			duration, err := time.ParseDuration(args[i])
			if err != nil {
				return err
			}
			cfg.SiegfriedTimeout = duration
		case strings.HasPrefix(arg, "--siegfried-timeout="):
			duration, err := time.ParseDuration(strings.TrimPrefix(arg, "--siegfried-timeout="))
			if err != nil {
				return err
			}
			cfg.SiegfriedTimeout = duration
		case arg == "--max-text-preview":
			i++
			if i >= len(args) {
				return fmt.Errorf("--max-text-preview requires a value")
			}
			value, err := strconv.Atoi(args[i])
			if err != nil {
				return err
			}
			cfg.MaxTextPreview = value
		case strings.HasPrefix(arg, "--max-text-preview="):
			value, err := strconv.Atoi(strings.TrimPrefix(arg, "--max-text-preview="))
			if err != nil {
				return err
			}
			cfg.MaxTextPreview = value
		case arg == "--no-tika":
			cfg.NoTika = true
		case arg == "--require-tika":
			cfg.RequireTika = true
		case arg == "--exiftool":
			cfg.UseExifTool = true
		case arg == "--ffprobe":
			cfg.UseFFProbe = true
		case arg == "--siegfried":
			cfg.UseSiegfried = true
		case arg == "--no-timestamp":
			cfg.NoTimestamp = true
		case arg == "--hash":
			cfg.Hash = true
		case strings.HasPrefix(arg, "--hash="):
			value, err := strconv.ParseBool(strings.TrimPrefix(arg, "--hash="))
			if err != nil {
				return err
			}
			cfg.Hash = value
		case strings.HasPrefix(arg, "-"):
			return fmt.Errorf("unknown scan flag %s", arg)
		}
	}
	return nil
}
