package ffprobe

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

func (Extractor) Name() evidence.EvidenceSource { return evidence.SourceFFProbe }

func (Extractor) Available(ctx context.Context) bool {
	return tools.Available("ffprobe")
}

func (e Extractor) Extract(ctx context.Context, path string) (evidence.PartialEvidence, error) {
	result, err := tools.Run(ctx, e.Timeout, "ffprobe", "-v", "error", "-print_format", "json", "-show_format", "-show_streams", path)
	if err != nil {
		message := strings.TrimSpace(string(result.Stderr))
		if message == "" {
			message = err.Error()
		}
		return evidence.PartialEvidence{}, fmt.Errorf("ffprobe failed: %s", message)
	}
	ev, err := Parse(path, result.Stdout)
	if err != nil {
		return evidence.PartialEvidence{}, err
	}
	return evidence.PartialEvidence{Source: evidence.SourceFFProbe, Evidence: ev}, nil
}

func Parse(path string, data []byte) (evidence.FileEvidence, error) {
	var payload ffprobePayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return evidence.FileEvidence{}, fmt.Errorf("parse ffprobe json: %w", err)
	}
	ev := evidence.FileEvidence{
		Path:     path,
		Metadata: map[string]string{},
		Sources:  []evidence.EvidenceSource{evidence.SourceFFProbe},
	}
	if payload.Format.FormatName != "" {
		ev.Metadata["ffprobe:format_name"] = payload.Format.FormatName
	}
	if payload.Format.FormatLongName != "" {
		ev.Metadata["ffprobe:format_long_name"] = payload.Format.FormatLongName
	}
	for key, value := range payload.Format.Tags {
		ev.Metadata[strings.ToLower(key)] = value
	}
	if payload.Format.FormatName != "" || payload.Format.FormatLongName != "" {
		ev.FormatIDs = append(ev.FormatIDs, evidence.FormatID{
			Source:     evidence.SourceFFProbe,
			Name:       firstNonEmpty(payload.Format.FormatLongName, payload.Format.FormatName),
			Confidence: 0.7,
		})
	}

	stream := bestStream(payload.Streams)
	tags := mergedTags(payload.Format.Tags, stream.Tags)
	ev.Media = &evidence.MediaEvidence{
		DurationSeconds: parseFloat(payload.Format.Duration),
		Codec:           stream.CodecName,
		Width:           stream.Width,
		Height:          stream.Height,
		Tags:            tags,
	}
	if ev.Media.DurationSeconds == 0 && ev.Media.Codec == "" && ev.Media.Width == 0 && ev.Media.Height == 0 && len(ev.Media.Tags) == 0 {
		ev.Media = nil
	}
	if title := firstTag(tags, "title"); title != "" {
		ev.TextSignals = append(ev.TextSignals, title)
	}
	if artist := firstTag(tags, "artist"); artist != "" {
		ev.TextSignals = append(ev.TextSignals, artist)
	}
	return ev, nil
}

type ffprobePayload struct {
	Format  ffprobeFormat   `json:"format"`
	Streams []ffprobeStream `json:"streams"`
}

type ffprobeFormat struct {
	FormatName     string            `json:"format_name"`
	FormatLongName string            `json:"format_long_name"`
	Duration       string            `json:"duration"`
	Tags           map[string]string `json:"tags"`
}

type ffprobeStream struct {
	CodecName string            `json:"codec_name"`
	CodecType string            `json:"codec_type"`
	Width     int               `json:"width"`
	Height    int               `json:"height"`
	Tags      map[string]string `json:"tags"`
}

func bestStream(streams []ffprobeStream) ffprobeStream {
	for _, stream := range streams {
		if stream.CodecType == "video" {
			return stream
		}
	}
	for _, stream := range streams {
		if stream.CodecType == "audio" {
			return stream
		}
	}
	if len(streams) > 0 {
		return streams[0]
	}
	return ffprobeStream{}
}

func mergedTags(maps ...map[string]string) map[string]string {
	out := map[string]string{}
	for _, values := range maps {
		for key, value := range values {
			if strings.TrimSpace(value) == "" {
				continue
			}
			out[strings.ToLower(key)] = strings.TrimSpace(value)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func firstTag(tags map[string]string, key string) string {
	for gotKey, value := range tags {
		if strings.EqualFold(gotKey, key) && strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func parseFloat(value string) float64 {
	n, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
	if err != nil {
		return 0
	}
	return n
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
