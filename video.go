package nigonigo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"net/url"
	"regexp"
	"strings"
)

const (
	errorMessageNotFound = "この動画は存在しないか、削除された可能性があります。"
	errorMessageExpired  = "お探しの動画は視聴可能期間が終了しています"
	errorMessageChannel  = "チャンネル会員専用動画です"
)

// JSON.parse($("#js-initial-watch-data").dataset.apiData);
type VideoData struct {
	Video struct {
		ContentID      string `json:"id"`
		Title          string `json:"title"`
		ThumbnailURL   string `json:"thumbnailURL"`
		Description    string `json:"description"`
		Duration       int    `json:"duration"`
		PostedDateTime string `json:"postedDateTime"`
		ViewCount      int    `json:"viewCount"`
		MylistCount    int    `json:"mylistCount"`
		// Deprecated?
		DMC struct {
			Quality struct {
				Audios []*SourceStream `json:"audios"`
				Videos []*SourceStream `json:"videos"`
			} `json:"quality"`
			SessionAPI map[string]interface{} `json:"session_api"`
			TrackingID string                 `json:"tracking_id"`
			Encryption map[string]interface{} `json:"encryption"`
		} `json:"dmcInfo"`
		// Deprecated?
		Smile struct {
			URL              string   `json:"url"`
			CurrentQualityID string   `json:"currentQualityId"`
			QualityIds       []string `json:"qualityIds"`
			IsSlowLine       bool     `json:"isSlowLine"`
		} `json:"smileInfo"`
	} `json:"video"`
	// temp1.media.delivery.movie.audios
	Media struct {
		Delivery struct {
			Movie struct {
				Audios  []*SourceStream2       `json:"audios"`
				Videos  []*SourceStream2       `json:"videos"`
				Session map[string]interface{} `json:"session"`
			} `json:"movie"`
			Encryption map[string]interface{} `json:"encryption"`
			TrackingID string                 `json:"trackingId"`
		} `json:"delivery"`
		DeliveryLegacy struct {
		} `json:"deliveryLegacy"`
	} `json:"media"`
	Thread  map[string]interface{} `json:"thread"`
	Owner   map[string]interface{} `json:"owner"`
	Channel map[string]interface{} `json:"channel"`
	Context map[string]interface{} `json:"context"`
}

type SourceStream struct {
	ID           string `json:"id"`
	Available    bool   `json:"available"`
	Bitrate      int    `json:"bitrate"`
	SamplingRate int    `json:"sampling_rate"`
	Resolution   struct {
		Width  int `json:"width"`
		Height int `json:"height"`
	} `json:"resolution"`
}

type SourceStream2 struct {
	ID        string `json:"id"`
	Available bool   `json:"isAvailable"`
	Metadata  struct {
		SamplingRate int `json:"samplingRate"`
		Bitrate      int `json:"bitrate"`
		Resolution   struct {
			Width  int `json:"width"`
			Height int `json:"height"`
		} `json:"resolution"`
	} `json:"metadata"`
}

func (s *SourceStream2) SourceStream() *SourceStream {
	return &SourceStream{
		ID:           s.ID,
		Available:    s.Available,
		Bitrate:      s.Metadata.Bitrate,
		SamplingRate: s.Metadata.SamplingRate,
		Resolution:   s.Metadata.Resolution,
	}
}

func (v *VideoData) IsDMC() bool {
	return v.GetAvailableAudio() != nil || v.GetAvailableVideo() != nil
}

func (v *VideoData) IsSmile() bool {
	return v.Video.Smile.URL != ""
}

func (v *VideoData) IsNeedPayment() bool {
	return v.Context["isNeedPayment"] == true
}

func (v *VideoData) GetSessionData() map[string]interface{} {
	if len(v.Media.Delivery.Movie.Session) > 0 {
		return v.Media.Delivery.Movie.Session
	}
	return v.Video.DMC.SessionAPI
}

func (v *VideoData) GetEncryption() map[string]interface{} {
	if len(v.Media.Delivery.Encryption) > 0 {
		return map[string]interface{}{
			"hls_encryption_v1": map[string]interface{}{
				"key_uri":       v.Media.Delivery.Encryption["keyUri"],
				"encrypted_key": v.Media.Delivery.Encryption["encryptedKey"],
			},
		}
	}
	return v.Video.DMC.Encryption
}

