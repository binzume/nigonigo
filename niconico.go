package nigonigo

import (
	"net/http"
)

const (
	topUrl        = "https://www.nicovideo.jp/"
	accountApiUrl = "https://account.nicovideo.jp/api/v1/"
	logoutUrl     = "https://account.nicovideo.jp/logout"
	searchApiUrl  = "https://api.search.nicovideo.jp/api/v2/video/contents/search"
	watchUrl      = "https://www.nicovideo.jp/watch/"
	httpOrigin    = "https://www.nicovideo.jp"
	nvApiUrl      = "https://nvapi.nicovideo.jp/v1/"
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
