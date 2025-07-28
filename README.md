# AI File Renamer

AI File Renamer is a CLI tool for automatically renaming files based on their content using AI. Inspired by a real life need, the primary usecase is for poorly named files and especially files recovered from a broken filesystem. It supports three backends: direct OpenAI API, local Ollama, or a remote Fly.io server.

## 📑 Table of Contents

- [🏆 Boot.Dev Hackathon](#bootdev-hackathon)
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

## Boot.Dev Hackathon

I completed this project for the [Boot.Dev](https://www.boot.dev/) July 2025 Hackathon.

### LLM Usage

This project was developed with AI assistance to accelerate development. While the initial Go code structure was generated using ChatGPT, the codebase has undergone significant evolution through iterative development, debugging, and feature expansion. Having completed all Go courses and projects on Boot.Dev, I approached this as coding efficiently rather than "vibe coding". I can explain and maintain all components of the codebase.

**Note**: AI tools proved less effective for DevOps configurations, with approximately 25% of responses being directly usable. Manual expertise was essential for the deployment and infrastructure components.

### Project Reflections

Initially, I envisioned a simple CLI tool to solve a personal problem: renaming poorly named files recovered from a damaged filesystem. What started as a basic file renamer evolved into a comprehensive solution with multiple AI backends, robust DevOps configurations, and production-ready deployment options.

**Key learnings from this hackathon:**

One unexpected discovery was that scope creep can actually be beneficial. Adding the server API and multiple deployment options transformed what could have been a simple utility into a versatile platform. The original use case of filesystem recovery naturally led to features like smart filtering and concurrent processing that make the tool more robust overall.

I also learned that investing time in proper infrastructure really pays off. The Docker, Kubernetes, and CI/CD configurations took significant effort but made the project feel more professional and maintainable. Privacy considerations became increasingly important as I worked with the tool. The local Ollama option addresses real concerns about sending sensitive file content to external APIs.

**Challenges and trade-offs:**

The main limitation as the deadline approached was file type support. While the plugin-based extractor system is designed for easy extensibility, I prioritized building a solid foundation over breadth of formats. The current implementation handles common text-based files well, but expanding to office documents and OCR for image-based PDFs (via Tesseract and Pandoc) remains on the roadmap.

I may have over-engineered the DevOps infrastructure, but Lane's emphasis on making projects "as easy as possible to run" resonated strongly. The comprehensive deployment options, ranging from simple remote server usage to full Kubernetes deployments, demonstrate production-readiness while maintaining simplicity for end users.

**Outcome:**

The final result exceeded my initial expectations, providing a tool that's not only useful for my original problem but could serve a broader community of developers dealing with file organization challenges. The modular architecture and multiple AI backend options create a foundation for future enhancements and community contributions.

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
