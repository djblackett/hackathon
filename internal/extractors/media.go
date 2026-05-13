package extractors

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

type mediaExtractor struct{}

func (mediaExtractor) CanHandle(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".mp3", ".mp4", ".m4a", ".mov", ".wav", ".flac", ".mkv", ".avi":
		return true
	default:
		return false
	}
}

func (mediaExtractor) CanHandleType(detectedType string) bool { return detectedType == "media" }

func (mediaExtractor) Extract(path string) (string, error) {
	info, err := mediaExtractor{}.ExtractInfo(path)
	if err != nil {
		return "", err
	}
	return info.RawContent, nil
}

func (mediaExtractor) ExtractInfo(path string) (ExtractedFileInfo, error) {
	info := NewExtractedFileInfo(path, "media", "")
	info.SuggestedExtension = strings.TrimPrefix(strings.ToLower(filepath.Ext(path)), ".")

	meta, warning := ffprobeMetadata(path)
	if warning != "" {
		info.Warnings = append(info.Warnings, warning)
	}

	for k, v := range meta {
		info.Metadata[k] = v
	}
	title := strings.Join(nonEmpty(meta["title"], meta["artist"], meta["album"]), " ")
	if title != "" {
		info.RawContent = title
		info.TextSamples = append(info.TextSamples, TextSample{
			Source: "media-tags",
			Text:   info.RawContent,
			Score:  0.9,
		})
	}

	dateText := firstNonEmpty(meta["creation_time"], meta["date"])
	if dateText != "" {
		info.TextSamples = append(info.TextSamples, TextSample{
			Source: "media-date",
			Text:   mediaKind(info.SuggestedExtension) + " " + dateText,
			Score:  0.62,
		})
	}

	if name := meaningfulMediaBasename(path); name != "" {
		info.TextSamples = append(info.TextSamples, TextSample{
			Source: "media-filename",
			Text:   name,
			Score:  mediaFilenameScore(name),
		})
	}

	if len(info.TextSamples) == 0 {
		info.RawContent = strings.Join(nonEmpty(mediaKind(info.SuggestedExtension), meta["format_long_name"], meta["duration"]), " ")
		if info.RawContent != "" {
			info.TextSamples = append(info.TextSamples, TextSample{
				Source: "media-properties",
				Text:   info.RawContent,
				Score:  0.35,
			})
		}
	}
	return info, nil
}

func ffprobeMetadata(path string) (map[string]string, string) {
	if _, err := exec.LookPath("ffprobe"); err != nil {
		return nil, "ffprobe not available; media metadata skipped"
	}

	out, err := exec.Command("ffprobe", "-v", "quiet", "-print_format", "json", "-show_format", path).Output()
	if err != nil {
		return nil, "ffprobe failed; media metadata skipped"
	}

	var payload struct {
		Format struct {
			FormatName     string            `json:"format_name"`
			FormatLongName string            `json:"format_long_name"`
			Duration       string            `json:"duration"`
			Tags           map[string]string `json:"tags"`
		} `json:"format"`
	}
	if err := json.Unmarshal(out, &payload); err != nil {
		return nil, "ffprobe output could not be parsed"
	}

	meta := map[string]string{}
	if payload.Format.FormatName != "" {
		meta["format_name"] = payload.Format.FormatName
	}
	if payload.Format.FormatLongName != "" {
		meta["format_long_name"] = payload.Format.FormatLongName
	}
	if payload.Format.Duration != "" {
		meta["duration"] = payload.Format.Duration
	}
	for k, v := range payload.Format.Tags {
		meta[strings.ToLower(k)] = fmt.Sprint(v)
	}
	return meta, ""
}

func nonEmpty(values ...string) []string {
	out := []string{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			out = append(out, value)
		}
	}
	return out
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func mediaKind(ext string) string {
	switch ext {
	case "mp3", "m4a", "wav", "flac":
		return "audio"
	case "mp4", "mov", "mkv", "avi":
		return "video"
	default:
		return "media"
	}
}

func meaningfulMediaBasename(path string) string {
	base := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	base = strings.TrimSpace(base)
	if base == "" {
		return ""
	}

	lower := strings.ToLower(base)
	if regexp.MustCompile(`^[a-f0-9_-]{8,}$`).MatchString(lower) {
		return ""
	}
	if regexp.MustCompile(`^\d{4}[-_]\d{2}[-_]\d{2}([_ -]\d{2}[-_]\d{2}[-_]\d{2})?$`).MatchString(lower) {
		return ""
	}
	if regexp.MustCompile(`^(img|dsc|mov|vid|pxl|audio|video|file)[-_ ]?\d+([_-]\d+)*$`).MatchString(lower) {
		return ""
	}
	if looksRandomMediaName(lower) {
		return ""
	}
	base = strings.NewReplacer("_", " ", "-", " ").Replace(base)
	return base
}

func mediaFilenameScore(name string) float64 {
	words := regexp.MustCompile(`[A-Za-z0-9]+`).FindAllString(name, -1)
	if len(words) == 1 && len(words[0]) < 4 {
		return 0.52
	}
	if len(words) <= 2 {
		return 0.72
	}
	return 0.82
}

func looksRandomMediaName(s string) bool {
	tokens := regexp.MustCompile(`[a-z0-9]+`).FindAllString(strings.ToLower(s), -1)
	for _, token := range tokens {
		if len(token) < 12 {
			continue
		}
		letters := 0
		vowels := 0
		digits := 0
		for _, r := range token {
			switch {
			case r >= '0' && r <= '9':
				digits++
			case r >= 'a' && r <= 'z':
				letters++
				if strings.ContainsRune("aeiou", r) {
					vowels++
				}
			}
		}
		if letters >= 8 && float64(vowels)/float64(letters) < 0.2 {
			return true
		}
		if digits >= 6 && letters >= 4 {
			return true
		}
	}
	return false
}

func init() { Register(mediaExtractor{}) }
