package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"

	"github.com/djblackett/bootdev-hackathon/internal/ai"
	"github.com/djblackett/bootdev-hackathon/internal/config"
	"github.com/djblackett/bootdev-hackathon/internal/report"
)

func TestClientDryRunReportRecoveredCorpus(t *testing.T) {
	root := repoRoot(t)
	outputDir := t.TempDir()
	reportPath := filepath.Join(t.TempDir(), "report.json")

	runClient(t, root,
		"--strategy", "metadata-only",
		"--dry-run",
		"--input", filepath.Join(root, "testdata/recovered"),
		"--output", outputDir,
		"--report", reportPath,
	)

	got := readReport(t, reportPath)
	if len(got.Entries) != 12 {
		t.Fatalf("report entries = %d, want 12", len(got.Entries))
	}
	if got.Summary.TotalFiles != 12 {
		t.Fatalf("summary total = %d, want 12", got.Summary.TotalFiles)
	}
	if got.Summary.PlannedCount != 12 || got.Summary.CopiedCount != 0 {
		t.Fatalf("unexpected dry-run summary: %+v", got.Summary)
	}
	if got.Summary.LowConfidenceCount != 1 {
		t.Fatalf("low confidence count = %d, want 1", got.Summary.LowConfidenceCount)
	}

	bySource := map[string]report.Entry{}
	destNames := map[string]bool{}
	for _, entry := range got.Entries {
		bySource[filepath.Base(entry.SourcePath)] = entry
		destNames[filepath.Base(entry.DestinationPath)] = true
		if !entry.DryRun {
			t.Fatalf("entry should be dry-run: %+v", entry)
		}
	}

	assertDestExt(t, bySource, "recovered-pdf.bin", ".pdf")
	assertDestExt(t, bySource, "file0007", ".json")
	assertDestExt(t, bySource, "recovered-doc", ".docx")
	assertDestExt(t, bySource, "unknown.dat", ".csv")
	assertDestExt(t, bySource, "message", ".eml")
	assertDestExt(t, bySource, "markdown-note", ".md")
	assertDestExt(t, bySource, "recovered-html", ".html")
	assertDestExt(t, bySource, "recovered-musicxml", ".musicxml")
	assertDestExt(t, bySource, "recovered-xml.bin", ".xml")
	assertSuggestedName(t, bySource, "customer-a.csv", "name-email-status.csv")
	assertSuggestedName(t, bySource, "customer-b.csv", "name-email-status.csv")
	assertSuggestedName(t, bySource, "unknown.dat", "customer-first-name-last-company-city-country-phone.csv")
	assertSuggestedName(t, bySource, "markdown-note", "incident-response-runbook.md")
	assertSuggestedName(t, bySource, "message", "customer-onboarding-checklist.eml")
	assertSuggestedName(t, bySource, "recovered-doc", "monumental-construction-panama-canal.docx")
	assertSuggestedName(t, bySource, "recovered-html", "basics-photosynthesis.html")
	assertSuggestedName(t, bySource, "recovered-musicxml", "you-my-sunshine.musicxml")
	assertSuggestedName(t, bySource, "recovered-xml.bin", "quarterly-safety-inspection-log.xml")
	assertSuggestedName(t, bySource, "random.txt", "unidentified-content.txt")

	if bySource["random.txt"].Confidence >= 0.4 {
		t.Fatalf("random.txt confidence = %.2f, want low confidence", bySource["random.txt"].Confidence)
	}
	if !destNames["name-email-status.csv"] || !destNames["name-email-status-2.csv"] {
		t.Fatalf("missing collision-safe duplicate csv destinations: %+v", destNames)
	}
	assertGoldenReport(t, got, map[string]goldenEntry{
		"customer-a.csv":     {Suggested: "name-email-status.csv", Method: "metadata", Evidence: "csv-headers"},
		"customer-b.csv":     {Suggested: "name-email-status.csv", Method: "metadata", Evidence: "csv-headers"},
		"file0007":           {Dest: "quiz-sport-q1-question-which-one-correct-team.json", Suggested: "quiz-sport-q1-question-which-one-correct-team.json", Method: "metadata", Evidence: "json-structured"},
		"markdown-note":      {Dest: "incident-response-runbook.md", Suggested: "incident-response-runbook.md", Method: "metadata", Evidence: "markdown-heading"},
		"message":            {Dest: "customer-onboarding-checklist.eml", Suggested: "customer-onboarding-checklist.eml", Method: "metadata", Evidence: "email-subject"},
		"random.txt":         {Dest: "unidentified-content.txt", Suggested: "unidentified-content.txt", Method: "metadata"},
		"recovered-doc":      {Dest: "monumental-construction-panama-canal.docx", Suggested: "monumental-construction-panama-canal.docx", Method: "metadata", Evidence: "office-heading"},
		"recovered-html":     {Dest: "basics-photosynthesis.html", Suggested: "basics-photosynthesis.html", Method: "metadata", Evidence: "html-title"},
		"recovered-musicxml": {Dest: "you-my-sunshine.musicxml", Suggested: "you-my-sunshine.musicxml", Method: "metadata", Evidence: "musicxml-work-title"},
		"recovered-xml.bin":  {Dest: "quarterly-safety-inspection-log.xml", Suggested: "quarterly-safety-inspection-log.xml", Method: "metadata", Evidence: "xml-title"},
		"recovered-pdf.bin":  {Dest: "some-things-only-actually-potentially.pdf", Suggested: "some-things-only-actually-potentially.pdf", Method: "metadata", Evidence: "pdf-first-text"},
		"unknown.dat":        {Dest: "customer-first-name-last-company-city-country-phone.csv", Suggested: "customer-first-name-last-company-city-country-phone.csv", Method: "metadata", Evidence: "csv-headers"},
	})
}

