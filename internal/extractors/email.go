package extractors

import (
	"io"
	"net/mail"
	"os"
	"strings"
)

type emailExtractor struct{}

func (emailExtractor) CanHandle(path string) bool {
	return strings.HasSuffix(strings.ToLower(path), ".eml")
}

func (emailExtractor) CanHandleType(detectedType string) bool { return detectedType == "email" }

func (emailExtractor) Extract(path string) (string, error) {
	info, err := emailExtractor{}.ExtractInfo(path)
	if err != nil {
		return "", err
	}
	return info.RawContent, nil
}

func (emailExtractor) ExtractInfo(path string) (ExtractedFileInfo, error) {
	f, err := os.Open(path)
	if err != nil {
		return ExtractedFileInfo{}, err
	}
	defer f.Close()

	msg, err := mail.ReadMessage(f)
	if err != nil {
		return ExtractedFileInfo{}, err
	}

	body, _ := io.ReadAll(io.LimitReader(msg.Body, 8192))
	info := NewExtractedFileInfo(path, "email", string(body))
	info.SuggestedExtension = "eml"

	for _, key := range []string{"Subject", "From", "To", "Date"} {
		value := strings.TrimSpace(msg.Header.Get(key))
		if value != "" {
			info.Metadata[strings.ToLower(key)] = value
		}
	}
	if subject := info.Metadata["subject"]; subject != "" {
		info.TextSamples = append([]TextSample{{
			Source: "email-subject",
			Text:   subject,
			Score:  0.95,
		}}, info.TextSamples...)
	}

	return info, nil
}

func init() { Register(emailExtractor{}) }
