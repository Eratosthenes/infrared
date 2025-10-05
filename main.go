package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	ir "example.com/infrared/search"
)

func main() {
	opts := ir.DocOpts{
		IndexPath:   "./example/index.gz",
		LoadPath:    "./example/docs",
		LoadContent: true,
	}

	// build the index
	start := time.Now()
	index := ir.NewIndex(ir.DefaultLoader, opts)
	elapsed := time.Since(start).Milliseconds()
	fmt.Printf("Index built in %d milliseconds.\n", elapsed)

	// save the index and print its size
	if err := index.Save(opts.IndexPath); err != nil {
		log.Fatalf("failed to save index: %v", err)
	}
	info, err := os.Stat(opts.IndexPath)
	if err != nil {
		log.Fatalf("failed to stat index file: %v", err)
	}
	sizeKB := float64(info.Size()) / 1024.0
	// print the size of the index file
	fmt.Printf("The index file is %.0f KB.\n\n", sizeKB)

	// clean up the index file
	if err := os.Remove(opts.IndexPath); err != nil {
		log.Fatalf("failed to remove index file: %v", err)
	}

	// print index metrics
	fmt.Printf("Documents: %d\n", index.DocCount())
	fmt.Printf("Indexed ngrams: %d\n", index.TermCount())
	fmt.Printf("Total words in corpus: %d\n", index.TotalWords())
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
