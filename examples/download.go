package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/binzume/nigonigo"
)

type ByteCounter struct {
	Count int64
}

func (w *ByteCounter) Write(p []byte) (int, error) {
	if w.Count != 0 {
		fmt.Fprint(os.Stderr, strings.Repeat("\b", 80))
	}
	w.Count += int64(len(p))
	fmt.Fprintf(os.Stderr, "Download %v MiB. ", w.Count/1024/1024)
	return len(p), nil
}

func download(client *nigonigo.Client, contentID string) {
	session, err := client.CreateDMCSessionById(contentID)
	if err != nil {
		log.Fatalf("Failed to create session: %v", err)
	}

	log.Printf("Start download %v", contentID)
	out, _ := os.Create(contentID + "." + session.FileExtension())
	defer out.Close()
	err = client.Download(context.Background(), session, io.MultiWriter(&ByteCounter{}, out))
	if err != nil {
		log.Fatalf("Failed to download: %v", err)
	}
	log.Println("ok")
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
}
