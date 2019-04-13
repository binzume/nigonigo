package nigonigo

import (
	"context"
	"io"
	"net/http"
)

type HTTPDownloader struct {
	HttpClient *http.Client
}

func NewHTTPDownloader(client *http.Client) *HTTPDownloader {
	if client == nil {
		client = &http.Client{}
	}
	return &HTTPDownloader{HttpClient: client}
}

func (c *HTTPDownloader) Downaload(ctx context.Context, url string, w io.Writer) error {
	req, err := NewGetReq(url, nil)
	if err != nil {
		return err
	}

	res, err := c.HttpClient.Do(req.WithContext(ctx))
	if err != nil {
		return err
	}

	defer res.Body.Close()
	_, err = io.Copy(w, res.Body)
	return err
}
