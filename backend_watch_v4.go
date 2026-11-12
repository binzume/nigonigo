package nigonigo

import (
	"errors"
	"net/http"
)

type WatchV4VideoData struct {
	// D map[string]any `json:"data"`
	Data struct {
		Client struct {
			NicoSID      string `json:"nicosid"`
			WatchId      string `json:"watchId"`
			WatchTrackId string `json:"watchTrackId"`
		} `json:"client"`
		Media struct {
			AccessRightKey string `json:"accessRightKey"`
			Contents       struct {
				Audios []*SourceStream3 `json:"audios"`
				Videos []*SourceStream3 `json:"videos"`
			} `json:"contents"`
			HLS         map[string]any `json:"hls"`
			LowDataMode map[string]any `json:"lowDataMode"`

			IsStoryboardAvailable bool `json:"isStoryboardAvailable"`
		} `json:"media"`
	} `json:"data"`
}

func CreateWatchV4SessionByVideoData(client *http.Client, data *VideoData) (*DomandSession, error) {
	if data.WatchV4 == nil {
		return nil, errors.New("watchv4 is not available")
	}

	// convert video params
	data.Media.Domand = &DomandVideoData{
		Videos:         data.WatchV4.Data.Media.Contents.Videos,
		Audios:         data.WatchV4.Data.Media.Contents.Videos,
		AccessRightKey: data.WatchV4.Data.Media.AccessRightKey,
	}
	data.Client = data.WatchV4.Data.Client

	return CreateDomandSessionByVideoData(client, data)
}
