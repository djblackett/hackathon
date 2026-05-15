package exiftool

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/djblackett/bootdev-hackathon/internal/evidence"
	"github.com/djblackett/bootdev-hackathon/internal/tools"
)

type Extractor struct {
	Timeout time.Duration
}

func (Extractor) Name() evidence.EvidenceSource { return evidence.SourceExifTool }

func (Extractor) Available(ctx context.Context) bool {
	return tools.Available("exiftool")
}

func (e Extractor) Extract(ctx context.Context, path string) (evidence.PartialEvidence, error) {
	result, err := tools.Run(ctx, e.Timeout, "exiftool", "-j", "-G", "-n", path)
	if err != nil {
		message := strings.TrimSpace(string(result.Stderr))
		if message == "" {
			message = err.Error()
		}
		return evidence.PartialEvidence{}, fmt.Errorf("exiftool failed: %s", message)
	}
	ev, err := Parse(path, result.Stdout)
	if err != nil {
		return evidence.PartialEvidence{}, err
	}
	return evidence.PartialEvidence{Source: evidence.SourceExifTool, Evidence: ev}, nil
}

func Parse(path string, data []byte) (evidence.FileEvidence, error) {
	var rows []map[string]any
	if err := json.Unmarshal(data, &rows); err != nil {
		return evidence.FileEvidence{}, fmt.Errorf("parse exiftool json: %w", err)
	}
	ev := evidence.FileEvidence{
		Path:     path,
		Metadata: map[string]string{},
		Sources:  []evidence.EvidenceSource{evidence.SourceExifTool},
	}
	if len(rows) == 0 {
		ev.Warnings = append(ev.Warnings, "exiftool returned no rows")
		return ev, nil
	}
	row := rows[0]
	for key, value := range row {
		if key == "SourceFile" {
			continue
		}
		if s := valueString(value); s != "" {
			ev.Metadata[key] = s
		}
	}

	mime := firstValue(row, "File:MIMEType", "MIMEType")
	ext := firstValue(row, "File:FileTypeExtension", "FileTypeExtension")
	if mime != "" {
		ev.DetectedMIME = mime
	}
	if ext != "" {
		ev.Extension = dotExt(ext)
	}
	if mime != "" || ext != "" {
		ev.FormatIDs = append(ev.FormatIDs, evidence.FormatID{
			Source:     evidence.SourceExifTool,
			Name:       firstValue(row, "File:FileType", "FileType", "File:FileTypeExtension", "FileTypeExtension"),
			MIME:       mime,
			Extension:  dotExt(ext),
			Confidence: 0.75,
		})
	}

	image := imageEvidence(row)
	if image != nil {
		ev.Image = image
	}
	media := mediaEvidence(row)
	if media != nil {
		ev.Media = media
	}
	if title := firstValue(row, "XMP:Title", "IPTC:ObjectName", "EXIF:ImageDescription", "QuickTime:Title", "ID3:Title", "Title"); title != "" {
		ev.TextSignals = append(ev.TextSignals, title)
	}
	return ev, nil
}

func imageEvidence(row map[string]any) *evidence.ImageEvidence {
	width := firstInt(row, "File:ImageWidth", "EXIF:ExifImageWidth", "Composite:ImageWidth", "ImageWidth")
	height := firstInt(row, "File:ImageHeight", "EXIF:ExifImageHeight", "Composite:ImageHeight", "ImageHeight")
	make := firstValue(row, "EXIF:Make", "Make")
	model := firstValue(row, "EXIF:Model", "Model")
	takenAt := firstValue(row, "EXIF:DateTimeOriginal", "EXIF:CreateDate", "QuickTime:CreateDate", "CreateDate")
	gpsDate := firstValue(row, "Composite:GPSDateTime", "EXIF:GPSDateStamp", "GPSDateTime", "GPSDateStamp")
	if width == 0 && height == 0 && make == "" && model == "" && takenAt == "" && gpsDate == "" {
		return nil
	}
	return &evidence.ImageEvidence{
		Width:       width,
		Height:      height,
		CameraMake:  make,
		CameraModel: model,
		TakenAt:     takenAt,
		GPSDate:     gpsDate,
		Tags:        selectedTags(row, "EXIF:", "XMP:", "IPTC:"),
	}
}

func mediaEvidence(row map[string]any) *evidence.MediaEvidence {
	duration := firstFloat(row, "QuickTime:Duration", "Composite:Duration", "Duration")
	codec := firstValue(row, "QuickTime:CompressorID", "QuickTime:VideoCodec", "Audio:Codec", "Codec")
	width := firstInt(row, "QuickTime:ImageWidth", "Track1:ImageWidth", "ImageWidth")
	height := firstInt(row, "QuickTime:ImageHeight", "Track1:ImageHeight", "ImageHeight")
	tags := selectedTags(row, "QuickTime:", "ID3:", "RIFF:")
	if duration == 0 && codec == "" && width == 0 && height == 0 && len(tags) == 0 {
		return nil
	}
	return &evidence.MediaEvidence{
		DurationSeconds: duration,
		Codec:           codec,
		Width:           width,
		Height:          height,
		Tags:            tags,
	}
}

func selectedTags(row map[string]any, prefixes ...string) map[string]string {
	tags := map[string]string{}
	for key, value := range row {
		for _, prefix := range prefixes {
			if strings.HasPrefix(key, prefix) {
				if s := valueString(value); s != "" {
					tags[key] = s
				}
				break
			}
		}
	}
	if len(tags) == 0 {
		return nil
	}
	return tags
}

func firstValue(row map[string]any, keys ...string) string {
	for _, key := range keys {
		if value, ok := row[key]; ok {
			if s := valueString(value); s != "" {
				return s
			}
		}
	}
	return ""
}

func firstInt(row map[string]any, keys ...string) int {
	for _, key := range keys {
		if value, ok := row[key]; ok {
			switch v := value.(type) {
			case float64:
				return int(v)
			case int:
				return v
			case string:
				n, err := strconv.Atoi(strings.TrimSpace(v))
				if err == nil {
					return n
				}
			}
		}
	}
	return 0
}

func firstFloat(row map[string]any, keys ...string) float64 {
	for _, key := range keys {
		if value, ok := row[key]; ok {
			switch v := value.(type) {
			case float64:
				return v
			case int:
				return float64(v)
			case string:
				n, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
				if err == nil {
					return n
				}
			}
		}
	}
	return 0
}

func valueString(value any) string {
	switch v := value.(type) {
	case string:
		return strings.TrimSpace(v)
	case float64, bool:
		return strings.TrimSpace(fmt.Sprint(v))
	case []any:
		parts := []string{}
		for _, item := range v {
			if s := valueString(item); s != "" {
				parts = append(parts, s)
			}
		}
		return strings.Join(parts, ", ")
	default:
		if value == nil {
			return ""
		}
		return strings.TrimSpace(fmt.Sprint(value))
	}
}

func dotExt(ext string) string {
	ext = strings.TrimSpace(strings.ToLower(ext))
	if ext == "" {
		return ""
	}
	if strings.HasPrefix(ext, ".") {
		return ext
	}
	return "." + ext
}
