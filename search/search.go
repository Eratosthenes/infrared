package search

import (
	"math"
	"sort"
	"strings"
)

/*
Index: {docs, tMap:{term: TermFreq:{idf, tfMap:{doc1: tf1, doc2: tf2, ...}}}}
*/
type Index struct {
	TMap       map[string]TermFreq `json:"t_map"` // term map
	docs       []Document
	normalizer Normalizer
	compressed bool
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

		if 1/tfreq.Idf >= idx.maxThreshold() {
			delete(idx.TMap, term)
		}
	}
}

// maxThreshold returns the maximum threshold for a term to be included in the index
func (idx Index) maxThreshold() float64 {
	docCount := math.Max(float64(idx.DocCount()), 10)
	f := 1 / math.Sqrt(docCount/10)
	if f < 0.05 {
		f = 0.05
	}
	return f
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

// docScore calculates the score of a document based on the weighted geometric mean of search terms scores
func (idx *Index) docScore(terms []string, doc *Document) SearchResult {
	weightedSum := 0.0
	weightTotal := 0.0
	for _, term := range buildNgrams(terms) {
		termScore := idx.tfLogIdf(strings.ToLower(term), doc.Name)
		if termScore > 0 {
			w := math.Log(idx.idf(term))
			weightedSum += w * math.Log(termScore)
			weightTotal += w
		}
	}

	var docScore float64
	if weightTotal == 0 {
		docScore = 0
	} else {
		docScore = math.Exp(weightedSum / weightTotal)
	}
	return SearchResult{Document: doc, Score: docScore}
}
