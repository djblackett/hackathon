# AI File Renamer

AI File Renamer is a comprehensive solution for automatically renaming files based on their content using AI. It consists of both a CLI tool for local file processing and a web server API for remote AI-powered filename suggestions.

---

## 🏗️ Architecture

This project provides two main components:

### 📱 CLI Client (`cmd/client/`)

A command-line tool that scans directories for poorly named files and uses AI to suggest better filenames based on content.

### 🌐 Server API (`cmd/server/`)

A web server that provides AI filename suggestion services via HTTP API, deployable to cloud platforms like Fly.io.

---

## ✨ Features

- 📁 **Multi-format support**: Scans `.txt`, `.md`, `.pdf`, `.json`, `.log`, `.cfg`, `.ini` files
- 🧠 **AI-powered**: Uses OpenAI GPT models or local Ollama for intelligent filename suggestions
- 🗂️ **Clean naming**: Generates kebab-case filenames following best practices
- 🛡️ **Privacy-focused**: Optional local LLM via Docker (Ollama) for sensitive documents
- 🔒 **OCR fallback**: Handles scanned PDFs using Tesseract OCR
- ⚙️ **Flexible deployment**: CLI for local use, server API for remote access
- 🔄 **Concurrent processing**: Efficient batch processing with configurable concurrency
- 🎯 **Smart filtering**: Skip already well-named files automatically

---

## 🚀 Quick Start

### Prerequisites

- [Go](https://go.dev/) 1.24+
- [`pdftotext`](https://poppler.freedesktop.org/) (Poppler utils) for PDF processing
- [`tesseract`](https://github.com/tesseract-ocr/tesseract) OCR engine for scanned PDFs
- [Docker](https://docker.com/) (optional, for local Ollama)

### 1. Clone and Setup

```bash
git clone https://github.com/djblackett/bootdev-hackathon.git
cd bootdev-hackathon
go mod download
```

### 2. Environment Configuration

Create a `.env` file with your API keys:

```bash
# For OpenAI (required for remote AI)
OPENAI_API_KEY=your_openai_api_key_here

# For local Ollama (optional)
OLLAMA_BASE_URL=http://localhost:11434
```

### 3. Start Local LLM (Optional)

For privacy-focused processing with local AI:

#### Using the integrated Ollama setup

```bash
# Start Ollama server
docker compose -f ollama.docker-compose.yaml up ollama -d

# Pull the mistral model (first time only)
docker exec -it ollama ollama pull mistral

# Now run the local client
docker compose -f ollama.docker-compose.yaml up client-local

# Or with debug mode
docker compose -f ollama.docker-compose.yaml up client-local-dev
```

#### Or using separate containers

```bash
docker compose up -d
docker exec -it ollama ollama pull mistral
```

### 4. Run the CLI

```bash
# Preview mode with local AI
go run ./cmd/client/main.go --dir ./files/input --local --model mistral --dry-run

# Actual renaming with OpenAI
go run ./cmd/client/main.go --dir ./files/input --model gpt-4o

# Build and run
go build -o ai-renamer ./cmd/client/
./ai-renamer --dir ./documents --dry-run
```

### 5. Deploy Server (Optional)

Deploy the API server to Fly.io:

```bash
fly deploy
```

Or run locally:

```bash
go run ./cmd/server/main.go
# Server starts on http://localhost:8080
```

---

## 🔧 CLI Options

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--dir` | `-d` | Directory to scan for files | *required* |
| `--types` | `-t` | Comma-separated file extensions | `txt,md,log,cfg,ini,pdf,json` |
| `--local` | `-l` | Use local Ollama instead of OpenAI | `false` |
| `--model` | `-m` | AI model name | `gpt-4o` (OpenAI) / `mistral` (Ollama) |
| `--dry-run` | `-n` | Preview changes without renaming | `false` |
| `--workers` | `-w` | Number of concurrent workers | `3` |
| `--server-url` | `-s` | Remote server URL for AI processing | - |

### Examples

```bash
# Basic usage with OpenAI
./ai-renamer --dir ./documents

# Privacy mode with local AI
./ai-renamer --dir ./sensitive-docs --local --model mistral

# Preview only, specific file types
./ai-renamer --dir ./logs --types "log,txt" --dry-run

# Use remote server
./ai-renamer --dir ./files --server-url https://your-app.fly.dev
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
curl -X POST https://your-app.fly.dev/suggest-filename \
  -H "Content-Type: application/json" \
  -d '{"content": "Meeting notes from quarterly review...", "model": "gpt-4o"}'
```

---

## 🔐 Privacy & Security

### Local Mode for Sensitive Files

When processing confidential documents, use local mode to keep all data on your machine:

```bash
./ai-renamer --dir ./confidential --local --model mistral
```

This ensures no file content is sent to external APIs.

### Supported AI Backends

- **OpenAI GPT** (default): Best accuracy, requires API key and internet
- **Local Ollama**: Privacy-focused, runs entirely offline after setup

---

## 📂 Project Structure

```text
├── cmd/
│   ├── client/          # CLI application
│   └── server/          # HTTP API server
├── internal/
│   ├── ai/              # AI clients (OpenAI, Ollama, HTTP)
│   ├── config/          # Environment configuration
│   ├── extractors/      # File content extraction (PDF, text, etc.)
│   └── utils/           # Utilities (logging, sanitization)
├── files/
│   ├── input/           # Sample input files
│   └── output/          # Renamed output files
├── docker-compose.yaml  # Ollama local AI setup
├── server.Dockerfile    # Server container
├── Dockerfile           # Client container
└── fly.toml            # Fly.io deployment config
```

---

## 🚀 Deployment

### Fly.io (Recommended)

1. Install [Fly CLI](https://fly.io/docs/flyctl/)
2. Deploy the server:

```bash
fly deploy
```

The server will be available at `https://your-app.fly.dev`

### Docker

```bash
# Build server image
docker build -f server.Dockerfile -t ai-renamer-server .

# Run server
docker run -p 8080:8080 -e OPENAI_API_KEY=your_key ai-renamer-server

# Build client image
docker build -f Dockerfile -t ai-renamer-client .
```

### Local Development

```bash
# Run server
go run ./cmd/server/main.go

# Run client
go run ./cmd/client/main.go --dir ./files/input --dry-run
```

---

## 🧪 Testing

```bash
# Run tests
go test ./...

# Test with sample files
go run ./cmd/client/main.go --dir ./files/input --dry-run

# Test server API
curl -X POST http://localhost:8080/suggest-filename \
  -H "Content-Type: application/json" \
  -d '{"content": "Sample content", "model": "gpt-4o"}'
```

---

## 🛠️ Roadmap

### Current Features ✅

- [x] CLI file renaming
- [x] OpenAI integration
- [x] Local Ollama support
- [x] PDF content extraction
- [x] OCR for scanned documents
- [x] HTTP API server
- [x] Fly.io deployment

### Planned Enhancements 🔄

- [ ] Web UI frontend
- [ ] Batch API endpoints
- [ ] File deduplication
- [ ] Automatic language detection
- [ ] Custom prompt templates
- [ ] Integration webhooks
- [ ] Electron desktop app

---

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push to branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

---

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## 🙏 Acknowledgments

- Built for the [Boot.dev](https://boot.dev) Hackathon ✨
- Powered by [OpenAI](https://openai.com) and [Ollama](https://ollama.com)
- OCR support via [Tesseract](https://github.com/tesseract-ocr/tesseract)

---

*Need help? Open an issue or check the [documentation](https://github.com/djblackett/bootdev-hackathon/wiki).*
