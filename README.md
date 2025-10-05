# InfraRed 🔎

[![Go Reference](https://pkg.go.dev/badge/github.com/Eratosthenes/infrared.svg)](https://pkg.go.dev/github.com/Eratosthenes/infrared)
[![Go Report Card](https://goreportcard.com/badge/github.com/Eratosthenes/infrared)](https://goreportcard.com/report/github.com/Eratosthenes/infrared)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

_A minimal, high-performance full-text search engine written in Go._

InfraRed is a lightweight information retrieval (IR) search tool that builds a TF–IDF-based index over plain-text documents (such as essays, notes, or code), supports n-gram tokenization, and returns ranked results in microseconds.

Unlike most traditional search engines, InfraRed produces globally-normalized scores that are directly comparable across different queries.

Engines like Lucene or Elasticsearch (BM25) normalize scores only within a single query—so a document scoring `2.0` for “moral law” and `0.5` for “freedom and law” doesn’t mean it’s four times more relevant to the first. Those values exist on different, query-dependent scales.

InfraRed normalizes each term’s contribution and combines them geometrically using an energy-minimizing formula, so every result falls on a consistent `[0, 1]` scale. A document scoring 0.8 for one query and 0.8 for another truly reflects equal relevance—making the engine useful not just for ranking results, but also for comparing conceptual similarity across searches.

---

## Features

- **Lightweight** — single binary, no external dependencies  
- **Fast** — microsecond-scale query times on small corpora  
- **Flexible** — pluggable loader and text normalizer functions  
- **Interpretable** — TF-IDF scoring with consistent normalization across queries

---

### ⚡ Performance

InfraRed builds its index in about 50-60 ms for four medium-length essays (~31,000 words total) and saves it as a 347 KB gzipped JSON file—using roughly 12 bytes per word in the corpus.

Search latency for these documents is in the range of 7–50 µs per query, returning ranked, normalized results.

InfraRed automatically prunes overly common terms based on corpus size: for small datasets it keeps all words, and for very large ones it limits terms that appear in more than about 5 percent of documents. This ensures the index stays compact and that search results remain discriminative even at scale.

---

## Example Usage

```text
$ go run main.go
Index built in 52 milliseconds.
The index file is 347 KB.

Documents: 4
Indexed ngrams: 55873
Total words in corpus: 31096
-------------------------
Search: [moral law]
civil_disobedience.txt                   (Score: 0.759)
self_reliance.txt                        (Score: 0.325)
how_much_land.txt                        (Score: 0.109)

Search completed in 26 microseconds.
-------------------------
Search: [human nature]
self_reliance.txt                        (Score: 0.482)
politics_and_the_english_language.txt    (Score: 0.267)
civil_disobedience.txt                   (Score: 0.218)

Search completed in 9 microseconds.
-------------------------
Search: [use of language]
politics_and_the_english_language.txt    (Score: 1.000)

Search completed in 11 microseconds.
-------------------------
Search: [freedom and law]
self_reliance.txt                        (Score: 0.806)
politics_and_the_english_language.txt    (Score: 0.756)
civil_disobedience.txt                   (Score: 0.617)
how_much_land.txt                        (Score: 0.109)

Search completed in 26 microseconds.
-------------------------
Search: [land]
how_much_land.txt                        (Score: 1.000)
civil_disobedience.txt                   (Score: 0.009)

Search completed in 19 microseconds.
-------------------------
```

---

### Scoring Methodology

InfraRed uses a TF-IDF approach with a few important twists for stability and interpretability.

Each term is weighted by how often it appears in a document (TF—term frequency) and how rare it is across all documents (IDF—inverse document frequency). Common words like _and_ or _the_ therefore carry almost no weight, while distinctive terms contribute strongly.

InfraRed also applies an L₂ normalization step that balances each term’s influence across the corpus.  This ensures that every term contributes proportionally to how informative it is.

When you search for multiple terms, InfraRed computes a relevance score for each term and then combines all non-zero scores using a weighted geometric mean. This rewards documents that match more of the query terms while still giving partial credit to those that contain only some of them.

The result is a compact, fast, and interpretable relevance model that produces rankings that "feel right" even on small text collections.