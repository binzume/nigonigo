package nigonigo

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"regexp"
)

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

type VideoData struct {
	Video struct {
		ID       string `json:"id"`
		Duration int    `json:"duration"`
		DMC      struct {
			Quality struct {
				Audios []*SourceStream `json:"audios"`
				Videos []*SourceStream `json:"videos"`
			} `json:"quality"`
			SessionAPI map[string]interface{} `json:"session_api"`
			TrackingID string                 `json:"tracking_id"`
		} `json:"dmcInfo"`
		Smile map[string]interface{} `json:"smileInfo"`
	} `json:"video"`
	Thread  map[string]interface{} `json:"thread"`
	Owner   map[string]interface{} `json:"owner"`
	Channel map[string]interface{} `json:"channel"`
	Context map[string]interface{} `json:"context"`
}

func (v *VideoData) IsDMC() bool {
	return v.GetAvailableAudio() != nil
}

func (v *VideoData) IsSmile() bool {
	return v.Video.Smile["url"] != nil
}

func (v *VideoData) GetAvailableAudio() *SourceStream {
	return v.GetAvailableSource(v.Video.DMC.Quality.Audios)
}

func (v *VideoData) GetAvailableVideo() *SourceStream {
	return v.GetAvailableSource(v.Video.DMC.Quality.Videos)
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
	res, err := c.get(watchUrl + contentId)
	if err != nil {
		return nil, err
	}

	re := regexp.MustCompile(`data-api-data="([^"]+)"`)
	match := re.FindStringSubmatch(res)
	if match == nil {
		return nil, errors.New("invalid response")
	}
	jsonString := html.UnescapeString(match[1])
	var data VideoData
	err = json.Unmarshal([]byte(jsonString), &data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

type DMCSession struct {
	ID         string `json:"id"`
	ContentURI string `json:"content_uri"`
	KeepMethod struct {
		Heartbeat *struct {
			LifetimeMs   int    `json:"lifetime"`
			OnetimeToken string `json:"onetime_token"`
		} `json:"heartbeat"`
	} `json:"keep_method"`
}

func (c *Client) GetDMCSession(data *VideoData, audio, video string) (*DMCSession, error) {
	type JSONObject map[string]interface{}
	type JSONArray []interface{}
	var audioIds []string
	var videoIds []string
	if audio != "" {
		audioIds = []string{audio}
	}
	if video != "" {
		videoIds = []string{video}
	}

	var httpParams = JSONObject{
		"http_output_download_parameters": JSONObject{
			"use_well_known_port": "yes",
			"use_ssl":             "yes",
			"transfer_preset":     "",
		},
	}
	httpAvailable := false
	for _, proto := range data.Video.DMC.SessionAPI["protocols"].([]interface{}) {
		if proto == "http" {
			httpAvailable = true
		}
	}
	if !httpAvailable {
		httpParams = JSONObject{
			"hls_parameters": JSONObject{
				"use_well_known_port": "yes",
				"use_ssl":             "yes",
				"segment_duration":    6000,
				"transfer_preset":     "",
			},
		}
	}

	var session = JSONObject{
		"session": JSONObject{
			"recipe_id":         data.Video.DMC.SessionAPI["recipe_id"],
			"content_id":        data.Video.DMC.SessionAPI["content_id"],
			"content_type":      "movie",
			"timing_constraint": "unlimited",
			"content_src_id_sets": JSONArray{
				JSONObject{
					"content_src_ids": JSONArray{
						JSONObject{
							"src_id_to_mux": JSONObject{
								"video_src_ids": videoIds,
								"audio_src_ids": audioIds,
							},
						},
					},
				},
			},
			"protocol": JSONObject{
				"name": "http",
				"parameters": JSONObject{
					"http_parameters": JSONObject{
						"parameters": httpParams,
					},
				},
			},
			"keep_method": JSONObject{
				"heartbeat": JSONObject{
					"lifetime": data.Video.DMC.SessionAPI["heartbeat_lifetime"],
				},
			},
			"session_operation_auth": JSONObject{
				"session_operation_auth_by_signature": JSONObject{
					"token":     data.Video.DMC.SessionAPI["token"],
					"signature": data.Video.DMC.SessionAPI["signature"],
				},
			},
			"content_auth": JSONObject{
				"auth_type":           "ht2",
				"service_id":          "nicovideo",
				"service_user_id":     data.Video.DMC.SessionAPI["service_user_id"],
				"content_key_timeout": data.Video.DMC.SessionAPI["content_key_timeout"],
			},
			"client_info": JSONObject{
				"player_id": data.Video.DMC.SessionAPI["player_id"],
			},
		},
	}
	sessionStr, err := json.Marshal(session)
	if err != nil {
		return nil, err
	}
	sessionApiURL := data.Video.DMC.SessionAPI["urls"].([]interface{})[0].(map[string]interface{})["url"].(string)

	req, err := http.NewRequest("POST", sessionApiURL+"?_format=json", bytes.NewReader(sessionStr))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := c.request(req)
	if err != nil {
		return nil, err
	}

	type sessionResponse struct {
		Meta struct {
			Status  int    `json:"status"`
			Message string `json:"message"`
		} `json:"meta"`
		Data struct {
			Session DMCSession `json:"session"`
		} `json:"data"`
	}
	var sessionRes sessionResponse
	err = json.Unmarshal([]byte(res), &sessionRes)
	if err != nil {
		return nil, err
	}
	if sessionRes.Meta.Status < 200 || sessionRes.Meta.Status >= 300 {
		log.Println(string(sessionStr))
		return nil, fmt.Errorf("Status: %v %v", sessionRes.Meta.Status, sessionRes.Meta.Message)
	}

	return &sessionRes.Data.Session, nil
}

func (c *Client) GetDMCSessionSimple(contentId string) (*DMCSession, error) {
	data, err := c.GetVideoData(contentId)
	if err != nil {
		return nil, err
	}

	if !data.IsDMC() {
		return nil, errors.New("DMC is not available for the video")
	}

	audio := data.GetAvailableAudio()
	video := data.GetAvailableVideo()
	if audio == nil || video == nil {
		return nil, errors.New("source not found")
	}
	return c.GetDMCSession(data, audio.ID, video.ID)
}

func (c *Client) StartDownload(session *DMCSession, w io.Writer) error {
	req, err := NewGetReq(session.ContentURI, nil)
	if err != nil {
		return err
	}

	res, err := c.HttpClient.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	if session.KeepMethod.Heartbeat != nil {
		log.Println("TODO: heartbeat...")
	}

	_, err = io.Copy(w, res.Body)
	return err
}

func (c *Client) SimpleDownload(contentId string, w io.Writer) error {
	session, err := c.GetDMCSessionSimple(contentId)
	if err != nil {
		return err
	}
	return c.StartDownload(session, w)
}
