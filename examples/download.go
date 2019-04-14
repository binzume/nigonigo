package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/binzume/nigonigo"
)

func download(client *nigonigo.Client, contentID string) {
	session, err := client.CreateDMCSessionById(contentID)
	if err != nil {
		log.Fatalf("Failed to create session: %v", err)
	}

	out, _ := os.Create(contentID + "." + session.FileExtension())
	defer out.Close()
	err = client.Download(context.Background(), session, out)
	if err != nil {
		log.Fatalf("Failed to download: %v", err)
	}
}

func main() {
	id := flag.String("i", "", "mail address")
	password := flag.String("p", "", "password")
	accountFile := flag.String("a", "", "account.json")
	sessionFile := flag.String("s", "", "session.json")
	flag.Parse()
	if flag.NArg() == 0 {
		flag.Usage()
		return
	}

	client := nigonigo.NewClient()
	var loginerr error
	if *sessionFile != "" {
		err := client.LoadLoginSession(*sessionFile)
		if err != nil {
			loginerr = err
		}
	}
	if client.Session == nil && *accountFile != "" {
		err := client.LoginWithJsonFile(*accountFile)
		if err != nil {
			loginerr = err
		}
	}
	if client.Session == nil && *id != "" {
		err := client.LoginWithPassword(*id, *password)
		if err != nil {
			loginerr = err
		}
	}
	if loginerr != nil && client.Session == nil {
		log.Fatalf("login failed: %v", loginerr)
	}

	for _, contentID := range flag.Args() {
		download(client, contentID)
	}
	log.Println("ok")
}
