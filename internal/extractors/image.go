package extractors

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type imageExtractor struct{}

func (imageExtractor) CanHandle(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".jpg", ".jpeg", ".png", ".gif":
		return true
	default:
		return false
	}
}

func (imageExtractor) CanHandleType(detectedType string) bool { return detectedType == "image" }

func (imageExtractor) Extract(path string) (string, error) {
	info, err := imageExtractor{}.ExtractInfo(path)
	if err != nil {
		return "", err
	}
	return info.RawContent, nil
}

func (imageExtractor) ExtractInfo(path string) (ExtractedFileInfo, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return ExtractedFileInfo{}, err
	}

	cfg, format, err := image.DecodeConfig(bytes.NewReader(b))
	if err != nil {
		return ExtractedFileInfo{}, err
	}

	content := fmt.Sprintf("%s image %dx%d", format, cfg.Width, cfg.Height)
	info := NewExtractedFileInfo(path, "image", content)
	info.SuggestedExtension = imageExtension(format, filepath.Ext(path))
	info.Metadata["format"] = format
	info.Metadata["dimensions"] = fmt.Sprintf("%dx%d", cfg.Width, cfg.Height)
	info.TextSamples = append([]TextSample{{
		Source: "image-properties",
		Text:   content,
		Score:  0.45,
	}}, info.TextSamples...)

	if meta, warning := exiftoolMetadata(path); warning != "" {
		info.Warnings = append(info.Warnings, warning)
	} else {
		for k, v := range meta {
			info.Metadata[k] = v
		}
		for _, key := range []string{"Title", "Description", "ImageDescription", "ObjectName"} {
			if value := strings.TrimSpace(meta[key]); value != "" {
				info.TextSamples = append([]TextSample{{
					Source: "image-exif-" + strings.ToLower(key),
					Text:   value,
					Score:  0.9,
				}}, info.TextSamples...)
				break
			}
		}
	}

	return info, nil
}

func exiftoolMetadata(path string) (map[string]string, string) {
	if _, err := exec.LookPath("exiftool"); err != nil {
		return nil, "exiftool not available; image EXIF metadata skipped"
	}

	out, err := exec.Command("exiftool", "-json", path).Output()
	if err != nil {
		return nil, "exiftool failed; image EXIF metadata skipped"
	}

	var rows []map[string]any
	if err := json.Unmarshal(out, &rows); err != nil || len(rows) == 0 {
		return nil, "exiftool output could not be parsed"
	}

	meta := map[string]string{}
	for k, v := range rows[0] {
		meta[k] = fmt.Sprint(v)
	}
	return meta, ""
}

func imageExtension(format, fallback string) string {
	switch format {
	case "jpeg":
		return "jpg"
	case "png", "gif":
		return format
	default:
		return strings.TrimPrefix(strings.ToLower(fallback), ".")
	}
}

func init() { Register(imageExtractor{}) }
