package search

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"log"
	"math"
	"os"
	"sort"
	"strings"
	"unicode"
)

/*
Index: {docs, tMap:{term: TermFreq:{idf, tfMap:{doc1: tf1, doc2: tf2, ...}}}}
*/
type Index struct {
	TMap       map[string]TermFreq `json:"t_map"` // term map
	docs       []Document
	normalizer Normalizer
}

// key: Document name, value: normalized tf-idf
type TermFreq struct {
	Idf   float64            `json:"idf"`
	TfMap map[string]float64 `json:"tf_map"` // key: doc name, value: tf in doc
}

// DocCount returns the number of documents in the index.
func (idx Index) DocCount() int {
	return len(idx.docs)
}

// TermCount returns the number of unique terms in the index.
func (idx Index) TermCount() int {
	return len(idx.TMap)
}

// Return the total number of words in all documents.
func (idx Index) TotalWords() int {
	total := 0
	for _, doc := range idx.docs {
		total += doc.Length
	}
	return total
}

// Search returns an ordering of the documents based on the search terms
func (idx Index) Search(terms []string) ([]SearchResult, error) {
	var results []SearchResult
	for i := range idx.docs {
		doc := idx.docs[i]
		sr := idx.docScore(terms, &doc)
		if sr.Score > 0 {
			results = append(results, sr)
		}
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	return results, nil
}

// Loader is a function that returns documents given some options.
type Loader func(opts DocOpts) ([]Document, error)

// DefaultLoader loads documents from the filesystem using the provided options.
func DefaultLoader(opts DocOpts) ([]Document, error) {
	// load documents from the LoadPath directory
	// create new docs for each file in the directory using NewDoc
	files, err := os.ReadDir(opts.LoadPath)
	if err != nil {
		return []Document{}, err
	}

	var docs []Document
	for _, file := range files {
		info, err := file.Info()
		if err != nil {
			return []Document{}, err
		}
		if info.IsDir() {
			continue
		}
		doc, err := NewDoc(file, opts)
		if err != nil {
			return []Document{}, err
		}
		docs = append(docs, doc)
	}
	return docs, nil
}

// Normalizer converts a raw document string into a cleaned version before tokenization (e.g. lowercase, strip punctuation, etc.).
type Normalizer func(text string) string

// DefaultNormalizer lowercases and strips punctuation.
func DefaultNormalizer(s string) string {
	s = strings.ToLower(s)
	s = strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.IsSpace(r) {
			return r
		}
		return -1
	}, s)
	return s
}

// NewIndex creates a new search index from the documents loaded using the provided loader function.
func NewIndex(loader Loader, docOpts DocOpts) *Index {
	idx := &Index{
		normalizer: DefaultNormalizer,
	}
	idx.populate(loader, docOpts)
	idx.build()
	return idx
}

// populate loads documents into the index using the provided loader function
func (idx *Index) populate(loader Loader, docOpts DocOpts) {
	docs, err := loader(docOpts)
	if err != nil {
		log.Fatal(err)
	}
	idx.docs = docs
}

// LoadIndex loads the index from a gzipped file.
func LoadIndex(loader Loader, docOpts DocOpts) *Index {
	file, err := os.Open(docOpts.IndexPath)
	if err != nil {
		log.Fatalf("failed to open index file: %v", err)
	}
	defer file.Close()

	// Wrap with gzip reader
	gz, err := gzip.NewReader(file)
	if err != nil {
		log.Fatalf("failed to create gzip reader: %v", err)
	}
	defer gz.Close()

	data, err := io.ReadAll(gz)
	if err != nil {
		log.Fatalf("failed to read gzipped data: %v", err)
	}

	var idx Index
	if err := json.Unmarshal(data, &idx); err != nil {
		log.Fatalf("failed to unmarshal index: %v", err)
	}

	idx.populate(loader, docOpts)
	return &idx
}

// Save saves the index to a gzipped JSON file.
func (idx *Index) Save(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// Create a gzip writer for compression
	gz := gzip.NewWriter(file)
	defer gz.Close()

	enc := json.NewEncoder(gz)
	if err := enc.Encode(idx); err != nil {
		return err
	}

	return nil
}

// ngrams generates n-grams from a slice of words.
func ngrams(words []string, n int) []string {
	if len(words) < n {
		return words
	}
	ngrams := make([]string, len(words)-n+1)
	for i := 0; i < len(words)-n+1; i++ {
		ngram := strings.Join(words[i:i+n], " ")
		ngrams[i] = ngram
	}
	return ngrams
}

// buildNgrams builds bigrams and trigrams from the content and appends them to the original words.
func buildNgrams(content []string) []string {
	bigrams := ngrams(content, 2)
	trigrams := ngrams(content, 3)
	content = append(content, bigrams...)
	content = append(content, trigrams...)
	return content
}

// build the search index from the documents
func (idx *Index) build() {
	// build the term map
	idx.TMap = make(map[string]TermFreq)
	for _, doc := range idx.docs {
		text := idx.normalizer(doc.Content)
		words := buildNgrams(strings.Fields(text))
		for _, word := range words {
			if _, ok := idx.TMap[word]; !ok {
				idx.TMap[word] = TermFreq{TfMap: make(map[string]float64)}
			}
			idx.TMap[word].TfMap[doc.Name] += 1.0 / float64(doc.Length)
		}
	}

	// calculate the idf for each term
	for term, tf := range idx.TMap {
		tfreq := idx.TMap[term]
		tfreq.Idf = float64(len(idx.docs)) / float64(len(tf.TfMap)) // always >= 1
		idx.TMap[term] = tfreq
	}
}

func (idx *Index) tfNorm(term string) float64 {
	normSum := 0.0
	tfreq := idx.TMap[term]
	for _, tf := range idx.TMap[term].TfMap {
		normSum += (math.Log(tfreq.Idf) * tf) * (math.Log(tfreq.Idf) * tf)
	}
	if normSum == 0 {
		return 1.0
	}
	return math.Sqrt(normSum)
}

func (idx *Index) tf(term, docName string) float64 {
	return idx.TMap[term].TfMap[docName]
}

func (idx *Index) idf(term string) float64 {
	if idx.TMap[term].Idf == 0 {
		return 1.0
	}
	return idx.TMap[term].Idf
}

func (idx *Index) tfLogIdf(term, docName string) float64 {
	return idx.tf(term, docName) * math.Log(idx.idf(term)) / idx.tfNorm(term)
}

// docScore calculates the score of a document based on the geometric mean of search terms scores
func (idx *Index) docScore(terms []string, doc *Document) SearchResult {
	score := 1.0
	nfound := 0.0
	for _, term := range buildNgrams(terms) {
		termScore := idx.tfLogIdf(strings.ToLower(term), doc.Name)
		if termScore > 0 {
			score *= termScore
			nfound++
		}
	}

	var docScore float64
	if nfound == 0 {
		docScore = 0
	} else {
		docScore = math.Pow(score, 1/nfound)
	}
	return SearchResult{Document: doc, Score: docScore}
}
