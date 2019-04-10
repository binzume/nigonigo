package nigonigo

import (
	"bytes"
	"encoding/json"
	"errors"
	"html"
	"net/http"
	"regexp"
)

type videoData struct {
	Video struct {
		ID       string `json:"id"`
		Duration int    `json:"duration"`
		DMC      struct {
			Quality struct {
				Audio []interface{} `json:"audio"`
				Video []interface{} `json:"video"`
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

func (c *Client) GetContentUrl(contentId string) (string, error) {
	res, err := c.get(watchUrl + contentId)
	if err != nil {
		return "", err
	}

	re := regexp.MustCompile(`data-api-data="([^"]+)"`)
	match := re.FindStringSubmatch(res)
	if match == nil {
		return "", errors.New("invalid response")
	}
	jsonString := html.UnescapeString(match[1])
	var data videoData
	err = json.Unmarshal([]byte(jsonString), &data)
	if err != nil {
		return "", err
	}

	// TODO
	video := "archive_h264_300kbps_360p"
	audio := "archive_aac_64kbps"

	type JSONObject map[string]interface{}
	type JSONArray []interface{}
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
								"video_src_ids": JSONArray{video},
								"audio_src_ids": JSONArray{audio},
							},
						},
					},
				},
			},
			"protocol": JSONObject{
				"name": "http",
				"parameters": JSONObject{
					"http_parameters": JSONObject{
						"parameters": JSONObject{
							"http_output_download_parameters": JSONObject{
								"use_well_known_port": "yes",
								"use_ssl":             "yes",
								"transfer_preset":     "",
							},
						},
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
		return "", err
	}
	sessionApiURL := data.Video.DMC.SessionAPI["urls"].([]interface{})[0].(map[string]interface{})["url"].(string)

	req, err := http.NewRequest("POST", sessionApiURL+"?_format=json", bytes.NewReader(sessionStr))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err = c.request(req)
	if err != nil {
		return "", err
	}

	type sessionResponse struct {
		Data struct {
			Session struct {
				ID         string `json:"id"`
				ContentURI string `json:"content_uri"`
			} `json:"session"`
		} `json:"data"`
	}
	var sessionRes sessionResponse
	err = json.Unmarshal([]byte(res), &sessionRes)
	if err != nil {
		return "", err
	}

	return sessionRes.Data.Session.ContentURI, nil
}
