# niconico API client for Golang
[![Build Status](https://github.com/binzume/nigonigo/actions/workflows/test.yaml/badge.svg)](https://github.com/binzume/nigonigo/actions)
[![Go Reference](https://pkg.go.dev/badge/github.com/binzume/nigonigo.svg)](https://pkg.go.dev/github.com/binzume/nigonigo)
[![license](https://img.shields.io/badge/license-MIT-4183c4.svg)](https://github.com/binzume/nigonigo/blob/master/LICENSE)

Experimental implementation of [niconico](https://www.nicovideo.jp/) API client for Golang.

## Features

- Login/Logout
- Search
- MyList
- Ranking
- Download video

## Usage

T.B.D.

```go
func main() {
	client := nigonigo.NewClient()

	contentID := "sm9"
	session, err := client.CreateVideoSession(contentID)
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

### Command line tool:

install:

```bash
go install github.com/binzume/nigonigo/cmd/nigo@latest
```

usage:

```bash
nigo search "ねこ"
nigo search -t "MMD 初音ミク"
nigo search -sort "-viewCounter" -limit 3  -t "ニコニコ技術部"
nigo search -series 96269
nigo search -id sm9
nigo ranking cooking
nigo auth -i "YOUR_MAILADDRESS"
 Password: ********
nigo download sm9
open sm9.mp4
```

## License

MIT License
