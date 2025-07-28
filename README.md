# AI File Renamer

AI File Renamer is a CLI tool for automatically renaming files based on their content using AI. It supports three backends: direct OpenAI API, local Ollama, or a remote Fly.io server.

## 📑 Table of Contents

- [🏗️ Architecture](#️-architecture)
- [✨ Features](#-features)
- [🚀 Quick Start](#-quick-start)
- [🔧 CLI Options](#-cli-options)
- [🌐 API Reference](#-api-reference)
- [🔐 Privacy & Security](#-privacy--security)
- [🚀 Deployment](#-deployment)
- [⚙️ DevOps](#️-devops)
- [🛠️ Roadmap](#️-roadmap)
- [🤝 Contributing](#-contributing)
- [📄 License](#-license)
- [🙏 Acknowledgments](#-acknowledgments)

---

## 🏗️ Architecture

This project provides a unified CLI tool with three different AI backend options:

### 📱 CLI Client (`cmd/client/`)

A command-line tool that scans directories for poorly named files and uses AI to suggest better filenames based on content. The CLI can operate in three modes:

1. **🔗 Direct OpenAI Mode**: When `OPENAI_API_KEY` is provided, communicates directly with OpenAI APIs
2. **🏠 Local Ollama Mode**: Use `--local` flag to process files with a local Ollama instance for privacy
3. **☁️ Remote Server Mode**: Defaults to using a remote AI service deployed on Fly.io when no API key is available

### 🌐 Server API (`cmd/server/`)

A web server that provides AI filename suggestion services via HTTP API, deployable to cloud platforms like Fly.io. This serves as the backend for the remote server mode.

---

## ✨ Features

- 📁 **Multi-format support**: Scans `.txt`, `.md`, `.pdf`, `.json`, `.log`, `.cfg`, `.ini` files (more to come)
- 🧠 **Flexible AI backends**: Three modes - Direct OpenAI, Local Ollama, or Remote Fly.io server
- 🗂️ **Clean naming**: Generates kebab-case filenames following best practices
- 🛡️ **Privacy-focused**: Local Ollama mode keeps all data on your machine
- 🚀 **GPU acceleration**: NVIDIA GPU support for faster local Ollama processing
- 🔌 **Plugin architecture**: Modular extractor system makes adding new file types straightforward
- ⚙️ **Easy deployment**: Simple CLI with automatic backend selection
- 🔄 **Concurrent processing**: Efficient batch processing with configurable concurrency
- 🎯 **Smart filtering**: Skip already well-named files automatically

---

## 🚀 Quick Start

### Prerequisites

- [Go](https://go.dev/) 1.24+
- [Docker](https://docker.com/)
- [jq](https://jqlang.org/) - if docker is not used
- [NVIDIA Container Toolkit](https://docs.nvidia.com/datacenter/cloud-native/container-toolkit/latest/index.html) - optional GPU acceleration (requires an NVIDIA GPU) - [Instructions](https://itsfoss.com/ollama-docker/)

### 1. Clone and Setup

```bash
git clone https://github.com/djblackett/bootdev-hackathon.git
cd bootdev-hackathon
go mod download
mv -n .env.example .env
```

### Quick Start (Recommended)

For the fastest experience, simply put your files in the `files/input/` folder and run:

```bash
# Put your files here (or use provided sample files)
cp /path/to/your/files/* files/input/

# Run with default settings (uses remote server)
go run ./cmd/client/main.go

# Renamed files will appear in files/output/
```

### 2. Choose Your Backend

#### Option A: Remote Server (No setup required)

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

> **💡 Tip**: For GPU acceleration or detailed privacy setup, see the [Privacy & Security](#-privacy--security) section.

#### Option C: Direct OpenAI

```bash
# Make sure OPENAI_API_KEY is set in .env
go run ./cmd/client/main.go --input ./files/input --dry-run
```

---

## 🔧 CLI Options

| Flag | Description | Default |
|------|-------------|---------|
| `--input` | Directory to scan for files | `files/input` |
| `--output` | Output directory for processed files | `files/output` |
| `--types` | File extensions to process (comma-separated) | `txt,md,log,cfg,ini,pdf,json` |
| `--local` | Use local Ollama instead of OpenAI | `false` |
| `--model` | AI model name | `gpt-3.5-turbo` (OpenAI) / `mistral` (Ollama) |
| `--dry-run` | Preview changes without processing | `false` |
| `--rename` | Rename files in place instead of copying to output | `false` |
| `--debug` | Return all errors joined together | `false` |
| `--flatten` | Flatten output directory structure | `false` |

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

# Copy to custom output directory with flattened structure
./ai-renamer --input ./files --output ./renamed --flatten

# Debug mode to see all errors
./ai-renamer --input ./documents --debug

# Rename files in place to save space (requires bravery)
./ai-renamer --input ./documents --rename
```

---

## 🌐 API Reference

The server provides RESTful endpoints for AI filename suggestions:

### POST `/suggest-filename`

Request filename suggestions based on file content.

**Request Body:**

```json
{
  "content": "File content to analyze...",
  "model": "gpt-4o"
}
```

**Response:**

```json
{
  "filename": "suggested-filename.txt",
  "error": ""
}
```

**Supported Models:**

- `gpt-3.5-turbo`
- `gpt-4`
- `gpt-4o`
- `gpt-4-1106-preview`

**Example:**

```bash
curl -X POST https://hackathon-rough-sunset-2856.fly.dev/suggest-filename \
  -H "Content-Type: application/json" \
  -d '{"content": "Meeting notes from quarterly review...", "model": "gpt-4o"}'
```

---

## 🔐 Privacy & Security

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

---

## 🚀 Deployment

### Server Deployment

```bash
# Deploy to Fly.io
fly deploy
```

---

## ⚙️ DevOps

This project includes comprehensive DevOps configurations:

- **🐳 Docker**: Multi-stage Dockerfiles for both client and server components
- **🔄 GitHub Actions**: Automated Docker Hub publishing on version tags
- **☸️ Kubernetes**: Ready-to-deploy K8s manifests in `k8s-deployment.yaml`
- **🐙 Docker Compose**: Local development with Ollama integration

---

## 🛠️ Roadmap

### Current Features ✅

- [x] CLI file renaming
- [x] OpenAI integration
- [x] Local Ollama support
- [x] PDF content extraction
- [x] HTTP API server
- [x] Fly.io deployment
- [x] Kubernetes deployment config

### Planned Enhancements 🔄

- [ ] OCR for scanned documents
- [ ] Whisper integration for parsing voice recordings
- [ ] Electron desktop app
- [ ] Batch API endpoints
- [ ] File deduplication
- [ ] Automatic language detection

---

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push to branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

---

## 📄 License

This project is licensed under the MIT License

---

## 🙏 Acknowledgments

- Built for the [Boot.dev](https://boot.dev) Hackathon ✨
- Powered by [OpenAI](https://openai.com) and [Ollama](https://ollama.com)
- OCR support via [Tesseract](https://github.com/tesseract-ocr/tesseract)
