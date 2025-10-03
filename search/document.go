package search

import (
	"io/fs"
	"os"
	"strings"
)

type DocOpts struct {
	LoadPath    string
	LoadContent bool
	LenPreview  int
}

type Document struct {
	Name    string `json:"name"`
	Date    string `json:"date"`
	Preview string `json:"preview"` // first N characters, using ellipsis if truncated
	Length  int    // number of words in the document
	Content string // full content, lowercase
}

type SearchResult struct {
	*Document
	Score float64
}

type MakeDoc func(file fs.DirEntry, opts DocOpts) (Document, error)

func NewDoc(file fs.DirEntry, opts DocOpts) (Document, error) {
	// create a new Document from the file
	var content string
	if opts.LoadContent {
		data, err := os.ReadFile(opts.LoadPath + "/" + file.Name())
		if err != nil {
			return Document{}, err
		}
		content = string(data)
	}

	preview := content
	if len(content) > opts.LenPreview {
		preview = content[:opts.LenPreview]
	}
	preview += "..."

	info, err := file.Info()
	if err != nil {
		return Document{}, err
	}

	doc := Document{
		Name:    file.Name(),
		Date:    info.ModTime().String(),
		Preview: preview,
		Length:  len(strings.Fields(content)),
		Content: content,
	}
	return doc, nil
}
