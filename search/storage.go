package search

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"log"
	"os"
	"strings"
	"unicode"
)

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
		compressed: docOpts.Compressed,
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

type indexLoader func(loader Loader, docOpts DocOpts) *Index

func jsonLoader(loader Loader, docOpts DocOpts) *Index {
	file, err := os.Open(docOpts.IndexPath)
	if err != nil {
		log.Fatalf("failed to open index file: %v", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		log.Fatalf("failed to read index file: %v", err)
	}

	var idx Index
	if err := json.Unmarshal(data, &idx); err != nil {
		log.Fatalf("failed to unmarshal index: %v", err)
	}

	idx.populate(loader, docOpts)
	return &idx
}

// gzipLoader loads the index from a gzipped file.
func gzipLoader(loader Loader, docOpts DocOpts) *Index {
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

func LoadIndex(loader Loader, opts DocOpts) *Index {
	var il indexLoader
	if opts.Compressed {
		il = gzipLoader
	} else {
		il = jsonLoader
	}
	return il(loader, opts)
}

// Save saves the index to a file.
func (idx *Index) Save(path string) error {
	var is indexSaver
	if idx.compressed {
		is = gzipSaver
	} else {
		is = jsonSaver
	}
	return is(idx, path)
}

type indexSaver func(idx *Index, path string) error

// jsonSaver saves the index to a JSON file.
func jsonSaver(idx *Index, path string) error {
	// Marshal the Index object into JSON
	jsonData, err := json.Marshal(idx)
	if err != nil {
		return err
	}

	// Write the JSON data to a file
	err = os.WriteFile(path, jsonData, 0644)
	if err != nil {
		return err
	}

	return nil
}

// gzipSaver saves the index to a gzipped JSON file.
func gzipSaver(idx *Index, path string) error {
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
