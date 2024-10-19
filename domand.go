package nigonigo

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

type DomandAccessData struct {
	ContentUrl string `json:"contentUrl"`
	CreateTime string `json:"createTime"`
	ExpireTime string `json:"expireTime"`
}

type DomandSession struct {
	VideoData *VideoData
	Data      DomandAccessData
}

func (s *DomandSession) FileExtension() string {
	return "m4v"
}

func (s *DomandSession) SubStreams() [][]string {
	return [][]string{{s.VideoData.Media.Domand.Audios[0].ID, "m4a"}}
}

func (s *DomandSession) Download(ctx context.Context, client *http.Client, w io.Writer, stream string) error {
	if s.Data.ContentUrl == "" {
		return errors.New("domand is not available")
	}
	hls := NewHLSDownloader(client)
	hls.TargetGroupID = stream
	return hls.Download(ctx, s.Data.ContentUrl, w)
}

func (c *Client) CreateDomandSessionByVideoData(data *VideoData) (*DomandSession, error) {
	if !data.IsDomand() || data.Client.WatchTrackId == "" {
		return nil, errors.New("domand is not available")
	}
	videos := data.Media.Domand.Videos
	audios := data.Media.Domand.Audios
	if len(videos) == 0 || len(audios) == 0 {
		return nil, errors.New("no domand streams")
	}
	vid := data.Client.WatchId
	trackid := data.Client.WatchTrackId

	reqsession := struct {
		Outputs [][]string `json:"outputs"`
	}{
		Outputs: [][]string{
			{videos[0].ID, audios[0].ID},
		},
	}

	sessionBytes, err := json.Marshal(reqsession)
	if err != nil {
		return nil, err
	}

	url := "https://nvapi.nicovideo.jp/v1/watch/" + vid + "/access-rights/hls?actionTrackId=" + trackid
	req, err := http.NewRequest("POST", url, bytes.NewReader(sessionBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-access-right-key", data.Media.Domand.AccessRightKey)
	req.Header.Set("x-frontend-id", "6")
	req.Header.Set("x-frontend-version", "0")
	req.Header.Set("x-niconico-language", "ja-jp")
	req.Header.Set("x-request-with", "nicovideo")

	res, err := doRequest(c.HttpClient, req)
	if err != nil {
		return nil, err
	}

	// log.Println(url)
	// log.Println(data.Media.Domand.AccessRightKey)
	// log.Println(string(sessionBytes))
	// log.Println(string(res))

	domainRes := struct {
		Meta map[string]int   `json:"meta"`
		Data DomandAccessData `json:"data"`
	}{}

	err = json.Unmarshal(res, &domainRes)
	if err != nil {
		return nil, err
	}

	return &DomandSession{VideoData: data, Data: domainRes.Data}, nil
}
