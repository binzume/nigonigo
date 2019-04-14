package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/binzume/nigonigo"
)

func main() {
	tag := flag.Bool("t", false, "tag search")
	offset := flag.Int("offset", 0, "offset")
	limit := flag.Int("limit", 100, "limit")
	flag.Parse()
	if flag.NArg() == 0 {
		log.Println("usage: go run search.go hoge")
		flag.Usage()
		return
	}
	q := flag.Arg(0)

	client := nigonigo.NewClient()

	var result *nigonigo.SearchResult
	var err error
	if *tag {
		result, err = client.SearchByTag(q, *offset, *limit)
	} else {
		result, err = client.SearchByKeyword(q, *offset, *limit)
	}
	if err != nil {
		log.Fatalf("Failed to request %v", err)
	}

	for _, item := range result.Items {
		fmt.Println(item.ContentID, item.Title)
	}
}
