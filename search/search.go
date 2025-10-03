package search

import (
	"encoding/json"
	"log"
	"math"
	"os"
	"sort"
	"strings"
)

/*
Index: {docs, tMap:{term: TermFreq:{idf, tfMap:{doc1: tf1, doc2: tf2, ...}}}}
*/
type Index struct {
	TMap map[string]TermFreq `json:"t_map"` // term map
	docs []Document
}

// key: Document name, value: normalized tf-idf
type TermFreq struct {
	Idf   float64            `json:"idf"`
	TfMap map[string]float64 `json:"tf_map"` // key: doc name, value: tf in doc
}

// Search returns an ordering of the documents based on the search terms
func (idx Index) Search(terms []string) ([]SearchResult, error) {
	docs := idx.docs
	var results []SearchResult
	for _, doc := range docs {
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

func LoadDocs(opts DocOpts) ([]Document, error) {
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

// buildIndex builds the index from the documents in the docs directory
func NewIndex(loader Loader, docOpts DocOpts) *Index {
	idx := &Index{}
	idx.populate(loader, docOpts)
	idx.build()
	return idx
}

func (idx *Index) populate(loader Loader, docOpts DocOpts) {
	docs, err := loader(docOpts)
	if err != nil {
		log.Fatal(err)
	}
	idx.docs = docs
}

func LoadIndex(loader Loader, docOpts DocOpts) *Index {
	// Read the JSON data from the file
	jsonData, err := os.ReadFile("index.json")
	if err != nil {
		log.Fatal(err)
	}

	// Unmarshal the JSON data into the Index object
	idx := Index{}
	err = json.Unmarshal(jsonData, &idx)
	if err != nil {
		log.Fatal(err)
	}

	idx.populate(loader, docOpts)
	return &idx
}

func (idx *Index) Save() error {
	// Marshal the Index object into JSON
	jsonData, err := json.Marshal(idx)
	if err != nil {
		return err
	}

	// Write the JSON data to a file
	err = os.WriteFile("index.json", jsonData, 0644)
	if err != nil {
		return err
	}

	return nil
}

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
		words := buildNgrams(strings.Fields(doc.Content))
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
// func (idx *Index) docScore(terms []string, docName string) float64 {
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
