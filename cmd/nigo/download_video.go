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
	"sync"

	"github.com/binzume/nigonigo"
)

type ByteCounter struct {
	Count int64
}

func join(sep string, values ...any) string {
	strs := make([]string, len(values))
	for i, v := range values {
		strs[i] = fmt.Sprint(v)
	}
	return strings.Join(strs, sep)
}

func (w *ByteCounter) Write(p []byte) (int, error) {
	if w.Count != 0 {
		fmt.Fprint(os.Stderr, "\r")
	}
	w.Count += int64(len(p))
	fmt.Fprintf(os.Stderr, "Download %v MiB", w.Count/1024/1024)
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

func downloadComment(client *nigonigo.Client, data *nigonigo.VideoData) error {

	outpath := data.Client.WatchId + ".tsv"

	nvComment := data.Comment.NvComment
	if nvComment == nil {
		return fmt.Errorf("failed to get comment params")
	}

	threads, err := client.GetComments(nvComment.Server, nvComment.Threadkey, nvComment.Params)
	if err != nil {
		return err
	}

	out, err := os.Create(outpath)
	if err != nil {
		return err
	}
	defer out.Close()

	for _, thread := range threads {
		for _, comment := range thread.Comments {
			out.WriteString(join("\t", thread.Fork, comment.ID, comment.No, comment.VposMs, comment.PostedAt, strings.ReplaceAll(comment.Body, "\n", " "), comment.Commands) + "\n")
		}
	}

	return nil
}

func download(client *nigonigo.Client, contentID string, saveThumbnail bool) {
	video, err := client.GetVideoData(contentID)
	if err != nil {
		log.Fatalf("Failed to get video info: %v", err)
	}

	if saveThumbnail {
		err = downloadThumbnail(video.Video.Thumbnail.Url, contentID+".jpg")
		if err != nil {
			log.Printf("Failed to get thumbnail: %v", err)
		}
	}

	session, err := client.CreateVideoSessionFromVideoData(video)
	if err != nil {
		log.Fatalf("Failed to create session: %v", err)
	}

	var wg sync.WaitGroup
	if domandSession, ok := session.(*nigonigo.DomandSession); ok {
		for _, s := range domandSession.SubStreams() {
			wg.Add(1)
			go func(id, ext string) {
				defer wg.Done()
				out, _ := os.Create(contentID + "." + ext)
				_ = domandSession.Download(context.Background(), client.HttpClient, out, id)
			}(s[0], s[1])
		}
	}

	log.Printf("Start download %v", contentID)
	out, _ := os.Create(contentID + "." + session.FileExtension())
	defer out.Close()
	err = client.Download(context.Background(), session, io.MultiWriter(&ByteCounter{}, out))
	if err != nil {
		log.Fatalf("Failed to download: %v", err)
	}
	wg.Wait()
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
	sessionFile := flag.String("s", defaultSessionFilePath, "session.json")
	saveThumbnail := flag.Bool("t", false, "save thumbnail")
	saveComments := flag.Bool("comment", false, "save comments")
	// flag.Parse()
	flag.CommandLine.Parse(os.Args[2:])

	if flag.NArg() == 0 {
		flag.Usage()
		return
	}

	client := nigonigo.NewClient()
	err := authLogin(client, *sessionFile, *accountFile, *id, *password)
	if err != nil {
		log.Println("Failed to login: ", err)
	}

	for _, contentID := range flag.Args() {
		if *saveComments {
			video, err := client.GetVideoData(contentID)
			if err != nil {
				log.Fatalf("Failed to get video info: %v", err)
			}
			downloadComment(client, video)
			continue
		}
		download(client, contentID, *saveThumbnail)
	}
}
