# Metadata-First Rename Milestone Plan

This plan extends AI File Renamer so it can identify recovered files using internal metadata and high-signal text before falling back to AI. The goal is to keep the default workflow non-destructive, reduce token usage, and avoid sending full file contents to an AI model unless it is genuinely needed.

## Milestone 1: Metadata-First Rename Pipeline

Goal: make the app useful without AI for many files, while preserving the current AI behavior as a fallback.

Deliverables:

- Add structured extraction output:
  - file path
  - detected extension/type
  - metadata fields
  - ranked text samples
  - warnings/errors
- Add local filename generation from extracted evidence.
- Add CLI strategy modes:
  - `--strategy ai-only`
  - `--strategy metadata-only`
  - `--strategy auto`
- In `auto`, use local metadata first and call AI only when confidence is low.
- Keep copying files to output as the default behavior.

Suggested files:

- `internal/extractors/types.go`
- `internal/extractors/extractors.go`
- `internal/analysis/evidence.go`
- `internal/analysis/filename.go`
- `internal/analysis/confidence.go`
- `cmd/client/main.go`
- `internal/ai/prompts.go`

Success criteria:

- Existing AI-only rename flow still works.
- `metadata-only` can rename simple text, Markdown, JSON, CSV, HTML, and PDF files without AI.
- `auto` only calls AI for weak or ambiguous evidence.
- Dry-run output shows method and confidence.

## Milestone 2: Better File Identification

Goal: handle recovered files whose extension may be missing or wrong.

Deliverables:

- Add basic file type detection independent of extension.
- Detect:
  - plain text
  - PDF
  - JSON
  - CSV-like text
  - HTML
  - Markdown-like text
  - ZIP-based Office files as unknown Office documents initially
- Update extractor selection to use detected type, not only file extension.
- Add fallback behavior for unknown files.

Suggested files:

- `internal/filetype/detect.go`
- `internal/filetype/detect_test.go`
- `internal/extractors/extractors.go`

Success criteria:

- A PDF renamed to `abc123.bin` can still be processed as PDF.
- A JSON file with no extension can still be processed.
- Unknown binary files are skipped cleanly with a warning.

## Milestone 3: Evidence Ranking

Goal: identify the most useful parts of a file locally before involving AI.

Deliverables:

- Rank evidence by source:
  - explicit title metadata
  - headings
  - first meaningful paragraph
  - JSON/YAML keys
  - CSV headers
  - PDF first-page text
- Penalize boilerplate, very short strings, repeated junk, and generic words.
- Produce compact evidence bundles for logs, reports, and AI fallback.

Suggested files:

- `internal/analysis/evidence.go`
- `internal/analysis/evidence_test.go`

Success criteria:

- Headings outrank random body text.
- CSV headers produce names like `customer-contact-list`.
- JSON keys produce names like `invoice-export-records` when obvious.
- Evidence sent to AI is small and predictable.

## Milestone 4: Compact AI Fallback

Goal: reduce token use and improve privacy by sending only high-signal evidence.

Deliverables:

- Replace the full-content prompt with a structured evidence prompt.
- Include only:
  - detected type
  - useful metadata
  - top ranked samples
  - filename constraints
- Add `--max-ai-chars` or a similar limit.
- Add logs showing when AI was used and why.

Suggested files:

- `internal/ai/prompts.go`
- `internal/ai/ai.go`
- `cmd/client/main.go`

Success criteria:

- AI requests no longer include full file contents by default.
- AI output still returns one sanitized filename.
- `metadata-only` never calls AI.
- `auto` calls AI only below the configured confidence threshold.

## Milestone 5: Reporting and Safety

Goal: make the tool auditable and safe for recovered file batches.

Deliverables:

- Add `--report report.json`.
- Report:
  - source path
  - destination path
  - suggested filename
  - method used
  - confidence
  - evidence sources
  - warnings
- Add collision handling by appending a counter or short hash.
- Keep non-destructive copy as the default.
- Treat `--rename` as advanced behavior.

Suggested files:

- `internal/report/report.go`
- `internal/utils/files.go`
- `cmd/client/main.go`

Success criteria:

- Batch runs produce a useful report.
- Two files with the same suggested name do not overwrite each other.
- Dry-run can preview all planned changes.

## Milestone 6: Expanded Format Support

Goal: add high-value file types after the pipeline is stable.

Priority order:

1. Office docs: `.docx`, `.xlsx`, `.pptx`
2. Images with EXIF
3. Email files: `.eml`
4. Audio/video metadata through `ffprobe`
5. OCR for scanned PDFs/images

Success criteria:

- Each new format plugs into the same structured extraction pipeline.
- No format-specific logic leaks into the main client flow.
- External tools are optional and fail gracefully when missing.

## Milestone 7: Quality and Review Workflow

Goal: reduce bad names, make batch runs auditable, and make real recovered-file runs safer.

Deliverables:

- Maintain a small recovered-file fixture corpus in `testdata/recovered/`. Done.
- Add CLI integration tests that run dry-run report generation against the fixture corpus. Done.
- Add collision tests for duplicate generated names. Done.
- Improve low-confidence handling so random-looking text becomes `unidentified-content` instead of a bogus filename. Done.
- Add report-driven review mode. Done:
  - generate a report with `--dry-run --report report.json`
  - inspect or edit the report
  - apply it later with `--apply-report report.json`
- Improve Office evidence extraction. Started:
  - workbook sheet names
  - spreadsheet first-row headers
  - presentation slide titles

Success criteria:

- Dry-run report generation is covered by an end-to-end test.
- Applying a report copies files to the planned destinations.
- Duplicate output names get stable suffixes such as `-2`.
- Random content stays low confidence and can trigger AI fallback in `auto`.
- Office files produce more useful local evidence before AI fallback.

## Milestone 8: Format Quality Expansion

Goal: improve local naming quality for common recovered formats without increasing AI use.

Completed deliverables:

- HTML quality pass:
  - extract `<title>`, first `<h1>`, meta description, OpenGraph title, and OpenGraph description
  - restore `.html` for extensionless or wrong-extension HTML
  - add recovered HTML fixture coverage
- XML and MusicXML support:
  - detect XML by content
  - detect MusicXML as a subtype
  - extract generic XML root/title/name/creator fields
  - extract MusicXML work title, movement title, creator/composer, and part names
  - restore `.xml` and `.musicxml`
- Media quality tuning:
  - use embedded title/artist/album tags when `ffprobe` is available
  - keep meaningful media basenames such as `alice.mp4`
  - reject timestamp, camera-style, and random-looking media names
  - use `unidentified-audio`, `unidentified-video`, and `unidentified-media` for weak local evidence
- Review workflow hardening:
  - add report summary stats
  - add `--min-confidence-to-copy`
  - mark skipped report entries with a reason
  - validate `--apply-report` entries before copying

Remaining candidates:

- Add OCR for scanned PDFs and images.
- Add realistic user-provided media fixtures with real title/artist/album/video metadata.
- Add realistic EXIF-rich photo fixtures.
- Add corrupted or partially recovered Office/PDF/media fixtures.
- Add more XLSX/PPTX real-world examples with messy sheet names and slide layouts.

## Starting Recommendation

Milestones 1 through 8 now form the baseline. The next practical work should focus on realistic sample coverage and OCR, not more broad architecture.
