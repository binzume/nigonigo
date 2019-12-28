package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/binzume/nigonigo"
)

func cmdSearch() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s search [options] [query]\n", os.Args[0])
		flag.PrintDefaults()
	}
	tag := flag.String("t", "", "tag")
	offset := flag.Int("offset", 0, "offset")
	limit := flag.Int("limit", 100, "limit")
	jsonFormat := flag.Bool("json", false, "json output")
	sortOrder := flag.String("sort", "-startTime", "sort order")
	filterGenre := flag.String("genre", "", "filter by genre")
	filterUserID := flag.String("user", "", "filter by userId")
	filterChannelID := flag.String("ch", "", "filter by channelId")
	filterViewCount := flag.String("viewCount", "0", "filter by viewCounter")
	// flag.Parse()
	flag.CommandLine.Parse(os.Args[2:])

	if flag.NArg() == 0 && *tag == "" {
		flag.Usage()
		return
	}
	q := flag.Arg(0)

	client := nigonigo.NewClient()

	var filters []nigonigo.SearchFilter
	if *filterGenre != "" {
		filters = append(filters, nigonigo.EqualFilter(nigonigo.SearchFieldGenre, *filterGenre))
	}
	if *filterUserID != "" {
		filters = append(filters, nigonigo.EqualFilter(nigonigo.SearchFieldUserID, *filterUserID))
	}
	if *filterChannelID != "" {
		filters = append(filters, nigonigo.EqualFilter(nigonigo.SearchFieldChannelID, *filterChannelID))
	}
	if *filterViewCount != "0" {
		filters = append(filters, nigonigo.RangeFilter(nigonigo.SearchFieldViewCounter, *filterViewCount, "", true))
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

	result, err := client.SearchVideo(q, searchField, nil, *sortOrder, *offset, *limit, filter)
	if err != nil {
		log.Fatalf("Failed to request %v", err)
	}

	if *jsonFormat {
		jsonStr, _ := json.MarshalIndent(result.Items, "", "   ")
		os.Stdout.Write(jsonStr)
	} else {
		for _, item := range result.Items {
			fmt.Printf("%v\t%v\t%v\n", item.ContentID, item.StartTime, item.Title)
		}
	}
}
