package native

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/djblackett/bootdev-hackathon/internal/evidence"
	"github.com/djblackett/bootdev-hackathon/internal/filetype"
)

const textPreviewLimit = 2000

type Extractor struct{}

func (Extractor) Name() evidence.EvidenceSource { return evidence.SourceNativeMIME }

func (Extractor) Available(ctx context.Context) bool { return true }

func (Extractor) Extract(ctx context.Context, path string) (evidence.PartialEvidence, error) {
	detection := filetype.Detect(path)
	ev := evidence.FileEvidence{
		Path:      path,
		Metadata:  map[string]string{},
		Sources:   []evidence.EvidenceSource{evidence.SourceNativeMIME},
		Warnings:  warnings(detection.Warning),
		Extension: dotExt(detection.CanonicalExtension),
	}

	if detection.Extension != "" {
		ev.Metadata["original_extension"] = "." + detection.Extension
	}
	if detection.Subtype != "" {
		ev.Metadata["detected_subtype"] = detection.Subtype
	}

	ev.DetectedMIME = mimeForDetection(detection)
	if ev.Extension == "" {
		ev.Extension = extensionForMIME(ev.DetectedMIME)
	}
	ev.FormatIDs = append(ev.FormatIDs, evidence.FormatID{
		Source:     evidence.SourceNativeMIME,
		Name:       formatName(detection),
		MIME:       ev.DetectedMIME,
		Extension:  ev.Extension,
		Confidence: nativeConfidence(detection),
	})

	if detection.Type == "image" {
		if img := imageEvidence(path); img != nil {
			ev.Image = img
			if ext := imageExtFromFormat(img.Tags["format"]); ext != "" && (ev.Extension == "" || ev.Extension == ".img") {
				ev.Extension = ext
			}
			if mime := imageMIMEFromFormat(img.Tags["format"]); mime != "" {
				ev.DetectedMIME = mime
				if len(ev.FormatIDs) > 0 {
					ev.FormatIDs[len(ev.FormatIDs)-1].MIME = mime
					ev.FormatIDs[len(ev.FormatIDs)-1].Extension = ev.Extension
				}
			}
			if ev.Extension == "" {
				ev.Extension = imageExtFromFormat(img.Tags["format"])
			}
		}
	}

	if isTextLike(detection.Type) {
		preview, err := textPreview(path, textPreviewLimit)
		if err != nil {
			ev.Warnings = append(ev.Warnings, "native text preview failed: "+err.Error())
		} else {
			ev.TextPreview = preview
			if signal := signalForDetection(detection, preview); signal != "" {
				ev.TextSignals = append(ev.TextSignals, signal)
			}
		}
	}

	return evidence.PartialEvidence{Source: evidence.SourceNativeMIME, Evidence: ev}, nil
}

func warnings(warning string) []string {
	if strings.TrimSpace(warning) == "" {
		return nil
	}
	return []string{warning}
}

func mimeForDetection(d filetype.Detection) string {
	switch d.Type {
	case "pdf":
		return "application/pdf"
	case "json":
		return "application/json"
	case "csv":
		return "text/csv"
	case "html":
		return "text/html"
	case "xml":
		if d.Subtype == "musicxml" {
			return "application/vnd.recordare.musicxml+xml"
		}
		return "application/xml"
	case "markdown":
		return "text/markdown"
	case "text":
		return "text/plain"
	case "email":
		return "message/rfc822"
	case "rtf":
		return "application/rtf"
	case "epub":
		return "application/epub+zip"
	case "office":
		return officeMIME(d.Subtype)
	case "opendocument":
		return openDocumentMIME(d.Subtype)
	case "archive":
		if d.Subtype == "tar" || d.Subtype == "tar.gz" {
			return "application/x-tar"
		}
		return "application/zip"
	case "image":
		return detectHTTPMIME(d.Extension)
	case "media":
		return mediaMIME(d.Extension)
	default:
		return detectHTTPMIME(d.Extension)
	}
}

func officeMIME(subtype string) string {
	switch subtype {
	case "docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case "xlsx":
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case "pptx":
		return "application/vnd.openxmlformats-officedocument.presentationml.presentation"
	default:
		return "application/zip"
	}
}

func openDocumentMIME(subtype string) string {
	switch subtype {
	case "odt":
		return "application/vnd.oasis.opendocument.text"
	case "ods":
		return "application/vnd.oasis.opendocument.spreadsheet"
	case "odp":
		return "application/vnd.oasis.opendocument.presentation"
	default:
		return "application/vnd.oasis.opendocument"
	}
}

func mediaMIME(ext string) string {
	switch strings.ToLower(ext) {
	case "mp3":
		return "audio/mpeg"
	case "wav":
		return "audio/wav"
	case "flac":
		return "audio/flac"
	case "mp4", "m4a":
		return "video/mp4"
	case "mov":
		return "video/quicktime"
	case "mkv":
		return "video/x-matroska"
	default:
		return "application/octet-stream"
	}
}

func detectHTTPMIME(ext string) string {
	switch strings.ToLower(ext) {
	case "jpg", "jpeg":
		return "image/jpeg"
	case "png":
		return "image/png"
	case "gif":
		return "image/gif"
	}
	return "application/octet-stream"
}

func extensionForMIME(mime string) string {
	switch mime {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/gif":
		return ".gif"
	case "application/pdf":
		return ".pdf"
	case "text/plain":
		return ".txt"
	default:
		return ""
	}
}

func dotExt(ext string) string {
	ext = strings.TrimSpace(ext)
	if ext == "" {
		return ""
	}
	if strings.HasPrefix(ext, ".") {
		return strings.ToLower(ext)
	}
	return "." + strings.ToLower(ext)
}

