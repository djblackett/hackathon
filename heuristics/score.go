package heuristics

func Score(content string) (filename string, confidence int) {
	// 100‑point scale
	// if meta := firstMetadataTitle(content); meta != "" {
	// 	return utils.Sanitize(meta), 90
	// }
	// if h1 := firstHeading(content); h1 != "" {
	// 	return utils.Sanitize(h1), 80
	// }
	// keywords := topKeywords(content)
	// if len(keywords) >= 3 {
	// 	return utils.Sanitize(strings.Join(keywords, "-")), 60
	// }
	return "", 20 // low confidence
}
