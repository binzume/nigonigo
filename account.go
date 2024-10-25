package nigonigo

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

type NicoSession struct {
	NiconicoID    string `json:"niconicoId"`
	IsPremium     bool   `json:"premium"`
	SessionString string `json:"user_session"`
}

type AccountStatus struct {
	Area     string `json:"area"`
	Timezone string `json:"timezone"`
	Language string `json:"language"`
	Locale   string `json:"locale"`
	UserID   string `json:"userId"`
	Nickname string `json:"nickname"`

	Description string `json:"description"`

	HasPremiumOrStrongerRights      bool `json:"hasPremiumOrStrongerRights"`
	HasSuperPremiumOrStrongerRights bool `json:"hasSuperPremiumOrStrongerRights"`
	IsExplicitlyLoginable           bool `json:"isExplicitlyLoginable"`

	Premium struct {
		Type string `json:"type"`
	} `json:"premium"`

	Icons struct {
		Urls map[string]string `json:"urls"` // "50x50", "150x150"
	} `json:"icons"`

	Existence map[string]any `json:"existence"` // residence, birthdat, sex
	Contacts  map[string]any `json:"contacts"`  // emails
}

func (c *Client) Login(id, password string) error {
	params := map[string]string{
		"mail_tel": id,
		"password": password,
	}
	req, err := newPostReq(accountApiUrl+"login?site=niconico", params)
	if err != nil {
		return err
	}

	res, err := c.HttpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if strings.Contains(res.Request.URL.String(), "/account.nicovideo.jp/") {
		return errors.New("login error")
	}

	return c.checkSessionStatus(res)
}

func (c *Client) LoginWithJsonFile(path string) error {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	var config struct {
		ID       string `json:"id"`
		Password string `json:"password"`
	}
	err = json.Unmarshal(buf, &config)
	if err != nil {
		return err
	}
	return c.Login(config.ID, config.Password)
}

func (c *Client) Logout() error {
	req, err := newGetReq(logoutUrl, nil)
	if err != nil {
		return err
	}
	res, err := c.HttpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	return nil
}

func (c *Client) SaveLoginSession(path string) error {
	json, err := json.Marshal(c.Session)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(path, json, 0644)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) LoadLoginSession(path string) error {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	var s NicoSession
	err = json.Unmarshal(buf, &s)
	if err != nil {
		return err
	}
	return c.SetSessionString(s.SessionString)
}

func (c *Client) SetSessionString(sessionStr string) error {
	cookie := &http.Cookie{Name: "user_session", Value: sessionStr, Domain: ".nicovideo.jp", Path: "/"}
	parsed, _ := url.Parse(topUrl)
	c.HttpClient.Jar.SetCookies(parsed, []*http.Cookie{cookie})

	return c.updateLoginSession()
}

func (c *Client) updateLoginSession() error {
	req, err := newGetReq(topUrl, nil)
	if err != nil {
		return err
	}
	res, err := c.HttpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	return c.checkSessionStatus(res)
}

func (c *Client) checkSessionStatus(res *http.Response) error {

	if res.Header.Get("X-Niconico-Id") == "" {
		return errors.New("failed to login")
	}

	cookie, err := res.Request.Cookie("user_session")
	if err != nil {
		return err
	}

	c.Session = &NicoSession{
		NiconicoID:    res.Header.Get("X-Niconico-Id"),
		IsPremium:     res.Header.Get("X-Niconico-Authflag") == "3",
		SessionString: cookie.Value,
	}
	return nil
}

func (c *Client) GetAvailableSessions() ([]string, error) {
	var sessions []string

	if c.Session != nil && c.Session.SessionString != "" {
		sessions = append(sessions, c.Session.SessionString)
	}

	body, err := getContent(c.HttpClient, "https://account.nicovideo.jp/my/history/login", nil)
	if err != nil {
		return sessions, err
	}
	re := regexp.MustCompile(`<div id="(user_session_[0-9a-f_-]+)"`)
	for _, match := range re.FindAllSubmatch(body, -1) {
		sessions = append(sessions, string(match[1]))
	}
	return sessions, nil
}

func (c *Client) GetAccountStatus() (*AccountStatus, error) {
	body, err := getContent(c.HttpClient, "https://account.nicovideo.jp/api/public/v2/user.json", nil)
	if err != nil {
		return nil, err
	}

	res := struct {
		Meta map[string]any
		Data AccountStatus
	}{}
	err = json.Unmarshal(body, &res)
	return &res.Data, err
}
