package main

import (
	"fmt"

	ir "example.com/infrared/search"
)

func main() {
	opts := ir.DocOpts{
		LoadPath:    "./docs",
		LoadContent: true,
		LenPreview:  200,
	}

	index := ir.NewIndex(ir.LoadDocs, opts) // build index with previews
	results, _ := index.Search([]string{"hello", "world"})
	for _, doc := range results {
		fmt.Println(doc.Name, doc.Score, doc.Preview)
	}
}