func (v *VideoData) concatStreams(srcs []*SourceStream, stream2 []*SourceStream2) []*SourceStream {
	for _, s := range stream2 {
		srcs = append(srcs, s.SourceStream())
	}
	return srcs
}

func (v *VideoData) GetAvailableAudio() *SourceStream {
	return v.GetAvailableSource(v.concatStreams(v.Video.DMC.Quality.Audios, v.Media.Delivery.Movie.Audios))
}

func (v *VideoData) GetAvailableVideo() *SourceStream {
	return v.GetAvailableSource(v.concatStreams(v.Video.DMC.Quality.Videos, v.Media.Delivery.Movie.Videos))
}

func (v *VideoData) SmileFileExtension() string {
	if strings.HasPrefix(v.Video.Smile.URL, "rtmp") {
		return "flv"
	}
	if strings.Contains(v.Video.Smile.URL, "?m=") {
		return "mp4"
	}
	if strings.Contains(v.Video.Smile.URL, "?v=") {
		return "flv"
	}
	if strings.Contains(v.Video.Smile.URL, "?s=") {
		return "swf"
	}
	return "bin"
}

func (v *VideoData) GetAvailableSource(sources []*SourceStream) *SourceStream {
	var ret *SourceStream
	var bitrate = -1
	for _, s := range sources {
		if s.Available && s.Bitrate > bitrate {
			bitrate = s.Bitrate
			ret = s
		}
	}
	return ret
}

type VideoSession interface {
	FileExtension() string
}

type SmileSession struct {
	VideoData *VideoData
}

func (s *SmileSession) FileExtension() string {
	return s.VideoData.SmileFileExtension()
}

