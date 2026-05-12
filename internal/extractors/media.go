package extractors

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
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
		return info, nil
	}

	for k, v := range meta {
		info.Metadata[k] = v
	}
	title := firstNonEmpty(meta["title"], meta["album"], meta["artist"])
	if title != "" {
		info.RawContent = strings.Join([]string{meta["title"], meta["artist"], meta["album"]}, " ")
		info.TextSamples = append(info.TextSamples, TextSample{
			Source: "media-tags",
			Text:   info.RawContent,
			Score:  0.9,
		})
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
			Duration string            `json:"duration"`
			Tags     map[string]string `json:"tags"`
		} `json:"format"`
	}
	if err := json.Unmarshal(out, &payload); err != nil {
		return nil, "ffprobe output could not be parsed"
	}

	meta := map[string]string{}
	if payload.Format.Duration != "" {
		meta["duration"] = payload.Format.Duration
	}
	for k, v := range payload.Format.Tags {
		meta[strings.ToLower(k)] = fmt.Sprint(v)
	}
	return meta, ""
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func init() { Register(mediaExtractor{}) }
