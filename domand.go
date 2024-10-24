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
	Data  DomandAccessData
	Video *SourceStream3
	Audio *SourceStream3
}

func (s *DomandSession) FileExtension() string {
	return "m4v"
}

func (s *DomandSession) SubStreams() [][]string {
	if s.Audio != nil {
		return [][]string{{s.Audio.ID, "m4a"}}
	}
	return nil
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
	video := data.GetPreferredVideo()
	audio := data.GetPreferredAudio()
	if video == nil && audio == nil {
		return nil, errors.New("no domand streams")
	}

	var streams []string
	if video != nil {
		streams = append(streams, video.ID)
	}
	if audio != nil {
		streams = append(streams, audio.ID)
	}

	reqsession := struct {
		Outputs [][]string `json:"outputs"`
	}{Outputs: [][]string{streams}}

	sessionBytes, err := json.Marshal(reqsession)
	if err != nil {
		return nil, err
	}

	trackid := data.Client.WatchTrackId
	url := nvApiUrl + "watch/" + data.Client.WatchId + "/access-rights/hls?actionTrackId=" + trackid
	req, err := http.NewRequest("POST", url, bytes.NewReader(sessionBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-access-right-key", data.Media.Domand.AccessRightKey)

	res, err := doRequest(c.HttpClient, req)
	if err != nil {
		return nil, err
	}

	domainRes := struct {
		Meta map[string]int   `json:"meta"`
		Data DomandAccessData `json:"data"`
	}{}

	err = json.Unmarshal(res, &domainRes)
	if err != nil {
		return nil, err
	}

	return &DomandSession{Video: video, Audio: audio, Data: domainRes.Data}, nil
}
