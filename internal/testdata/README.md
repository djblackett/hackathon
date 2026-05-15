# Test Data Notes

Keep fixtures small and license-clear. The `recovername` tests should prefer tiny synthetic files under the repository's existing `testdata/` tree unless a real-world fixture is needed to cover a parser edge case.

Good fixture categories:

- extensionless PDF, DOCX, PNG/JPEG, CSV, JSON, HTML, XML, Markdown, and plain text files
- random or damaged files that should produce low-confidence names
- duplicate files or duplicate metadata that should trigger deterministic conflict suffixes
- optional-tool fixtures only when the tool is not required for the normal test suite

Do not vendor large corpora. If using samples from external corpora, document the source and license next to the fixture subset.
