# [WIP] niconico API client for Golang

Experimental implementation of niconico API client for Golang.

## Features

- Login
- Search
- My list
- Download video

## Usage

T.B.D.

```go
func main() {
	client := nigonigo.NewClient()

	contentID := "sm9"
	video, err := client.GetVideoData(contentID)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	out, _ := os.Create(contentID + "." + video.SmileFileExtension())
	defer out.Close()
	err = client.DownloadFromSmile(ctx, video, out)
	if err != nil {
		t.Errorf("Failed to download: %v", err)
	}
	log.Println("ok")
}
```

```go
func main() {
	client := nigonigo.NewClient()

	contentID := "sm9"
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
```

## License

MIT License
