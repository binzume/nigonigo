package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/binzume/nigonigo"
)

func cmdRanking() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s ranking [genre] [term]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  genre: all, game, anime... \n")
		flag.PrintDefaults()
	}
	jsonFormat := flag.Bool("json", false, "json output")
	page := flag.Int("p", 1, "page")
	term := flag.String("t", "24h", "term(24h or hour)")

	// flag.Parse()
	flag.CommandLine.Parse(os.Args[2:])

	client := nigonigo.NewClient()

	genre := flag.Arg(0)
	if genre == "" {
		genre = "all"
	}

	result, err := client.GetVideoRanking(genre, *term, *page)
	if err != nil {
		log.Fatalf("Failed to request %v", err)
	}
	if *jsonFormat {
		jsonStr, _ := json.MarshalIndent(result, "", "   ")
		os.Stdout.Write(jsonStr)
	} else {
		for _, item := range result.Items {
			fmt.Printf("%v\t%v\t%v\t%v\n",
				item.ContentID,
				item.Title,
				item.RegisteredAt,
				item.Thumbnail.Url)
		}
	}
}
