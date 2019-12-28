# [WIP] niconico API client for Golang

[![GoDoc](https://godoc.org/github.com/binzume/nigonigo?status.svg)](https://godoc.org/github.com/binzume/nigonigo) [![license](https://img.shields.io/badge/license-MIT-4183c4.svg)](https://github.com/binzume/nigonigo/blob/master/LICENSE)

Experimental implementation of [niconico](https://www.nicovideo.jp/) API client for Golang.

## Features

- Login/Logout
- Search
- MyList
- Download video

## Usage

T.B.D.

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

### Command line tool:

install:

```bash
go get -u github.com/binzume/nigonigo
go install github.com/binzume/nigonigo/cmd/nigo
```

usage:

```bash
nigo search -limit 10 -t "アニメ"
nigo search -ch 1 "ねこ"
nigo auth -i "YOUR_MAILADDRESS" -s session.json
 Password: ********
nigo download -s session.json sm9
open sm9.mp4
```

## License

MIT License
