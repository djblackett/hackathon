package extractors

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type archiveExtractor struct{}

func (archiveExtractor) CanHandle(path string) bool {
	return archiveExtension(path) != ""
}

func (archiveExtractor) CanHandleType(detectedType string) bool { return detectedType == "archive" }

func (archiveExtractor) Extract(path string) (string, error) {
	info, err := archiveExtractor{}.ExtractInfo(path)
	if err != nil {
		return "", err
	}
	return info.RawContent, nil
}

func (archiveExtractor) ExtractInfo(path string) (ExtractedFileInfo, error) {
	ext := archiveExtension(path)
	info := NewExtractedFileInfo(path, "archive", "")
	info.SuggestedExtension = ext

	names, err := archiveNames(path, ext)
	if err != nil {
		info.Warnings = append(info.Warnings, "archive entries could not be read")
		return info, nil
	}
	if len(names) == 0 {
		info.Warnings = append(info.Warnings, "archive contains no files")
		return info, nil
	}

	evidence := archiveEvidence(names)
	info.RawContent = evidence
	if evidence != "" {
		info.TextSamples = append(info.TextSamples, TextSample{
			Source: "archive-contents",
			Text:   evidence,
			Score:  0.68,
		})
	}
	return info, nil
}

func archiveExtension(path string) string {
	lower := strings.ToLower(path)
	switch {
	case strings.HasSuffix(lower, ".tar.gz"):
		return "tar.gz"
	case strings.HasSuffix(lower, ".tgz"):
		return "tgz"
	case strings.HasSuffix(lower, ".tar"):
		return "tar"
	case strings.HasSuffix(lower, ".zip"):
		return "zip"
	default:
		return ""
	}
}

func archiveNames(path, ext string) ([]string, error) {
	switch ext {
	case "zip":
		zr, err := zip.OpenReader(path)
		if err != nil {
			return nil, err
		}
		defer zr.Close()
		var names []string
		for _, file := range zr.File {
			if !file.FileInfo().IsDir() {
				names = append(names, file.Name)
			}
		}
		return names, nil
	case "tar", "tar.gz", "tgz":
		return tarNames(path, ext == "tar.gz" || ext == "tgz")
	default:
		return nil, nil
	}
}

func tarNames(path string, gzipped bool) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var r io.Reader = f
	if gzipped {
		gr, err := gzip.NewReader(f)
		if err != nil {
			return nil, err
		}
		defer gr.Close()
		r = gr
	}

	tr := tar.NewReader(r)
	var names []string
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return names, err
		}
		if header.Typeflag == tar.TypeReg {
			names = append(names, header.Name)
		}
	}
	return names, nil
}

func archiveEvidence(names []string) string {
	top := map[string]int{}
	bases := map[string]int{}
	for _, name := range names {
		clean := strings.Trim(strings.ReplaceAll(name, "\\", "/"), "/")
		if clean == "" {
			continue
		}
		parts := strings.Split(clean, "/")
		top[parts[0]]++
		base := strings.TrimSuffix(filepath.Base(clean), filepath.Ext(clean))
		if base != "" {
			bases[base]++
		}
	}
	if winner := majorityKey(top, len(names)); winner != "" {
		return winner
	}
	return strings.Join(topKeys(bases, 6), " ")
}

func majorityKey(counts map[string]int, total int) string {
	type pair struct {
		key   string
		count int
	}
	var pairs []pair
	for key, count := range counts {
		pairs = append(pairs, pair{key: key, count: count})
	}
	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].count == pairs[j].count {
			return pairs[i].key < pairs[j].key
		}
		return pairs[i].count > pairs[j].count
	})
	if len(pairs) > 0 && pairs[0].count >= 2 && pairs[0].count*2 >= total {
		return pairs[0].key
	}
	return ""
}

func topKeys(counts map[string]int, limit int) []string {
	keys := make([]string, 0, len(counts))
	for key := range counts {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	if len(keys) > limit {
		keys = keys[:limit]
	}
	return keys
}

func init() { Register(archiveExtractor{}) }
