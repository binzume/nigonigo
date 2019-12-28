package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
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

func downloadThumbnail(url, outpath string) (err error) {
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	out, err := os.Create(outpath)
	if err != nil {
		return
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return
}

func download(client *nigonigo.Client, contentID string, saveThumbnail bool) {
	video, err := client.GetVideoData(contentID)
	if err != nil {
		log.Fatalf("Failed to get video info: %v", err)
	}

	if saveThumbnail {
		err = downloadThumbnail(video.Video.ThumbnailURL, contentID+".jpg")
		if err != nil {
			log.Printf("Failed to get thumbnail: %v", err)
		}
	}

	session, err := client.CreateDMCSessionByVideoData(video)
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

func cmdDownload() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s download [options] contentId\n", os.Args[0])
		flag.PrintDefaults()
	}
	id := flag.String("i", "", "mail address")
	password := flag.String("p", "", "password")
	accountFile := flag.String("a", "account.json", "account.json")
	sessionFile := flag.String("s", "session.json", "session.json")
	saveThumbnail := flag.Bool("t", false, "save thumbnail")
	// flag.Parse()
	flag.CommandLine.Parse(os.Args[2:])

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
		err := client.Login(*id, *password)
		if err != nil {
			loginerr = err
		}
	}
	if loginerr != nil && client.Session == nil {
		log.Fatalf("login failed: %v", loginerr)
	}

	for _, contentID := range flag.Args() {
		download(client, contentID, *saveThumbnail)
	}
}
