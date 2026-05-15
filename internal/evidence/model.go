package evidence

import "context"

type EvidenceSource string

const (
	SourceNativeMIME EvidenceSource = "native-mimetype"
	SourceTika       EvidenceSource = "tika"
	SourceExifTool   EvidenceSource = "exiftool"
	SourceFFProbe    EvidenceSource = "ffprobe"
	SourceSiegfried  EvidenceSource = "siegfried"
	SourceTesseract  EvidenceSource = "tesseract"
	SourceJHOVE      EvidenceSource = "jhove"
)

type FileEvidence struct {
	Path      string `json:"path"`
	SizeBytes int64  `json:"sizeBytes"`
	SHA256    string `json:"sha256,omitempty"`

	DetectedMIME string `json:"detectedMime,omitempty"`
	Extension    string `json:"extension,omitempty"`

	FormatIDs   []FormatID        `json:"formatIds,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	TextPreview string            `json:"textPreview,omitempty"`
	TextSignals []string          `json:"textSignals,omitempty"`

	Media      *MediaEvidence    `json:"media,omitempty"`
	Image      *ImageEvidence    `json:"image,omitempty"`
	Validation *ValidationResult `json:"validation,omitempty"`

	Sources  []EvidenceSource `json:"sources"`
	Warnings []string         `json:"warnings,omitempty"`
	Errors   []ToolError      `json:"errors,omitempty"`
}

type FormatID struct {
	Source     EvidenceSource `json:"source"`
	ID         string         `json:"id,omitempty"`
	Name       string         `json:"name,omitempty"`
	Version    string         `json:"version,omitempty"`
	MIME       string         `json:"mime,omitempty"`
	Extension  string         `json:"extension,omitempty"`
	Confidence float64        `json:"confidence,omitempty"`
}

type MediaEvidence struct {
	DurationSeconds float64           `json:"durationSeconds,omitempty"`
	Codec           string            `json:"codec,omitempty"`
	Width           int               `json:"width,omitempty"`
	Height          int               `json:"height,omitempty"`
	Tags            map[string]string `json:"tags,omitempty"`
}

type ImageEvidence struct {
	Width       int               `json:"width,omitempty"`
	Height      int               `json:"height,omitempty"`
	CameraMake  string            `json:"cameraMake,omitempty"`
	CameraModel string            `json:"cameraModel,omitempty"`
	TakenAt     string            `json:"takenAt,omitempty"`
	GPSDate     string            `json:"gpsDate,omitempty"`
	Tags        map[string]string `json:"tags,omitempty"`
}

type ValidationResult struct {
	Source   EvidenceSource `json:"source"`
	Valid    *bool          `json:"valid,omitempty"`
	Status   string         `json:"status,omitempty"`
	Warnings []string       `json:"warnings,omitempty"`
}

type ToolError struct {
	Source  EvidenceSource `json:"source"`
	Message string         `json:"message"`
}

type PartialEvidence struct {
	Source   EvidenceSource
	Evidence FileEvidence
}

type Extractor interface {
	Name() EvidenceSource
	Available(ctx context.Context) bool
	Extract(ctx context.Context, path string) (PartialEvidence, error)
}
