# AI File Renamer

AI File Renamer is a CLI tool that scans directories for unhelpfully named documents (like `file001.txt`, `untitled.pdf`) and uses an AI model to rename them based on their content.

---

## ✨ Features

- 📁 Scans `.txt`, `.md`, and `.pdf` files
- 🧠 Uses AI to infer a relevant filename from the file's contents
- 🗂️ Renames files using clean, kebab-case naming conventions
- 🛡️ Privacy-focused: optionally use a local LLM via Docker (Ollama)
- 🔒 Fallback to OCR (Tesseract) for scanned PDFs
- ⚙️ CLI flags for preview mode, filtering by type, and AI backend selection

---

## 🚀 Quick Start

### 1. Clone the repository

```bash
git clone https://github.com/your-username/ai-file-renamer.git
cd ai-file-renamer
```

### 2. Start the local LLM (optional)

```bash
docker compose up -d
docker exec -it ollama ollama pull mistral
```

### 3. Run the CLI (once built)

```bash
go run ./cmd/main.go --dir ./docs --local --model mistral --dry-run
```

---

## 🔧 CLI Options

| Flag         | Description                                     |
|--------------|-------------------------------------------------|
| `--dir`      | Directory to scan for files                     |
| `--types`    | Comma-separated list of file types to include   |
| `--local`    | Use local LLM via Ollama                        |
| `--model`    | Model to use for local LLM (`mistral`, etc.)    |
| `--dry-run`  | Show suggested filenames without renaming files |

---

## 📦 Dependencies

- [Go](https://go.dev/) 1.20+
- [`pdftotext`](https://poppler.freedesktop.org/) (Poppler utils)
- [`tesseract`](https://github.com/tesseract-ocr/tesseract) OCR engine
- [Ollama](https://ollama.com) (optional for local AI)

---

## 🔐 Local Mode for Private Files

If your documents contain sensitive or private data, enable local mode:

```bash
--local --model mistral
```

This avoids sending any file content to external APIs.

---

## 📂 Folder Structure

- `cmd/` – CLI entry point
- `ai/` – LLM clients and prompt logic
- `extractors/` – File type handlers (PDF, text, OCR)
- `utils/` – Filename sanitizers, logging

---

## 🛠️ Roadmap (Post-MVP Ideas)

- Embedding-based deduplication
- Automatic language detection
- Frontend (Web/Electron) wrapper
- Drag-and-drop UI

---

## 📄 License

MIT (or choose your preferred license)

---

Built for the Boot.dev Hackathon ✨
