# [WIP] niconico API client for Golang

Experimental implementation of niconico API client for Golang.

## Features

- Login
- Search
- Download video

## Usage

T.B.D.

```
func main() {
    contentID := "sm9"
	client := nigonigo.NewClient()
	session, err := client.CreateDMCSessionById(contentID)
	if err != nil {
		t.Errorf("Failed to create session: %v", err)
	}

	out, _ := os.Create(contentID + "." + session.FileExtension())
	defer out.Close()
	err = client.Download(context.Background(), session, out)
	if err != nil {
		t.Errorf("Failed to download: %v", err)
	}
}
```


## License

MIT License
