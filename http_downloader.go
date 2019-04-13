package nigonigo

import (
	"context"
	"errors"
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

type wrappedReader struct {
	reader io.Reader
	done   <-chan struct{}
}

func (r *wrappedReader) Read(p []byte) (n int, err error) {
	select {
	case <-r.done:
		return 0, errors.New("cancelled")
	default:
		return r.reader.Read(p)
	}
}

func (c *HTTPDownloader) Downaload(ctx context.Context, url string, w io.Writer) error {
	req, err := NewGetReq(url, nil)
	if err != nil {
		return err
	}

	res, err := c.HttpClient.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	_, err = io.Copy(w, &wrappedReader{res.Body, ctx.Done()})
	return err
}