func TestClientCopyModeCopiesFilesWithoutMutatingSources(t *testing.T) {
	root := repoRoot(t)
	inputDir := filepath.Join(root, "testdata/recovered")
	outputDir := t.TempDir()

	runClient(t, root,
		"--strategy", "metadata-only",
		"--input", inputDir,
		"--output", outputDir,
		"--types", "csv,markdown,email",
	)

	for _, rel := range []string{
		"name-email-status.csv",
		"name-email-status-2.csv",
		"incident-response-runbook.md",
		"customer-onboarding-checklist.eml",
	} {
		if _, err := os.Stat(filepath.Join(outputDir, rel)); err != nil {
			t.Fatalf("expected copied file %s: %v", rel, err)
		}
	}

	for _, source := range []string{"customer-a.csv", "customer-b.csv", "markdown-note", "message"} {
		if _, err := os.Stat(filepath.Join(inputDir, source)); err != nil {
			t.Fatalf("source file was mutated or removed: %s: %v", source, err)
		}
	}
}

func TestClientCopyModeSkipsLowConfidenceFiles(t *testing.T) {
	root := repoRoot(t)
	outputDir := t.TempDir()
	reportPath := filepath.Join(t.TempDir(), "report.json")

	runClient(t, root,
		"--strategy", "metadata-only",
		"--input", filepath.Join(root, "testdata/recovered"),
		"--output", outputDir,
		"--types", "text,markdown",
		"--min-confidence-to-copy", "0.75",
		"--report", reportPath,
	)

	got := readReport(t, reportPath)
	bySource := entriesByBase(got)
	random := bySource["random.txt"]
	if !random.Skipped {
		t.Fatalf("random.txt should be skipped: %+v", random)
	}
	if random.ReviewStatus != "pending" {
		t.Fatalf("review status = %q, want pending", random.ReviewStatus)
	}
	if !strings.Contains(random.SkipReason, "below copy threshold") {
		t.Fatalf("skip reason = %q", random.SkipReason)
	}
	if _, err := os.Stat(random.DestinationPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("skipped file should not be copied, stat err=%v", err)
	}
	heading := bySource["markdown-note"]
	if heading.Skipped {
		t.Fatalf("markdown-note should not be skipped: %+v", heading)
	}
	if _, err := os.Stat(heading.DestinationPath); err != nil {
		t.Fatalf("expected markdown copy: %v", err)
	}
	if got.Summary.SkippedCount != 1 {
		t.Fatalf("summary skipped = %d, want 1", got.Summary.SkippedCount)
	}
	if got.Summary.CopiedCount != 1 {
		t.Fatalf("summary copied = %d, want 1", got.Summary.CopiedCount)
	}
}