func (c *Client) GetVideoData(contentId string) (*VideoData, error) {
	res, err := getContent(c.HttpClient, watchUrl+contentId, nil)
	if err != nil {
		return nil, err
	}

	body := string(res)
	re := regexp.MustCompile(`data-api-data="([^"]+)"`)
	match := re.FindStringSubmatch(body)
	if match == nil {
		if strings.Contains(body, errorMessageNotFound) {
			return nil, errors.New("video not found")
		}
		if strings.Contains(body, errorMessageExpired) {
			return nil, errors.New("expired")
		}
		if strings.Contains(body, errorMessageChannel) {
			return nil, errors.New("member only")
		}
		return nil, errors.New("invalid response(data-api-data)")
	}
	jsonString := html.UnescapeString(match[1])
	var data VideoData
	err = json.Unmarshal([]byte(jsonString), &data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (c *Client) CreateVideoSession(contentID string) (VideoSession, error) {
	data, err := c.GetVideoData(contentID)
	if err != nil {
		return nil, err
	}
	if data.IsDMC() {
		return c.CreateDMCSessionByVideoData(data)
	} else if data.IsSmile() {
		return c.CreateSmileSessionByVideoData(data)
	}
	return nil, fmt.Errorf("No available video source")
}

func (c *Client) Download(ctx context.Context, session VideoSession, w io.Writer) error {
	if dmcSession, ok := session.(*DMCSession); ok {
		return c.DownloadFromDMC(ctx, dmcSession, w)
	}
	smileSession, _ := session.(*SmileSession)
	return c.DownloadFromSmile(ctx, smileSession.VideoData, w)
}

type jsonObject map[string]interface{}
type jsonArray []interface{}

func (data *VideoData) GenDMCSessionReq(audio, video string) (jsonObject, error) {
	if !data.IsDMC() {
		return nil, errors.New("DMC is not available for the video")
	}

	var audioIds []string
	var videoIds []string
	if audio != "" {
		audioIds = []string{audio}
	}
	if video != "" {
		videoIds = []string{video}
	}

	transferPreset := ""
	transferPresets := data.GetSessionData()["transferPresets"].([]interface{})
	if len(transferPresets) > 0 {
		transferPreset = transferPresets[0].(string)
	}

	var httpParams = jsonObject{
		"http_output_download_parameters": jsonObject{
			"use_well_known_port": "yes",
			"use_ssl":             "yes",
			"transfer_preset":     transferPreset,
			//"file_extension":      "flv",
		},
	}
	httpAvailable := false
	for _, proto := range data.GetSessionData()["protocols"].([]interface{}) {
		if proto == "http" {
			httpAvailable = true
		}
	}
	if !httpAvailable {
		httpParams = jsonObject{
			"hls_parameters": jsonObject{
				"use_well_known_port": "yes",
				"use_ssl":             "yes",
				"segment_duration":    6000,
				"transfer_preset":     transferPreset,
				"encryption":          data.GetEncryption(),
			},
		}
	}

	var reqsession = jsonObject{
		"session": jsonObject{
			"recipe_id":         data.GetSessionData()["recipeId"],
			"content_id":        data.GetSessionData()["contentId"],
			"content_type":      "movie",
			"timing_constraint": "unlimited",
			"content_src_id_sets": jsonArray{
				jsonObject{
					"content_src_ids": jsonArray{
						jsonObject{
							"src_id_to_mux": jsonObject{
								"video_src_ids": videoIds,
								"audio_src_ids": audioIds,
							},
						},
					},
				},
			},
			"protocol": jsonObject{
				"name": "http",
				"parameters": jsonObject{
					"http_parameters": jsonObject{
						"parameters": httpParams,
					},
				},
			},
			"keep_method": jsonObject{
				"heartbeat": jsonObject{
					"lifetime": data.GetSessionData()["heartbeatLifetime"],
				},
			},
			"session_operation_auth": jsonObject{
				"session_operation_auth_by_signature": jsonObject{
					"token":     data.GetSessionData()["token"],
					"signature": data.GetSessionData()["signature"],
				},
			},
			"content_auth": jsonObject{
				"auth_type":           "ht2",
				"service_id":          "nicovideo",
				"service_user_id":     data.GetSessionData()["serviceUserId"],
				"content_key_timeout": data.GetSessionData()["contentKeyTimeout"],
			},
			"client_info": jsonObject{
				"player_id": data.GetSessionData()["playerId"],
			},
			"priority": data.GetSessionData()["priority"],
		},
	}
	return reqsession, nil
}

func (c *Client) prepareLicense(data *VideoData) error {
	if len(data.GetEncryption()) > 0 {
		// Prepare encryption key.
		// See: https://github.com/tor4kichi/Hohoema/issues/778
		url := nvApiUrl + "2ab0cbaa/watch?t=" + url.QueryEscape(data.Video.DMC.TrackingID+data.Media.Delivery.TrackingID)
		req, err := newGetReq(url, nil)
		if err != nil {
			return err
		}
		req.Header.Set("X-Frontend-Id", "6")
		req.Header.Set("X-Frontend-Version", "0")

		_, err = doRequest(c.HttpClient, req)
		if err != nil {
			return err
		}
	}
	return nil
}
func (c *Client) CreateDMCSessionById(contentID string) (*DMCSession, error) {
	data, err := c.GetVideoData(contentID)
	if err != nil {
		return nil, err
	}
	return c.CreateDMCSessionByVideoData(data)
}

func (c *Client) CreateDMCSessionByVideoData(data *VideoData) (*DMCSession, error) {
	err := c.prepareLicense(data)
	if err != nil {
		return nil, err
	}

	audio := data.GetAvailableAudio()
	video := data.GetAvailableVideo()
	if audio == nil || video == nil {
		if data.IsNeedPayment() {
			return nil, errors.New("need payment")
		}
		return nil, errors.New("source not found")
	}

	reqsession, err := data.GenDMCSessionReq(audio.ID, video.ID)
	if err != nil {
		return nil, err
	}
	sessionApiURL := data.GetSessionData()["urls"].([]interface{})[0].(map[string]interface{})["url"].(string)
	return c.CreateDMCSession(reqsession, sessionApiURL)
}

func (c *Client) CreateSmileSessionByVideoData(data *VideoData) (*SmileSession, error) {
	return &SmileSession{VideoData: data}, nil
}

func (c *Client) DownloadFromSmile(ctx context.Context, data *VideoData, w io.Writer) error {
	if !data.IsSmile() {
		return errors.New("SMILE is not available")
	}

	if !strings.HasPrefix(data.Video.Smile.URL, "http") {
		return fmt.Errorf("unsported protocol : %v", data.Video.Smile.URL)
	}
	return NewHTTPDownloader(c.HttpClient).Download(ctx, data.Video.Smile.URL, w)
}
