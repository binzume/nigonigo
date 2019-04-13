package nigonigo

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
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

func (v *VideoData) IsDMC() bool {
	return v.GetAvailableAudio() != nil || v.GetAvailableVideo() != nil
}

func (v *VideoData) IsSmile() bool {
	return v.Video.Smile.URL != ""
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
	res, err := c.GetContent(watchUrl + contentId)
	if err != nil {
		return nil, err
	}

	re := regexp.MustCompile(`data-api-data="([^"]+)"`)
	match := re.FindStringSubmatch(string(res))
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

	var httpParams = jsonObject{
		"http_output_download_parameters": jsonObject{
			"use_well_known_port": "yes",
			"use_ssl":             "yes",
			"transfer_preset":     "",
			// "file_extension": "flv",
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
				"transfer_preset":     "",
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

type DMCSession struct {
	ID         string `json:"id"`
	ContentURI string `json:"content_uri"`
	Protocol   struct {
		Name       string                 `json:"name"`
		Parameters map[string]interface{} `json:"parameters"`
	} `json:"protocol"`
	KeepMethod struct {
		Heartbeat *struct {
			LifetimeMs   int    `json:"lifetime"`
			OnetimeToken string `json:"onetime_token"`
		} `json:"heartbeat"`
	} `json:"keep_method"`
	url      string
	fullData jsonObject
}

// IsHLS returns true if protocol is hls
func (s *DMCSession) IsHLS() bool {
	httpParams := s.Protocol.Parameters["http_parameters"]
	if httpParams == nil {
		return false
	}
	hlsParams := httpParams.(map[string]interface{})["parameters"].(map[string]interface{})["hls_parameters"]
	return hlsParams != nil
}

// IsHTTP returns true if protocol is http or https
func (s *DMCSession) IsHTTP() bool {
	return strings.HasPrefix(s.ContentURI, "http")
}

// IsRTMP returns true if protocol is rtmp
func (s *DMCSession) IsRTMP() bool {
	return strings.HasPrefix(s.ContentURI, "rtmp:")
}

func (s *DMCSession) FileExtension() string {
	if s.IsHLS() {
		// TODO: segment formant
		return "ts"
	}
	// TODO: http_output_download_parameters.file_extension
	return "mp4"
}

func (c *Client) GetDMCSession(data *VideoData, audio, video string) (*DMCSession, error) {
	reqsession, err := data.GenDMCSessionReq(audio, video)
	if err != nil {
		return nil, err
	}
	sessionApiURL := data.Video.DMC.SessionAPI["urls"].([]interface{})[0].(map[string]interface{})["url"].(string)

	sessionBytes, err := json.Marshal(reqsession)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", sessionApiURL+"?_format=json", bytes.NewReader(sessionBytes))
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
		log.Println(string(sessionBytes))
		return nil, fmt.Errorf("Status: %v %v", sessionRes.Meta.Status, sessionRes.Meta.Message)
	}

	var untyped struct {
		Data map[string]interface{} `json:"data"`
	}
	session := &sessionRes.Data.Session
	session.url = sessionApiURL + "/" + sessionRes.Data.Session.ID
	err = json.Unmarshal([]byte(res), &untyped)
	if err == nil {
		session.fullData = untyped.Data
	}
	return session, nil
}

func (c *Client) Heartbeat(session *DMCSession) error {
	sessionStr, err := json.Marshal(session.fullData)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", session.url+"?_format=json&_method=PUT", bytes.NewReader(sessionStr))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := c.request(req)
	if err != nil {
		return err
	}

	type sessionResponse struct {
		Meta struct {
			Status  int    `json:"status"`
			Message string `json:"message"`
		} `json:"meta"`
		Data map[string]interface{} `json:"data"`
	}
	var sessionRes sessionResponse
	err = json.Unmarshal([]byte(res), &sessionRes)
	if err != nil {
		return err
	}
	if sessionRes.Meta.Status < 200 || sessionRes.Meta.Status >= 300 {
		return fmt.Errorf("Status: %v %v", sessionRes.Meta.Status, sessionRes.Meta.Message)
	}
	session.fullData = sessionRes.Data
	return nil
}

func (c *Client) PrepareLicense(data *VideoData) error {
	if data.Video.DMC.Encryption["hls_encryption_v1"] != nil {
		// Prepare encryption key.
		// See: https://github.com/tor4kichi/Hohoema/issues/778
		log.Println(data.Video.DMC.Encryption)
		log.Println(data.Video.DMC.TrackingID)
		url := nvApiUrl + "2ab0cbaa/watch?t=" + url.QueryEscape(data.Video.DMC.TrackingID)
		req, err := NewGetReq(url, nil)
		if err != nil {
			return err
		}
		req.Header.Set("X-Frontend-Id", "6")
		req.Header.Set("X-Frontend-Version", "0")

		result, err := c.request(req)
		if err != nil {
			return err
		}
		log.Println(result)
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
		return nil, errors.New("source not found")
	}
	return c.GetDMCSession(data, audio.ID, video.ID)
}

func (c *Client) StartHeartbeat(ctx context.Context, session *DMCSession, errorLimit int) {
	go func() {
		errorCount := 0
		interval := time.Duration(session.KeepMethod.Heartbeat.LifetimeMs) / 2 * time.Millisecond
		for {
			select {
			case <-ctx.Done():
				log.Printf("finished heartbeat")
				return
			case <-time.After(interval):
				err := c.Heartbeat(session)
				if err != nil {
					log.Printf("hearbeat error :%v", err)
					errorCount++
					if errorCount > errorLimit {
						return
					}
				}
			}
		}
	}()
}

func (c *Client) Download(ctx context.Context, session *DMCSession, w io.Writer) error {
	if session.Protocol.Name != "http" {
		return fmt.Errorf("unsupported protocol : %v", session.Protocol.Name)
	}

	if session.KeepMethod.Heartbeat != nil {
		hbCtx, cancel := context.WithCancel(ctx)
		defer cancel()
		c.StartHeartbeat(hbCtx, session, 2)
	}

	if session.IsHLS() {
		return NewHLSDownloader(c.HttpClient).Downaload(ctx, session.ContentURI, w)
	} else if session.IsHTTP() {
		return NewHTTPDownloader(c.HttpClient).Downaload(ctx, session.ContentURI, w)
	} else {
		return fmt.Errorf("unsupported protocol url: %v", session.ContentURI)
	}
}
