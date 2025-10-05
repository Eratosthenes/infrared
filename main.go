package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	ir "example.com/infrared/search"
)

func main() {
	opts := ir.DocOpts{
		LoadPath:    "./example/docs",
		LoadContent: true,
	}

	// build the index

	start := time.Now()
	index := ir.NewIndex(ir.DefaultLoader, opts)
	elapsed := time.Since(start).Milliseconds()
	fmt.Printf("Index built in %d milliseconds.\n\n", elapsed)

	// print index metrics
	fmt.Printf("Documents: %d\n", index.DocCount())
	fmt.Printf("Indexed ngrams: %d\n", index.TermCount())
	fmt.Printf("Total terms in all documents: %d\n", index.TotalTerms())
	fmt.Println("-------------------------")

	searchAndPrint := func(s string, index *ir.Index) {
		terms := strings.Fields(s)

		// perform the search
		fmt.Println("Search:", terms)

		// time the search
		start := time.Now()
		results, err := index.Search(terms)
		if err != nil {
			log.Fatal(err)
		}
		elapsed := time.Since(start).Microseconds()

		// print the results
		for _, doc := range results {
			fmt.Printf("%-40s (Score: %.3f)\n", doc.Name, doc.Score)
		}
		fmt.Printf("\nSearch completed in %d microseconds.\n", elapsed)
		fmt.Println("-------------------------")
	}

	searchAndPrint("moral law", index)
	searchAndPrint("human nature", index)
	searchAndPrint("use of language", index)
	searchAndPrint("freedom and law", index)
	searchAndPrint("land", index)
}
