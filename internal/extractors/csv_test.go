package extractors

import (
	"path/filepath"
	"testing"
)

func TestCSVExtractorUsesHeadersFromWrongExtension(t *testing.T) {
	path := filepath.Join("..", "..", "testdata", "recovered", "unknown.dat")

	info, err := csvExtractor{}.ExtractInfo(path)
	if err != nil {
		t.Fatal(err)
	}

	if !hasSample(info, "csv-headers", "Index Customer Id First Name Last Name Company City Country Phone 1 Phone 2 Email Subscription Date Website") {
		t.Fatalf("missing csv header sample: %+v", info.TextSamples)
	}
}