func TestClientWritesReviewReport(t *testing.T) {
	root := repoRoot(t)
	outputDir := t.TempDir()
	reportPath := filepath.Join(t.TempDir(), "report.json")
	reviewPath := filepath.Join(t.TempDir(), "review.md")

	runClient(t, root,
		"--strategy", "metadata-only",
		"--input", filepath.Join(root, "testdata/recovered"),
		"--output", outputDir,
		"--types", "text,markdown",
		"--min-confidence-to-copy", "0.75",
		"--report", reportPath,
		"--review-report", reviewPath,
	)

	data, err := os.ReadFile(reviewPath)
	if err != nil {
		t.Fatal(err)
	}
	got := string(data)
	for _, want := range []string{"# File Rename Review", "random.txt", "pending"} {
		if !strings.Contains(got, want) {
			t.Fatalf("review report missing %q:\n%s", want, got)
		}
	}
}

func TestClientApplyReportCopiesPlannedFiles(t *testing.T) {
	root := repoRoot(t)
	outputDir := t.TempDir()
	reportPath := filepath.Join(t.TempDir(), "report.json")

	runClient(t, root,
		"--strategy", "metadata-only",
		"--dry-run",
		"--input", filepath.Join(root, "testdata/recovered"),
		"--output", outputDir,
		"--types", "csv",
		"--report", reportPath,
	)
	runClient(t, root, "--apply-report", reportPath)

	got := readReport(t, reportPath)
	for _, entry := range got.Entries {
		if _, err := os.Stat(entry.DestinationPath); err != nil {
			t.Fatalf("expected copied file %s: %v", entry.DestinationPath, err)
		}
	}
}

func TestApplyReportDryRunDoesNotCopy(t *testing.T) {
	root := repoRoot(t)
	outputDir := t.TempDir()
	reportPath := filepath.Join(t.TempDir(), "report.json")

	runClient(t, root,
		"--strategy", "metadata-only",
		"--dry-run",
		"--input", filepath.Join(root, "testdata/recovered"),
		"--output", outputDir,
		"--types", "csv",
		"--report", reportPath,
	)
	runClient(t, root, "--apply-report", reportPath, "--dry-run")

	got := readReport(t, reportPath)
	for _, entry := range got.Entries {
		if _, err := os.Stat(entry.DestinationPath); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("dry-run apply should not copy %s, stat err=%v", entry.DestinationPath, err)
		}
	}
}