func formatName(d filetype.Detection) string {
	if d.Subtype != "" {
		return d.Type + "/" + d.Subtype
	}
	return d.Type
}

func nativeConfidence(d filetype.Detection) float64 {
	if d.Type == "unknown" {
		return 0
	}
	if d.CanonicalExtension != "" {
		return 0.8
	}
	return 0.55
}

func imageEvidence(path string) *evidence.ImageEvidence {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()
	cfg, format, err := image.DecodeConfig(f)
	if err != nil {
		return nil
	}
	return &evidence.ImageEvidence{
		Width:  cfg.Width,
		Height: cfg.Height,
		Tags: map[string]string{
			"format": format,
		},
	}
}

func imageExtFromFormat(format string) string {
	switch strings.ToLower(format) {
	case "jpeg":
		return ".jpg"
	case "png", "gif":
		return "." + strings.ToLower(format)
	default:
		return ""
	}
}

func imageMIMEFromFormat(format string) string {
	switch strings.ToLower(format) {
	case "jpeg":
		return "image/jpeg"
	case "png":
		return "image/png"
	case "gif":
		return "image/gif"
	default:
		return ""
	}
}

func isTextLike(detectedType string) bool {
	switch detectedType {
	case "text", "markdown", "csv", "json", "html", "xml", "email":
		return true
	default:
		return false
	}
}

func textPreview(path string, limit int) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, int64(limit*2)))
	if err != nil {
		return "", err
	}
	data = bytes.TrimPrefix(data, []byte{0xEF, 0xBB, 0xBF})
	if !utf8.Valid(data) {
		return "", nil
	}
	text := strings.ReplaceAll(string(data), "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	text = strings.TrimSpace(text)
	if len(text) > limit {
		text = strings.TrimSpace(text[:limit])
	}
	return text, nil
}

func signalForDetection(detection filetype.Detection, preview string) string {
	switch detection.Type {
	case "csv":
		return csvSignal(preview)
	case "json":
		return jsonSignal(preview)
	case "markdown":
		return markdownSignal(preview)
	case "html":
		return htmlSignal(preview)
	case "xml":
		return xmlSignal(preview)
	case "email":
		return emailSignal(preview)
	default:
		return firstMeaningfulLine(preview)
	}
}

func csvSignal(preview string) string {
	reader := csv.NewReader(strings.NewReader(preview))
	reader.FieldsPerRecord = -1
	record, err := reader.Read()
	if err != nil {
		return firstMeaningfulLine(preview)
	}
	fields := make([]string, 0, len(record))
	for _, field := range record {
		field = strings.TrimSpace(field)
		if field != "" {
			fields = append(fields, field)
		}
	}
	return strings.Join(fields, " ")
}

func jsonSignal(preview string) string {
	var value any
	if err := json.Unmarshal([]byte(preview), &value); err != nil {
		return firstMeaningfulLine(preview)
	}
	keys := map[string]struct{}{}
	collectJSONKeys(value, keys)
	if len(keys) == 0 {
		return firstMeaningfulLine(preview)
	}
	sorted := make([]string, 0, len(keys))
	for key := range keys {
		sorted = append(sorted, key)
	}
	sort.Strings(sorted)
	if len(sorted) > 10 {
		sorted = sorted[:10]
	}
	return strings.Join(sorted, " ")
}

func collectJSONKeys(value any, keys map[string]struct{}) {
	switch v := value.(type) {
	case map[string]any:
		for key, child := range v {
			keys[key] = struct{}{}
			collectJSONKeys(child, keys)
		}
	case []any:
		for _, child := range v {
			collectJSONKeys(child, keys)
		}
	}
}

func markdownSignal(preview string) string {
	for _, line := range strings.Split(preview, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") {
			return strings.TrimSpace(strings.TrimLeft(line, "#"))
		}
	}
	return firstMeaningfulLine(preview)
}

var (
	htmlTitlePattern = regexp.MustCompile(`(?is)<title[^>]*>(.*?)</title>`)
	xmlTitlePattern  = regexp.MustCompile(`(?is)<(?:[^:>]+:)?title[^>]*>(.*?)</(?:[^:>]+:)?title>`)
	tagPattern       = regexp.MustCompile(`(?is)<[^>]+>`)
)

func htmlSignal(preview string) string {
	if match := htmlTitlePattern.FindStringSubmatch(preview); len(match) > 1 {
		return cleanMarkupText(match[1])
	}
	return firstMeaningfulLine(tagPattern.ReplaceAllString(preview, " "))
}

func xmlSignal(preview string) string {
	if match := xmlTitlePattern.FindStringSubmatch(preview); len(match) > 1 {
		return cleanMarkupText(match[1])
	}
	return firstMeaningfulLine(tagPattern.ReplaceAllString(preview, " "))
}

func emailSignal(preview string) string {
	for _, line := range strings.Split(preview, "\n") {
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(key), "subject") {
			return strings.TrimSpace(value)
		}
	}
	return firstMeaningfulLine(preview)
}

func cleanMarkupText(value string) string {
	value = tagPattern.ReplaceAllString(value, " ")
	return strings.Join(strings.Fields(value), " ")
}

func firstMeaningfulLine(text string) string {
	for _, line := range strings.Split(text, "\n") {
		line = strings.Join(strings.Fields(line), " ")
		if line != "" {
			return line
		}
	}
	if len(text) > 160 {
		return strings.TrimSpace(text[:160])
	}
	return strings.TrimSpace(text)
}

func DetectContentType(data []byte) string {
	return http.DetectContentType(data)
}
