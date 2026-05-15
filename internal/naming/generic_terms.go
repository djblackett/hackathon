package naming

import "strings"

var genericTerms = map[string]struct{}{
	"untitled":                  {},
	"document":                  {},
	"document1":                 {},
	"scan":                      {},
	"scan0001":                  {},
	"image":                     {},
	"img":                       {},
	"photo":                     {},
	"microsoft word document":   {},
	"microsoft word - document": {},
	"adobe photoshop":           {},
	"canon":                     {},
	"epson scan":                {},
	"hp scan":                   {},
}

func IsGenericTitle(value string) bool {
	normalized := strings.ToLower(strings.TrimSpace(value))
	normalized = strings.Join(strings.Fields(strings.NewReplacer("_", " ", "-", " ").Replace(normalized)), " ")
	if normalized == "" {
		return true
	}
	if _, ok := genericTerms[normalized]; ok {
		return true
	}
	if strings.HasPrefix(normalized, "scan ") || strings.HasPrefix(normalized, "image ") {
		return true
	}
	return false
}
