package naming

import (
	"path/filepath"
	"regexp"
	"strings"
)

const maxFilenameBaseLength = 80

var (
	controlChars      = regexp.MustCompile(`[\x00-\x1f\x7f]`)
	unsafeChars       = regexp.MustCompile(`[^a-z0-9._-]+`)
	repeatedHyphens   = regexp.MustCompile(`-+`)
	windowsReserved   = regexp.MustCompile(`^(con|prn|aux|nul|com[1-9]|lpt[1-9])$`)
	leadingSeparators = regexp.MustCompile(`^[._-]+`)
)

func SanitizeBase(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.ReplaceAll(value, string(filepath.Separator), " ")
	value = strings.ReplaceAll(value, "/", " ")
	value = strings.ReplaceAll(value, "\\", " ")
	value = controlChars.ReplaceAllString(value, " ")
	value = strings.Join(strings.Fields(value), "-")
	value = unsafeChars.ReplaceAllString(value, "-")
	value = repeatedHyphens.ReplaceAllString(value, "-")
	value = strings.Trim(value, " .-_")
	value = leadingSeparators.ReplaceAllString(value, "")
	if len(value) > maxFilenameBaseLength {
		value = strings.Trim(value[:maxFilenameBaseLength], " .-_")
	}
	if value == "" {
		value = "unidentified"
	}
	if windowsReserved.MatchString(value) {
		value += "-file"
	}
	return value
}

func NormalizeExtension(ext string) string {
	ext = strings.ToLower(strings.TrimSpace(ext))
	if ext == "" {
		return ".bin"
	}
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	ext = strings.ReplaceAll(ext, "/", "")
	ext = strings.ReplaceAll(ext, "\\", "")
	if ext == "." {
		return ".bin"
	}
	return ext
}

func WithExtension(base, ext string) string {
	return SanitizeBase(base) + NormalizeExtension(ext)
}
