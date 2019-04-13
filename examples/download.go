package main

import (
	"context"
	"log"
	"os"

	"github.com/binzume/nigonigo"
)

func main() {
	client := nigonigo.NewClient()

	contentID := "sm9"
	if len(os.Args) > 1 {
		contentID = os.Args[1]
	}
	if len(os.Args) > 2 {
		client.LoginWithJsonFile(os.Args[2])
	}
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
	log.Println("ok")
}
