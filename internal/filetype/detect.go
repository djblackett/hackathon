package filetype

import (
	"archive/zip"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

const sampleSize = 8192

type Detection struct {
	Type               string
	Subtype            string
	Extension          string
	CanonicalExtension string
	Warning            string
}

func Detect(path string) Detection {
	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(path)), ".")
	detection := Detection{
		Type:      "unknown",
		Extension: ext,
	}

	f, err := os.Open(path)
	if err != nil {
		detection.Warning = err.Error()
		return detection
	}
	defer f.Close()

	buf := make([]byte, sampleSize)
	n, err := f.Read(buf)
	if err != nil && err != io.EOF {
		detection.Warning = err.Error()
		return detection
	}
	sample := buf[:n]

	switch {
	case bytes.HasPrefix(sample, []byte("%PDF-")):
		detection.Type = "pdf"
		detection.CanonicalExtension = "pdf"
	case bytes.HasPrefix(sample, []byte("PK\x03\x04")):
		detection.Type, detection.Subtype = detectZipContainer(path)
	default:
		detection.Type = detectTextLike(sample, ext)
	}
	if detection.CanonicalExtension == "" {
		detection.CanonicalExtension = canonicalExtension(detection.Type, detection.Subtype, ext)
	}

	return detection
}

func detectZipContainer(path string) (string, string) {
	zr, err := zip.OpenReader(path)
	if err != nil {
		return "zip", ""
	}
	defer zr.Close()

	for _, file := range zr.File {
		name := strings.ToLower(file.Name)
		switch {
		case name == "word/document.xml":
			return "office", "docx"
		case name == "xl/workbook.xml":
			return "office", "xlsx"
		case name == "ppt/presentation.xml":
			return "office", "pptx"
		}
	}

	return "zip", ""
}

func detectTextLike(sample []byte, ext string) string {
	trimmed := bytes.TrimSpace(bytes.TrimPrefix(sample, []byte{0xEF, 0xBB, 0xBF}))
	if len(trimmed) == 0 {
		return extensionFallback(ext)
	}
	if !utf8.Valid(trimmed) || mostlyBinary(trimmed) {
		return "unknown"
	}

	lower := strings.ToLower(string(trimmed))

	switch {
	case json.Valid(trimmed):
		return "json"
	case looksHTML(lower):
		return "html"
	case ext == "md" || looksMarkdown(lower):
		return "markdown"
	case ext == "csv" || looksCSV(string(trimmed)):
		return "csv"
	default:
		return extensionFallback(ext)
	}
}

func extensionFallback(ext string) string {
	switch ext {
	case "txt", "log", "cfg", "ini":
		return "text"
	case "md":
		return "markdown"
	case "csv":
		return "csv"
	case "json":
		return "json"
	case "html", "htm":
		return "html"
	case "pdf":
		return "pdf"
	default:
		return "text"
	}
}

func mostlyBinary(sample []byte) bool {
	if len(sample) == 0 {
		return false
	}

	control := 0
	for _, b := range sample {
		if b == 0 {
			return true
		}
		if b < 0x20 && b != '\n' && b != '\r' && b != '\t' {
			control++
		}
	}

	return float64(control)/float64(len(sample)) > 0.05
}

func looksHTML(s string) bool {
	return strings.HasPrefix(s, "<!doctype html") ||
		strings.HasPrefix(s, "<html") ||
		strings.Contains(s, "<html") ||
		strings.Contains(s, "<head") ||
		strings.Contains(s, "<body")
}

func looksMarkdown(s string) bool {
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		return strings.HasPrefix(line, "# ") ||
			strings.HasPrefix(line, "## ") ||
			strings.HasPrefix(line, "---")
	}
	return false
}

func looksCSV(s string) bool {
	reader := csv.NewReader(strings.NewReader(s))
	reader.FieldsPerRecord = -1

	records := 0
	for records < 3 {
		fields, err := reader.Read()
		if err != nil {
			break
		}
		if len(fields) < 2 {
			return false
		}
		records++
	}

	return records > 0
}

func canonicalExtension(detectedType, subtype, originalExt string) string {
	switch detectedType {
	case "pdf":
		return "pdf"
	case "json":
		return "json"
	case "csv":
		return "csv"
	case "html":
		return "html"
	case "markdown":
		return "md"
	case "office":
		return subtype
	case "text":
		if originalExt == "log" || originalExt == "cfg" || originalExt == "ini" || originalExt == "txt" {
			return originalExt
		}
		return "txt"
	default:
		return originalExt
	}
}
