# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

AI File Renamer: a Go CLI + HTTP server that renames poorly-named files based on their content. Built for the Boot.dev July 2025 hackathon. Go 1.24.

## Commands

```bash
# Run the CLI (defaults: files/input -> files/output, remote Fly.io backend)
go run ./cmd/client/main.go

# Run the relay server (cmd/server) locally
go run ./cmd/server/main.go   # listens on $PORT or :8080

# Build binaries
go build -o filename-fixer ./cmd/client/main.go
go build -o relay-server ./cmd/server/main.go

# Dependency hygiene
go mod download
go mod tidy

# Docker builds (multi-stage)
docker build -f Dockerfile -t ai-file-renamer .         # client
docker build -f server.Dockerfile -t relay-server .     # server

# Local stack with Ollama
docker compose -f ollama.docker-compose.yaml up ollama -d
docker exec -it ollama ollama pull mistral
go run ./cmd/client/main.go --local --model mistral --dry-run

# Deploy server to Fly.io
fly deploy
```

There is **no test suite** in this repo — `go test ./...` finds no tests. If adding tests, follow standard Go conventions (`*_test.go` alongside source).

## Architecture

Three runtime modes, all selected at startup in [internal/ai/ai.go](internal/ai/ai.go) by `NewClient(cfg, local, model)`:

1. `--local` flag → `OllamaClient` (talks to `OLLAMA_HOST`, default `http://localhost:11434/api/generate`)
2. `OPENAI_API_KEY` set → `OpenAIClient` (direct OpenAI calls)
3. Otherwise → `HTTPClient` pointing at `AI_SERVER_URL` (the relay server in `cmd/server/`)

All three implement the `ai.Client` interface (single method `SuggestFilename(content string) (string, error)`). The relay server (`cmd/server/main.go`) is itself a thin wrapper that re-invokes `ai.NewClient` server-side with the user's chosen OpenAI model — API keys never leave the server. The server validates models against an allowlist (`validOpenAIModels`).

### Extractor plugin system

[internal/extractors/extractors.go](internal/extractors/extractors.go) defines an `Extractor` interface (`CanHandle(path) bool`, `Extract(path) (string, error)`) and a package-level `registered` slice populated via `Register(...)` in each extractor file's `init()`. `Walk(dir, types, fn)` traverses the input tree, filters by extension allowlist (the `--types` flag), and dispatches to the first registered extractor that claims the file. **To add a new file type:** create a file in `internal/extractors/`, implement the interface, and call `Register` in `init()`. Also extend the `defaultFileTypes` slice in [cmd/client/main.go](cmd/client/main.go) if it should be on by default.

### Concurrency model

[cmd/client/main.go](cmd/client/main.go) spawns one unbounded goroutine per file inside the `Walk` callback, coordinated by a `sync.WaitGroup` and a buffered error channel (capacity 100). There is currently no semaphore — if processing very large trees, this can overwhelm the AI backend. Errors are collected and either the first error is returned, or with `--debug` all errors are joined via `errors.Join`.

### Flag flow

CLI flags (`urfave/cli/v2`) → `Action` closure → passed to `ai.NewClient`, `extractors.Walk`, and `utils.CopyFile`/`utils.RenameFile`. Default model is chosen based on `--local` (`mistral`) vs. remote (`gpt-3.5-turbo`).

## Config

`internal/config/env.go` reads three env vars: `OPENAI_API_KEY`, `OLLAMA_HOST`, `AI_SERVER_URL`. `.env` is auto-loaded by `godotenv` if present (errors ignored — important for Docker where env is injected directly).

## DevOps

- `Dockerfile` builds the **client** (multi-stage alpine, CGO disabled).
- `server.Dockerfile` builds the **server** for Fly.io.
- `docker-compose.yaml` runs the client; `ollama.docker-compose.yaml` runs Ollama (optionally with NVIDIA GPU — uncomment the `deploy` section).
- `k8s-deployment.yaml` is a ready-to-apply manifest for the relay server.
- GitHub Actions in `.github/workflows/` publishes images to Docker Hub on version tags.
