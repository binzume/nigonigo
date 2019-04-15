package nigonigo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

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
	// http_parameters.http_output_download_parameters.file_extension
	if httpParams := s.Protocol.Parameters["http_parameters"]; httpParams != nil {
		httpParams = httpParams.(map[string]interface{})["parameters"]
		if dlParams := httpParams.(map[string]interface{})["http_output_download_parameters"]; dlParams != nil {
			if ext := dlParams.(map[string]interface{})["file_extension"]; ext != nil && ext != "" {
				return ext.(string)
			}
		}
	}
	return "mp4"
}

func (c *Client) CreateDMCSession(reqsession jsonObject, sessionApiURL string) (*DMCSession, error) {
	sessionBytes, err := json.Marshal(reqsession)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", sessionApiURL+"?_format=json", bytes.NewReader(sessionBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := DoRequest(c.HttpClient, req)
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
	err = json.Unmarshal(res, &sessionRes)
	if err != nil {
		return nil, err
	}
	if sessionRes.Meta.Status < 200 || sessionRes.Meta.Status >= 300 {
		Logger.Println(string(sessionBytes))
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

	res, err := DoRequest(c.HttpClient, req)
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
	err = json.Unmarshal(res, &sessionRes)
	if err != nil {
		return err
	}
	if sessionRes.Meta.Status < 200 || sessionRes.Meta.Status >= 300 {
		return fmt.Errorf("Status: %v %v", sessionRes.Meta.Status, sessionRes.Meta.Message)
	}
	session.fullData = sessionRes.Data
	return nil
}

func (c *Client) StartHeartbeat(ctx context.Context, session *DMCSession, errorLimit int) {
	go func() {
		errorCount := 0
		interval := time.Duration(session.KeepMethod.Heartbeat.LifetimeMs) / 2 * time.Millisecond
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(interval):
				err := c.Heartbeat(session)
				if err != nil {
					Logger.Printf("hearbeat error :%v", err)
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
		return NewHLSDownloader(c.HttpClient).Download(ctx, session.ContentURI, w)
	} else if session.IsHTTP() {
		return NewHTTPDownloader(c.HttpClient).Download(ctx, session.ContentURI, w)
	} else {
		return fmt.Errorf("unsupported protocol url: %v", session.ContentURI)
	}
}
