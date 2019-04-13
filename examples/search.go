package main

import (
	"fmt"
	"log"
	"os"

	"github.com/binzume/nigonigo"
)

func main() {
	client := nigonigo.NewClient()

	if len(os.Args) < 1 {
		log.Fatalf("usage: go run search.go hoge")
	}

	result, err := client.SearchByKeyword(os.Args[1], 0, 100)
	if err != nil {
		log.Fatalf("Failed to request %v", err)
	}

	for _, item := range result.Items {
		fmt.Println(item.ContentID, item.Title)
	}
}
