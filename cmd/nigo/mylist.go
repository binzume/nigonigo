package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/binzume/nigonigo"
)

func cmdMylist() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s mylist [options] [mylistID|default]\n", os.Args[0])
		flag.PrintDefaults()
	}
	sessionFile := flag.String("s", defaultSessionFilePath, "session.json")
	jsonFormat := flag.Bool("json", false, "json output")
	// flag.Parse()
	flag.CommandLine.Parse(os.Args[2:])

	client := nigonigo.NewClient()
	err := authLogin(client, *sessionFile, "", "", "")
	if err != nil {
		log.Fatalf("Failed to login: %v", err)
	}

	if flag.Arg(0) == "" {
		result, err := client.GetMyLists()
		if err != nil {
			log.Fatalf("Failed to request %v", err)
		}
		if *jsonFormat {
			jsonStr, _ := json.MarshalIndent(result, "", "   ")
			os.Stdout.Write(jsonStr)
		} else {
			for _, item := range result {
				fmt.Printf("%v\t%v\t%v\t%v\n",
					item.ID,
					item.Name,
					item.CreatedAt,
					item.IsPublic)
			}
		}
	} else {
		result, err := client.GetMyList(flag.Arg(0))
		if err != nil {
			log.Fatalf("Failed to request %v", err)
		}
		if *jsonFormat {
			jsonStr, _ := json.MarshalIndent(result, "", "   ")
			os.Stdout.Write(jsonStr)
		} else {
			for _, item := range result.Items {
				fmt.Printf("%v\t%v\t%v\t%v\n",
					item.ItemID,
					item.Data.ContentID,
					item.Data.Title,
					item.AddedAt)
			}
		}
	}

}
