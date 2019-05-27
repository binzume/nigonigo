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

var (
	errorMessageNotFound = "この動画は存在しないか、削除された可能性があります。"
	errorMessageExpired  = "お探しの動画は視聴可能期間が終了しています"
	errorMessageChannel  = "チャンネル会員専用動画です"
)

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
		DMC            struct {
			Quality struct {
				Audios []*SourceStream `json:"audios"`
				Videos []*SourceStream `json:"videos"`
			} `json:"quality"`
			SessionAPI map[string]interface{} `json:"session_api"`
			TrackingID string                 `json:"tracking_id"`
			Encryption map[string]interface{} `json:"encryption"`
		} `json:"dmcInfo"`
		Smile struct {
			URL              string   `json:"url"`
			CurrentQualityID string   `json:"currentQualityId"`
			QualityIds       []string `json:"qualityIds"`
			IsSlowLine       bool     `json:"isSlowLine"`
		} `json:"smileInfo"`
	} `json:"video"`
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

func (v *VideoData) IsDMC() bool {
	return v.GetAvailableAudio() != nil || v.GetAvailableVideo() != nil
}

func (v *VideoData) IsSmile() bool {
	return v.Video.Smile.URL != ""
}

func (v *VideoData) IsNeedPayment() bool {
	return v.Context["isNeedPayment"] == true
}

func (v *VideoData) GetAvailableAudio() *SourceStream {
	return v.GetAvailableSource(v.Video.DMC.Quality.Audios)
}

func (v *VideoData) GetAvailableVideo() *SourceStream {
	return v.GetAvailableSource(v.Video.DMC.Quality.Videos)
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

func (c *Client) GetVideoData(contentId string) (*VideoData, error) {
	res, err := GetContent(c.HttpClient, watchUrl+contentId, nil)
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
	transferPresets := data.Video.DMC.SessionAPI["transfer_presets"].([]interface{})
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
	for _, proto := range data.Video.DMC.SessionAPI["protocols"].([]interface{}) {
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
				"encryption":          data.Video.DMC.Encryption,
			},
		}
	}

	var reqsession = jsonObject{
		"session": jsonObject{
			"recipe_id":         data.Video.DMC.SessionAPI["recipe_id"],
			"content_id":        data.Video.DMC.SessionAPI["content_id"],
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
					"lifetime": data.Video.DMC.SessionAPI["heartbeat_lifetime"],
				},
			},
			"session_operation_auth": jsonObject{
				"session_operation_auth_by_signature": jsonObject{
					"token":     data.Video.DMC.SessionAPI["token"],
					"signature": data.Video.DMC.SessionAPI["signature"],
				},
			},
			"content_auth": jsonObject{
				"auth_type":           "ht2",
				"service_id":          "nicovideo",
				"service_user_id":     data.Video.DMC.SessionAPI["service_user_id"],
				"content_key_timeout": data.Video.DMC.SessionAPI["content_key_timeout"],
			},
			"client_info": jsonObject{
				"player_id": data.Video.DMC.SessionAPI["player_id"],
			},
			"priority": data.Video.DMC.SessionAPI["priority"],
		},
	}
	return reqsession, nil
}

func (c *Client) PrepareLicense(data *VideoData) error {
	if data.Video.DMC.Encryption["hls_encryption_v1"] != nil {
		// Prepare encryption key.
		// See: https://github.com/tor4kichi/Hohoema/issues/778
		Logger.Println(data.Video.DMC.Encryption)
		url := nvApiUrl + "2ab0cbaa/watch?t=" + url.QueryEscape(data.Video.DMC.TrackingID)
		req, err := NewGetReq(url, nil)
		if err != nil {
			return err
		}
		req.Header.Set("X-Frontend-Id", "6")
		req.Header.Set("X-Frontend-Version", "0")

		_, err = DoRequest(c.HttpClient, req)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) CreateDMCSessionById(contentId string) (*DMCSession, error) {
	data, err := c.GetVideoData(contentId)
	if err != nil {
		return nil, err
	}

	err = c.PrepareLicense(data)
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
	sessionApiURL := data.Video.DMC.SessionAPI["urls"].([]interface{})[0].(map[string]interface{})["url"].(string)
	return c.CreateDMCSession(reqsession, sessionApiURL)
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
