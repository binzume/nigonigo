package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/binzume/nigonigo"
)

func cmdSearch() {
	channelID := flag.String("ch", "", "channelId")
	userID := flag.String("user", "", "userId")
	tag := flag.String("t", "", "tag")
	offset := flag.Int("offset", 0, "offset")
	limit := flag.Int("limit", 100, "limit")
	// flag.Parse()
	flag.CommandLine.Parse(os.Args[2:])

	if flag.NArg() == 0 && *tag == "" {
		log.Println("usage: go run search.go hoge")
		flag.Usage()
		return
	}
	q := flag.Arg(0)

	client := nigonigo.NewClient()

	var filters []nigonigo.SearchFilter
	if *channelID != "" {
		filters = append(filters, nigonigo.EqualFilter(nigonigo.SearchFieldChannelID, *channelID))
	}
	if *userID != "" {
		filters = append(filters, nigonigo.EqualFilter(nigonigo.SearchFieldUserID, *userID))
	}
	var filter nigonigo.SearchFilter
	if len(filters) == 1 {
		filter = filters[0]
	} else if len(filters) > 1 {
		filter = nigonigo.AndFilter(filters)
	}

	var searchField []nigonigo.SearchField
	if *tag != "" {
		searchField = []nigonigo.SearchField{"tagsExact,categoryTags"}
		q = *tag
	} else {
		searchField = []nigonigo.SearchField{"description,title"}
	}

	result, err := client.SearchVideo(q, searchField, nil, "-startTime", *offset, *limit, filter)
	if err != nil {
		log.Fatalf("Failed to request %v", err)
	}

	for _, item := range result.Items {
		fmt.Printf("%v\t%v\t%v\n", item.ContentID, item.StartTime, item.Title)
	}
}