func TestApplyReportReturnsMissingSourceError(t *testing.T) {
	reportPath := filepath.Join(t.TempDir(), "report.json")
	missing := filepath.Join(t.TempDir(), "missing.txt")
	if err := report.Write(reportPath, []report.Entry{{
		SourcePath:      missing,
		DestinationPath: filepath.Join(t.TempDir(), "out.txt"),
	}}); err != nil {
		t.Fatal(err)
	}

	err := applyReport(reportPath, false, false, "")
	if err == nil {
		t.Fatal("expected missing source error")
	}
	if !strings.Contains(err.Error(), "source file does not exist") || !strings.Contains(err.Error(), missing) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestApplyReportValidatesEmptyEntries(t *testing.T) {
	reportPath := filepath.Join(t.TempDir(), "report.json")
	source := filepath.Join(t.TempDir(), "source.txt")
	if err := os.WriteFile(source, []byte("source"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := report.Write(reportPath, []report.Entry{
		{},
		{SourcePath: "", DestinationPath: filepath.Join(t.TempDir(), "out.txt")},
		{SourcePath: source, DestinationPath: ""},
	}); err != nil {
		t.Fatal(err)
	}

	err := applyReport(reportPath, false, false, "")
	if err == nil {
		t.Fatal("expected empty source error")
	}
	if !strings.Contains(err.Error(), "empty source path") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestApplyReportSkipsReportSkippedEntries(t *testing.T) {
	source := filepath.Join(t.TempDir(), "source.txt")
	dest := filepath.Join(t.TempDir(), "dest.txt")
	reportPath := filepath.Join(t.TempDir(), "report.json")
	if err := os.WriteFile(source, []byte("source"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := report.Write(reportPath, []report.Entry{{
		SourcePath:      source,
		DestinationPath: dest,
		Skipped:         true,
		SkipReason:      "low confidence",
	}}); err != nil {
		t.Fatal(err)
	}

	if err := applyReport(reportPath, false, false, ""); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(dest); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("skipped report entry should not copy destination, stat err=%v", err)
	}
}

func TestApplyReportIncludesAcceptedSkippedEntries(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "source.txt")
	acceptedDest := filepath.Join(dir, "accepted.txt")
	pendingDest := filepath.Join(dir, "pending.txt")
	reportPath := filepath.Join(dir, "report.json")
	if err := os.WriteFile(source, []byte("source"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := report.Write(reportPath, []report.Entry{
		{
			SourcePath:      source,
			DestinationPath: acceptedDest,
			Skipped:         true,
			SkipReason:      "reviewed",
			ReviewStatus:    "accepted",
		},
		{
			SourcePath:      source,
			DestinationPath: pendingDest,
			Skipped:         true,
			SkipReason:      "not reviewed",
			ReviewStatus:    "pending",
		},
	}); err != nil {
		t.Fatal(err)
	}

	if err := applyReport(reportPath, false, true, ""); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(acceptedDest); err != nil {
		t.Fatalf("accepted skipped entry should copy: %v", err)
	}
	if _, err := os.Stat(pendingDest); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("pending skipped entry should not copy, stat err=%v", err)
	}
}

func TestApplyReportCanWriteReviewReport(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "source.txt")
	dest := filepath.Join(dir, "dest.txt")
	reportPath := filepath.Join(dir, "report.json")
	reviewPath := filepath.Join(dir, "review.md")
	if err := os.WriteFile(source, []byte("source"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := report.Write(reportPath, []report.Entry{{
		SourcePath:      source,
		DestinationPath: dest,
		Skipped:         true,
		ReviewStatus:    "pending",
	}}); err != nil {
		t.Fatal(err)
	}

	if err := applyReport(reportPath, true, false, reviewPath); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(reviewPath); err != nil {
		t.Fatalf("expected review report: %v", err)
	}
}

func TestAutoFallbackUsesFakeAIForLowConfidenceOnly(t *testing.T) {
	root := repoRoot(t)
	outputDir := t.TempDir()
	reportPath := filepath.Join(t.TempDir(), "report.json")
	fake := &fakeClient{filename: "ai-named-random"}
	withFakeAI(t, fake)

	if err := runApp([]string{
		"ai-file-renamer",
		"--strategy", "auto",
		"--dry-run",
		"--input", filepath.Join(root, "testdata/recovered"),
		"--output", outputDir,
		"--types", "text,markdown",
		"--report", reportPath,
	}); err != nil {
		t.Fatal(err)
	}

	got := readReport(t, reportPath)
	bySource := entriesByBase(got)
	assertSuggestedName(t, bySource, "random.txt", "ai-named-random.txt")
	assertSuggestedName(t, bySource, "markdown-note", "incident-response-runbook.md")
	if fake.evidenceCalls != 1 {
		t.Fatalf("fake evidence calls = %d, want 1", fake.evidenceCalls)
	}
	if fake.rawCalls != 0 {
		t.Fatalf("fake raw calls = %d, want 0", fake.rawCalls)
	}
	if !strings.Contains(fake.lastEvidence, "detected_type: text") {
		t.Fatalf("expected compact evidence, got %q", fake.lastEvidence)
	}
}

func TestMetadataOnlyDoesNotCreateAIClient(t *testing.T) {
	root := repoRoot(t)
	reportPath := filepath.Join(t.TempDir(), "report.json")
	created := false
	newAIClient = func(config.Config, bool, string) (ai.Client, error) {
		created = true
		return &fakeClient{filename: "unused"}, nil
	}
	t.Cleanup(func() { newAIClient = ai.NewClient })

	if err := runApp([]string{
		"ai-file-renamer",
		"--strategy", "metadata-only",
		"--dry-run",
		"--input", filepath.Join(root, "testdata/recovered"),
		"--types", "text,markdown",
		"--report", reportPath,
	}); err != nil {
		t.Fatal(err)
	}
	if created {
		t.Fatal("metadata-only should not create an AI client")
	}
}

func TestClientLargeFixtureSmoke(t *testing.T) {
	root := repoRoot(t)
	inputDir := t.TempDir()
	outputDir := t.TempDir()
	reportPath := filepath.Join(t.TempDir(), "report.json")

	for i := 0; i < 250; i++ {
		path := filepath.Join(inputDir, fmt.Sprintf("recovered-%03d.txt", i))
		content := fmt.Sprintf("# Batch Recovery File %03d\n\nOperational notes for generated fixture %03d.\n", i, i)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	runClient(t, root,
		"--strategy", "metadata-only",
		"--input", inputDir,
		"--output", outputDir,
		"--types", "text,markdown",
		"--report", reportPath,
	)

	got := readReport(t, reportPath)
	if len(got.Entries) != 250 {
		t.Fatalf("report entries = %d, want 250", len(got.Entries))
	}
	for _, entry := range got.Entries {
		if _, err := os.Stat(entry.DestinationPath); err != nil {
			t.Fatalf("expected copied file %s: %v", entry.DestinationPath, err)
		}
	}
}

func runClient(t *testing.T, root string, args ...string) {
	t.Helper()

	cmdArgs := append([]string{"run", "./cmd/client"}, args...)
	cmd := exec.Command("go", cmdArgs...)
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "GOCACHE="+filepath.Join(t.TempDir(), "go-cache"))

	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go %s failed: %v\n%s", strings.Join(cmdArgs, " "), err, out)
	}
}

func readReport(t *testing.T, path string) report.Report {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	var got report.Report
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	return got
}

func entriesByBase(got report.Report) map[string]report.Entry {
	entries := map[string]report.Entry{}
	for _, entry := range got.Entries {
		entries[filepath.Base(entry.SourcePath)] = entry
	}
	return entries
}

func assertSuggestedName(t *testing.T, entries map[string]report.Entry, source, want string) {
	t.Helper()

	entry, ok := entries[source]
	if !ok {
		t.Fatalf("missing source %s", source)
	}
	if entry.SuggestedName != want {
		t.Fatalf("%s suggested name = %q, want %q", source, entry.SuggestedName, want)
	}
}

type goldenEntry struct {
	Dest      string
	Suggested string
	Method    string
	Evidence  string
}

func assertGoldenReport(t *testing.T, got report.Report, want map[string]goldenEntry) {
	t.Helper()

	seen := map[string]goldenEntry{}
	for _, entry := range got.Entries {
		key := filepath.Base(entry.SourcePath)
		evidence := ""
		if len(entry.Evidence) > 0 {
			evidence = entry.Evidence[0]
		}
		seen[key] = goldenEntry{
			Dest:      filepath.Base(entry.DestinationPath),
			Suggested: entry.SuggestedName,
			Method:    entry.Method,
			Evidence:  evidence,
		}
	}

	if len(seen) != len(want) {
		t.Fatalf("golden entry count = %d, want %d; got keys %v", len(seen), len(want), sortedKeys(seen))
	}
	for source, expected := range want {
		actual, ok := seen[source]
		if !ok {
			t.Fatalf("missing golden source %s; got keys %v", source, sortedKeys(seen))
		}
		if expected.Dest == "" {
			actual.Dest = ""
		}
		if actual != expected {
			t.Fatalf("%s golden entry = %+v, want %+v", source, actual, expected)
		}
	}
}

func sortedKeys(entries map[string]goldenEntry) []string {
	keys := make([]string, 0, len(entries))
	for key := range entries {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func assertDestExt(t *testing.T, entries map[string]report.Entry, source, ext string) {
	t.Helper()

	entry, ok := entries[source]
	if !ok {
		t.Fatalf("missing source %s", source)
	}
	if filepath.Ext(entry.DestinationPath) != ext {
		t.Fatalf("%s destination = %s, want extension %s", source, entry.DestinationPath, ext)
	}
}

func repoRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not determine caller path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "../.."))
}

type fakeClient struct {
	filename      string
	rawCalls      int
	evidenceCalls int
	lastEvidence  string
}

func (f *fakeClient) SuggestFilename(string) (string, error) {
	f.rawCalls++
	return f.filename, nil
}

func (f *fakeClient) SuggestFilenameFromEvidence(evidence string) (string, error) {
	f.evidenceCalls++
	f.lastEvidence = evidence
	return f.filename, nil
}

func withFakeAI(t *testing.T, fake *fakeClient) {
	t.Helper()
	newAIClient = func(config.Config, bool, string) (ai.Client, error) {
		return fake, nil
	}
	t.Cleanup(func() { newAIClient = ai.NewClient })
}
