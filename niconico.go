package nigonigo

import (
	"net/http"
)

const (
	loginApiUrl  = "https://account.nicovideo.jp/api/v1/login?site=niconico"
	searchApiUrl = "http://api.search.nicovideo.jp/api/v2/video/contents/search"
	topUrl       = "https://www.nicovideo.jp/"
	watchUrl     = "https://www.nicovideo.jp/watch/"
	httpOrigin   = "https://www.nicovideo.jp"
	nvApiUrl     = "https://nvapi.nicovideo.jp/v1/"
)

type Client struct {
	HttpClient *http.Client
	Session    *NicoSession
}

func NewClient() *Client {
	client, err := NewHttpClient()
	if err != nil {
		panic(err)
	}
	return &Client{client, nil}
}
