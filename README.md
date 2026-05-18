# AI File Renamer

AI File Renamer is a CLI tool for automatically renaming files based on their content. It was inspired by a real-life need: cleaning up poorly named files, especially files recovered from a broken filesystem.

It supports three backends: the OpenAI API, local Ollama, and a remote Fly.io server.

## Table of Contents

- [Boot.Dev Hackathon](#bootdev-hackathon)
- [Architecture](#architecture)
- [Features](#features)
- [Quick Start](#quick-start)
- [CLI Options](#cli-options)
- [API Reference](#api-reference)
- [Privacy and Security](#privacy-and-security)
- [Deployment](#deployment)
- [DevOps](#devops)
- [Roadmap](#roadmap)
- [Contributing](#contributing)
- [License](#license)
- [Acknowledgments](#acknowledgments)

## Boot.Dev Hackathon

I completed this project for the [Boot.Dev](https://www.boot.dev/) July 2025 Hackathon.

### LLM Usage

This project was developed with AI assistance to accelerate development. While the initial Go code structure was generated using ChatGPT, the codebase has undergone significant evolution through iterative development, debugging, and feature expansion. Having completed all Go courses and projects on Boot.Dev, I approached this as coding efficiently rather than "vibe coding". I can explain and maintain all components of the codebase.

**Note:** AI tools were less effective for DevOps configuration, with roughly 25% of responses being directly usable. Manual expertise was essential for the deployment and infrastructure components.

### Project Reflections

Initially, I envisioned a simple CLI tool to solve a personal problem: renaming poorly named files recovered from a damaged filesystem. What started as a basic file renamer evolved into a metadata-first recovery tool with optional AI fallback, multiple AI backends, robust DevOps configurations, and production-ready deployment options.

**Key learnings from this hackathon:**

One unexpected discovery was that scope creep can actually be beneficial. Adding the server API and multiple deployment options transformed what could have been a simple utility into a versatile platform. The original use case of filesystem recovery naturally led to features like smart filtering and concurrent processing that make the tool more robust overall.

I also learned that investing time in proper infrastructure really pays off. The Docker, Kubernetes, and CI/CD configurations took significant effort but made the project feel more professional and maintainable. Privacy considerations became increasingly important as I worked with the tool. The local Ollama option addresses real concerns about sending sensitive file content to external APIs.

**Challenges and trade-offs:**

The main limitation as the deadline approached was file type support. While the plugin-based extractor system is designed for easy extensibility, I prioritized building a solid foundation over breadth of formats. The current implementation handles common text files, PDFs, JSON, CSV, HTML, XML, MusicXML, Office documents, email, image metadata, and media metadata. OCR for scanned PDFs and image-based documents remains on the roadmap.

Another consideration is that AI calls should be used sparingly. The current client now extracts local metadata and high-signal snippets first, then sends compact evidence to AI only when the local confidence is low. This keeps token use down and avoids sending full file contents by default.

I may have over-engineered the DevOps infrastructure, but Lane's emphasis on making projects "as easy as possible to run" resonated strongly. The comprehensive deployment options, ranging from simple remote server usage to full Kubernetes deployments, demonstrate production-readiness while maintaining simplicity for end users.

**Outcome:**

The final result exceeded my initial expectations, providing a tool that's not only useful for my original problem but could serve a broader community of developers dealing with file organization challenges. The modular architecture and multiple AI backend options create a foundation for future enhancements and community contributions.

## Architecture

This project provides a unified CLI tool with three different AI backend options:

### CLI Client (`cmd/client/`)

A command-line tool that scans directories for poorly named files and suggests better filenames based on local metadata, content evidence, and optional AI fallback. The CLI can operate in three modes:

1. **Direct OpenAI mode:** When `OPENAI_API_KEY` is provided, the CLI communicates directly with the OpenAI API.
2. **Local Ollama mode:** Use the `--local` flag to process files with a local Ollama instance.
3. **Remote server mode:** When no API key is available, the CLI defaults to a remote AI service deployed on Fly.io.

### Server API (`cmd/server/`)

A web server that provides AI filename suggestion services via HTTP API, deployable to cloud platforms like Fly.io. This serves as the backend for the remote server mode.

## Features

- **Multi-format support:** Scans text, Markdown, RTF, CSV, PDF, JSON, Jupyter notebooks, EPUB, OpenDocument files, archives, HTML, XML, MusicXML, config/log files, Office documents, email files, image metadata, and media metadata.
- **Metadata-first naming:** Can rename many recovered files without AI by using internal metadata, headings, document properties, CSV headers, XML fields, and other local evidence.
- **Wrong-extension recovery:** Detects common file types from content, so files such as extensionless PDFs or `.bin` JSON/XML files can still be processed.
- **Flexible AI backends:** Supports direct OpenAI, local Ollama, and a remote Fly.io server.
- **Clean naming:** Generates kebab-case filenames based on file content.
- **Privacy-focused local mode:** Local Ollama mode keeps all file content on your machine.
- **GPU acceleration:** Supports NVIDIA GPUs for faster local Ollama processing.
- **Plugin architecture:** The modular extractor system makes adding new file types straightforward.
- **Easy deployment:** The CLI selects a backend automatically based on configuration.
- **Concurrent processing:** Batch processing supports configurable concurrency.
- **Review workflow:** Dry-run and copy reports can be inspected, exported to Markdown for review, edited, and later applied with `--apply-report`.
- **Safety controls:** Default behavior copies files instead of renaming in place, handles collisions, and can skip low-confidence copies.

## Quick Start

### Prerequisites

- [Go](https://go.dev/) 1.24+
- [Docker](https://docker.com/)
- [jq](https://jqlang.org/), if Docker is not used
- [NVIDIA Container Toolkit](https://docs.nvidia.com/datacenter/cloud-native/container-toolkit/latest/index.html), optional for GPU acceleration with an NVIDIA GPU

### 1. Clone and Set Up

```bash
git clone https://github.com/djblackett/bootdev-hackathon.git
cd bootdev-hackathon
go mod download
mv -n .env.example .env
```

### Recommended Path

For the fastest path, put files in `files/input/` and run the client:

```bash
# Put your files here, or use the provided sample files
cp /path/to/your/files/* files/input/

# Run metadata-only first. This does not call AI.
go run ./cmd/client/main.go \
  --strategy metadata-only \
  --dry-run \
  --min-confidence-to-copy 0.75 \
  --report report.json \
  --review-report review.md

# Inspect review.md, then remove --dry-run or apply the reviewed report.
```

`report.json` is the machine-readable audit trail. `review.md` is the human-readable checklist with confidence bands, evidence sources, reasons, warnings, and notes.

### Recovername v0.1 Preview

The newer scan-only workflow is available as `recovername`. It is designed for recovered files where the original names, extensions, or metadata may be unreliable. This command writes a reviewable JSON plan and optional Markdown review, and does not copy, rename, or delete files.

```bash
go run ./cmd/recovername scan ./testdata/recovered \
  --out rename-plan.json \
  --review-out rename-review.md \
  --no-timestamp
```

Use `cmd/client` when you want to copy files into `files/output/`, apply a reviewed report, or use AI fallback. Use `recovername scan` when you want the safer recovered-files workflow: scan first, inspect the plan, then decide how to apply it.

Optional Tika evidence can be added without making Tika mandatory:

```bash
go run ./cmd/recovername scan ./recovered \
  --out rename-plan.json \
  --tika-url http://127.0.0.1:9998
```

Use `--require-tika` only when a missing Tika server should fail the whole scan. Otherwise Tika failures are recorded as per-file warnings and the plan is still written.

Siegfried can be enabled as an optional format-identification evidence provider:

```bash
go run ./cmd/recovername scan ./recovered \
  --out rename-plan.json \
  --siegfried \
  --siegfried-timeout 10s
```

If `sf` is not installed, the scan still completes and records a warning in the plan.

ExifTool can be enabled for image/media metadata evidence:

```bash
go run ./cmd/recovername scan ./recovered \
  --out rename-plan.json \
  --exiftool \
  --exiftool-timeout 15s
```

ExifTool only runs for files that native detection identifies as image or media-like. If `exiftool` is not installed, the scan still completes and records warnings for those files.

ffprobe can be enabled for audio/video technical metadata:

```bash
go run ./cmd/recovername scan ./recovered \
  --out rename-plan.json \
  --ffprobe \
  --ffprobe-timeout 15s
```

ffprobe only runs for files that native detection identifies as audio/video-like. Tool failures are recorded per file and do not stop the batch.

JHOVE validation can be enabled with:

```bash
go run ./cmd/recovername scan ./recovered \
  --out rename-plan.json \
  --validate \
  --jhove-timeout 30s
```

JHOVE does not generate names. It only adds validation status and warnings to the evidence model.

Tesseract OCR can be enabled for image-like recovered files:

```bash
go run ./cmd/recovername scan ./recovered \
  --out rename-plan.json \
  --ocr \
  --ocr-lang eng \
  --ocr-timeout 60s
```

OCR is disabled by default and is treated as cautious evidence because recognized text can be noisy. Missing or failing Tesseract is recorded per file and does not stop the batch.

### Optional Tool Setup

The app works without external metadata tools, but recovered media, image, and scanned-document batches improve when these are installed:

| Tool | Used For | Enables | Typical package |
|------|----------|---------|-----------------|
| Apache Tika | Broad document text and metadata fallback | `--tika-url` / `TIKA_URL` | Docker sidecar |
| ExifTool | Image and media embedded metadata | image titles, camera timestamps, GPS dates | `exiftool` / `libimage-exiftool-perl` |
| ffprobe | Audio/video technical metadata and tags | media duration, codecs, title/date tags | `ffmpeg` |
| Siegfried | Format identification for recovered files | PRONOM-style format IDs | `siegfried` / `sf` |
| JHOVE | Preservation validation | validation status and warnings | `jhove` |
| Tesseract | OCR for image-like files | cautious text evidence from scans | `tesseract-ocr` |

On Debian/Ubuntu-style systems, the common local tools are usually:

```bash
sudo apt-get install libimage-exiftool-perl ffmpeg tesseract-ocr
```

On macOS with Homebrew:

```bash
brew install exiftool ffmpeg tesseract siegfried
```

Warnings such as `exiftool not available; image EXIF metadata skipped` mean the scan continued safely, but that evidence source was unavailable. Install the relevant tool or omit that optional flag if the warning is expected.

### 2. Choose Your Backend

#### Option A: Remote Server

```bash
# Uses remote Fly.io server automatically
go run ./cmd/client/main.go --input ./files/input --dry-run
```

#### Option B: Local Ollama (Privacy-focused)

```bash
# Start Ollama with Docker Compose
docker compose -f ollama.docker-compose.yaml up ollama -d

# Pull the model (if not already downloaded)
docker exec -it ollama ollama pull mistral

# Run with local backend
go run ./cmd/client/main.go --input ./files/input --local --model mistral --dry-run
```

For GPU acceleration or a more detailed privacy setup, see [Privacy and Security](#privacy-and-security).

#### Option C: Direct OpenAI

```bash
# Make sure OPENAI_API_KEY is set in .env
go run ./cmd/client/main.go --input ./files/input --dry-run
```

#### Optional: Apache Tika Fallback

Apache Tika can run as a Docker sidecar to broaden document parsing without replacing the app's higher-signal local extractors.

```bash
# Start the Tika sidecar
docker compose up tika -d

# Use Tika as fallback extraction for weak or unsupported local parses
TIKA_URL=http://localhost:9998 go run ./cmd/client/main.go --input ./files/input --strategy metadata-only --dry-run
```

When running the client through the default Docker Compose file, `TIKA_URL=http://tika:9998` is already configured.

## CLI Options

| Flag | Description | Default |
|------|-------------|---------|
| `--input` | Directory to scan for files | `files/input` |
| `--output` | Output directory for processed files | `files/output` |
| `--types` | File extensions or detected content types to process (comma-separated) | `txt,text,md,markdown,rtf,csv,pdf,json,ipynb,notebook,epub,odt,ods,odp,opendocument,zip,tar,tgz,archive,html,xml,musicxml,log,cfg,ini,docx,xlsx,pptx,office,eml,email,image,media` |
| `--local` | Use local Ollama instead of OpenAI | `false` |
| `--model` | AI model name | `gpt-3.5-turbo` (OpenAI) / `mistral` (Ollama) |
| `--dry-run` | Preview changes without processing | `false` |
| `--rename` | Rename files in place instead of copying to output | `false` |
| `--debug` | Return all errors joined together | `false` |
| `--quiet` | Suppress progress logs and human-readable summaries | `false` |
| `--json-summary` | Print machine-readable JSON summaries | `false` |
| `--flatten` | Flatten output directory structure | `false` |
| `--strategy` | Rename strategy: `auto`, `metadata-only`, or `ai-only` | `auto` |
| `--confidence-threshold` | Minimum local confidence before `auto` skips AI fallback | `0.75` |
| `--max-ai-chars` | Maximum compact evidence characters sent to AI in `auto` mode | `2000` |
| `--min-confidence-to-copy` | Minimum confidence required before copying files; `0` disables copy skipping | `0` |
| `--report` | Write a JSON report of processed files | none |
| `--apply-report` | Copy files using destinations from a previous JSON report | none |
| `--undo-report` | Delete copied destination files listed in a JSON report | none |
| `--apply-accepted` | Copy planned files and accepted skipped entries from a JSON report | none |
| `--list-pending` | Print pending review entries from a JSON report | none |
| `--set-review-status` | Update review status values in a JSON report | none |
| `--review-entry` | Review update in `source=status` form; repeatable | none |
| `--review-note` | Review note update in `source=note` form; repeatable | none |
| `--explain` | Explain the metadata filename suggestion for one file | none |
| `--include-skipped` | When applying a report, also copy skipped entries marked `review_status=accepted` | `false` |
| `--review-report` | Write a Markdown review file for all report entries | none |
| `--tika-url` | Optional Apache Tika server URL for fallback extraction | `TIKA_URL` |
| `--disable-tika` | Disable Apache Tika fallback even when `TIKA_URL` is set | `false` |

### Examples

```bash
# Mode 1: Direct OpenAI (when OPENAI_API_KEY is set)
./ai-renamer --input ./documents

# Mode 2: Local Ollama (privacy-focused)
./ai-renamer --input ./sensitive-docs --local --model mistral

# Mode 3: Remote Fly.io server (default when no API key)
./ai-renamer --input ./documents

# Preview only, specific file types
./ai-renamer --input ./logs --types "log,txt" --dry-run

# Use local metadata and text evidence only
./ai-renamer --input ./documents --strategy metadata-only --dry-run

# Use metadata first, then compact AI fallback for ambiguous files
./ai-renamer --input ./documents --strategy auto --confidence-threshold 0.8

# Preview changes and write an audit report
./ai-renamer --input ./documents --strategy metadata-only --dry-run --report report.json

# Apply a reviewed dry-run report
./ai-renamer --apply-report report.json

# Apply a report with script-friendly output
./ai-renamer --apply-report report.json --quiet --json-summary

# Copy only confident local matches; weak matches stay in the report as skipped
./ai-renamer --input ./recovered --strategy metadata-only --min-confidence-to-copy 0.75 --report report.json

# Generate a Markdown review file for all entries
./ai-renamer --input ./recovered --strategy metadata-only --min-confidence-to-copy 0.75 --report report.json --review-report review.md

# Print pending review entries from a report
./ai-renamer --list-pending report.json

# Mark a pending entry accepted without hand-editing JSON
./ai-renamer --set-review-status report.json --review-entry files/input/foo.txt=accepted --review-note "files/input/foo.txt=looks right"

# Explain why one file got its suggested name
./ai-renamer --explain files/input/foo.txt

# Apply planned entries plus skipped entries marked accepted
./ai-renamer --apply-accepted report.json

# Undo copied destination files listed in a report
./ai-renamer --undo-report report.json

# Copy to custom output directory with flattened structure
./ai-renamer --input ./files --output ./renamed --flatten

# Debug mode to see all errors
./ai-renamer --input ./documents --debug

# Rename files in place to save space (requires bravery)
./ai-renamer --input ./documents --rename
```

## API Reference

The server provides RESTful endpoints for AI filename suggestions:

### POST `/suggest-filename`

Request filename suggestions based on file content.

**Request Body:**

```json
{
  "content": "File content to analyze...",
  "model": "gpt-4o",
  "evidence_only": false
}
```

Set `evidence_only` to `true` when `content` contains compact metadata and ranked snippets instead of full file contents.

**Response:**

```json
{
  "filename": "suggested-filename.txt",
  "error": ""
}
```

**Supported Models:**

The server accepts the model name supplied by the client or environment configuration. Exact model availability depends on the configured backend.

**Example:**

```bash
curl -X POST https://hackathon-rough-sunset-2856.fly.dev/suggest-filename \
  -H "Content-Type: application/json" \
  -d '{"content": "Meeting notes from quarterly review...", "model": "gpt-4o"}'
```

## Privacy and Security

### Local Mode for Sensitive Files

When processing confidential documents, use local mode to keep all data on your machine:

```bash
# Run ollama and client together
docker compose -f ollama.docker-compose.yaml up

# For NVIDIA GPU acceleration, uncomment the deploy section in ollama.docker-compose.yaml

# For maximum performance, run separately (see note in ollama.docker-compose.yaml)
docker compose -f ollama.docker-compose.yaml up client-local
docker compose -f ollama.docker-compose.yaml up ollama
```

This ensures no file content is sent to external APIs.

For an even stricter local workflow, start with:

```bash
go run ./cmd/client/main.go --input ./files/input --strategy metadata-only --dry-run --report report.json
```

This mode does not call OpenAI, Ollama, or the remote server.

## Confidence and Review

Reports include both numeric confidence and a scan-friendly confidence band:

| Band | Numeric Range | Meaning |
|------|---------------|---------|
| `high` | `>= 0.85` | Strong local metadata or content evidence. |
| `medium` | `>= 0.75` and `< 0.85` | Good enough for the default confident-copy threshold. |
| `review` | `>= 0.40` and `< 0.75` | Plausible but should be checked before copying. |
| `low` | `< 0.40` | Weak, random, missing, or damaged evidence. |

The default examples use `--min-confidence-to-copy 0.75`, which means `high` and `medium` entries are planned while `review` and `low` entries stay pending unless accepted later.

## Deployment

### Server Deployment

```bash
# Deploy to Fly.io
fly deploy
```

## DevOps

This project includes comprehensive DevOps configurations:

- **Docker:** Multi-stage Dockerfiles for both client and server components.
- **GitHub Actions:** Automated Docker Hub publishing on version tags.
- **Kubernetes:** Ready-to-deploy Kubernetes manifests in `k8s-deployment.yaml`.
- **Docker Compose:** Local development with Ollama integration.

## Roadmap

### Current Features

- [x] CLI file renaming
- [x] OpenAI integration
- [x] Local Ollama support
- [x] PDF content extraction
- [x] Metadata-only rename strategy
- [x] Wrong-extension file detection
- [x] JSON, CSV, HTML, XML, MusicXML, Office, email, image metadata, and media metadata extraction
- [x] Dry-run JSON reports and report application
- [x] Markdown review reports for client reports and recovername plans
- [x] Confidence bands and source-specific report reasons
- [x] Collision-safe non-destructive copying
- [x] HTTP API server
- [x] Fly.io deployment
- [x] Kubernetes deployment config

### Planned Enhancements

- [ ] Broader OCR coverage for scanned PDFs
- [ ] Whisper integration for parsing voice recordings
- [ ] Electron desktop app
- [ ] Batch API endpoints
- [ ] File deduplication
- [ ] Automatic language detection

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push to branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License.

## Acknowledgments

- Built for the [Boot.dev](https://boot.dev) Hackathon.
- Powered by [OpenAI](https://openai.com) and [Ollama](https://ollama.com).
- OCR support via [Tesseract](https://github.com/tesseract-ocr/tesseract).
