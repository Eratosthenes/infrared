package search

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestSearchEngine(t *testing.T) {
	opts := DocOpts{
		LoadPath:    "../example/docs",
		LoadContent: true,
	}

	index := NewIndex(DefaultLoader, opts)
	if index.DocCount() == 0 {
		t.Fatalf("expected >0 documents, got %d", index.DocCount())
	}

	tests := []struct {
		query    string
		expected string
	}{
		{"moral law", "civil_disobedience.txt"},
		{"human nature", "self_reliance.txt"},
		{"use of language", "politics_and_the_english_language.txt"},
		{"land", "how_much_land.txt"},
	}

	for _, tt := range tests {
		results, err := index.Search(strings.Fields(tt.query), SearchOpts{Limit: 5})
		if err != nil {
			t.Fatalf("search error for %q: %v", tt.query, err)
		}
		if len(results) == 0 {
			t.Errorf("no results for query %q", tt.query)
			continue
		}

		got := results[0].Name
		if got != tt.expected {
			t.Errorf("query %q: expected top result %q, got %q", tt.query, tt.expected, got)
		}
	}
}

func TestNormalizationConsistency(t *testing.T) {
	opts := DocOpts{
		LoadPath:    "../example/docs",
		LoadContent: true,
	}
	index := NewIndex(DefaultLoader, opts)

	// Ensure normalization produces comparable scores
	sopts := SearchOpts{Limit: 5}
	r1, _ := index.Search(strings.Fields("freedom and law"), sopts)
	r2, _ := index.Search(strings.Fields("moral law"), sopts)

	if len(r1) == 0 || len(r2) == 0 {
		t.Skip("not enough results for comparison")
	}

	if r1[0].Score < 0 || r1[0].Score > 1 {
		t.Errorf("expected normalized score in [0,1], got %.3f", r1[0].Score)
	}
	if r2[0].Score < 0 || r2[0].Score > 1 {
		t.Errorf("expected normalized score in [0,1], got %.3f", r2[0].Score)
	}
}

func TestSaveLoadSearch(t *testing.T) {
	opts := DocOpts{
		IndexPath:   "test_index.json",
		LoadPath:    "../example/docs",
		LoadContent: true,
	}

	// --- Build index
	idx := NewIndex(DefaultLoader, opts)
	if idx.DocCount() == 0 {
		t.Fatal("expected non-empty index")
	}

	// --- Save to a temporary file
	tmpFile := "test_index.json"
	defer os.Remove(tmpFile)

	if err := idx.Save(tmpFile); err != nil {
		t.Fatalf("failed to save index: %v", err)
	}

	// --- Load from disk
	loaded := LoadIndex(DefaultLoader, opts)
	if loaded.DocCount() != idx.DocCount() {
		t.Errorf("doc count mismatch: got %d, want %d", loaded.DocCount(), idx.DocCount())
	}
	if len(loaded.TMap) != len(idx.TMap) {
		t.Errorf("term map size mismatch: got %d, want %d", len(loaded.TMap), len(idx.TMap))
	}

	// --- Run a sample query
	sopts := SearchOpts{Limit: 5}
	results, err := loaded.Search([]string{"moral", "law"}, sopts)
	if err != nil {
		t.Fatalf("search on loaded index failed: %v", err)
	}
	if len(results) == 0 {
		t.Fatalf("expected results from loaded index, got 0")
	}

	// --- Verify top result stability
	top := results[0].Name
	if top != "civil_disobedience.txt" {
		t.Errorf("unexpected top result after reload: got %q, want %q", top, "civil_disobedience.txt")
	}
}

func BenchmarkBuildIndex(b *testing.B) {
	opts := DocOpts{
		LoadPath:    "../example/docs",
		LoadContent: true,
	}

	for i := 0; i < b.N; i++ {
		start := time.Now()
		NewIndex(DefaultLoader, opts)
		elapsed := time.Since(start)
		b.ReportMetric(float64(elapsed.Milliseconds()), "ms/index")
	}
}

func BenchmarkSearch(b *testing.B) {
	opts := DocOpts{
		LoadPath:    "../example/docs",
		LoadContent: true,
	}
	index := NewIndex(DefaultLoader, opts)

	queries := [][]string{
		{"moral", "law"},
		{"human", "nature"},
		{"use", "of", "language"},
		{"freedom", "and", "law"},
		{"land"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q := queries[i%len(queries)]
		results, _ := index.Search(q, SearchOpts{Limit: 5})
		if len(results) == 0 {
			b.Fatalf("no results for %v", q)
		}
	}
}

func BenchmarkIndexSize(b *testing.B) {
	opts := DocOpts{
		LoadPath:    "../example/docs",
		LoadContent: true,
		Compressed:  true,
	}
	index := NewIndex(DefaultLoader, opts)

	tmpfile := "bench_index.json.gz"
	defer os.Remove(tmpfile)

	start := time.Now()
	if err := index.Save(tmpfile); err != nil {
		b.Fatalf("failed to save index: %v", err)
	}
	elapsed := time.Since(start)

	info, err := os.Stat(tmpfile)
	if err != nil {
		b.Fatalf("failed to stat index file: %v", err)
	}

	sizeBytes := float64(info.Size())
	sizeKB := sizeBytes / 1024.0
	totalTerms := float64(index.TotalWords())
	bytesPerTerm := sizeBytes / totalTerms

	b.ReportMetric(sizeKB, "KB")
	b.ReportMetric(bytesPerTerm, "B/term")
	b.ReportMetric(float64(elapsed.Milliseconds()), "ms/save")
}
